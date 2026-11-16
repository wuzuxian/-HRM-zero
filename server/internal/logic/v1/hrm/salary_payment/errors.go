package salary_payment

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 业务错误码便捷包装（1071x 段）。

func statusNotFound(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPaymentNotFoundCode, msg)
}

func statusInvalid(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPaymentInvalidCode, msg)
}

func statusBatchNotConfirmed(msg string) error {
	return status.ErrorMessage(errcodes.SalaryPaymentBatchNotConfirmedCode, msg)
}
