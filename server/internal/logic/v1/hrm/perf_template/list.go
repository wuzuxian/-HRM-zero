package perf_template

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_template"
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

func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmPerfTemplate.PageByCondition(l.ctx, nil, condition.NewChain().
		Page(req.Current, req.Size).
		OrderByDesc(hrm_perf_template.CreateTime).
		Like(hrm_perf_template.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
		Equal(hrm_perf_template.Status, req.Status, condition.WithSkip(req.Status == "")).
		Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmPerfTemplate, 0, len(rows))
	for _, row := range rows {
		// 列表行不带指标权重配置；进入编辑配置时用 getPerfTemplateDetail 取明细
		records = append(records, templateRecord(row, nil))
	}
	return &types.ListResponse{
		PageResponse: types.PageResponse{
			Current: req.Current,
			Size:    req.Size,
			Total:   total,
		},
		Records: records,
	}, nil
}

// templateRecord renders one 考核模板 row (with optional 指标权重配置) for the API.
func templateRecord(row *hrm_perf_template.HrmPerfTemplate, items []types.HrmPerfTemplateItem) types.HrmPerfTemplate {
	if items == nil {
		items = []types.HrmPerfTemplateItem{}
	}
	return types.HrmPerfTemplate{
		Uuid:       row.Uuid,
		Name:       row.Name,
		Status:     row.Status,
		Remark:     row.Remark,
		Items:      items,
		CreateTime: row.CreateTime.Format(time.DateTime),
		UpdateTime: row.UpdateTime.Format(time.DateTime),
	}
}
