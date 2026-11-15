package leavebalance

import (
	"github.com/pkg/errors"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero/core/status"
)

func errEmployeeNotFound(uuid string) error {
	return status.ErrorMessage(errcodes.LeaveBalanceEmployeeNotFoundCode, "员工不存在: "+uuid)
}

func errBalanceNotFound() error {
	return status.ErrorMessage(errcodes.LeaveBalanceNotFoundCode, "余额行不存在")
}

var errNotFound = errors.New("balance row not found")
