package bill

import (
	"context"

	"github.com/jzero-io/jzero-admin/core-engine/helper/auth"
)

// currentOperator builds the audit Operator from the JWT context.
// The username is snapshotted into the approval record as 操作人.
func currentOperator(ctx context.Context) Operator {
	op := Operator{}
	if info, err := auth.Info(ctx); err == nil {
		op.UserUuid = info.Uuid
		op.Name = info.Username
	}
	return op
}
