package overtimerule

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

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Edit 编辑加班规则（覆盖全部可编辑字段）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.OvertimeRuleNotFoundCode, "加班规则不存在")
	}
	fields := ruleFieldsFrom(req.RuleName, req.Workday, req.Restday, req.Holiday,
		req.MinUnitMinutes, req.MonthlyLimitHours, req.GroupUuids, req.OrgUuid, req.Status)
	if err = validateRuleBasic(l.svcCtx, fields); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmOvertimeRule.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_overtime_rule.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OvertimeRuleNotFoundCode, "加班规则不存在")
		}
		return nil, err
	}
	row.RuleName = req.RuleName
	row.WorkdayApprovalRequired = req.Workday.ApprovalRequired
	row.WorkdayPunchRequired = req.Workday.PunchRequired
	row.WorkdayCompMultiplier = req.Workday.CompMultiplier
	row.RestdayApprovalRequired = req.Restday.ApprovalRequired
	row.RestdayPunchRequired = req.Restday.PunchRequired
	row.RestdayCompMultiplier = req.Restday.CompMultiplier
	row.HolidayApprovalRequired = req.Holiday.ApprovalRequired
	row.HolidayPunchRequired = req.Holiday.PunchRequired
	row.HolidayCompMultiplier = req.Holiday.CompMultiplier
	row.MinUnitMinutes = minUnitOrDefault(req.MinUnitMinutes)
	row.MonthlyLimitHours = req.MonthlyLimitHours
	row.GroupUuids = joinCsv(req.GroupUuids)
	row.OrgUuid = req.OrgUuid
	row.Status = req.Status
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmOvertimeRule.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}
	return &types.EditResponse{}, nil
}
