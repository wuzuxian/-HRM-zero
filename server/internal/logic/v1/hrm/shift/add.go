package shift

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_shift"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/shift"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Add 新增班次（GLOSSARY「班次」）: 跨天标记与标准工时由上下班/休息时间推导.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.ShiftName == "" {
		return nil, status.ErrorMessage(errcodes.ShiftTimeInvalidCode, "班次名称不能为空")
	}
	if req.Status == "" {
		req.Status = shiftStatusEnabled
	}
	if err = validateShiftTimes(req.StartTime, req.EndTime); err != nil {
		return nil, status.ErrorMessage(errcodes.ShiftTimeInvalidCode, err.Error())
	}
	if err = validateRestTimes(req.RestStart, req.RestEnd); err != nil {
		return nil, status.ErrorMessage(errcodes.ShiftRestTimeInvalidCode, err.Error())
	}
	if !validShiftStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.ShiftStatusInvalidCode, "班次状态必须是 启用 或 停用")
	}

	_, err = l.svcCtx.Model.HrmShift.Insert(l.ctx, nil, &hrm_shift.HrmShift{
		Uuid:            uuid.New().String(),
		ShiftName:       req.ShiftName,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		CrossDay:        computeCrossDay(req.StartTime, req.EndTime),
		RestStart:       req.RestStart,
		RestEnd:         req.RestEnd,
		StandardHours:   computeStandardHours(req.StartTime, req.EndTime, req.RestStart, req.RestEnd),
		FlexibleMinutes: req.FlexibleMinutes,
		Status:          req.Status,
		Remark:          req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
