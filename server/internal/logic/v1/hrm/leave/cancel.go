package leave

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
)

type Cancel struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewCancel(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Cancel {
	return &Cancel{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Cancel 销假：审批通过后回填 实际开始/结束时间 + 实际天数（自动计算），
// 余额按差额多退少补（同一事务内）：
//   - 实际 < 预计：leavebalance.RefundLeave 退回差额（used -= diff，refund 流水）
//   - 实际 > 预计：leavebalance.ConsumeLeave 再扣差额（可用不足 → 10513 整体回滚）
//
// 幂等：settled=1（已销假）后再次销假 → 10542。
func (l *Cancel) Cancel(req *types.CancelRequest) (resp *types.CancelResponse, err error) {
	actualDays, err := CalcLeaveDays(req.ActualStartTime, req.ActualEndTime)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.LeaveInvalidCode, err.Error())
	}

	row, err := l.svcCtx.Model.HrmLeaveBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_leave_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.LeaveNotFoundCode, "请假单不存在")
		}
		return nil, err
	}
	if row.Status != StatusApproved {
		return nil, status.ErrorMessage(errcodes.LeaveNotEditableCode, "仅 审批通过 的请假单可销假")
	}
	if row.Settled == 1 {
		return nil, status.ErrorMessage(errcodes.LeaveNotEditableCode, "单据已销假，不能重复销假")
	}

	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := settleBalance(ctx, session, l.svcCtx, row, actualDays); err != nil {
			return err
		}
		return l.svcCtx.Model.HrmLeaveBill.UpdateFieldsByCondition(ctx, session, map[string]any{
			"actual_start_time": req.ActualStartTime,
			"actual_end_time":   req.ActualEndTime,
			"actual_days":       actualDays,
			"settled":           1,
		}, conditionChainByUuid(req.Uuid)...)
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("leave: 请假单销假完成 billNo=%s planDays=%.1f actualDays=%.1f (差额 %.1f 天多退少补)",
		row.BillNo, row.PlanDays, actualDays, actualDays-row.PlanDays)
	return &types.CancelResponse{}, nil
}
