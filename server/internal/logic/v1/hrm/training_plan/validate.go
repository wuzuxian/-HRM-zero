package training_plan

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// 计划类型（演示站 培训计划 计划类型筛选）
const (
	PlanTypeSpecial = "special"  // 专项
	PlanTypeNewHire = "new_hire" // 新员工培养
	PlanTypeSafety  = "safety"   // 安全培训
	PlanTypeSkill   = "skill"    // 技能提升
)

// PlanTypes lists all valid 计划类型 codes.
var PlanTypes = []string{PlanTypeSpecial, PlanTypeNewHire, PlanTypeSafety, PlanTypeSkill}

// dateLayout is the 起止日期 format (YYYY-MM-DD).
const dateLayout = time.DateOnly

// isValidPlanType reports whether s is a known 计划类型.
func isValidPlanType(s string) bool {
	for _, t := range PlanTypes {
		if s == t {
			return true
		}
	}
	return false
}

// isValidPlanYear reports whether s is a 4-digit year.
func isValidPlanYear(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isValidDate reports whether s parses as YYYY-MM-DD.
func isValidDate(s string) bool {
	_, err := time.ParseInLocation(dateLayout, s, time.Local)
	return err == nil
}

// editableStatuses are the 单据状态 a 培训计划单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked, mirroring the demo site).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(status string) bool {
	_, ok := editableStatuses[status]
	return ok
}

// validatePlan is the pure parameter validation for 培训计划单 Add/Edit:
// 名称必填、年度 4 位、计划类型枚举、起止日期格式与先后顺序.
func validatePlan(planName, planYear, planType, startDate, endDate string) error {
	if planName == "" {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode, "计划名称必填")
	}
	if !isValidPlanYear(planYear) {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode,
			fmt.Sprintf("年度非法（须为 4 位年份）: %s", planYear))
	}
	if !isValidPlanType(planType) {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode,
			"计划类型必须是 专项/新员工培养/安全培训/技能提升")
	}
	if startDate == "" || endDate == "" {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode, "开始/结束日期必填")
	}
	sd, err := time.ParseInLocation(dateLayout, startDate, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode,
			fmt.Sprintf("开始日期格式非法（须为 YYYY-MM-DD）: %s", startDate))
	}
	ed, err := time.ParseInLocation(dateLayout, endDate, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode,
			fmt.Sprintf("结束日期格式非法（须为 YYYY-MM-DD）: %s", endDate))
	}
	if ed.Before(sd) {
		return status.ErrorMessage(errcodes.TrainingPlanInvalidCode,
			"结束日期不能早于开始日期")
	}
	return nil
}
