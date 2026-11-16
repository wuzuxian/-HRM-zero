package perf_assessment

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type GetDetail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetDetail {
	return &GetDetail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetDetail 考核任务详情：模板快照 items + 各阶段填报记录（逐项得分/说明/时间/操作人）。
func (l *GetDetail) GetDetail(req *types.GetDetailRequest) (resp *types.GetDetailResponse, err error) {
	row, err := FindAssessmentByUuid(l.ctx, l.svcCtx.SqlxConn, req.Uuid)
	if err != nil {
		return nil, wrapBizErr(err)
	}

	snapshot, err := parseItemsSnapshot(row.ItemsSnapshot)
	if err != nil {
		return nil, wrapBizErr(err)
	}
	selfItems, err := parseScoreItems(row.SelfItems)
	if err != nil {
		return nil, wrapBizErr(err)
	}
	reviewItems, err := parseScoreItems(row.ReviewItems)
	if err != nil {
		return nil, wrapBizErr(err)
	}

	period, err := findPeriod(l.ctx, l.svcCtx.SqlxConn, row.PeriodUuid)
	if err != nil {
		return nil, wrapBizErr(err)
	}
	employeeName := ""
	if employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, row.EmployeeUuid).Build()...); err == nil && len(employees) > 0 {
		employeeName = employees[0].Name
	}
	templateName := ""
	if templates, err := l.svcCtx.Model.HrmPerfTemplate.FindByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_perf_template.Uuid, row.TemplateUuid).Build()...); err == nil && len(templates) > 0 {
		templateName = templates[0].Name
	}
	appealStatus, err := latestAppealStatus(l.ctx, l.svcCtx.SqlxConn, row.Uuid)
	if err != nil {
		return nil, err
	}

	return &types.GetDetailResponse{
		HrmPerfAssessmentDetail: types.HrmPerfAssessmentDetail{
			HrmPerfAssessment: assessmentRecord(row, itemsSnapshotToTypes(snapshot),
				period, employeeName, templateName, appealStatus),
			PlanTime:      row.PlanTime,
			PlanBy:        row.PlanBy,
			SelfItems:     scoreItemsToTypes(selfItems),
			SelfComment:   row.SelfComment,
			SelfTime:      row.SelfTime,
			SelfBy:        row.SelfBy,
			ReviewItems:   scoreItemsToTypes(reviewItems),
			ReviewComment: row.ReviewComment,
			ReviewTime:    row.ReviewTime,
			ReviewBy:      row.ReviewBy,
			CalibComment:  row.CalibComment,
			CalibTime:     row.CalibTime,
			CalibBy:       row.CalibBy,
			ResultTime:    row.ResultTime,
			ResultBy:      row.ResultBy,
		},
	}, nil
}
