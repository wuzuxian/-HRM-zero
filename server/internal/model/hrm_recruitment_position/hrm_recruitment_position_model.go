package hrm_recruitment_position

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmRecruitmentPositionModel = (*customHrmRecruitmentPositionModel)(nil)

type (
	// HrmRecruitmentPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmRecruitmentPositionModel.
	HrmRecruitmentPositionModel interface {
		hrmRecruitmentPositionModel
	}

	customHrmRecruitmentPositionModel struct {
		*defaultHrmRecruitmentPositionModel
	}
)

// NewHrmRecruitmentPositionModel returns a model for the database table.
func NewHrmRecruitmentPositionModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmRecruitmentPositionModel {
	return &customHrmRecruitmentPositionModel{
		defaultHrmRecruitmentPositionModel: newHrmRecruitmentPositionModel(conn, op...),
	}
}
