package bill

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/bill"
)

type Revoke struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRevoke(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Revoke {
	return &Revoke{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Revoke) Revoke(req *types.BillActionRequest) (resp *types.BillActionResponse, err error) {
	err = wrapBizErr(RevokeBill(l.ctx, l.svcCtx.SqlxConn, req.BillType, req.BillUuid, currentOperator(l.ctx), req.Opinion))
	if err != nil {
		return nil, err
	}

	return &types.BillActionResponse{}, nil
}
