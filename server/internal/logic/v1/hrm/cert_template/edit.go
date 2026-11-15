package cert_template

import (
	"context"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cert_template"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 编辑证书模板：前缀变更时保持前缀唯一（撞车即拒绝）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateTemplate(req.TemplateName, req.CertType, req.Source, req.NoPrefix); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmTrainingCertTemplate.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_training_cert_template.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.CertTemplateNotFoundCode, "证书模板不存在")
		}
		return nil, err
	}

	prefix := strings.ToUpper(strings.TrimSpace(req.NoPrefix))
	if prefix != row.NoPrefix {
		if owner, err := l.svcCtx.Model.HrmTrainingCertTemplate.FindOneByNoPrefix(l.ctx, nil, prefix); err == nil && owner.Uuid != row.Uuid {
			return nil, status.ErrorMessage(errcodes.CertTemplatePrefixDuplicateCode,
				"编号前缀已存在: "+prefix)
		} else if err != nil && err != hrm_training_cert_template.ErrNotFound {
			return nil, err
		}
	}

	if err = l.svcCtx.Model.HrmTrainingCertTemplate.Update(l.ctx, nil, &hrm_training_cert_template.HrmTrainingCertTemplate{
		Id:           row.Id,
		Uuid:         row.Uuid,
		TemplateName: req.TemplateName,
		CertType:     req.CertType,
		Source:       req.Source,
		NoPrefix:     prefix,
		Status:       row.Status,
		Remark:       req.Remark,
	}); err != nil {
		return nil, err
	}

	l.Infof("cert_template: 证书模板已更新 template=%s prefix=%s", req.TemplateName, prefix)
	return &types.EditResponse{}, nil
}
