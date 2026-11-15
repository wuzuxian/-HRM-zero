package entry

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
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

// Delete 删除/批量删除入职单：仅 未提交/审批不通过/已撤销 状态可删除.
// 已审批通过的单据已自动建档（employee_uuid 非空），随 审批通过 状态一并锁定，
// 不可删除 — 档案的生命周期归 员工档案管理 (HRM-03).
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	rows := make([]*hrm_entry_bill.HrmEntryBill, 0, len(req.Uuids))
	for _, billUuid := range req.Uuids {
		row, err := l.svcCtx.Model.HrmEntryBill.FindOneByUuid(l.ctx, nil, billUuid)
		if err != nil {
			if err == hrm_entry_bill.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.EntryNotFoundCode, "入职单不存在")
			}
			return nil, err
		}
		if !canEdit(row.Status) {
			return nil, status.ErrorMessage(errcodes.EntryNotEditableCode,
				"单据 "+row.BillNo+" 当前状态不允许删除（仅 未提交/审批不通过/已撤销 可删）")
		}
		if row.EmployeeUuid != "" {
			return nil, status.ErrorMessage(errcodes.EntryArchivedCode,
				"单据 "+row.BillNo+" 已自动建档，不允许删除")
		}
		rows = append(rows, row)
	}

	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, row := range rows {
			if err := deleteChildren(ctx, session, l.svcCtx, row.Uuid); err != nil {
				return err
			}
		}
		return l.svcCtx.Model.HrmEntryBill.DeleteByCondition(l.ctx, session, condition.Condition{
			Field:    hrm_entry_bill.Uuid,
			Operator: condition.In,
			Value:    req.Uuids,
		})
	}); txErr != nil {
		return nil, txErr
	}

	return &types.DeleteResponse{}, nil
}
