package cardreplacement

import (
	"context"

	"github.com/jzero-io/jzero/core/status"
	"github.com/google/uuid"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

func employeeNotFound(employeeUuid string) error {
	return status.ErrorMessage(errcodes.CardReplacementEmployeeNotFoundCode, "员工不存在: "+employeeUuid)
}

func uuidNew() string {
	return uuid.New().String()
}

var _ = context.Background
