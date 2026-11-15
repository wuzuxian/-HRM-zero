package hrm_resignation_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmResignationBillModel = (*customHrmResignationBillModel)(nil)

type (
	// HrmResignationBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmResignationBillModel.
	HrmResignationBillModel interface {
		hrmResignationBillModel
	}

	customHrmResignationBillModel struct {
		*defaultHrmResignationBillModel
	}
)

// NewHrmResignationBillModel returns a model for the database table.
func NewHrmResignationBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmResignationBillModel {
	return &customHrmResignationBillModel{
		defaultHrmResignationBillModel: newHrmResignationBillModel(conn, op...),
	}
}
