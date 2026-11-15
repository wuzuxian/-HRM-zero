package hrm_recruitment_requirement

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmRecruitmentRequirementModel = (*customHrmRecruitmentRequirementModel)(nil)

type (
	// HrmRecruitmentRequirementModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmRecruitmentRequirementModel.
	HrmRecruitmentRequirementModel interface {
		hrmRecruitmentRequirementModel
	}

	customHrmRecruitmentRequirementModel struct {
		*defaultHrmRecruitmentRequirementModel
	}
)

// NewHrmRecruitmentRequirementModel returns a model for the database table.
func NewHrmRecruitmentRequirementModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmRecruitmentRequirementModel {
	return &customHrmRecruitmentRequirementModel{
		defaultHrmRecruitmentRequirementModel: newHrmRecruitmentRequirementModel(conn, op...),
	}
}
