package hrm_recruitment_candidate

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmRecruitmentCandidateModel = (*customHrmRecruitmentCandidateModel)(nil)

type (
	// HrmRecruitmentCandidateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmRecruitmentCandidateModel.
	HrmRecruitmentCandidateModel interface {
		hrmRecruitmentCandidateModel
	}

	customHrmRecruitmentCandidateModel struct {
		*defaultHrmRecruitmentCandidateModel
	}
)

// NewHrmRecruitmentCandidateModel returns a model for the database table.
func NewHrmRecruitmentCandidateModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmRecruitmentCandidateModel {
	return &customHrmRecruitmentCandidateModel{
		defaultHrmRecruitmentCandidateModel: newHrmRecruitmentCandidateModel(conn, op...),
	}
}
