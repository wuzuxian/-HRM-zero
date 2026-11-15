package training_session

import (
	"context"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
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

// Edit 修改培训班次：班次编号/状态由系统维护；若改关联计划，目标计划同样须已发布.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateSession(req.SessionName, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmTrainingSession.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TrainingSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}

	if req.PlanUuid != row.PlanUuid && strings.TrimSpace(req.PlanUuid) != "" {
		plan, perr := l.svcCtx.Model.HrmTrainingPlan.FindOneByUuid(l.ctx, nil, req.PlanUuid)
		if perr != nil {
			if perr == hrm_training_plan.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.TrainingSessionPlanNotPublishedCode, "关联的培训计划不存在")
			}
			return nil, perr
		}
		if plan.Published != 1 {
			return nil, status.ErrorMessage(errcodes.TrainingSessionPlanNotPublishedCode,
				"关联的培训计划未发布（仅已发布计划可关联班次）")
		}
	}

	fields := map[string]any{
		"session_name": req.SessionName,
		"plan_uuid":    req.PlanUuid,
		"lecturer":     req.Lecturer,
		"location":     req.Location,
		"hours":        req.Hours,
		"start_time":   mustParseTime(req.StartTime),
		"end_time":     mustParseTime(req.EndTime),
		"remark":       req.Remark,
	}
	if err = l.svcCtx.Model.HrmTrainingSession.UpdateFieldsByCondition(l.ctx, nil, fields,
		sessionConditionByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
