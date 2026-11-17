package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type UnlinkUser struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUnlinkUser(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *UnlinkUser {
	return &UnlinkUser{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// UnlinkUser 解除关联 (HRM-51): 只清档案 user_uuid 指针。
// 账号行与 manage_user_role 角色绑定原样保留 (ADR 0007: 解绑保留账号,
// 禁用/删除账号属账号侧生命周期管理, 不在解绑语义内)。未关联时幂等成功。
func (l *UnlinkUser) UnlinkUser(req *types.UnlinkUserRequest) (resp *types.UnlinkUserResponse, err error) {
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		emp, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, req.EmployeeUuid)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return status.ErrorMessage(errcodes.EmployeeUserNotFoundCode, "员工不存在: "+req.EmployeeUuid)
			}
			return err
		}
		if emp.UserUuid == "" {
			return nil
		}
		emp.UserUuid = ""
		if err = l.svcCtx.Model.HrmEmployee.Update(ctx, session, emp); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.UnlinkUserResponse{}, nil
}
