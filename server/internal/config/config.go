package config

import "github.com/jzero-io/jzero-admin/core-engine/config"

// HrmConf 人力资源模块配置.
type HrmConf struct {
	// InitialUserPassword 员工一键生成系统账号 (HRM-04) 的初始密码, 默认 123456
	// (与 1_init.up.sql 种子 admin 账号一致; 密码在本系统为明文存储的既有约定).
	InitialUserPassword string `json:",default=123456"`
}

type Config struct {
	config.Config
	Hrm HrmConf `json:",optional"`
}
