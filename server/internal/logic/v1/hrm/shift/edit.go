package shift

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_shift"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/shift"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Edit 编辑班次: 跨天标记与标准工时按新的上下班/休息时间重新推导.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.ShiftName == "" {
		return nil, status.ErrorMessage(errcodes.ShiftTimeInvalidCode, "班次名称不能为空")
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

	if _, err = l.svcCtx.Model.HrmShift.FindOneByUuid(l.ctx, nil, req.Uuid); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.ShiftNotFoundCode, "班次不存在")
		}
		return nil, err
	}

	err = l.svcCtx.Model.HrmShift.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_shift.ShiftName):       req.ShiftName,
		string(hrm_shift.StartTime):       req.StartTime,
		string(hrm_shift.EndTime):         req.EndTime,
		string(hrm_shift.CrossDay):        computeCrossDay(req.StartTime, req.EndTime),
		string(hrm_shift.RestStart):       req.RestStart,
		string(hrm_shift.RestEnd):         req.RestEnd,
		string(hrm_shift.StandardHours):   computeStandardHours(req.StartTime, req.EndTime, req.RestStart, req.RestEnd),
		string(hrm_shift.FlexibleMinutes): req.FlexibleMinutes,
		string(hrm_shift.Status):          req.Status,
		string(hrm_shift.Remark):          req.Remark,
	}, condition.Condition{
		Field:    hrm_shift.Uuid,
		Operator: condition.Equal,
		Value:    req.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
