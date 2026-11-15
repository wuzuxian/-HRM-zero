package candidate

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
)

type Import struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewImport(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Import {
	return &Import{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// importHeader is the excelize template shared by Import, mirroring the
// demo-inventory §5 候选人中心 columns. AI 入库通道（邮箱线索入库/文本识别补录）
// 按 ADR-0002 裁剪，仅保留 Excel 导入.
var importHeader = []string{"姓名", "手机号", "邮箱", "来源渠道", "最高学历", "当前公司", "当前岗位", "期望薪资"}

// Import 导入候选人 (spec 用户故事 45): reads the shared xlsx template (one row
// per candidate, first column must be 姓名); all rows share one transaction —
// any invalid row rolls the whole import back (mirrors employee import UX).
func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	raw, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.CandidateImportInvalidCode, "导入文件内容不是合法的 base64")
	}

	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.CandidateImportInvalidCode, "导入文件必须是 xlsx 格式")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.CandidateImportInvalidCode, "导入文件读取失败")
	}
	if len(rows) < 2 {
		return nil, status.ErrorMessage(errcodes.CandidateImportInvalidCode, "导入文件没有数据行")
	}

	var successCount int64
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, row := range rows {
			name := strings.TrimSpace(cellAt(row, 0))
			if name == "" || name == importHeader[0] {
				continue // blank line or header
			}

			sourceChannel, err := parseSourceChannel(cellAt(row, 3))
			if err != nil {
				return err
			}
			education, err := parseEducation(cellAt(row, 4))
			if err != nil {
				return err
			}

			if err = l.svcCtx.Model.HrmRecruitmentCandidate.InsertV2(ctx, session, &hrm_recruitment_candidate.HrmRecruitmentCandidate{
				Uuid:            uuid.New().String(),
				Name:            name,
				Phone:           strings.TrimSpace(cellAt(row, 1)),
				Email:           strings.TrimSpace(cellAt(row, 2)),
				SourceChannel:   sourceChannel,
				Education:       education,
				CurrentCompany:  strings.TrimSpace(cellAt(row, 5)),
				CurrentPosition: strings.TrimSpace(cellAt(row, 6)),
				ExpectSalary:    strings.TrimSpace(cellAt(row, 7)),
				Status:          CandidateStatusNew,
			}); err != nil {
				return err
			}
			successCount++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.ImportResponse{SuccessCount: successCount}, nil
}

func parseSourceChannel(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "内部推荐":
		return SourceChannelRecommend, nil
	case "招聘网站":
		return SourceChannelWebsite, nil
	case "校园招聘":
		return SourceChannelCampus, nil
	case "猎头":
		return SourceChannelHeadhunter, nil
	case "其他":
		return SourceChannelOther, nil
	case "":
		return SourceChannelOther, nil
	case SourceChannelRecommend, SourceChannelWebsite, SourceChannelCampus,
		SourceChannelHeadhunter, SourceChannelOther:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.CandidateImportInvalidCode,
			"来源渠道必须是 内部推荐/招聘网站/校园招聘/猎头/其他")
	}
}

func parseEducation(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "博士":
		return EducationDoctor, nil
	case "硕士":
		return EducationMaster, nil
	case "本科":
		return EducationBachelor, nil
	case "大专":
		return EducationAssociate, nil
	case "高中":
		return EducationSenior, nil
	case "其他":
		return EducationOther, nil
	case "":
		return "", nil
	case EducationDoctor, EducationMaster, EducationBachelor,
		EducationAssociate, EducationSenior, EducationOther:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.CandidateImportInvalidCode,
			"最高学历必须是 博士/硕士/本科/大专/高中/其他")
	}
}

func cellAt(row []string, i int) string {
	if i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}
