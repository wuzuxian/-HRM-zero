package salary_payslip

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
)

type Read struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRead(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Read {
	return &Read{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Read 员工本人查看工资条记 已读+已读时间（幂等不覆盖首次）。
func (l *Read) Read(req *types.ReadRequest) (resp *types.ReadResponse, err error) {
	if err := MarkRead(l.ctx, l.svcCtx.SqlxConn, req.Uuid); err != nil {
		return nil, err
	}
	return &types.ReadResponse{}, nil
}
