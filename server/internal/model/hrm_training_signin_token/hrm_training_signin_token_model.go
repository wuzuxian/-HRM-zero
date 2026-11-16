package hrm_training_signin_token

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmTrainingSigninTokenModel = (*customHrmTrainingSigninTokenModel)(nil)

type (
	// HrmTrainingSigninTokenModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmTrainingSigninTokenModel.
	HrmTrainingSigninTokenModel interface {
		hrmTrainingSigninTokenModel
	}

	customHrmTrainingSigninTokenModel struct {
		*defaultHrmTrainingSigninTokenModel
	}
)

// NewHrmTrainingSigninTokenModel returns a model for the database table.
func NewHrmTrainingSigninTokenModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmTrainingSigninTokenModel {
	return &customHrmTrainingSigninTokenModel{
		defaultHrmTrainingSigninTokenModel: newHrmTrainingSigninTokenModel(conn, op...),
	}
}
