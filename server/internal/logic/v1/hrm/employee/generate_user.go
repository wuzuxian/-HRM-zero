package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
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

// GenerateUser 员工一键生成系统账号 (HRM-04, spec 用户故事 12): 单个/批量;
// 用户名=工号, 昵称=姓名, 初始密码取配置 (Hrm.InitialUserPassword);
// 已生成过账号 (user_uuid 非空) 的员工自动跳过 (幂等).
// 全部生成动作在单个事务内, 任一员工失败整体回滚.
func (l *GenerateUser) GenerateUser(req *types.GenerateUserRequest) (resp *types.GenerateUserResponse, err error) {
	resp = &types.GenerateUserResponse{}
	if len(req.Uuids) == 0 {
		return resp, nil
	}

	initialPassword := l.svcCtx.MustGetConfig().Hrm.InitialUserPassword
	// 兜底: configcenter 装载路径不应用 struct 的 default 标签 (实测 InitialUserPassword 为空),
	// 空密码会让生成的账号永远无法登录; 与配置注释的默认值保持一致
	if initialPassword == "" {
		initialPassword = "123456"
	}

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, employeeUuid := range req.Uuids {
			row, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, employeeUuid)
			if err != nil {
				if err == sqlx.ErrNotFound {
					return status.ErrorMessage(errcodes.EmployeeUserNotFoundCode, "员工不存在: "+employeeUuid)
				}
				return err
			}

			// 幂等: 已生成过账号的员工跳过
			if row.UserUuid != "" {
				resp.Skipped++
				continue
			}

			// 用户名=工号; manage_user.username UNIQUE, 撞名 (如历史脏数据/手工建过账号) 报错
			_, err = l.svcCtx.Model.ManageUser.FindOneByUsername(ctx, session, row.EmployeeNo)
			if err == nil {
				return status.ErrorMessage(errcodes.EmployeeUserNoDuplicateCode,
					"系统账号用户名已存在: "+row.EmployeeNo+" (工号与已有账号冲突)")
			}
			if err != sqlx.ErrNotFound {
				return err
			}

			userUuid := uuidNew()
			if err = l.svcCtx.Model.ManageUser.InsertV2(ctx, session, &manage_user.ManageUser{
				Uuid:     userUuid,
				Username: row.EmployeeNo,
				Password: initialPassword,
				Nickname: row.Name,
				Gender:   row.Gender,
				Phone:    row.Phone,
				Status:   "1",
				Email:    row.Email,
			}); err != nil {
				return err
			}

			row.UserUuid = userUuid
			if err = l.svcCtx.Model.HrmEmployee.Update(ctx, session, row); err != nil {
				return err
			}
			resp.Generated++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
