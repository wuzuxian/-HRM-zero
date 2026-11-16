package scheduleboard

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type SaveCell struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSaveCell(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SaveCell {
	return &SaveCell{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *SaveCell) SaveCell(req *types.CellRequest) (resp *types.CellResponse, err error) {
	// 单格保存（矩阵单元格点击保存）：一人一天一行，同键覆盖；shiftUuid 空=休/清空。
	if req.Date == "" {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, "日期必填")
	}
	if _, err = validateDate(req.Date); err != nil {
		return nil, err
	}
	_, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}
	if !isMember(members, req.EmployeeUuid) {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardEmployeeNotFoundCode, "员工不在该考勤组适用范围内")
	}
	if err = checkShifts(l.ctx, l.svcCtx, []string{req.ShiftUuid}); err != nil {
		return nil, err
	}
	if err = upsertEntry(l.ctx, l.svcCtx, nil, req.EmployeeUuid, req.Date, req.ShiftUuid, req.GroupUuid, false); err != nil {
		return nil, err
	}
	return &types.CellResponse{}, nil
}
