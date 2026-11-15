package hrm_tax_opening

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTaxOpeningModel = (*customHrmTaxOpeningModel)(nil)

type (
	// HrmTaxOpeningModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTaxOpeningModel.
	HrmTaxOpeningModel interface {
		hrmTaxOpeningModel
	}

	customHrmTaxOpeningModel struct {
		*defaultHrmTaxOpeningModel
	}
)

// NewHrmTaxOpeningModel returns a model for the database table.
func NewHrmTaxOpeningModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTaxOpeningModel {
	return &customHrmTaxOpeningModel{
		defaultHrmTaxOpeningModel: newHrmTaxOpeningModel(conn, op...),
	}
}
