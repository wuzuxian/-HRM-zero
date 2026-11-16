package salary_batch

import (
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 业务错误码便捷包装（logic 层统一走这些小包装，便于测试断言与后续重编码）。

func statusNotFound(msg string) error {
	return status.ErrorMessage(errcodes.SalaryBatchNotFoundCode, msg)
}

func statusInvalid(msg string) error {
	return status.ErrorMessage(errcodes.SalaryBatchInvalidCode, msg)
}

func statusSubjectNotFound(msg string) error {
	return status.ErrorMessage(errcodes.SalaryBatchSubjectNotFoundCode, msg)
}

func statusDuplicate(msg string) error {
	return status.ErrorMessage(errcodes.SalaryBatchDuplicateCode, msg)
}

func statusTrialFailed(msg string) error {
	return status.ErrorMessage(errcodes.SalaryBatchTrialFailedCode, msg)
}

func statusNoResult(msg string) error {
	return status.ErrorMessage(errcodes.SalaryBatchNoResultCode, msg)
}
