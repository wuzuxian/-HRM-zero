package hrm_reward_bill

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmRewardBillModel = (*customHrmRewardBillModel)(nil)

type (
	// HrmRewardBillModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmRewardBillModel.
	HrmRewardBillModel interface {
		hrmRewardBillModel
	}

	customHrmRewardBillModel struct {
		*defaultHrmRewardBillModel
	}
)

// NewHrmRewardBillModel returns a model for the database table.
func NewHrmRewardBillModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmRewardBillModel {
	return &customHrmRewardBillModel{
		defaultHrmRewardBillModel: newHrmRewardBillModel(conn, op...),
	}
}
