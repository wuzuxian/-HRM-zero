package training

import (
	"context"

	"github.com/jzero-io/jzero-admin/core-engine/helper/auth"
)

// currentUserUuid extracts the current login user's uuid from the JWT context
// (helper/auth.Info reads ctx "uuid"). Empty when not authenticated.
func currentUserUuid(ctx context.Context) string {
	if info, err := auth.Info(ctx); err == nil {
		return info.Uuid
	}
	return ""
}
