package hrm_overtime_rule

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmOvertimeRuleModel = (*customHrmOvertimeRuleModel)(nil)

type (
	// HrmOvertimeRuleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmOvertimeRuleModel.
	HrmOvertimeRuleModel interface {
		hrmOvertimeRuleModel
	}

	customHrmOvertimeRuleModel struct {
		*defaultHrmOvertimeRuleModel
	}
)

// NewHrmOvertimeRuleModel returns a model for the database table.
func NewHrmOvertimeRuleModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmOvertimeRuleModel {
	return &customHrmOvertimeRuleModel{
		defaultHrmOvertimeRuleModel: newHrmOvertimeRuleModel(conn, op...),
	}
}
