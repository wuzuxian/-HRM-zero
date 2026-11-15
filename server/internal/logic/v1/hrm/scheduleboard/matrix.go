package scheduleboard

// Matrix 排班矩阵读取（demo-inventory 4.3）：按考勤组 + 月份，员工×当月每日班次矩阵。
// 行成员 = 考勤组适用范围展开；每行 cells 与 monthDays 一一对应，无排班行=空班次。

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
)

type Matrix struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMatrix(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Matrix {
	return &Matrix{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Matrix) Matrix(req *types.MatrixRequest) (resp *types.MatrixResponse, err error) {
	group, members, err := groupAndMembers(l.ctx, l.svcCtx, req.GroupUuid)
	if err != nil {
		return nil, err
	}

	first, next, err := parseMonth(req.Month)
	if err != nil {
		return nil, err
	}
	days := monthDays(first, next)

	// 当月该考勤组成员的全部排班行，按 employee+date 索引
	entries, err := l.svcCtx.Model.HrmScheduleEntry.FindByCondition(l.ctx, nil,
		condition.NewChain().
			GreaterEqualThan(hrm_schedule_entry.ScheduleDate, days[0]).
			LessEqualThan(hrm_schedule_entry.ScheduleDate, days[len(days)-1]).
			In(hrm_schedule_entry.EmployeeUuid, members).
			Build()...)
	if err != nil {
		return nil, err
	}
	type cellKey struct{ emp, date string }
	entryMap := make(map[cellKey]*hrm_schedule_entry.HrmScheduleEntry, len(entries))
	for _, e := range entries {
		entryMap[cellKey{e.EmployeeUuid, e.ScheduleDate}] = e
	}

	rows := make([]types.HrmScheduleMatrixRow, 0, len(members))
	for _, empUuid := range members {
		employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, empUuid)
		if err != nil {
			if err == hrm_employee.ErrNotFound {
				continue // 指定人员已被删除，跳过（与 Members 端点同口径）
			}
			return nil, err
		}
		orgName := ""
		if employee.OrgUuid != "" {
			org, err := l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, employee.OrgUuid)
			if err != nil {
				if err == hrm_org.ErrNotFound {
					org = nil
				} else {
					return nil, err
				}
			}
			if org != nil {
				orgName = org.OrgName
			}
		}

		cells := make([]types.HrmScheduleCell, 0, len(days))
		publishedCount := int64(0)
		for _, day := range days {
			e := entryMap[cellKey{empUuid, day}]
			cell := types.HrmScheduleCell{
				EmployeeUuid: empUuid,
				Date:         day,
				ShiftUuid:    "",
				Published:    "0",
			}
			if e != nil {
				cell.ShiftUuid = e.ShiftUuid
				cell.Published = e.Published
				if e.Published == publishedYes {
					publishedCount++
				}
			}
			cells = append(cells, cell)
		}
		rows = append(rows, types.HrmScheduleMatrixRow{
			EmployeeUuid:   empUuid,
			EmployeeNo:     employee.EmployeeNo,
			EmployeeName:   employee.Name,
			OrgName:        orgName,
			MonthDays:      days,
			Cells:          cells,
			PublishedCount: publishedCount,
		})
	}

	// 姓名/工号/部门筛选（内存过滤——矩阵是全量展开视图，非分页列表）
	if req.Name != "" || req.EmployeeNo != "" || req.OrgName != "" {
		filtered := make([]types.HrmScheduleMatrixRow, 0, len(rows))
		for _, row := range rows {
			if req.Name != "" && !containsStr(row.EmployeeName, req.Name) {
				continue
			}
			if req.EmployeeNo != "" && !containsStr(row.EmployeeNo, req.EmployeeNo) {
				continue
			}
			if req.OrgName != "" && !containsStr(row.OrgName, req.OrgName) {
				continue
			}
			filtered = append(filtered, row)
		}
		rows = filtered
	}

	return &types.MatrixResponse{
		GroupUuid: group.Uuid,
		GroupName: group.GroupName,
		GroupType: group.GroupType,
		Month:     req.Month,
		MonthDays: days,
		Rows:      rows,
	}, nil
}
