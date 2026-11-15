package overtimerule

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_rule"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimerule"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Add 新增加班规则（GLOSSARY「加班规则」）: 三种当天类型策略/最小单位/月上限/考勤组绑定/启停.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = statusEnabled
	}
	fields := ruleFieldsFrom(req.RuleName, req.Workday, req.Restday, req.Holiday,
		req.MinUnitMinutes, req.MonthlyLimitHours, req.GroupUuids, req.OrgUuid, req.Status)
	if err = validateRuleBasic(l.svcCtx, fields); err != nil {
		return nil, err
	}

	if _, err = l.svcCtx.Model.HrmOvertimeRule.Insert(l.ctx, nil, &hrm_overtime_rule.HrmOvertimeRule{
		Uuid:                    uuid.New().String(),
		RuleName:                req.RuleName,
		WorkdayApprovalRequired: req.Workday.ApprovalRequired,
		WorkdayPunchRequired:    req.Workday.PunchRequired,
		WorkdayCompMultiplier:   req.Workday.CompMultiplier,
		RestdayApprovalRequired: req.Restday.ApprovalRequired,
		RestdayPunchRequired:    req.Restday.PunchRequired,
		RestdayCompMultiplier:   req.Restday.CompMultiplier,
		HolidayApprovalRequired: req.Holiday.ApprovalRequired,
		HolidayPunchRequired:    req.Holiday.PunchRequired,
		HolidayCompMultiplier:   req.Holiday.CompMultiplier,
		MinUnitMinutes:          minUnitOrDefault(req.MinUnitMinutes),
		MonthlyLimitHours:       req.MonthlyLimitHours,
		GroupUuids:              joinCsv(req.GroupUuids),
		OrgUuid:                 req.OrgUuid,
		Status:                  req.Status,
		Remark:                  req.Remark,
	}); err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
