package salary_scheme

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 薪酬方案分页列表，按创建时间倒序，支持按方案名称/主体/状态筛选；
// 每行带 项目数 与 配置完整度（6 项检查，口径见 completeness.go）.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmSalaryScheme.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_salary_scheme.SchemeName, "%"+req.SchemeName+"%", condition.WithSkip(req.SchemeName == "")).
			Equal(hrm_salary_scheme.SubjectUuid, req.SubjectUuid, condition.WithSkip(req.SubjectUuid == "")).
			Equal(hrm_salary_scheme.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_salary_scheme.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmSalaryScheme, 0, len(rows))
	for _, row := range rows {
		record := schemeRowToRecord(row)
		completeness, itemCount, err := completenessFor(l.ctx, l.svcCtx.SqlxConn, row.Uuid)
		if err != nil {
			return nil, err
		}
		record.ItemCount = itemCount
		record.Completeness = completeness
		records = append(records, record)
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
