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

type GenerateUser struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGenerateUser(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GenerateUser {
	return &GenerateUser{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GenerateUser 员工一键生成系统账号 (HRM-04): 单个/批量; 全部生成动作在单个事务内,
// 任一员工失败整体回滚。
//
// HRM-49 起 open-account 规则 (用户名=工号、初始密码取配置、挂默认角色、
// 离职跳过、已关联幂等、工号撞账号用户名报错) 全部收敛到唯一接缝
// ProvisionAccount (ADR 0007), 本端点只是批量驱动它, 不自带一份实现。
func (l *GenerateUser) GenerateUser(req *types.GenerateUserRequest) (resp *types.GenerateUserResponse, err error) {
	resp = &types.GenerateUserResponse{}
	if len(req.Uuids) == 0 {
		return resp, nil
	}

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, employeeUuid := range req.Uuids {
			// 先看档案: 已关联账号 → 幂等跳过 (接缝也会幂等, 这里只为计数语义)
			emp, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, employeeUuid)
			if err != nil {
				if err == sqlx.ErrNotFound {
					return status.ErrorMessage(errcodes.EmployeeUserNotFoundCode, "员工不存在: "+employeeUuid)
				}
				return err
			}

			userUuid, err := provisionAccount(ctx, l.svcCtx, session, employeeUuid)
			if err != nil {
				return err
			}
			switch {
			case userUuid == "":
				// 离职跳过: 不计入任何计数
			case emp.UserUuid != "":
				// 幂等: 本次调用前已关联账号
				resp.Skipped++
			default:
				resp.Generated++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
