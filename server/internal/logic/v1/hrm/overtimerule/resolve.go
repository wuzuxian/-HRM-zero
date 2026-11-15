package overtimerule

// 本文件是加班规则解析（消费契约）的唯一实现。
// HRM-21 加班申请单（后续工单）按 加班日期当天类型 调 ResolveOvertimeRule
// 判定 需审批/免审批、需打卡/免打卡、调休折算倍数 —— 不要重新推导。

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_rule"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimerule"
)

type Resolve struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewResolve(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Resolve {
	return &Resolve{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// ResolvedOvertimeRule ResolveOvertimeRule 的解析结果.
type ResolvedOvertimeRule struct {
	Found             bool
	Source            string // RuleSourceGroup | RuleSourceOrg | RuleSourceGlobal
	SourceUuid        string
	SourceName        string
	DayType           string
	ApprovalRequired  bool
	PunchRequired     bool
	CompMultiplier    float64
	MinUnitMinutes    int64
	MonthlyLimitHours float64
}

// ResolveOvertimeRule 加班规则解析唯一入口（HRM-21 加班申请单消费契约）:
//
//	ResolveOvertimeRule(ctx, svcCtx, groupUuid, orgUuid, dayType) (*ResolvedOvertimeRule, error)
//
// 参数:
//   - groupUuid: 员工所在考勤组 uuid（HRM-12 ResolveEffectiveRule 命中 group 时带回的 SourceUuid）；
//     空表示员工未绑定考勤组
//   - orgUuid: 员工所属组织 uuid（用于组织默认规则回退）；空跳过组织级
//   - dayType: 加班日期当天类型 DayTypeWorkday("1")/DayTypeRestday("2")/DayTypeHoliday("3")
//
// 解析顺序（考勤组绑定优先，未绑定考勤组回退组织默认/全局默认）:
//  1. 启用状态 且 GroupUuids 绑定了 groupUuid 的规则（按 id 升序取第一条）
//  2. 启用状态 且 OrgUuid = orgUuid 的组织默认规则（groupUuid == "" 时不要求，命中即用）
//  3. 启用状态 且 GroupUuids 与 OrgUuid 均为空 的全局兜底规则
//  4. 全空 → Found=false（调用方自行兜底，例如按 需审批+倍数1 处理）
//
// dayType 非法返回 OvertimeRuleDayTypeInvalidCode。
func ResolveOvertimeRule(ctx context.Context, svcCtx *svc.ServiceContext, groupUuid, orgUuid, dayType string) (*ResolvedOvertimeRule, error) {
	switch dayType {
	case DayTypeWorkday, DayTypeRestday, DayTypeHoliday:
	case "":
		return nil, status.ErrorMessage(errcodes.OvertimeRuleDayTypeInvalidCode, "加班当天类型不能为空")
	default:
		return nil, status.ErrorMessage(errcodes.OvertimeRuleDayTypeInvalidCode, "加班当天类型必须是 工作日/休息日/法定节假日")
	}

	rows, err := svcCtx.Model.HrmOvertimeRule.FindByCondition(ctx, nil,
		conditionNewChainEnabled()...)
	if err != nil {
		return nil, err
	}

	// 第 1 级: 考勤组绑定规则
	if groupUuid != "" {
		for _, row := range rows {
			if containsStr(splitCsv(row.GroupUuids), groupUuid) {
				return resolvedFrom(row, groupUuid, dayType), nil
			}
		}
	}
	// 第 2 级: 组织默认规则
	if orgUuid != "" {
		for _, row := range rows {
			if row.OrgUuid == orgUuid && len(splitCsv(row.GroupUuids)) == 0 {
				return resolvedFrom(row, orgUuid, dayType), nil
			}
		}
	}
	// 第 3 级: 全局兜底规则
	for _, row := range rows {
		if len(splitCsv(row.GroupUuids)) == 0 && row.OrgUuid == "" {
			return resolvedFrom(row, row.Uuid, dayType), nil
		}
	}

	return &ResolvedOvertimeRule{Found: false, DayType: dayType}, nil
}

// resolvedFrom row → 结果（按 dayType 摘出当天策略）.
func resolvedFrom(row *hrm_overtime_rule.HrmOvertimeRule, sourceUuid, dayType string) *ResolvedOvertimeRule {
	out := &ResolvedOvertimeRule{
		Found:             true,
		SourceUuid:        sourceUuid,
		SourceName:        row.RuleName,
		DayType:           dayType,
		MinUnitMinutes:    row.MinUnitMinutes,
		MonthlyLimitHours: row.MonthlyLimitHours,
	}
	switch dayType {
	case DayTypeWorkday:
		out.Source = sourceKindOf(row, sourceUuid)
		out.ApprovalRequired = row.WorkdayApprovalRequired == policyRequired
		out.PunchRequired = row.WorkdayPunchRequired == policyRequired
		out.CompMultiplier = row.WorkdayCompMultiplier
	case DayTypeRestday:
		out.Source = sourceKindOf(row, sourceUuid)
		out.ApprovalRequired = row.RestdayApprovalRequired == policyRequired
		out.PunchRequired = row.RestdayPunchRequired == policyRequired
		out.CompMultiplier = row.RestdayCompMultiplier
	case DayTypeHoliday:
		out.Source = sourceKindOf(row, sourceUuid)
		out.ApprovalRequired = row.HolidayApprovalRequired == policyRequired
		out.PunchRequired = row.HolidayPunchRequired == policyRequired
		out.CompMultiplier = row.HolidayCompMultiplier
	}
	return out
}

// sourceKindOf 由规则的绑定形态推导来源：
// 绑了考勤组 → group; 挂了组织 → org; 全空 → global.
// sourceUuid 为第 2 级的 orgUuid 时按组织来源算.
func sourceKindOf(row *hrm_overtime_rule.HrmOvertimeRule, sourceUuid string) string {
	if len(splitCsv(row.GroupUuids)) > 0 {
		return ruleSourceGroup
	}
	if row.OrgUuid != "" {
		return ruleSourceOrg
	}
	return ruleSourceGlobal
}

func containsStr(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

// Resolve HTTP 入口：把 ResolveOvertimeRule 结果映射到 api 类型.
func (l *Resolve) Resolve(req *types.ResolveRequest) (resp *types.ResolveResponse, err error) {
	out, err := ResolveOvertimeRule(l.ctx, l.svcCtx, req.GroupUuid, req.OrgUuid, req.DayType)
	if err != nil {
		return nil, err
	}
	return &types.ResolveResponse{
		Found:             out.Found,
		Source:            out.Source,
		SourceUuid:        out.SourceUuid,
		SourceName:        out.SourceName,
		DayType:           out.DayType,
		ApprovalRequired:  out.ApprovalRequired,
		PunchRequired:     out.PunchRequired,
		CompMultiplier:    out.CompMultiplier,
		MinUnitMinutes:    out.MinUnitMinutes,
		MonthlyLimitHours: out.MonthlyLimitHours,
	}, nil
}
