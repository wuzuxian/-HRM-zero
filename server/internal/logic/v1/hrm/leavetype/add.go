package leavetype

import (
	"context"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavetype"
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

// HRM 假期类型（GLOSSARY「假期类型」）: 1=启用 0=停用; 类型编码须为小写字母+下划线
// （annual_leave/comp_leave/sick_leave/other_leave…）.
var typeCodeRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func validLeaveTypeStatus(v string) bool {
	return v == "1" || v == "0"
}

func validateTypeCode(code string) error {
	if !typeCodeRegexp.MatchString(code) {
		return status.ErrorMessage(errcodes.LeaveTypeInvalidCode, "类型编码必须为小写字母/数字/下划线且以字母开头")
	}
	return nil
}

// Add 新增假期类型.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.TypeName == "" {
		return nil, status.ErrorMessage(errcodes.LeaveTypeInvalidCode, "类型名称不能为空")
	}
	if req.Status == "" {
		req.Status = "1"
	}
	if err = validateTypeCode(req.TypeCode); err != nil {
		return nil, err
	}
	if !validLeaveTypeStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.LeaveTypeInvalidCode, "假期类型状态必须是 启用 或 停用")
	}

	// 类型编码唯一
	_, err = l.svcCtx.Model.HrmLeaveType.FindOneByTypeCode(l.ctx, nil, req.TypeCode)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	} else if err == nil {
		return nil, status.ErrorMessage(errcodes.LeaveTypeCodeDuplicateCode, "类型编码 "+req.TypeCode+" 已存在")
	}

	_, err = l.svcCtx.Model.HrmLeaveType.Insert(l.ctx, nil, &hrm_leave_type.HrmLeaveType{
		Uuid:               uuid.New().String(),
		TypeCode:           req.TypeCode,
		TypeName:           req.TypeName,
		TypeValue:          req.TypeValue,
		BalanceEnabled:     req.BalanceEnabled,
		AttachmentRequired: req.AttachmentRequired,
		DefaultDays:        req.DefaultDays,
		OrderNum:           req.OrderNum,
		Status:             req.Status,
		Remark:             req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
