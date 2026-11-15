package attendancegroup

import (
	"context"
	"strings"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_config"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 本文件是三级回退（考勤组 > 本部门规则 > 最近上级组织规则 > 无规则兜底）的唯一实现。
// 日结/打卡等后续工单（HRM-15/16…）应复用这里的 ResolveEffectiveRule /
// resolveOrgUuidChain / expandGroupMembers，不要重新推导。

// RuleSource 生效规则命中来源.
const (
	RuleSourceGroup = ruleSourceGroup
	RuleSourceOrg   = ruleSourceOrg
)

// splitCsv splits a comma-separated uuid/name column value, dropping empties.
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

func containsStr(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// resolveOrgUuidChain walks the hrm_org parent chain bottom-up and returns
// [orgUuid, parent, grandparent, ...]. Cycles are guarded by a visited set.
func resolveOrgUuidChain(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) ([]string, error) {
	chain := make([]string, 0, 8)
	visited := make(map[string]bool)
	current := orgUuid
	for current != "" && !visited[current] {
		visited[current] = true
		chain = append(chain, current)
		row, err := svcCtx.Model.HrmOrg.FindOneByUuid(ctx, nil, current)
		if err != nil {
			if err == hrm_org.ErrNotFound {
				break
			}
			return nil, err
		}
		current = row.ParentUuid
	}
	return chain, nil
}

// findAllSubOrgUuids returns orgUuid plus every descendant uuid (部门含下级).
func findAllSubOrgUuids(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) ([]string, error) {
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

// expandGroupMembers 解析考勤组适用范围（部门含下级 + 指定人员，去重），
// 返回命中员工的 uuid 列表。
func expandGroupMembers(ctx context.Context, svcCtx *svc.ServiceContext, group *hrm_attendance_group.HrmAttendanceGroup) ([]string, error) {
	seen := make(map[string]bool)
	var members []string

	for _, orgUuid := range splitCsv(group.OrgUuids) {
		orgUuids, err := findAllSubOrgUuids(ctx, svcCtx, orgUuid)
		if err != nil {
			return nil, err
		}
		for _, u := range orgUuids {
			rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
				condition.NewChain().Equal(hrm_employee.OrgUuid, u).Build()...)
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				if !seen[row.Uuid] {
					seen[row.Uuid] = true
					members = append(members, row.Uuid)
				}
			}
		}
	}

	for _, memberUuid := range splitCsv(group.MemberUuids) {
		if !seen[memberUuid] {
			seen[memberUuid] = true
			members = append(members, memberUuid)
		}
	}
	return members, nil
}

// matchEnabledGroup returns the enabled group covering the employee, or nil.
// 适用范围: employee.OrgUuid ∈ group.OrgUuids(含下级) 或 employee.Uuid ∈ group.MemberUuids.
func matchEnabledGroup(ctx context.Context, svcCtx *svc.ServiceContext, employee *hrm_employee.HrmEmployee) (*hrm_attendance_group.HrmAttendanceGroup, error) {
	groups, err := svcCtx.Model.HrmAttendanceGroup.FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_attendance_group.Status, statusEnabled).
			OrderByAsc(hrm_attendance_group.Id).
			Build()...)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}

	// 员工所属组织链（含下级在部门匹配时逐组展开）
	for _, group := range groups {
		// 指定人员
		if containsStr(splitCsv(group.MemberUuids), employee.Uuid) {
			return group, nil
		}
		// 部门含下级: group 的任一适用部门是员工部门的祖先（含本部门）
		ancestorChain, err := resolveOrgUuidChain(ctx, svcCtx, employee.OrgUuid)
		if err != nil {
			return nil, err
		}
		for _, scopeOrg := range splitCsv(group.OrgUuids) {
			if containsStr(ancestorChain, scopeOrg) {
				return group, nil
			}
		}
	}
	return nil, nil
}

// findOrgConfig returns the default rule of one org, or nil (not found is normal).
func findOrgConfig(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) (*hrm_attendance_config.HrmAttendanceConfig, error) {
	row, err := svcCtx.Model.HrmAttendanceConfig.FindOneByOrgUuid(ctx, nil, orgUuid)
	if err != nil {
		if err == hrm_attendance_config.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

// ResolvedRule is the outcome of the 三级回退 resolution.
type ResolvedRule struct {
	Found      bool
	Source     string // RuleSourceGroup | RuleSourceOrg
	SourceUuid string
	SourceName string
	GroupType  string
	StartTime  string
	EndTime    string
	WorkDays   []int64
	FlexibleMinutes int64
}

// ResolveEffectiveRule 三级回退唯一入口:
//  1. 员工命中考勤组（部门含下级 + 指定人员）→ 考勤组规则
//  2. 本部门默认规则（hrm_attendance_config.org_uuid = employee.OrgUuid）
//  3. 最近上级组织默认规则（沿 hrm_org 父链向上第一条）
//  4. 全空 → Found=false（无规则兜底）
//
// employee 为 nil 时按工号加载；员工不存在返回 ErrNotFound 由调用方决定错误码。
func ResolveEffectiveRule(ctx context.Context, svcCtx *svc.ServiceContext, employee *hrm_employee.HrmEmployee) (*ResolvedRule, error) {
	// 第 1 级: 考勤组
	group, err := matchEnabledGroup(ctx, svcCtx, employee)
	if err != nil {
		return nil, err
	}
	if group != nil {
		return &ResolvedRule{
			Found:           true,
			Source:          RuleSourceGroup,
			SourceUuid:      group.Uuid,
			SourceName:      group.GroupName,
			GroupType:       group.GroupType,
			StartTime:       group.StartTime,
			EndTime:         group.EndTime,
			WorkDays:        csvToInt64s(group.WorkDays),
			FlexibleMinutes: group.FlexibleMinutes,
		}, nil
	}

	// 第 2/3 级: 本部门 → 最近上级组织
	chain, err := resolveOrgUuidChain(ctx, svcCtx, employee.OrgUuid)
	if err != nil {
		return nil, err
	}
	for _, orgUuid := range chain {
		config, err := findOrgConfig(ctx, svcCtx, orgUuid)
		if err != nil {
			return nil, err
		}
		if config != nil {
			return &ResolvedRule{
				Found:      true,
				Source:     RuleSourceOrg,
				SourceUuid: config.OrgUuid,
				SourceName: config.Remark,
				StartTime:  config.StartTime,
				EndTime:    config.EndTime,
				WorkDays:   csvToInt64s(config.WorkDays),
			}, nil
		}
	}

	return &ResolvedRule{}, nil
}

func csvToInt64s(v string) []int64 {
	parts := splitCsv(v)
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		var n int64
		var ok bool
		n, ok = parseInt64(p)
		if ok {
			out = append(out, n)
		}
	}
	return out
}

func parseInt64(s string) (int64, bool) {
	var n int64
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int64(c-'0')
	}
	return n, true
}
