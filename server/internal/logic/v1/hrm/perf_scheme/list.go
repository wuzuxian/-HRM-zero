package perf_scheme

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_scheme"
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
	rows, total, err := l.svcCtx.Model.HrmPerfScheme.PageByCondition(l.ctx, nil, condition.NewChain().
		Page(req.Current, req.Size).
		OrderByDesc(hrm_perf_scheme.CreateTime).
		Like(hrm_perf_scheme.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
		Equal(hrm_perf_scheme.Status, req.Status, condition.WithSkip(req.Status == "")).
		Build()...)
	if err != nil {
		return nil, err
	}

	// 列表行带关联模板名称快照（一次批量查询，无 N+1）
	templateUuids := make([]string, 0, len(rows))
	for _, row := range rows {
		templateUuids = append(templateUuids, row.TemplateUuid)
	}
	templateNames := make(map[string]string, len(templateUuids))
	if len(templateUuids) > 0 {
		templates, err := l.svcCtx.Model.HrmPerfTemplate.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_perf_template.Uuid, templateUuids).Build()...)
		if err != nil {
			return nil, err
		}
		for _, tpl := range templates {
			templateNames[tpl.Uuid] = tpl.Name
		}
	}

	records := make([]types.HrmPerfScheme, 0, len(rows))
	for _, row := range rows {
		records = append(records, schemeRecord(row, templateNames[row.TemplateUuid]))
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

// schemeRecord renders one 绩效方案 row for the API (templateName = 关联模板名称快照).
func schemeRecord(row *hrm_perf_scheme.HrmPerfScheme, templateName string) types.HrmPerfScheme {
	return types.HrmPerfScheme{
		Uuid:           row.Uuid,
		Name:           row.Name,
		TemplateUuid:   row.TemplateUuid,
		TemplateName:   templateName,
		PeriodType:     row.PeriodType,
		AssessmentMode: row.AssessmentMode,
		ScoreMode:      row.ScoreMode,
		PayLinkMode:    row.PayLinkMode,
		Status:         row.Status,
		Remark:         row.Remark,
		CreateTime:     row.CreateTime.Format(time.DateTime),
		UpdateTime:     row.UpdateTime.Format(time.DateTime),
	}
}
