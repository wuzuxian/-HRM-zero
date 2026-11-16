package leavebalance

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
)

type Adjust struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdjust(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Adjust {
	return &Adjust{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Adjust 手工调整余额（记原因，demo-inventory 4.4 行操作「调整」）：
// days ≠ 0 增减应发 granted；可用 = granted-used-frozen 不得为负（负数导致可用 < 0 → 10513 拒绝）。
func (l *Adjust) Adjust(req *types.AdjustRequest) (resp *types.AdjustResponse, err error) {
	if req.BalanceUuid == "" {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "余额行不能为空")
	}
	if req.Days == 0 {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "调整天数不能为 0")
	}

	user := hrmuser.Current(l.ctx)
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		m := modelOf(l.svcCtx, session)
		row, findErr := m.FindOneByCondition(l.ctx, session, balanceByUuidCondition(req.BalanceUuid).Build()...)
		if findErr != nil {
			if findErr == sqlx.ErrNotFound {
				return errBalanceNotFound()
			}
			return findErr
		}

		newGranted := round1(row.Granted + req.Days)
		if round1(newGranted-row.Used-row.Frozen) < 0 {
			return status.ErrorMessage(errcodes.LeaveBalanceInsufficientCode, "调整后可用余额不能为负")
		}

		if updErr := updateBalanceFields(l.ctx, session, l.svcCtx, row, map[string]any{"granted": newGranted}); updErr != nil {
			return updErr
		}
		row.Granted = newGranted

		operatorUuid := ""
		operatorName := ""
		if user.Uuid != "" {
			operatorUuid = user.Uuid
			operatorName = user.Username
		}
		return writeLog(l.ctx, session, l.svcCtx, row, ChangeTypeAdjust, round1(req.Days),
			req.Reason, "", operatorUuid, operatorName)
	}); txErr != nil {
		return nil, txErr
	}

	return &types.AdjustResponse{}, nil
}
