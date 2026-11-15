package leavebalance

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_type"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/leavebalance"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// List 假期余额分页列表：余额行 (year, employee, leave_type) 维度展示，
// 附员工工号/姓名/部门/入职日期与假期类型名称快照。
// 筛选: 年度/姓名(模糊)/部门含下级/假期类型。可用 = granted - used - frozen。
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	chain := condition.NewChain().
		Page(req.Current, req.Size).
		Equal(hrm_leave_balance.Year, req.Year, condition.WithSkip(req.Year <= 0)).
		Equal(hrm_leave_balance.LeaveTypeCode, req.LeaveTypeCode, condition.WithSkip(req.LeaveTypeCode == ""))

	if req.Name != "" || req.OrgUuid != "" {
		// 姓名(模糊)/部门(含下级) 命中 → 员工 uuid 集合再 In 过滤余额行
		employeeUuids, err := matchEmployeeUuids(l.ctx, l.svcCtx, req.Name, req.OrgUuid)
		if err != nil {
			return nil, err
		}
		if len(employeeUuids) == 0 {
			return &types.ListResponse{
				PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: 0},
				Records:      []types.HrmLeaveBalance{},
			}, nil
		}
		chain = chain.In(hrm_leave_balance.EmployeeUuid, employeeUuids)
	}

	rows, total, err := l.svcCtx.Model.HrmLeaveBalance.PageByCondition(l.ctx, nil,
		chain.OrderByDesc(hrm_leave_balance.Year).
			OrderByAsc(hrm_leave_balance.EmployeeNo).
			OrderByAsc(hrm_leave_balance.LeaveTypeCode).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmLeaveBalance, 0, len(rows))
	for _, row := range rows {
		employeeNo, employeeName, orgUuid, orgName, hireDate := employeeSnapshot(l.ctx, l.svcCtx, row.EmployeeUuid)
		records = append(records, types.HrmLeaveBalance{
			Uuid:          row.Uuid,
			Year:          row.Year,
			EmployeeUuid:  row.EmployeeUuid,
			EmployeeNo:    employeeNo,
			EmployeeName:  employeeName,
			OrgUuid:       orgUuid,
			OrgName:       orgName,
			HireDate:      hireDate,
			LeaveTypeCode: row.LeaveTypeCode,
			LeaveTypeName: leaveTypeName(l.ctx, l.svcCtx, row.LeaveTypeCode),
			Granted:       row.Granted,
			Used:          row.Used,
			Frozen:        row.Frozen,
			Available:     round1(row.Granted - row.Used - row.Frozen),
			CreateTime:    row.CreateTime.Format(time.DateTime),
			UpdateTime:    row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.ListResponse{
		PageResponse: types.PageResponse{Current: req.Current, Size: req.Size, Total: total},
		Records:      records,
	}, nil
}

// matchEmployeeUuids 解析 姓名(模糊)/部门(含下级) 筛选 → 员工 uuid 集合。
func matchEmployeeUuids(ctx context.Context, svcCtx *svc.ServiceContext, name, orgUuid string) ([]string, error) {
	chain := condition.NewChain()
	if name != "" {
		chain = chain.Like(hrm_employee.Name, "%"+name+"%")
	}
	if orgUuid != "" {
		orgUuids, err := subOrgUuids(ctx, svcCtx, orgUuid)
		if err != nil {
			return nil, err
		}
		chain = chain.In(hrm_employee.OrgUuid, orgUuids)
	}
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Uuid)
	}
	return out, nil
}

// subOrgUuids 部门含下级（orgUuid 自身 + 全部后代，与 attendancegroup 同款展开）。
func subOrgUuids(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) ([]string, error) {
	all, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil,
		condition.NewChain().OrderByAsc(hrm_org.Id).Build()...)
	if err != nil {
		return nil, err
	}
	childrenOf := make(map[string][]string, len(all))
	for _, row := range all {
		childrenOf[row.ParentUuid] = append(childrenOf[row.ParentUuid], row.Uuid)
	}
	out := []string{orgUuid}
	stack := []string{orgUuid}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, child := range childrenOf[cur] {
			out = append(out, child)
			stack = append(stack, child)
		}
	}
	return out, nil
}

func employeeSnapshot(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid string) (employeeNo, employeeName, orgUuid, orgName, hireDate string) {
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil || row == nil {
		return "", "", "", "", ""
	}
	if row.OrgUuid != "" {
		org, orgErr := svcCtx.Model.HrmOrg.FindOneByCondition(ctx, nil,
			condition.NewChain().Equal(hrm_org.Uuid, row.OrgUuid).Build()...)
		if orgErr == nil && org != nil {
			orgName = org.OrgName
		}
	}
	return row.EmployeeNo, row.Name, row.OrgUuid, orgName, row.HireDate
}

func leaveTypeName(ctx context.Context, svcCtx *svc.ServiceContext, typeCode string) string {
	row, err := svcCtx.Model.HrmLeaveType.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_leave_type.TypeCode, typeCode).Build()...)
	if err != nil || row == nil {
		return typeCode
	}
	return row.TypeName
}
