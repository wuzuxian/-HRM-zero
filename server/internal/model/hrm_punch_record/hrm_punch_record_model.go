package hrm_punch_record

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPunchRecordModel = (*customHrmPunchRecordModel)(nil)

type (
	// HrmPunchRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPunchRecordModel.
	HrmPunchRecordModel interface {
		hrmPunchRecordModel
	}

	customHrmPunchRecordModel struct {
		*defaultHrmPunchRecordModel
	}
)

// NewHrmPunchRecordModel returns a model for the database table.
func NewHrmPunchRecordModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPunchRecordModel {
	return &customHrmPunchRecordModel{
		defaultHrmPunchRecordModel: newHrmPunchRecordModel(conn, op...),
	}
}
