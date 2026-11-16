package leavebalance

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
)

type InitPreview struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewInitPreview(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *InitPreview {
	return &InitPreview{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// InitPreview 初始化预览：展示按默认天数将初始化哪些 (员工 × 假期类型) 行、
// 哪些已存在将被跳过（skipped=1），供「初始化本年度额度」弹窗确认。
func (l *InitPreview) InitPreview(req *types.InitPreviewRequest) (resp *types.InitPreviewResponse, err error) {
	if req.Year <= 0 {
		return nil, status.ErrorMessage(errcodes.LeaveBalanceInvalidCode, "年度必须大于 0")
	}

	employeeUuids, _, err := resolveInitEmployees(l.ctx, l.svcCtx, req.EmployeeUuids)
	if err != nil {
		return nil, err
	}
	leaveTypes, err := managedLeaveTypes(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmInitPreviewItem, 0, len(employeeUuids)*len(leaveTypes))
	for _, employeeUuid := range employeeUuids {
		employeeNo, employeeName, _, _, _ := employeeSnapshot(l.ctx, l.svcCtx, employeeUuid)
		for _, lt := range leaveTypes {
			existing, findErr := l.svcCtx.Model.HrmLeaveBalance.FindOneByCondition(l.ctx, nil,
				balanceCondition(req.Year, employeeUuid, lt.TypeCode).Build()...)
			skipped := int64(0)
			if findErr == nil && existing != nil {
				skipped = 1
			}
			records = append(records, types.HrmInitPreviewItem{
				EmployeeUuid:  employeeUuid,
				EmployeeNo:    employeeNo,
				EmployeeName:  employeeName,
				LeaveTypeCode: lt.TypeCode,
				LeaveTypeName: lt.TypeName,
				DefaultDays:   lt.DefaultDays,
				Skipped:       skipped,
			})
		}
	}

	return &types.InitPreviewResponse{Records: records}, nil
}
