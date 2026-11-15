package training_session

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
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

// List 培训班次分页列表（管理端），支持按班次名称/培训计划/状态筛选；
// 关联计划名称随行返回（单条 IN 查询快照，无 N+1）。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmTrainingSession.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Like(hrm_training_session.SessionName, "%"+req.SessionName+"%", condition.WithSkip(req.SessionName == "")).
			Equal(hrm_training_session.PlanUuid, req.PlanUuid, condition.WithSkip(req.PlanUuid == "")).
			Equal(hrm_training_session.Status, req.Status, condition.WithSkip(req.Status == "")).
			Page(req.Current, req.Size).
			OrderByDesc(hrm_training_session.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	// plan names for display (one batched IN query)
	planUuids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.PlanUuid != "" {
			planUuids = append(planUuids, row.PlanUuid)
		}
	}
	planNames := make(map[string]string, len(planUuids))
	if len(planUuids) > 0 {
		plans, perr := l.svcCtx.Model.HrmTrainingPlan.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_training_plan.Uuid, planUuids).Build()...)
		if perr != nil {
			return nil, perr
		}
		for _, p := range plans {
			planNames[p.Uuid] = p.PlanName
		}
	}

	records := make([]types.HrmTrainingSession, 0, len(rows))
	for _, row := range rows {
		records = append(records, sessionRecord(row, planNames[row.PlanUuid]))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func sessionRecord(row *hrm_training_session.HrmTrainingSession, planName string) types.HrmTrainingSession {
	return types.HrmTrainingSession{
		Uuid:        row.Uuid,
		SessionNo:   row.SessionNo,
		SessionName: row.SessionName,
		PlanUuid:    row.PlanUuid,
		PlanName:    planName,
		Lecturer:    row.Lecturer,
		Location:    row.Location,
		Hours:       row.Hours,
		StartTime:   row.StartTime.Format(time.DateTime),
		EndTime:     row.EndTime.Format(time.DateTime),
		Status:      row.Status,
		Remark:      row.Remark,
		CreateTime:  row.CreateTime.Format(time.DateTime),
		UpdateTime:  row.UpdateTime.Format(time.DateTime),
	}
}
