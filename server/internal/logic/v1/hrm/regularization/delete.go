package regularization

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_regularization_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/regularization"
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

// Delete 删除/批量删除转正单：仅 未提交/审批不通过/已撤销 状态可删除.
// 审批通过的单据已回写员工档案（applied=1），随 审批通过 状态一并锁定，不可删除.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	for _, billUuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmRegularizationBill.FindOneByUuid(l.ctx, nil, billUuid)
		if err != nil {
			if err == hrm_regularization_bill.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.RegularizationNotFoundCode, "转正单不存在")
			}
			return nil, err
		}
		if !canEdit(row.Status) {
			return nil, status.ErrorMessage(errcodes.RegularizationNotEditableCode,
				"单据 "+row.BillNo+" 当前状态不允许删除（仅 未提交/审批不通过/已撤销 可删）")
		}
	}

	err = l.svcCtx.Model.HrmRegularizationBill.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_regularization_bill.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
