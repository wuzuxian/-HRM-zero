package employee

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type LinkUser struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLinkUser(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *LinkUser {
	return &LinkUser{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// LinkUser 关联已有系统账号 (HRM-51): 把账号 uuid 写入档案 user_uuid 指针。
// 一对一在应用层校验 (ADR 0007 — 单列指针 + 应用层校验, 不加库级唯一索引):
//   - 账号已被其他档案持有 → EmployeeUserOccupiedCode, 错误消息报出占用者 (工号+姓名)
//   - 员工已关联其他账号 → EmployeeUserAlreadyLinkedCode
//   - 账号不存在 → EmployeeUserAccountNotFoundCode
//   - 关联到当前已持有的同一账号 → 幂等成功
//
// 校验与写入在同一事务内, 拒绝时不留半截状态。
func (l *LinkUser) LinkUser(req *types.LinkUserRequest) (resp *types.LinkUserResponse, err error) {
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		emp, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, req.EmployeeUuid)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return status.ErrorMessage(errcodes.EmployeeUserNotFoundCode, "员工不存在: "+req.EmployeeUuid)
			}
			return err
		}

		// 已持有该账号 → 幂等成功 (不做任何写入)
		if emp.UserUuid == req.UserUuid && req.UserUuid != "" {
			return nil
		}

		// 已关联其他账号 → 拒绝 (唯一改绑路径: 先解绑再关联)
		if emp.UserUuid != "" {
			current, userErr := l.svcCtx.Model.ManageUser.FindOneByUuid(ctx, session, emp.UserUuid)
			if userErr == nil {
				return status.ErrorMessage(errcodes.EmployeeUserAlreadyLinkedCode,
					fmt.Sprintf("员工 %s(%s) 已关联系统账号 %s, 请先解除关联", emp.EmployeeNo, emp.Name, current.Username))
			}
			return status.ErrorMessage(errcodes.EmployeeUserAlreadyLinkedCode,
				fmt.Sprintf("员工 %s(%s) 已关联其他系统账号, 请先解除关联", emp.EmployeeNo, emp.Name))
		}

		// 账号必须存在
		user, err := l.svcCtx.Model.ManageUser.FindOneByUuid(ctx, session, req.UserUuid)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return status.ErrorMessage(errcodes.EmployeeUserAccountNotFoundCode,
					"系统账号不存在: "+req.UserUuid)
			}
			return err
		}

		// 占用校验: 指针已被其他档案持有 → 拒绝并报出占用者
		holder, err := l.svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, session,
			condition.NewChain().Equal(hrm_employee.UserUuid, req.UserUuid).Build()...)
		switch {
		case err == nil && holder.Uuid != emp.Uuid:
			return status.ErrorMessage(errcodes.EmployeeUserOccupiedCode,
				fmt.Sprintf("系统账号 %s 已被员工 %s(%s) 关联", user.Username, holder.EmployeeNo, holder.Name))
		case err != nil && err != sqlx.ErrNotFound:
			return err
		}

		emp.UserUuid = req.UserUuid
		if err = l.svcCtx.Model.HrmEmployee.Update(ctx, session, emp); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.LinkUserResponse{}, nil
}
