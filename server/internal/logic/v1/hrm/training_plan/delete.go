package training_plan

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_plan"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Delete 删除/批量删除培训计划单：仅 未提交/审批不通过/已撤销 状态可删除，
// 已发布（审批通过）的计划同样锁定.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	for _, billUuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmTrainingPlan.FindOneByUuid(l.ctx, nil, billUuid)
		if err != nil {
			if err == hrm_training_plan.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.TrainingPlanNotFoundCode, "培训计划单不存在")
			}
			return nil, err
		}
		if !canEdit(row.Status) {
			return nil, status.ErrorMessage(errcodes.TrainingPlanNotEditableCode,
				"单据 "+row.BillNo+" 当前状态不允许删除（仅 未提交/审批不通过/已撤销 可删）")
		}
		if row.Published == 1 {
			return nil, status.ErrorMessage(errcodes.TrainingPlanPublishedGuardCode,
				"已发布的培训计划不允许删除")
		}
	}

	err = l.svcCtx.Model.HrmTrainingPlan.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_training_plan.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}

func conditionChainByUuid(billUuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_training_plan.Uuid, billUuid).Build()
}
