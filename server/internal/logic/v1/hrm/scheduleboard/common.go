package scheduleboard

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// status values（HRM 约定 '1'=启用 '0'=停用；published '1'=已发布 '0'=未发布）
const (
	statusEnabled  = "1"
	statusDisabled = "0"

	publishedYes = "1"
	publishedNo  = "0"

	dateLayout     = "2006-01-02"
	monthLayout    = "2006-01"
	weekAlignWarn  = "周起始日必须是周一"
	monthLenLayout = "月份必须是 yyyy-MM"
)

// groupAndMembers 校验考勤组存在且启用，并展开适用成员（部门含下级 + 指定人员，去重）。
// 排班表按考勤组维护 —— 矩阵行成员即此列表（与 attendancegroup Members 端点同口径）。
func groupAndMembers(ctx context.Context, svcCtx *svc.ServiceContext, groupUuid string) (*hrm_attendance_group.HrmAttendanceGroup, []string, error) {
	if groupUuid == "" {
		return nil, nil, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, "考勤组必填（排班表按考勤组维护）")
	}
	group, err := svcCtx.Model.HrmAttendanceGroup.FindOneByUuid(ctx, nil, groupUuid)
	if err != nil {
		if err == hrm_attendance_group.ErrNotFound {
			return nil, nil, status.ErrorMessage(errcodes.ScheduleBoardGroupNotFoundCode, "考勤组不存在")
		}
		return nil, nil, err
	}
	if group.Status != statusEnabled {
		return nil, nil, status.ErrorMessage(errcodes.ScheduleBoardGroupNotFoundCode, "考勤组已停用: "+group.GroupName)
	}

	seen := make(map[string]bool)
	members, err := expandMembers(ctx, svcCtx, group, seen)
	if err != nil {
		return nil, nil, err
	}
	return group, members, nil
}

// splitCsv splits a comma-separated uuid column value, dropping empties.
func splitCsv(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseMonth "yyyy-MM" → (月初, 下月初)；非法月份报 ScheduleBoardInvalidCode.
func parseMonth(month string) (time.Time, time.Time, error) {
	first, err := time.ParseInLocation(monthLayout, month, time.Local)
	if err != nil || len(month) != 7 {
		return time.Time{}, time.Time{}, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, monthLenLayout+": "+month)
	}
	next := first.AddDate(0, 1, 0)
	return first, next, nil
}

// monthDays 生成 [月初, 下月初) 的逐日 "yyyy-MM-dd" 列表.
func monthDays(first, next time.Time) []string {
	var days []string
	for cur := first; cur.Before(next); cur = cur.AddDate(0, 0, 1) {
		days = append(days, cur.Format(dateLayout))
	}
	return days
}

// validateDate yyyy-MM-dd 校验.
func validateDate(date string) (time.Time, error) {
	t, err := time.ParseInLocation(dateLayout, date, time.Local)
	if err != nil || len(date) != 10 {
		return time.Time{}, status.ErrorMessage(errcodes.ScheduleBoardInvalidCode, "日期必须为 yyyy-MM-dd: "+date)
	}
	return t, nil
}

// checkShifts 批量校验班次 uuid 存在且启用（shiftUuid "" 允许 = 休/清空）。
func checkShifts(ctx context.Context, svcCtx *svc.ServiceContext, shiftUuids []string) error {
	seen := make(map[string]bool)
	for _, su := range shiftUuids {
		if su == "" || seen[su] {
			continue
		}
		seen[su] = true
		shift, err := svcCtx.Model.HrmShift.FindOneByUuid(ctx, nil, su)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return status.ErrorMessage(errcodes.ScheduleBoardShiftInvalidCode, "班次不存在")
			}
			return err
		}
		if shift.Status != statusEnabled {
			return status.ErrorMessage(errcodes.ScheduleBoardShiftInvalidCode, "班次已停用: "+shift.ShiftName)
		}
	}
	return nil
}

// upsertEntry 写一个排班单元格（一人一天一行，同键覆盖；shiftUuid "" 表示休/清空）。
// 已发布行默认拒绝覆盖（撤销发布后才能改）。
func upsertEntry(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, employeeUuid, date, shiftUuid, groupUuid string, allowOverwritePublished bool) error {
	existing, err := svcCtx.Model.HrmScheduleEntry.FindOneByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_schedule_entry.ScheduleDate, date).
			Equal(hrm_schedule_entry.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil && err != hrm_schedule_entry.ErrNotFound {
		return err
	}
	if existing != nil && existing.Published == publishedYes && !allowOverwritePublished {
		return status.ErrorMessage(errcodes.ScheduleBoardPublishedLockedCode,
			"该日期排班已发布，请先撤销发布: "+existing.ScheduleDate)
	}
	if existing != nil {
		return svcCtx.Model.HrmScheduleEntry.UpdateFieldsByCondition(ctx, session,
			map[string]any{"shift_uuid": shiftUuid, "group_uuid": groupUuid},
			condition.NewChain().
				Equal(hrm_schedule_entry.ScheduleDate, date).
				Equal(hrm_schedule_entry.EmployeeUuid, employeeUuid).
				Build()...)
	}
	_, err = svcCtx.Model.HrmScheduleEntry.Insert(ctx, session, &hrm_schedule_entry.HrmScheduleEntry{
		Uuid:         newUuid(),
		ScheduleDate: date,
		EmployeeUuid: employeeUuid,
		ShiftUuid:    shiftUuid,
		GroupUuid:    groupUuid,
		Published:    publishedNo,
	})
	return err
}

// containsStr 大小写不敏感子串匹配.
func containsStr(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// joinCsv joins uuid list into a comma-separated column value.
func joinCsv(v []string) string {
	return strings.Join(v, ",")
}

// isMember 判断员工 uuid 是否在考勤组成员列表内.
func isMember(members []string, employeeUuid string) bool {
	for _, u := range members {
		if u == employeeUuid {
			return true
		}
	}
	return false
}

// xlsxBytes base64 → bytes（导入文件内容解码；非法输入报 ScheduleBoardImportInvalidCode）.
func xlsxBytes(content string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.ScheduleBoardImportInvalidCode, "导入文件内容不是合法的 base64")
	}
	return raw, nil
}
