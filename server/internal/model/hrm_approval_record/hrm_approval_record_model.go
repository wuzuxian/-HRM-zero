package hrm_approval_record

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmApprovalRecordModel = (*customHrmApprovalRecordModel)(nil)

type (
	// HrmApprovalRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmApprovalRecordModel.
	HrmApprovalRecordModel interface {
		hrmApprovalRecordModel
	}

	customHrmApprovalRecordModel struct {
		*defaultHrmApprovalRecordModel
	}
)

// NewHrmApprovalRecordModel returns a model for the database table.
func NewHrmApprovalRecordModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmApprovalRecordModel {
	return &customHrmApprovalRecordModel{
		defaultHrmApprovalRecordModel: newHrmApprovalRecordModel(conn, op...),
	}
}
