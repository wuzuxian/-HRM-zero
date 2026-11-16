package leavebalance

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
)

type Init struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewInit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Init {
	return &Init{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Init 初始化本年度额度 (demo-inventory 4.4): 对启用了余额管理的假期类型，
// 按类型默认天数初始化全员或指定员工，事务内逐行落账。
// 幂等规则：目标 (year, employee, leave_type) 已有余额行 → 跳过不覆盖（重复初始化拒绝/跳过）。
func (l *Init) Init(req *types.InitRequest) (resp *types.InitResponse, err error) {
	if req.Year <= 0 {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "年度必须大于 0")
	}

	// 目标员工集合：指定 uuids 或全体在册员工（person_status NOT IN ('5','6')）
	employeeUuids, employeeNames, err := resolveInitEmployees(l.ctx, l.svcCtx, req.EmployeeUuids)
	if err != nil {
		return nil, err
	}
	if len(employeeUuids) == 0 {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceEmployeeNotFoundCode, "未找到目标员工")
	}

	// 启用余额管理的假期类型
	leaveTypes, err := managedLeaveTypes(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if len(leaveTypes) == 0 {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceTypeNotManagedCode, "没有启用余额管理的假期类型")
	}

	var skipped int64
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, employeeUuid := range employeeUuids {
			for _, lt := range leaveTypes {
				existing, findErr := modelOf(l.svcCtx, session).FindOneByCondition(ctx, session,
					balanceCondition(req.Year, employeeUuid, lt.TypeCode).Build()...)
				if findErr == nil && existing != nil {
					skipped++ // 幂等：已有余额行跳过
					continue
				}
				if findErr != nil && findErr != sqlx.ErrNotFound {
					return findErr
				}
				if err := initOneBalance(ctx, session, l.svcCtx, req.Year, employeeUuid,
					employeeNames[employeeUuid], lt); err != nil {
					return err
				}
			}
		}
		return nil
	}); txErr != nil {
		return nil, txErr
	}

	if skipped > 0 {
		l.Infof("init leave balance year=%d skipped=%d (already initialized)", req.Year, skipped)
	}

	return &types.InitResponse{}, nil
}
