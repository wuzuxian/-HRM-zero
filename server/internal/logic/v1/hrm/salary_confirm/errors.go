package salary_confirm

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 业务错误码便捷包装（1071x 段，与 salary_batch 同款小包装模式）。

func statusNotFound(msg string) error {
	return status.ErrorMessage(errcodes.SalaryConfirmNotFoundCode, msg)
}

func statusInvalid(msg string) error {
	return status.ErrorMessage(errcodes.SalaryConfirmInvalidCode, msg)
}
