package overtimerule

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_rule"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimerule"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// List 加班规则分页列表，支持按规则名称/状态筛选.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmOvertimeRule.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_overtime_rule.RuleName, "%"+req.RuleName+"%", condition.WithSkip(req.RuleName == "")).
			Equal(hrm_overtime_rule.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_overtime_rule.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmOvertimeRule, 0, len(rows))
	for _, row := range rows {
		records = append(records, ruleToType(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

// ruleToType row → api type（列表/编辑回显共用）.
func ruleToType(row *hrm_overtime_rule.HrmOvertimeRule) types.HrmOvertimeRule {
	return types.HrmOvertimeRule{
		Uuid:     row.Uuid,
		RuleName: row.RuleName,
		Workday: types.DayOvertimePolicy{
			ApprovalRequired: row.WorkdayApprovalRequired,
			PunchRequired:    row.WorkdayPunchRequired,
			CompMultiplier:   row.WorkdayCompMultiplier,
		},
		Restday: types.DayOvertimePolicy{
			ApprovalRequired: row.RestdayApprovalRequired,
			PunchRequired:    row.RestdayPunchRequired,
			CompMultiplier:   row.RestdayCompMultiplier,
		},
		Holiday: types.DayOvertimePolicy{
			ApprovalRequired: row.HolidayApprovalRequired,
			PunchRequired:    row.HolidayPunchRequired,
			CompMultiplier:   row.HolidayCompMultiplier,
		},
		MinUnitMinutes:    row.MinUnitMinutes,
		MonthlyLimitHours: row.MonthlyLimitHours,
		GroupUuids:        splitCsv(row.GroupUuids),
		OrgUuid:           row.OrgUuid,
		Status:            row.Status,
		Remark:            row.Remark,
		CreateTime:        row.CreateTime.Format(time.DateTime),
		UpdateTime:        row.UpdateTime.Format(time.DateTime),
	}
}
