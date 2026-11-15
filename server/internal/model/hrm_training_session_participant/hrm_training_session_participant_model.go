package hrm_training_session_participant

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTrainingSessionParticipantModel = (*customHrmTrainingSessionParticipantModel)(nil)

type (
	// HrmTrainingSessionParticipantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTrainingSessionParticipantModel.
	HrmTrainingSessionParticipantModel interface {
		hrmTrainingSessionParticipantModel
	}

	customHrmTrainingSessionParticipantModel struct {
		*defaultHrmTrainingSessionParticipantModel
	}
)

// NewHrmTrainingSessionParticipantModel returns a model for the database table.
func NewHrmTrainingSessionParticipantModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTrainingSessionParticipantModel {
	return &customHrmTrainingSessionParticipantModel{
		defaultHrmTrainingSessionParticipantModel: newHrmTrainingSessionParticipantModel(conn, op...),
	}
}
