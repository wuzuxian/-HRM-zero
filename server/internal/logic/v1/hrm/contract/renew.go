package contract

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/contract"
)

type Renew struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRenew(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Renew {
	return &Renew{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Renew 合同续签（spec 用户故事 15）：续签次数 +1、生成新到期日
// （显式指定则取指定值；缺省 = 原到期日 +1 年）。已解除合同不允许续签。
// 幂等性说明：续签次数严格按当前值 +1（read-modify-write 单条 Update），同一请求
// 不重复提交多次则只加一次；重复续签是合法业务动作（每次续签次数再 +1），
// 与入职单「已建档不可重复」的幂等语义不同。
func (l *Renew) Renew(req *types.RenewRequest) (resp *types.RenewResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.ContractNotFoundCode, "uuid 不能为空")
	}
	row, err := findContractByUuid(l.ctx, l.svcCtx, req.Uuid)
	if err != nil {
		return nil, err
	}
	if row.Status == ContractStatusTerminated {
		return nil, status.ErrorMessage(errcodes.ContractNotRenewableCode, "已解除合同不允许续签")
	}

	newEndDate, err := NextExpiryDate(row.EndDate, req.EndDate)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.ContractInvalidCode, err.Error())
	}

	row.EndDate = newEndDate
	row.RenewCount = row.RenewCount + 1
	if row.Status == ContractStatusExpired {
		row.Status = ContractStatusActive // 续签后回到有效
	}

	if err = l.svcCtx.Model.HrmEmployeeContract.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	return &types.RenewResponse{}, nil
}
