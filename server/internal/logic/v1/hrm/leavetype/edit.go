package leavetype

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavetype"
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

// Edit 编辑假期类型: 类型编码保持唯一（不允许改成别人的编码）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.TypeName == "" {
		return nil, status.ErrorMessage(errcodes.LeaveTypeInvalidCode, "类型名称不能为空")
	}
	if err = validateTypeCode(req.TypeCode); err != nil {
		return nil, err
	}
	if !validLeaveTypeStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.LeaveTypeInvalidCode, "假期类型状态必须是 启用 或 停用")
	}

	row, err := l.svcCtx.Model.HrmLeaveType.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.LeaveTypeNotFoundCode, "假期类型不存在")
		}
		return nil, err
	}

	// 类型编码唯一（排除自身）
	duplicated, err := l.svcCtx.Model.HrmLeaveType.FindOneByTypeCode(l.ctx, nil, req.TypeCode)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}
	if duplicated != nil && duplicated.Uuid != row.Uuid {
		return nil, status.ErrorMessage(errcodes.LeaveTypeCodeDuplicateCode, "类型编码 "+req.TypeCode+" 已存在")
	}

	err = l.svcCtx.Model.HrmLeaveType.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_leave_type.TypeCode):           req.TypeCode,
		string(hrm_leave_type.TypeName):           req.TypeName,
		string(hrm_leave_type.TypeValue):          req.TypeValue,
		string(hrm_leave_type.BalanceEnabled):     req.BalanceEnabled,
		string(hrm_leave_type.AttachmentRequired): req.AttachmentRequired,
		string(hrm_leave_type.DefaultDays):        req.DefaultDays,
		string(hrm_leave_type.OrderNum):           req.OrderNum,
		string(hrm_leave_type.Status):             req.Status,
		string(hrm_leave_type.Remark):             req.Remark,
	}, condition.Condition{
		Field:    hrm_leave_type.Uuid,
		Operator: condition.Equal,
		Value:    req.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
