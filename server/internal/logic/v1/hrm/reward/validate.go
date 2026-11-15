package reward

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// 奖惩类型（demo-inventory §7 / spec 用户故事 53）：奖励 / 处分.
const (
	RewardTypeReward     = "reward"     // 奖励 → 薪资「奖惩奖励」（应发项数据源）
	RewardTypePunishment = "punishment" // 处分 → 薪资「奖惩扣款」（应扣项数据源）
)

var rewardTypes = map[string]struct{}{
	RewardTypeReward:     {},
	RewardTypePunishment: {},
}

// 奖惩级别（demo-inventory §7）：公司级 / 部门级 / 通报 / 其他.
const (
	RewardLevelCompany = "company" // 公司级
	RewardLevelDept    = "dept"    // 部门级
	RewardLevelNotice  = "notice"  // 通报
	RewardLevelOther   = "other"   // 其他
)

var rewardLevels = map[string]struct{}{
	RewardLevelCompany: {},
	RewardLevelDept:    {},
	RewardLevelNotice:  {},
	RewardLevelOther:   {},
}

// editableStatuses are the 单据状态 a 奖惩单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked, mirroring the demo site).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// validateReward checks the user-supplied 奖惩单 fields and maps failures to
// the 1041x errcode range.
func validateReward(employeeNo, employeeName, rewardType, rewardLevel, occurDate, salaryMonth string, amount float64) error {
	if employeeNo == "" || employeeName == "" {
		return status.ErrorMessage(errcodes.RewardEmployeeRequiredCode, "员工工号与姓名为必填")
	}
	if _, ok := rewardTypes[rewardType]; !ok {
		return status.ErrorMessage(errcodes.RewardInvalidCode,
			"奖惩类型必须是 奖励/处分")
	}
	if _, ok := rewardLevels[rewardLevel]; !ok {
		return status.ErrorMessage(errcodes.RewardInvalidCode,
			"奖惩级别必须是 公司级/部门级/通报/其他")
	}
	if amount < 0 {
		return status.ErrorMessage(errcodes.RewardInvalidCode, "金额不能为负数")
	}
	if occurDate == "" {
		return status.ErrorMessage(errcodes.RewardInvalidCode, "发生日期必填")
	}
	if _, err := time.Parse(time.DateOnly, occurDate); err != nil {
		return status.ErrorMessage(errcodes.RewardInvalidCode,
			fmt.Sprintf("发生日期格式非法（须为 YYYY-MM-DD）: %s", occurDate))
	}
	// 入账薪资年月（薪资数据源契约）：填写时须为 YYYY-MM，供 HRM-34/HRM-37 按员工+年月取数。
	if salaryMonth != "" {
		if _, err := time.Parse("2006-01", salaryMonth); err != nil {
			return status.ErrorMessage(errcodes.RewardInvalidCode,
				fmt.Sprintf("入账薪资年月格式非法（须为 YYYY-MM）: %s", salaryMonth))
		}
	}
	return nil
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(status string) bool {
	_, ok := editableStatuses[status]
	return ok
}
