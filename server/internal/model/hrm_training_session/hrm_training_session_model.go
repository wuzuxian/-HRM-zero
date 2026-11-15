package hrm_training_session

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTrainingSessionModel = (*customHrmTrainingSessionModel)(nil)

type (
	// HrmTrainingSessionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTrainingSessionModel.
	HrmTrainingSessionModel interface {
		hrmTrainingSessionModel
	}

	customHrmTrainingSessionModel struct {
		*defaultHrmTrainingSessionModel
	}
)

// NewHrmTrainingSessionModel returns a model for the database table.
func NewHrmTrainingSessionModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTrainingSessionModel {
	return &customHrmTrainingSessionModel{
		defaultHrmTrainingSessionModel: newHrmTrainingSessionModel(conn, op...),
	}
}
