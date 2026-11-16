package config

import "github.com/jzero-io/jzero-admin/core-engine/config"

// HrmConf 人力资源模块配置.
type HrmConf struct {
	// InitialUserPassword 员工一键生成系统账号 (HRM-04) 的初始密码, 默认 123456
	// (与 1_init.up.sql 种子 admin 账号一致; 密码在本系统为明文存储的既有约定).
	InitialUserPassword string `json:",default=123456"`
	// AttendanceLockDay 考勤月报自动锁定日 (spec 定时任务: 次月锁定日可配置, 如 5)。
	// 每天 AttendanceAutoLockJob 检查: today 日 == AttendanceLockDay 时自动锁定
	// 上一个自然月的月报 (已锁定的月跳过, 幂等)。默认 5。
	AttendanceLockDay int `json:",default=5"`
}

type Config struct {
	config.Config
	Hrm HrmConf `json:",optional"`
}
