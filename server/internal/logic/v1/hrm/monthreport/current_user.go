package monthreport

// 当前登录人（JWT ctx；logic 测试无认证上下文时返回空）。

import (
	"context"

	"github.com/jzero-io/jzero-admin/core-engine/helper/auth"
)

// currentUserUuid 当前登录用户 uuid。
func currentUserUuid(ctx context.Context) string {
	if info, err := auth.Info(ctx); err == nil {
		return info.Uuid
	}
	return ""
}

// currentUserName 当前登录用户名（登录账号；查无返回空）。
func currentUserName(ctx context.Context) string {
	if info, err := auth.Info(ctx); err == nil {
		return info.Username
	}
	return ""
}
