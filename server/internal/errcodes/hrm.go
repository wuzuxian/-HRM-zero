package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 组织管理 error codes (1032x range, manage.go uses 10311).
	OrgParentNotFoundCode = 10321
	OrgNotFoundCode       = 10322
	OrgHasChildrenCode    = 10323
	OrgImportInvalidCode  = 10324
)

func init() {
	status.Register(OrgParentNotFoundCode)
	status.Register(OrgNotFoundCode)
	status.Register(OrgHasChildrenCode)
	status.Register(OrgImportInvalidCode)
}
