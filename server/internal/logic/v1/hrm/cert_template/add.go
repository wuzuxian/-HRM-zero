package cert_template

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cert_template"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增证书模板：编号前缀唯一（前缀决定证书号=前缀-年-流水），初始状态 1=启用.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateTemplate(req.TemplateName, req.CertType, req.Source, req.NoPrefix); err != nil {
		return nil, err
	}

	prefix := strings.ToUpper(strings.TrimSpace(req.NoPrefix))

	// 前缀唯一：撞车即拒绝（证书号按前缀分配流水，前缀重复会导致两套编号互相踩号）
	if _, err = l.svcCtx.Model.HrmTrainingCertTemplate.FindOneByNoPrefix(l.ctx, nil, prefix); err == nil {
		return nil, status.ErrorMessage(errcodes.CertTemplatePrefixDuplicateCode,
			"编号前缀已存在: "+prefix)
	} else if err != hrm_training_cert_template.ErrNotFound {
		return nil, err
	}

	if _, err = l.svcCtx.Model.HrmTrainingCertTemplate.Insert(l.ctx, nil, &hrm_training_cert_template.HrmTrainingCertTemplate{
		Uuid:         uuid.New().String(),
		TemplateName: req.TemplateName,
		CertType:     req.CertType,
		Source:       req.Source,
		NoPrefix:     prefix,
		Status:       TemplateStatusEnabled,
		Remark:       req.Remark,
	}); err != nil {
		return nil, err
	}

	l.Infof("cert_template: 证书模板已创建 template=%s prefix=%s", req.TemplateName, prefix)
	return &types.AddResponse{}, nil
}

var _ = time.Now
