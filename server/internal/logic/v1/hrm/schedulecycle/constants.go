package schedulecycle

import (
	"strings"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_cycle"
)

// status values（HRM 约定 '1'=启用 '0'=停用）
const (
	statusEnabled  = "1"
	statusDisabled = "0"
)

// 查询启用排班周期的公共条件链（HRM-15 排班表消费方按需复用）.
func conditionNewChainEnabled() []condition.Condition {
	return condition.NewChain().
		Equal(hrm_schedule_cycle.Status, statusEnabled).
		OrderByAsc(hrm_schedule_cycle.Id).
		Build()
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

func joinCsv(v []string) string {
	return strings.Join(v, ",")
}
