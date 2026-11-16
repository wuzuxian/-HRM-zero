package hrm_perf_period_participant

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfPeriodParticipantModel = (*customHrmPerfPeriodParticipantModel)(nil)

type (
	// HrmPerfPeriodParticipantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfPeriodParticipantModel.
	HrmPerfPeriodParticipantModel interface {
		hrmPerfPeriodParticipantModel
	}

	customHrmPerfPeriodParticipantModel struct {
		*defaultHrmPerfPeriodParticipantModel
	}
)

// NewHrmPerfPeriodParticipantModel returns a model for the database table.
func NewHrmPerfPeriodParticipantModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfPeriodParticipantModel {
	return &customHrmPerfPeriodParticipantModel{
		defaultHrmPerfPeriodParticipantModel: newHrmPerfPeriodParticipantModel(conn, op...),
	}
}
