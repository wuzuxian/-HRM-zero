package manage_user

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound

// manage_user.status 枚举 (ADR 0007): '1'=启用 / '2'=禁用。全仓库共用同一常量,
// 不再散落字面量: 写入方 = 开号接缝 (hrm/employee ProvisionAccount, 启用) 与
// 离职禁用规则 (hrm/employee DisableLinkedAccount); 消费方 = 登录状态校验 (auth 包)。
const (
	StatusEnabled  = "1"
	StatusDisabled = "2"
)
