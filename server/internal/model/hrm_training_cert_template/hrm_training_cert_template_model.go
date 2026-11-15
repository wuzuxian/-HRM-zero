package hrm_training_cert_template

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTrainingCertTemplateModel = (*customHrmTrainingCertTemplateModel)(nil)

type (
	// HrmTrainingCertTemplateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTrainingCertTemplateModel.
	HrmTrainingCertTemplateModel interface {
		hrmTrainingCertTemplateModel
	}

	customHrmTrainingCertTemplateModel struct {
		*defaultHrmTrainingCertTemplateModel
	}
)

// NewHrmTrainingCertTemplateModel returns a model for the database table.
func NewHrmTrainingCertTemplateModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTrainingCertTemplateModel {
	return &customHrmTrainingCertTemplateModel{
		defaultHrmTrainingCertTemplateModel: newHrmTrainingCertTemplateModel(conn, op...),
	}
}
