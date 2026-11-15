package perf_scheme

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// wrapBizErr maps 绩效方案 sentinel errors to registered business error codes;
// unknown errors pass through unchanged.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrSchemeNotFound):
		return status.Wrap(errcodes.HrmPerfSchemeNotFoundCode, err)
	case errors.Is(err, ErrSchemeTemplateNotFound):
		return status.Wrap(errcodes.HrmPerfSchemeTemplateNotFoundCode, err)
	case errors.Is(err, ErrSchemeNameRequired),
		errors.Is(err, ErrSchemeTemplateRequired),
		errors.Is(err, ErrSchemeInvalid):
		return status.Wrap(errcodes.HrmPerfSchemeInvalidCode, err)
	default:
		return err
	}
}
