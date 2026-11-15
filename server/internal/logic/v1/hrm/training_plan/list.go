package training_plan

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_plan"
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

// List 培训计划单分页列表，支持按计划名称/计划类型/单据状态/发布状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTrainingPlan.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Like(hrm_training_plan.PlanName, "%"+req.PlanName+"%", condition.WithSkip(req.PlanName == "")).
			Equal(hrm_training_plan.PlanType, req.PlanType, condition.WithSkip(req.PlanType == "")).
			Equal(hrm_training_plan.Status, req.Status, condition.WithSkip(req.Status == "")).
			Equal(hrm_training_plan.Published, req.Published, condition.WithSkip(req.Published == "")).
			Page(req.Current, req.Size).
			OrderByDesc(hrm_training_plan.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmTrainingPlan, 0, len(rows))
	for _, row := range rows {
		records = append(records, planRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func planRecord(row *hrm_training_plan.HrmTrainingPlan) types.HrmTrainingPlan {
	return types.HrmTrainingPlan{
		Uuid:        row.Uuid,
		BillNo:      row.BillNo,
		Status:      row.Status,
		PlanName:    row.PlanName,
		Theme:       row.Theme,
		PlanYear:    row.PlanYear,
		PlanType:    row.PlanType,
		StartDate:   row.StartDate,
		EndDate:     row.EndDate,
		Published:   row.Published,
		CreatorUuid: row.CreateBy,
		CreatorName: row.CreatorName,
		CreateTime:  row.CreateTime.Format(time.DateTime),
		UpdateTime:  row.UpdateTime.Format(time.DateTime),
	}
}
