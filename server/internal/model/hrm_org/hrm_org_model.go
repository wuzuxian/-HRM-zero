package hrm_org

import (
	"github.com/eddieowens/opts"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HrmOrgModel = (*customHrmOrgModel)(nil)

type (
	// HrmOrgModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHrmOrgModel.
	HrmOrgModel interface {
		hrmOrgModel
	}

	customHrmOrgModel struct {
		*defaultHrmOrgModel
	}
)

// NewHrmOrgModel returns a model for the database table.
func NewHrmOrgModel(conn sqlx.SqlConn, op ...opts.Opt[modelx.ModelOpts]) HrmOrgModel {
	return &customHrmOrgModel{
		defaultHrmOrgModel: newHrmOrgModel(conn, op...),
	}
}
