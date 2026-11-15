package hrm_tax_profile

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTaxProfileModel = (*customHrmTaxProfileModel)(nil)

type (
	// HrmTaxProfileModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTaxProfileModel.
	HrmTaxProfileModel interface {
		hrmTaxProfileModel
	}

	customHrmTaxProfileModel struct {
		*defaultHrmTaxProfileModel
	}
)

// NewHrmTaxProfileModel returns a model for the database table.
func NewHrmTaxProfileModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTaxProfileModel {
	return &customHrmTaxProfileModel{
		defaultHrmTaxProfileModel: newHrmTaxProfileModel(conn, op...),
	}
}
