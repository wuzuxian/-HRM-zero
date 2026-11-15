package certificate

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

type BatchIssue struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewBatchIssue(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *BatchIssue {
	return &BatchIssue{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// BatchIssue 批量发证（班次结业）：按班次 + 模板为班次学员批量生成证书
// （source='1'，template_uuid 回填，证书号=前缀-年-流水，发证即置 status='1' 已发）。
// 幂等：同一班次+模板下已有有效（非已撤销）证书的学员跳过，不重复发证；
// 重复调用同一请求 IssuedCount=0 / SkippedCount=学员数。模板停用（status='2'）不可发证。
func (l *BatchIssue) BatchIssue(req *types.BatchIssueRequest) (resp *types.BatchIssueResponse, err error) {
	if req.SessionUuid == "" || req.TemplateUuid == "" {
		return nil, status.ErrorMessage(errcodes.CertificateInvalidCode, "班次与证书模板必填")
	}

	// 模板须存在且启用（停用后不可发证）
	tpl, err := l.svcCtx.Model.HrmTrainingCertTemplate.FindOneByUuid(l.ctx, nil, req.TemplateUuid)
	if err != nil {
		if err == hrm_training_cert_template.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CertTemplateNotFoundCode, "证书模板不存在")
		}
		return nil, err
	}
	if tpl.Status != templateEnabled {
		return nil, status.ErrorMessage(errcodes.CertificateTemplateDisabledCode,
			"证书模板已停用，不可发证: "+tpl.TemplateName)
	}

	// 班次须存在
	if _, err = l.svcCtx.Model.HrmTrainingSession.FindOneByUuid(l.ctx, nil, req.SessionUuid); err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TrainingSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}

	// 发证名单：缺省=班次全部学员
	participants, err := l.svcCtx.Model.HrmTrainingSessionParticipant.FindByCondition(l.ctx, nil,
		conditionEq(hrm_training_session_participant.SessionUuid, req.SessionUuid)...)
	if err != nil {
		return nil, err
	}
	if len(req.UserUuids) > 0 {
		want := make(map[string]struct{}, len(req.UserUuids))
		for _, u := range req.UserUuids {
			want[u] = struct{}{}
		}
		filtered := participants[:0]
		for _, p := range participants {
			if _, ok := want[p.UserUuid]; ok {
				filtered = append(filtered, p)
			}
		}
		participants = filtered
	}

	// 已有该班次+模板的有效证书的学员 → 幂等跳过
	existing, err := l.svcCtx.Model.HrmTrainingCertificate.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_training_certificate.SessionUuid, req.SessionUuid).
			Equal(hrm_training_certificate.TemplateUuid, req.TemplateUuid).
			In(hrm_training_certificate.Status, []string{CertStatusPending, CertStatusIssued}).
			Build()...)
	if err != nil {
		return nil, err
	}
	hasCert := make(map[string]struct{}, len(existing))
	for _, c := range existing {
		hasCert[c.UserUuid] = struct{}{}
	}

	var issuedCount, skippedCount int64
	for _, p := range participants {
		if _, ok := hasCert[p.UserUuid]; ok {
			skippedCount++
			continue
		}

		var certNo string
		if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			// 流水分配在事务内（select-then-upsert on hrm_bill_sequence, type=前缀, date=年）
			no, err := GenerateCertNo(ctx, session, tpl.NoPrefix, time.Now().Year())
			if err != nil {
				return err
			}
			certNo = no
			return l.svcCtx.Model.HrmTrainingCertificate.InsertV2(ctx, session, &hrm_training_certificate.HrmTrainingCertificate{
				Uuid:         uuid.New().String(),
				CertNo:       no,
				CertName:     tpl.TemplateName,
				CertType:     tpl.CertType,
				TemplateUuid: tpl.Uuid,
				Source:       CertSourceSession,
				SessionUuid:  req.SessionUuid,
				UserUuid:     p.UserUuid,
				IssueTime:    nullTime(time.Now()),
				Status:       CertStatusIssued,
			})
		}); txErr != nil {
			return nil, txErr
		}
		issuedCount++
		l.Infof("certificate: 批量发证 certNo=%s session=%s template=%s", certNo, req.SessionUuid, tpl.TemplateName)
	}

	return &types.BatchIssueResponse{IssuedCount: issuedCount, SkippedCount: skippedCount}, nil
}
