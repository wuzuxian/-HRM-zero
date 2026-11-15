package overtimerule

import (
	"context"
	"strings"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimerule"
)

// ruleFields Add/Edit 共用的校验输入（结构一致, 仅 Status 语义略有差别）.
type ruleFields struct {
	RuleName          string
	Workday           types.DayOvertimePolicy
	Restday           types.DayOvertimePolicy
	Holiday           types.DayOvertimePolicy
	MinUnitMinutes    int64
	MonthlyLimitHours float64
	GroupUuids        []string
	OrgUuid           string
	Status            string
}

func ruleFieldsFrom(ruleName string, workday, restday, holiday types.DayOvertimePolicy,
	minUnit int64, monthlyLimit float64, groupUuids []string, orgUuid, ruleStatus string) *ruleFields {
	return &ruleFields{
		RuleName:          ruleName,
		Workday:           workday,
		Restday:           restday,
		Holiday:           holiday,
		MinUnitMinutes:    minUnit,
		MonthlyLimitHours: monthlyLimit,
		GroupUuids:        groupUuids,
		OrgUuid:           orgUuid,
		Status:            ruleStatus,
	}
}

// validateRuleBasic 加班规则字段校验（Add/Edit 共用）.
func validateRuleBasic(svcCtx *svc.ServiceContext, f *ruleFields) error {
	if f.RuleName == "" {
		return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, "规则名称不能为空")
	}
	policies := map[string]types.DayOvertimePolicy{
		"工作日":    f.Workday,
		"休息日":    f.Restday,
		"法定节假日": f.Holiday,
	}
	for name, p := range policies {
		if p.ApprovalRequired != policyRequired && p.ApprovalRequired != policyNotRequired {
			return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, name+"的需审批/免审批取值非法")
		}
		if p.PunchRequired != policyRequired && p.PunchRequired != policyNotRequired {
			return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, name+"的需打卡/免打卡取值非法")
		}
		if p.CompMultiplier <= 0 {
			return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, name+"的调休倍数必须大于0")
		}
	}
	if !validMinUnits[minUnitOrDefault(f.MinUnitMinutes)] {
		return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, "最小计算单位必须是 30 或 60 分钟")
	}
	if f.MonthlyLimitHours < 0 {
		return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, "月上限（小时）不能为负数，0 表示不限制")
	}
	if f.OrgUuid != "" {
		_, err := svcCtx.Model.HrmOrg.FindOneByUuid(context.Background(), nil, f.OrgUuid)
		if err != nil {
			if err == hrm_org.ErrNotFound {
				return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, "所属组织不存在")
			}
			return err
		}
	}
	if err := validateGroupBindings(svcCtx, f.GroupUuids); err != nil {
		return err
	}
	if f.Status != statusEnabled && f.Status != statusDisabled {
		return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, "加班规则状态必须是 启用 或 停用")
	}
	return nil
}

// validateGroupBindings 绑定的考勤组必须真实存在.
func validateGroupBindings(svcCtx *svc.ServiceContext, groupUuids []string) error {
	if len(groupUuids) == 0 {
		return nil
	}
	for _, g := range groupUuids {
		_, err := svcCtx.Model.HrmAttendanceGroup.FindOneByUuid(context.Background(), nil, g)
		if err != nil {
			if err == hrm_attendance_group.ErrNotFound {
				return status.ErrorMessage(errcodes.OvertimeRuleInvalidCode, "绑定的考勤组不存在")
			}
			return err
		}
	}
	return nil
}

// minUnitOrDefault 0 → 默认 30 分钟.
func minUnitOrDefault(v int64) int64 {
	if v == 0 {
		return minUnitDefaultMinutes
	}
	return v
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
