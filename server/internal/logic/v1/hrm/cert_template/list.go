package cert_template

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cert_template"
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

// List 证书模板分页列表：支持名称（模糊）/证书类型/来源/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTrainingCertTemplate.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			OrderByDesc(hrm_training_cert_template.Id).
			Like(hrm_training_cert_template.TemplateName, "%"+strings.TrimSpace(req.TemplateName)+"%",
				condition.WithSkip(strings.TrimSpace(req.TemplateName) == "")).
			Equal(hrm_training_cert_template.CertType, req.CertType, condition.WithSkip(req.CertType == "")).
			Equal(hrm_training_cert_template.Source, req.Source, condition.WithSkip(req.Source == "")).
			Equal(hrm_training_cert_template.Status, req.Status, condition.WithSkip(req.Status == "")).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmCertTemplate, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmCertTemplate{
			Uuid:         row.Uuid,
			TemplateName: row.TemplateName,
			CertType:     row.CertType,
			Source:       row.Source,
			NoPrefix:     row.NoPrefix,
			Status:       row.Status,
			Remark:       row.Remark,
			CreateTime:   row.CreateTime.Format(time.DateTime),
			UpdateTime:   row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
