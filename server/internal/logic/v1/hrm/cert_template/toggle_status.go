package cert_template

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cert_template"
)

type ToggleStatus struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewToggleStatus(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *ToggleStatus {
	return &ToggleStatus{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// ToggleStatus 启用/停用（行开关）：停用后不可发证（CertificateTemplateDisabledCode）.
func (l *ToggleStatus) ToggleStatus(req *types.ToggleStatusRequest) (resp *types.ToggleStatusResponse, err error) {
	if req.Status != TemplateStatusEnabled && req.Status != TemplateStatusDisabled {
		return nil, status.ErrorMessage(errcodes.CertTemplateInvalidCode, "状态非法（1=启用 2=停用）")
	}

	if _, err = l.svcCtx.Model.HrmTrainingCertTemplate.FindOneByUuid(l.ctx, nil, req.Uuid); err != nil {
		if err == hrm_training_cert_template.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CertTemplateNotFoundCode, "证书模板不存在")
		}
		return nil, err
	}

	if err = l.svcCtx.Model.HrmTrainingCertTemplate.UpdateFieldsByCondition(l.ctx, nil,
		map[string]any{"status": req.Status},
		conditionUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	l.Infof("cert_template: 证书模板状态已切换 status=%s", req.Status)
	return &types.ToggleStatusResponse{}, nil
}
