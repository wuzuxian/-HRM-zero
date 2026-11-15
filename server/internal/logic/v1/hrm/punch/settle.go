package punch

// Settle 手动日结（demo-inventory 4.4 工具栏「手动日结」）：批量员工 × 日期区间
// 逐员工逐日调 SettleDay（复用三级回退 ResolveEffectiveRule），返回结构化日结结果。
// 未指定员工 = 全体员工；无打卡行的工作日也产出 missing 结果（空行落库固化缺卡）。

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/punch"
)

type Settle struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSettle(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Settle {
	return &Settle{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Settle) Settle(req *types.SettleRequest) (resp *types.SettleResponse, err error) {
	if req.StartDate == "" {
		return nil, status.ErrorMessage(errcodes.SettleInvalidCode, "日结开始日期必填")
	}
	start, parseErr := time.ParseInLocation(DateLayout, req.StartDate, time.Local)
	if parseErr != nil {
		return nil, status.ErrorMessage(errcodes.SettleInvalidCode, "日结日期必须为 yyyy-MM-dd: "+req.StartDate)
	}
	end := start
	if req.EndDate != "" {
		end, parseErr = time.ParseInLocation(DateLayout, req.EndDate, time.Local)
		if parseErr != nil {
			return nil, status.ErrorMessage(errcodes.SettleInvalidCode, "日结日期必须为 yyyy-MM-dd: "+req.EndDate)
		}
	}
	if end.Before(start) {
		return nil, status.ErrorMessage(errcodes.SettleInvalidCode, "日结结束日期不能早于开始日期")
	}
	if end.Sub(start).Hours() > 24*62 {
		return nil, status.ErrorMessage(errcodes.SettleInvalidCode, "日结日期区间不能超过两个月")
	}

	employees, err := l.resolveEmployees(req.EmployeeUuids)
	if err != nil {
		return nil, err
	}
	if len(employees) == 0 {
		return nil, status.ErrorMessage(errcodes.SettleInvalidCode, "没有可日结的员工")
	}

	records := make([]types.HrmSettleResult, 0, len(employees))
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format(DateLayout)
		for _, employee := range employees {
			result, settleErr := SettleDay(l.ctx, l.svcCtx, employee.Uuid, dateStr)
			if settleErr != nil {
				return nil, settleErr
			}
			records = append(records, types.HrmSettleResult{
				EmployeeUuid: employee.Uuid,
				EmployeeName: employee.Name,
				PunchDate:    dateStr,
				SettleStatus: result.Status,
				LateMinutes:  result.LateMinutes,
				EarlyMinutes: result.EarlyMinutes,
				WorkHours:    result.WorkHours,
			})
		}
	}
	return &types.SettleResponse{Records: records}, nil
}

// resolveEmployees 指定员工 uuid 或全体员工（跳过无部门员工）。
func (l *Settle) resolveEmployees(employeeUuids []string) ([]*hrm_employee.HrmEmployee, error) {
	if len(employeeUuids) > 0 {
		rows, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_employee.Uuid, employeeUuids).Build()...)
		if err != nil {
			return nil, err
		}
		return rows, nil
	}
	rows, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
		condition.NewChain().NotEqual(hrm_employee.OrgUuid, "").Build()...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
