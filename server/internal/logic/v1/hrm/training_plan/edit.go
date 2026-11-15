package training_plan

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_plan"
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

// Edit 修改培训计划单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则），
// 已发布（审批通过）的计划同样锁定。单据编号/状态由系统维护，不允许通过编辑改动.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validatePlan(req.PlanName, req.PlanYear, req.PlanType, req.StartDate, req.EndDate); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmTrainingPlan.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_training_plan.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TrainingPlanNotFoundCode, "培训计划单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.TrainingPlanNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}
	if row.Published == 1 {
		return nil, status.ErrorMessage(errcodes.TrainingPlanPublishedGuardCode,
			"已发布的培训计划不允许修改")
	}

	fields := map[string]any{
		"plan_name":  req.PlanName,
		"theme":      req.Theme,
		"plan_year":  req.PlanYear,
		"plan_type":  req.PlanType,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
	}
	if err = l.svcCtx.Model.HrmTrainingPlan.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
