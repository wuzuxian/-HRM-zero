package hrm_training_certificate

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTrainingCertificateModel = (*customHrmTrainingCertificateModel)(nil)

type (
	// HrmTrainingCertificateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTrainingCertificateModel.
	HrmTrainingCertificateModel interface {
		hrmTrainingCertificateModel
	}

	customHrmTrainingCertificateModel struct {
		*defaultHrmTrainingCertificateModel
	}
)

// NewHrmTrainingCertificateModel returns a model for the database table.
func NewHrmTrainingCertificateModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTrainingCertificateModel {
	return &customHrmTrainingCertificateModel{
		defaultHrmTrainingCertificateModel: newHrmTrainingCertificateModel(conn, op...),
	}
}
