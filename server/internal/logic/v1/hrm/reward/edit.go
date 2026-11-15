package reward

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/reward"
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

// Edit 修改奖惩单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）.
// 单据编号与状态由系统维护，不允许通过编辑改动.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateReward(req.EmployeeNo, req.EmployeeName, req.RewardType, req.RewardLevel,
		req.OccurDate, req.SalaryMonth, req.Amount); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmRewardBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_reward_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.RewardNotFoundCode, "奖惩单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.RewardNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	fields := map[string]any{
		"employee_uuid": req.EmployeeUuid,
		"employee_no":   req.EmployeeNo,
		"employee_name": req.EmployeeName,
		"reward_type":   req.RewardType,
		"reward_level":  req.RewardLevel,
		"doc_no":        req.DocNo,
		"occur_date":    req.OccurDate,
		"amount":        req.Amount,
		"reason":        req.Reason,
		"salary_month":  normalizeSalaryMonth(req.SalaryMonth, req.OccurDate),
	}
	if err = l.svcCtx.Model.HrmRewardBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}

func conditionChainByUuid(billUuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_reward_bill.Uuid, billUuid).Build()
}
