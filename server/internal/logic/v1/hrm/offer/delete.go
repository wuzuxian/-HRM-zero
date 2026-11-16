package offer

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_offer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/offer"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Delete 删除/批量删除录用审批单：仅 未提交/审批不通过/已撤销 状态可删除
// （审批通过的单据已联动 职位已录用/候选人 hired/入职草稿单，不可删）.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	for _, billUuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmOfferBill.FindOneByUuid(l.ctx, nil, billUuid)
		if err != nil {
			if err == hrm_offer_bill.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.OfferNotFoundCode, "录用审批单不存在")
			}
			return nil, err
		}
		if !canEdit(row.Status) {
			return nil, status.ErrorMessage(errcodes.OfferNotEditableCode,
				"单据 "+row.BillNo+" 当前状态不允许删除（仅 未提交/审批不通过/已撤销 可删）")
		}
	}

	err = l.svcCtx.Model.HrmOfferBill.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_offer_bill.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
