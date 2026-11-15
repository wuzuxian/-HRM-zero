package hrm_tax_special_deduction

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTaxSpecialDeductionModel = (*customHrmTaxSpecialDeductionModel)(nil)

type (
	// HrmTaxSpecialDeductionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTaxSpecialDeductionModel.
	HrmTaxSpecialDeductionModel interface {
		hrmTaxSpecialDeductionModel
	}

	customHrmTaxSpecialDeductionModel struct {
		*defaultHrmTaxSpecialDeductionModel
	}
)

// NewHrmTaxSpecialDeductionModel returns a model for the database table.
func NewHrmTaxSpecialDeductionModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTaxSpecialDeductionModel {
	return &customHrmTaxSpecialDeductionModel{
		defaultHrmTaxSpecialDeductionModel: newHrmTaxSpecialDeductionModel(conn, op...),
	}
}
