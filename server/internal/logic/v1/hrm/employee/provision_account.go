package employee

import (
	"context"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user_role"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// Package-internal seam documentation (ADR 0007):
// 账号是员工档案的投影——档案是唯一写入方。本文件是全仓库唯一的开号接缝:
// 任何要把一份员工档案变成一个可登录系统账号的代码路径 (批量「生成用户」、
// 后续票的新增/导入/入职单审批建档) 都必须经 ProvisionAccount, 不允许自带实现。
//
// ProvisionAccount 在给定事务会话 (session 为 nil 时开独立事务) 内为一份员工档案
// 开通系统账号并返回账号 uuid。全部规则集中于此:
//   - 用户名 = 工号; 昵称/性别/手机/邮箱取自档案; 状态启用; 初始密码取配置 (兜底 123456)
//   - 自动挂默认角色 (hrm.defaultRoleCode, 兜底 R_EMPLOYEE); 角色不存在时降级:
//     只建号不挂角色并打 warn, 不返回错误 (ADR 0007: 开通流程不因配置缺失而失败)
//   - 离职 (person_status='5') 跳过开号, 返回空串
//   - 幂等: 已关联账号 (user_uuid 非空) 的档案直接返回既有 user_uuid
//   - 工号撞已有账号用户名报错 (EmployeeUserNoDuplicateCode, 1046x 段)

// DefaultRoleCodeFallback is the code-level default for hrm.defaultRoleCode.
// configcenter 装载路径不应用 struct 的 json default 标签, 空值兜底必须在代码里
// (与 InitialUserPassword 的既有兜底同一原因, 见 ticket HRM-49)。
const DefaultRoleCodeFallback = "R_EMPLOYEE"

// defaultInitialPassword 开号初始密码兜底, 与 HrmConf.InitialUserPassword 的
// json default 一致 (configcenter 不应用 struct default 的既有实测结论)。
const defaultInitialPassword = "123456"

// ProvisionAccount 为员工档案开通系统账号, 返回账号 uuid。
// session 非 nil 时在调用方事务内执行 (失败随外层整体回滚), nil 时开独立事务。
func ProvisionAccount(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, employeeUuid string) (string, error) {
	if session != nil {
		return provisionAccount(ctx, svcCtx, session, employeeUuid)
	}
	var userUuid string
	err := svcCtx.SqlxConn.TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		u, err := provisionAccount(ctx, svcCtx, tx, employeeUuid)
		userUuid = u
		return err
	})
	if err != nil {
		return "", err
	}
	return userUuid, nil
}

// accountOverrides 开号字段的调用方覆写 (HRM-51 新建并关联: 用户名/密码可自定义)。
// 空串 = 沿用接缝默认 (用户名=工号, 密码取配置兜底)。
type accountOverrides struct {
	username string
	password string
}

func provisionAccount(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, employeeUuid string) (string, error) {
	return provisionAccountWith(ctx, svcCtx, session, employeeUuid, accountOverrides{})
}

func provisionAccountWith(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, employeeUuid string, ov accountOverrides) (string, error) {
	row, err := findEmployeeOrNotFound(ctx, session, svcCtx, employeeUuid)
	if err != nil {
		return "", err
	}

	// 幂等: 已关联账号的档案直接返回既有指针
	if row.UserUuid != "" {
		return row.UserUuid, nil
	}

	// 离职 (补录历史员工) 不开号
	if row.PersonStatus == PersonStatusResigned {
		return "", nil
	}

	// 用户名默认=工号, 调用方可覆写; manage_user.username UNIQUE, 撞名 (历史脏数据/手工建过账号) 报错
	username := ov.username
	if username == "" {
		username = row.EmployeeNo
	}
	if _, err = svcCtx.Model.ManageUser.FindOneByUsername(ctx, session, username); err == nil {
		return "", status.ErrorMessage(errcodes.EmployeeUserNoDuplicateCode,
			"系统账号用户名已存在: "+username+" (与已有账号冲突)")
	} else if err != sqlx.ErrNotFound {
		return "", err
	}

	initialPassword := ov.password
	if initialPassword == "" {
		initialPassword = svcCtx.MustGetConfig().Hrm.InitialUserPassword
	}
	if initialPassword == "" {
		initialPassword = defaultInitialPassword
	}

	userUuid := uuidNew()
	if err = svcCtx.Model.ManageUser.InsertV2(ctx, session, &manage_user.ManageUser{
		Uuid:     userUuid,
		Username: username,
		Password: initialPassword,
		Nickname: row.Name,
		Gender:   row.Gender,
		Phone:    row.Phone,
		Status:   manage_user.StatusEnabled,
		Email:    row.Email,
	}); err != nil {
		return "", err
	}

	// 挂默认角色: 角色不存在时降级——只建号不挂角色, warn 不阻断 (ADR 0007)
	roleCode := svcCtx.MustGetConfig().Hrm.DefaultRoleCode
	if roleCode == "" {
		roleCode = DefaultRoleCodeFallback
	}
	role, roleErr := svcCtx.Model.ManageRole.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(manage_role.Code, roleCode).Build()...)
	switch {
	case roleErr == nil:
		if err = svcCtx.Model.ManageUserRole.InsertV2(ctx, session, &manage_user_role.ManageUserRole{
			Uuid:     uuidNew(),
			UserUuid: userUuid,
			RoleUuid: role.Uuid,
		}); err != nil {
			return "", err
		}
	case roleErr == sqlx.ErrNotFound:
		logx.WithContext(ctx).Errorf("hrm: 默认角色 %q 不存在, 账号 %s 已创建但未挂角色 (开号降级, 不阻断)", roleCode, row.EmployeeNo)
	default:
		return "", roleErr
	}

	row.UserUuid = userUuid
	if err = svcCtx.Model.HrmEmployee.Update(ctx, session, row); err != nil {
		return "", err
	}
	return userUuid, nil
}
