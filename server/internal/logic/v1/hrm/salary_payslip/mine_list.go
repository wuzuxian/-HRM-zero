package salary_payslip

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payslip"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payslip"
)

// actorEmployeeUuid resolves the JWT 登录账号 to the bound 员工 uuid; "" when
// unauthenticated or no employee row (mine 契约与 perf_result 同款).
func actorEmployeeUuid(ctx context.Context, conn sqlx.SqlConn) (string, error) {
	user := hrmuser.Current(ctx)
	if user.Uuid == "" {
		return "", nil
	}
	rows, err := hrm_employee.NewHrmEmployeeModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.UserUuid, user.Uuid).Build()...)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0].Uuid, nil
}

type MineList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMineList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *MineList {
	return &MineList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// MineList 我的工资条（员工本人端点；JWT 登录账号 → hrm_employee.user_uuid 反查；
// mine 契约与 perf_result 同款）。查看即记已读由前端对每张工资条调用 Read 端点完成。
func (l *MineList) MineList(req *types.MineListRequest) (resp *types.MineListResponse, err error) {
	employeeUuid, err := actorEmployeeUuid(l.ctx, l.svcCtx.SqlxConn)
	if err != nil {
		return nil, err
	}
	if employeeUuid == "" {
		return &types.MineListResponse{PageResponse: types.PageResponse{
			Current: req.Current, Size: req.Size, Total: 0,
		}, Records: []types.HrmSalaryPayslip{}}, nil
	}

	chain := condition.NewChain().
		Equal(hrm_salary_payslip.EmployeeUuid, employeeUuid).
		Equal(hrm_salary_payslip.Published, 1).
		OrderByDesc(hrm_salary_payslip.BatchYear).
		OrderByDesc(hrm_salary_payslip.BatchMonth).
		OrderByDesc(hrm_salary_payslip.Id).
		Page(int(req.Current), int(req.Size))

	if req.YearMonth != "" {
		year, month, ok := splitYearMonth(req.YearMonth)
		if ok {
			chain = chain.Equal(hrm_salary_payslip.BatchYear, year).Equal(hrm_salary_payslip.BatchMonth, month)
		}
	}

	rows, total, err := l.svcCtx.Model.HrmSalaryPayslip.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]types.HrmSalaryPayslip, 0, len(rows))
	for _, row := range rows {
		records = append(records, payslipRowToRecord(row))
	}
	return &types.MineListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
