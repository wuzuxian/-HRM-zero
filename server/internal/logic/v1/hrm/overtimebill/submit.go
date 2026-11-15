package overtimebill

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
)

type Submit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSubmit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Submit {
	return &Submit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Submit 提交加班单（HR217 专属端点，区别于共享底座 submitBill）：
// 按命中加班规则判定 需审批（→ 审批中）/ 免审批（同一事务内 SubmitBill+ApproveBill
// 底座自动通过 → 审批通过，不落核定）。返回最终单据状态.
func (l *Submit) Submit(req *types.SubmitRequest) (resp *types.SubmitResponse, err error) {
	user := hrmuser.Current(l.ctx)
	st, err := SubmitOvertimeBill(WithSvcCtx(l.ctx, l.svcCtx), l.svcCtx.SqlxConn, l.svcCtx, req.Uuid,
		bill.Operator{UserUuid: user.Uuid, Name: user.Username}, "")
	if err != nil {
		return nil, err
	}
	l.Infof("overtimebill: 加班单已提交 billUuid=%s finalStatus=%s", req.Uuid, st)
	return &types.SubmitResponse{Status: st}, nil
}
