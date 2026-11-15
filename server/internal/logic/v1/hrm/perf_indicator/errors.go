package perf_indicator

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// wrapBizErr maps 指标库 sentinel errors to registered business error codes;
// unknown errors pass through unchanged.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrIndicatorCodeDuplicated):
		return status.Wrap(errcodes.HrmPerfIndicatorCodeDuplicatedCode, err)
	case errors.Is(err, ErrIndicatorNotFound):
		return status.Wrap(errcodes.HrmPerfIndicatorNotFoundCode, err)
	case errors.Is(err, ErrIndicatorInvalid):
		return status.Wrap(errcodes.HrmPerfIndicatorInvalidCode, err)
	case errors.Is(err, ErrIndicatorInUse):
		return status.Wrap(errcodes.HrmPerfIndicatorInUseCode, err)
	default:
		return err
	}
}
