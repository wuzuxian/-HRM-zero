package perf_template

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// wrapBizErr maps 考核模板 sentinel errors to registered business error codes;
// unknown errors pass through unchanged.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrTemplateNotFound):
		return status.Wrap(errcodes.HrmPerfTemplateNotFoundCode, err)
	case errors.Is(err, ErrTemplateItemInvalid):
		return status.Wrap(errcodes.HrmPerfTemplateItemInvalidCode, err)
	case errors.Is(err, ErrTemplateItemWeightSum):
		return status.Wrap(errcodes.HrmPerfTemplateWeightSumCode, err)
	default:
		return err
	}
}
