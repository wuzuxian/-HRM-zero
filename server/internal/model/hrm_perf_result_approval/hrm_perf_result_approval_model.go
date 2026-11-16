package hrm_perf_result_approval

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmPerfResultApprovalModel = (*customHrmPerfResultApprovalModel)(nil)

type (
	// HrmPerfResultApprovalModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmPerfResultApprovalModel.
	HrmPerfResultApprovalModel interface {
		hrmPerfResultApprovalModel
	}

	customHrmPerfResultApprovalModel struct {
		*defaultHrmPerfResultApprovalModel
	}
)

// NewHrmPerfResultApprovalModel returns a model for the database table.
func NewHrmPerfResultApprovalModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmPerfResultApprovalModel {
	return &customHrmPerfResultApprovalModel{
		defaultHrmPerfResultApprovalModel: newHrmPerfResultApprovalModel(conn, op...),
	}
}
