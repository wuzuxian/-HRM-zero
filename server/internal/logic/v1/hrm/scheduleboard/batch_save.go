package scheduleboard

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type BatchSave struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewBatchSave(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *BatchSave {
	return &BatchSave{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *BatchSave) BatchSave(req *types.BatchSaveRequest) (resp *types.BatchSaveResponse, err error) {
	// 批量保存（矩阵整表保存）：事务内逐格 upsert，同键覆盖；任一格非法整批回滚。
	if len(req.Cells) == 0 {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, "没有要保存的单元格")
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}
	memberSet := make(map[string]bool, len(members))
	for _, u := range members {
		memberSet[u] = true
	}
	shiftUuids := make([]string, 0, len(req.Cells))
	for _, c := range req.Cells {
		if !memberSet[c.EmployeeUuid] {
			return nil, status.ErrorMessage(errcodes.ScheduleBoardEmployeeNotFoundCode, "员工不在该考勤组适用范围内")
		}
		if _, err = validateDate(c.Date); err != nil {
			return nil, err
		}
		shiftUuids = append(shiftUuids, c.ShiftUuid)
	}
	if err = checkShifts(l.ctx, l.svcCtx, shiftUuids); err != nil {
		return nil, err
	}

	var saved int64
	if err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, c := range req.Cells {
			if err := upsertEntry(ctx, l.svcCtx, session, c.EmployeeUuid, c.Date, c.ShiftUuid, req.GroupUuid, false); err != nil {
				return err
			}
			saved++
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.BatchSaveResponse{SavedCount: saved}, nil
}
