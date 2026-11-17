package config

import "github.com/jzero-io/jzero-admin/core-engine/config"

// HrmConf 人力资源模块配置.
type HrmConf struct {
	// InitialUserPassword 员工一键生成系统账号 (HRM-04) 的初始密码, 默认 123456
	// (与 1_init.up.sql 种子 admin 账号一致; 密码在本系统为明文存储的既有约定).
	InitialUserPassword string `json:",default=123456"`
	// DefaultRoleCode 员工开号 (HRM-49, ADR 0007) 自动挂载的默认角色 code, 默认 R_EMPLOYEE.
	// 注意: configcenter 装载路径不应用 struct 的 default 标签, 消费方必须对空值做代码级兜底
	// (见 employee.ProvisionAccount), 不能只靠 json:",default=".
	DefaultRoleCode string `json:",default=R_EMPLOYEE"`
	// AttendanceLockDay 考勤月报自动锁定日 (spec 定时任务: 次月锁定日可配置, 如 5)。
	// 每天 AttendanceAutoLockJob 检查: today 日 == AttendanceLockDay 时自动锁定
	// 上一个自然月的月报 (已锁定的月跳过, 幂等)。默认 5。
	AttendanceLockDay int `json:",default=5"`
}

type Config struct {
	config.Config
	Hrm HrmConf `json:",optional"`
}
