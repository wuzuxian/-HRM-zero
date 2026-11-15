package leave

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leave"
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

// Delete 删除/批量删除请假销假单：仅 未提交/审批不通过/已撤销 状态可删除.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	for _, billUuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmLeaveBill.FindOneByUuid(l.ctx, nil, billUuid)
		if err != nil {
			if err == hrm_leave_bill.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.LeaveNotFoundCode, "请假单不存在")
			}
			return nil, err
		}
		if !canEdit(row.Status) {
			return nil, status.ErrorMessage(errcodes.LeaveNotEditableCode,
				"单据 "+row.BillNo+" 当前状态不允许删除（仅 未提交/审批不通过/已撤销 可删）")
		}
	}

	err = l.svcCtx.Model.HrmLeaveBill.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_leave_bill.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
