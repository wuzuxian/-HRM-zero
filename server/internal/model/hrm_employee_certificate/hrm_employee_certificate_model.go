package hrm_employee_certificate

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmEmployeeCertificateModel = (*customHrmEmployeeCertificateModel)(nil)

type (
	// HrmEmployeeCertificateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmEmployeeCertificateModel.
	HrmEmployeeCertificateModel interface {
		hrmEmployeeCertificateModel
	}

	customHrmEmployeeCertificateModel struct {
		*defaultHrmEmployeeCertificateModel
	}
)

// NewHrmEmployeeCertificateModel returns a model for the database table.
func NewHrmEmployeeCertificateModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmEmployeeCertificateModel {
	return &customHrmEmployeeCertificateModel{
		defaultHrmEmployeeCertificateModel: newHrmEmployeeCertificateModel(conn, op...),
	}
}
