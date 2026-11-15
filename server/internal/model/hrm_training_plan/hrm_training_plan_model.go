package hrm_training_plan

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTrainingPlanModel = (*customHrmTrainingPlanModel)(nil)

type (
	// HrmTrainingPlanModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTrainingPlanModel.
	HrmTrainingPlanModel interface {
		hrmTrainingPlanModel
	}

	customHrmTrainingPlanModel struct {
		*defaultHrmTrainingPlanModel
	}
)

// NewHrmTrainingPlanModel returns a model for the database table.
func NewHrmTrainingPlanModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTrainingPlanModel {
	return &customHrmTrainingPlanModel{
		defaultHrmTrainingPlanModel: newHrmTrainingPlanModel(conn, op...),
	}
}
