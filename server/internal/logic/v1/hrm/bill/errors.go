package bill

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// wrapBizErr maps the shared lifecycle sentinel errors to registered business
// error codes; unknown errors pass through unchanged.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrBillTypeNotRegistered):
		return status.Wrap(errcodes.HrmBillTypeNotRegisteredCode, err)
	case errors.Is(err, ErrBillNotFound):
		return status.Wrap(errcodes.HrmBillNotFoundCode, err)
	case errors.Is(err, ErrInvalidTransition):
		return status.Wrap(errcodes.HrmBillInvalidTransitionCode, err)
	case errors.Is(err, ErrInvalidTypeCode), errors.Is(err, ErrBillNoSeqOutOfRange):
		return status.Wrap(errcodes.HrmBillInvalidTypeCode, err)
	default:
		return err
	}
}
