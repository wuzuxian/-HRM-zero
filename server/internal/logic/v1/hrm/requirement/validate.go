package requirement

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
)

// 岗位类别（demo-inventory §5 / spec 用户故事 43）.
const (
	PostCategoryTech      = "tech"      // 技术类
	PostCategoryProduct   = "product"   // 产品类
	PostCategoryDesign    = "design"    // 设计类
	PostCategoryOperation = "operation" // 运营类
	PostCategoryMarket    = "market"    // 市场类
	PostCategoryFunction  = "function"  // 职能类
	PostCategoryOther     = "other"     // 其他
)

var postCategories = map[string]struct{}{
	PostCategoryTech:      {},
	PostCategoryProduct:   {},
	PostCategoryDesign:    {},
	PostCategoryOperation: {},
	PostCategoryMarket:    {},
	PostCategoryFunction:  {},
	PostCategoryOther:     {},
}

// 紧急度（demo-inventory §5）：普通 / 紧急.
const (
	UrgencyNormal = "normal" // 普通
	UrgencyUrgent = "urgent" // 紧急
)

var urgencies = map[string]struct{}{
	UrgencyNormal: {},
	UrgencyUrgent: {},
}

// editableStatuses are the 单据状态 a 招聘需求单 can still be modified/deleted in:
// 未提交/审批不通过/已撤销 (审批中/审批通过 are locked, mirroring the demo site).
var editableStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {}, bill.StatusRejected: {}, bill.StatusRevoked: {},
}

// validateRequirement checks the user-supplied 招聘需求单 fields and maps failures
// to the 1043x errcode range.
func validateRequirement(postName, postCategory, urgency, expectDate string, headcount int64, orgUuid string) error {
	if postName == "" {
		return status.ErrorMessage(errcodes.RequirementInvalidCode, "岗位名称必填")
	}
	if _, ok := postCategories[postCategory]; !ok {
		return status.ErrorMessage(errcodes.RequirementInvalidCode,
			"岗位类别必须是 技术类/产品类/设计类/运营类/市场类/职能类/其他")
	}
	if headcount < 1 {
		return status.ErrorMessage(errcodes.RequirementInvalidCode, "需求人数必须大于等于 1")
	}
	if _, ok := urgencies[urgency]; !ok {
		return status.ErrorMessage(errcodes.RequirementInvalidCode, "紧急度必须是 普通/紧急")
	}
	if orgUuid == "" {
		return status.ErrorMessage(errcodes.RequirementOrgRequiredCode, "所属部门必填")
	}
	if expectDate != "" {
		if _, err := time.Parse(time.DateOnly, expectDate); err != nil {
			return status.ErrorMessage(errcodes.RequirementInvalidCode,
				fmt.Sprintf("期望到岗日期格式非法（须为 YYYY-MM-DD）: %s", expectDate))
		}
	}
	return nil
}

// canEdit reports whether the 单据状态 allows modification/deletion.
func canEdit(status string) bool {
	_, ok := editableStatuses[status]
	return ok
}
