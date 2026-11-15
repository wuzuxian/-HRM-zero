package overtimebill

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
)

type Confirm struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewConfirm(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Confirm {
	return &Confirm{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Confirm 核定加班单（审批人在 审批中 状态操作）：核定时长（>0 且 ≤ 申请时长，
// 折算口径按最小计算单位向下取整）+ 补偿方式（调休/其他）。
// 核定结果在 审批通过 时折算入账（bill.go OnApproved）.
func (l *Confirm) Confirm(req *types.ConfirmRequest) (resp *types.ConfirmResponse, err error) {
	switch req.CompType {
	case CompTypeLeave, CompTypeOther:
	default:
		return nil, status.ErrorMessage(errcodes.OvertimeInvalidCode, "补偿方式必须是 调休/其他")
	}
	if req.ConfirmHours <= 0 {
		return nil, status.ErrorMessage(errcodes.OvertimeInvalidCode, "核定时长必须大于 0")
	}

	row, err := l.svcCtx.Model.HrmOvertimeBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OvertimeNotFoundCode, "加班单不存在")
		}
		return nil, err
	}
	if row.Status != StatusApproving {
		return nil, status.ErrorMessage(errcodes.OvertimeNotEditableCode,
			"仅 审批中 状态的加班单可以核定")
	}
	// 核定时长 ≤ 申请时长（口径校验用申请时长原值；显示折算按最小单位向下取整）.
	if req.ConfirmHours > row.ApplyHours+1e-9 {
		return nil, status.ErrorMessage(errcodes.OvertimeConfirmHoursExceededCode,
			"核定时长不能超过申请时长")
	}
	confirmHours := RoundHoursToUnit(req.ConfirmHours, resolveMinUnitMinutes(l.ctx, l.svcCtx, row.EmployeeUuid, row.DayType))
	if confirmHours <= 0 {
		return nil, status.ErrorMessage(errcodes.OvertimeInvalidCode,
			"核定时长不足一个最小计算单位（30 分钟）")
	}

	fields := map[string]any{
		"confirm_hours": confirmHours,
		"comp_type":     req.CompType,
	}
	if err = l.svcCtx.Model.HrmOvertimeBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		conditionChainByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	l.Infof("overtimebill: 加班单核定 billNo=%s confirmHours=%.1f compType=%s", row.BillNo, confirmHours, req.CompType)
	return &types.ConfirmResponse{}, nil
}
