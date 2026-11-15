package position

import (
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 岗位类别 — 与招聘需求单（HRM-24 requirement）同一枚举（demo-inventory §5）.
var postCategories = map[string]struct{}{
	"tech":      {}, // 技术类
	"product":   {}, // 产品类
	"design":    {}, // 设计类
	"operation": {}, // 运营类
	"market":    {}, // 市场类
	"function":  {}, // 职能类
	"other":     {}, // 其他
}

// validatePosition checks the user-supplied 招聘职位 fields and maps failures to
// the 1058x errcode range. 来源需求单（requirementUuid）必填 — 审批状态约束由
// ensureRequirementApproved 在事务内校验（HRM-24 seam 契约）.
func validatePosition(postName, postCategory, publishDate string, headcount int64, requirementUuid string) error {
	if postName == "" {
		return status.ErrorMessage(errcodes.PositionInvalidCode, "岗位名称必填")
	}
	if requirementUuid == "" {
		return status.ErrorMessage(errcodes.PositionRequirementRequiredCode, "来源需求单必填")
	}
	if _, ok := postCategories[postCategory]; !ok {
		return status.ErrorMessage(errcodes.PositionInvalidCode,
			"岗位类别必须是 技术类/产品类/设计类/运营类/市场类/职能类/其他")
	}
	if headcount < 1 {
		return status.ErrorMessage(errcodes.PositionInvalidCode, "招聘人数必须大于等于 1")
	}
	if publishDate != "" {
		if _, err := time.Parse(time.DateOnly, publishDate); err != nil {
			return status.ErrorMessage(errcodes.PositionInvalidCode,
				fmt.Sprintf("发布日期格式非法（须为 YYYY-MM-DD）: %s", publishDate))
		}
	}
	return nil
}
