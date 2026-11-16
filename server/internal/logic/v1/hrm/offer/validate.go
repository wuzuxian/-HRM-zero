package offer

import (
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// BillTypeCode is the 录用审批 类型码 (HR208, demo-inventory §5 录用审批).
const BillTypeCode = "208"

// RankLevel 职级枚举（demo-inventory §5：职级 初级/中级/高级/专家）.
const (
	RankLevelJunior       = "junior"       // 初级
	RankLevelIntermediate = "intermediate" // 中级
	RankLevelSenior       = "senior"       // 高级
	RankLevelExpert       = "expert"       // 专家
)

var rankLevels = map[string]struct{}{
	RankLevelJunior:       {},
	RankLevelIntermediate: {},
	RankLevelSenior:       {},
	RankLevelExpert:       {},
}

// timeNow 可测试钩子（单据编号日期部分用当天）。
var timeNow = time.Now

// validateOffer Add/Edit 共用校验：候选人/职位/计划入职日期必填 + 职级枚举合法（空默认 junior）.
func validateOffer(candidateUuid, positionUuid, planEntryDate, rankLevel string) error {
	if candidateUuid == "" {
		return status.ErrorMessage(errcodes.OfferInvalidCode, "候选人必填")
	}
	if positionUuid == "" {
		return status.ErrorMessage(errcodes.OfferInvalidCode, "招聘职位必填")
	}
	if planEntryDate == "" {
		return status.ErrorMessage(errcodes.OfferInvalidCode, "计划入职日期必填")
	}
	if rankLevel != "" {
		if _, ok := rankLevels[rankLevel]; !ok {
			return status.ErrorMessage(errcodes.OfferInvalidCode, "职级必须是 初级/中级/高级/专家")
		}
	}
	return nil
}

// canEdit 与共享单据状态机一致的「可编辑」状态集（未提交/审批不通过/已撤销）.
func canEdit(statusCode string) bool {
	switch statusCode {
	case "unsubmitted", "rejected", "revoked":
		return true
	}
	return false
}
