package hrm_shift

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmShiftModel = (*customHrmShiftModel)(nil)

type (
	// HrmShiftModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmShiftModel.
	HrmShiftModel interface {
		hrmShiftModel
	}

	customHrmShiftModel struct {
		*defaultHrmShiftModel
	}
)

// NewHrmShiftModel returns a model for the database table.
func NewHrmShiftModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmShiftModel {
	return &customHrmShiftModel{
		defaultHrmShiftModel: newHrmShiftModel(conn, op...),
	}
}
