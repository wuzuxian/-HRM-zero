package perf_assessment

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_appeal"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
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

// List 考核任务列表（五个填报列表页共用：当前节点/任务状态/当前得分/当前等级列；
// 展示快照批量 IN 查询，无 N+1）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := ListTasks(l.ctx, l.svcCtx, req.PeriodUuid, req.EmployeeUuid, req.CurrentNode, req.TaskStatus,
		req.Current, req.Size)
	if err != nil {
		return nil, err
	}

	// 展示快照：批次（编号/名称）
	periodUuids := make([]string, 0, len(rows))
	employeeUuids := make([]string, 0, len(rows))
	templateUuids := make([]string, 0, len(rows))
	seenPeriod := map[string]struct{}{}
	seenEmployee := map[string]struct{}{}
	seenTemplate := map[string]struct{}{}
	for _, row := range rows {
		if _, ok := seenPeriod[row.PeriodUuid]; !ok {
			seenPeriod[row.PeriodUuid] = struct{}{}
			periodUuids = append(periodUuids, row.PeriodUuid)
		}
		if _, ok := seenEmployee[row.EmployeeUuid]; !ok {
			seenEmployee[row.EmployeeUuid] = struct{}{}
			employeeUuids = append(employeeUuids, row.EmployeeUuid)
		}
		if _, ok := seenTemplate[row.TemplateUuid]; !ok {
			seenTemplate[row.TemplateUuid] = struct{}{}
			templateUuids = append(templateUuids, row.TemplateUuid)
		}
	}

	periodNames := map[string]*hrm_perf_period.HrmPerfPeriod{}
	if len(periodUuids) > 0 {
		periods, err := l.svcCtx.Model.HrmPerfPeriod.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_perf_period.Uuid, periodUuids).Build()...)
		if err != nil {
			return nil, err
		}
		for _, p := range periods {
			periodNames[p.Uuid] = p
		}
	}

	employeeNames := map[string]string{}
	if len(employeeUuids) > 0 {
		employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_employee.Uuid, employeeUuids).Build()...)
		if err != nil {
			return nil, err
		}
		for _, e := range employees {
			employeeNames[e.Uuid] = e.Name
		}
	}

	templateNames := map[string]string{}
	if len(templateUuids) > 0 {
		templates, err := l.svcCtx.Model.HrmPerfTemplate.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_perf_template.Uuid, templateUuids).Build()...)
		if err != nil {
			return nil, err
		}
		for _, t := range templates {
			templateNames[t.Uuid] = t.Name
		}
	}

	// 申诉状态快照（一次 IN 查询最近申诉；演示口径按 assessment 分组取最新）
	appealStatus := map[string]string{}
	if len(rows) > 0 {
		assessmentUuids := make([]string, 0, len(rows))
		for _, row := range rows {
			assessmentUuids = append(assessmentUuids, row.Uuid)
		}
		appeals, err := l.svcCtx.Model.HrmPerfAppeal.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_perf_appeal.AssessmentUuid, assessmentUuids).
				OrderByDesc(hrm_perf_appeal.Id).Build()...)
		if err != nil {
			return nil, err
		}
		for _, a := range appeals {
			if _, ok := appealStatus[a.AssessmentUuid]; !ok {
				appealStatus[a.AssessmentUuid] = a.Status
			}
		}
	}

	records := make([]types.HrmPerfAssessment, 0, len(rows))
	for _, row := range rows {
		// 列表行不带 items 快照详情（详情接口返回），仅展示各阶段得分/等级
		records = append(records, assessmentRecord(row, []types.AssessmentItemSnapshot{},
			periodNames[row.PeriodUuid], employeeNames[row.EmployeeUuid],
			templateNames[row.TemplateUuid], appealStatus[row.Uuid]))
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
