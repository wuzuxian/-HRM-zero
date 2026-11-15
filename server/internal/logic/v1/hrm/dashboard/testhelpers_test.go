package dashboard

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendanceconfig"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_leave_balance"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	configtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendanceconfig"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"

	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/dashboard"
)

const testUserUuid = "test-user-uuid"

// newDashboardCtx 看板测试环境: Model + SqlxConn（punch.SettleDay 事务写日结行）。
func newDashboardCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// mustResp 员工端视角调用看板端点（当前登录人 test-user-uuid）并断言成功。
func mustResp(t *testing.T, svcCtx *svc.ServiceContext, month string) *types.GetMyDashboardResponse {
	t.Helper()
	ctx := context.WithValue(context.Background(), "uuid", testUserUuid)
	resp, err := NewGetMyDashboard(ctx, svcCtx, nil).GetMyDashboard(&types.GetMyDashboardRequest{Month: month})
	if err != nil {
		t.Fatalf("GetMyDashboard(%s): %v", month, err)
	}
	return resp
}

// seedEmployeeViaLogic 建公司A→技术部（组织默认规则 09:00-18:00 周一~周五）+ 员工张三
// 并把 user_uuid 关联到 test-user-uuid（语义同 GenerateUser 回写）。返回员工 uuid。
func seedEmployeeViaLogic(t *testing.T, svcCtx *svc.ServiceContext) string {
	t.Helper()
	ctx := context.Background()

	if _, err := org.NewAdd(ctx, svcCtx, nil).Add(&orgtypes.AddRequest{OrgName: "公司A", OrgType: "1", OrderNum: 1, Status: "1"}); err != nil {
		t.Fatalf("add company: %v", err)
	}
	if _, err := org.NewAdd(ctx, svcCtx, nil).Add(&orgtypes.AddRequest{OrgName: "技术部", OrgType: "2", ParentUuid: mustOrgUuid(t, svcCtx, "公司A"), OrderNum: 1, Status: "1"}); err != nil {
		t.Fatalf("add dept: %v", err)
	}
	orgTech := mustOrgUuid(t, svcCtx, "技术部")

	if _, err := attendanceconfig.NewSave(ctx, svcCtx, nil).Save(&configtypes.SaveRequest{
		OrgUuid: orgTech, StartTime: "09:00", EndTime: "18:00", WorkDays: []int64{1, 2, 3, 4, 5},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if _, err := employee.NewAdd(ctx, svcCtx, nil).Add(&emptypes.AddRequest{
		Name: "张三", EmployeeNo: "E0001", OrgUuid: orgTech, PersonStatus: "1",
	}); err != nil {
		t.Fatalf("add employee: %v", err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0001").Build()...)
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if err = svcCtx.Model.HrmEmployee.UpdateFieldsByCondition(ctx, nil,
		map[string]any{"user_uuid": testUserUuid},
		condition.NewChain().Equal(hrm_employee.Uuid, row.Uuid).Build()...); err != nil {
		t.Fatalf("link user: %v", err)
	}
	return row.Uuid
}

// seedEmployeeViaLogicNoOrg 员工张三（无任何组织/规则 → 三级回退全空 no_rule）。
func seedEmployeeViaLogicNoOrg(t *testing.T, svcCtx *svc.ServiceContext) string {
	t.Helper()
	ctx := context.Background()
	if _, err := employee.NewAdd(ctx, svcCtx, nil).Add(&emptypes.AddRequest{
		Name: "张三", EmployeeNo: "E0001", PersonStatus: "1",
	}); err != nil {
		t.Fatalf("add employee: %v", err)
	}
	row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.EmployeeNo, "E0001").Build()...)
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	if err = svcCtx.Model.HrmEmployee.UpdateFieldsByCondition(ctx, nil,
		map[string]any{"user_uuid": testUserUuid},
		condition.NewChain().Equal(hrm_employee.Uuid, row.Uuid).Build()...); err != nil {
		t.Fatalf("link user: %v", err)
	}
	return row.Uuid
}

func mustOrgUuid(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	rows, err := svcCtx.Model.HrmOrg.FindByCondition(context.Background(), nil)
	if err != nil {
		t.Fatalf("find org %s: %v", name, err)
	}
	for _, row := range rows {
		if row.OrgName == name {
			return row.Uuid
		}
	}
	t.Fatalf("org %s not found", name)
	return ""
}

// seedSettledDay 落一行已日结打卡记录（直写 settle_* 快照，口径与 punch.SettleDay 落库一致）。
func seedSettledDay(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date, settleStatus string,
	lateMin, earlyMin int64, workHours float64, overtime string) {
	t.Helper()
	day, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		t.Fatalf("bad date %s: %v", date, err)
	}
	row := &hrm_punch_record.HrmPunchRecord{
		Uuid:               uuid.New().String(),
		PunchDate:          date,
		EmployeeUuid:       employeeUuid,
		EmployeeNo:         "E0001",
		EmployeeName:       "张三",
		ClockInTime:        sql.NullTime{Time: day.Add(9 * time.Hour), Valid: true},
		ClockInStatus:      "normal",
		ClockOutTime:       sql.NullTime{Time: day.Add(18 * time.Hour), Valid: true},
		ClockOutStatus:     "normal",
		Overtime:           overtime,
		Source:             "app",
		SettleStatus:       settleStatus,
		SettleLateMinutes:  lateMin,
		SettleEarlyMinutes: earlyMin,
		SettleWorkHours:    workHours,
		SettleTime:         sql.NullTime{Time: day.Add(18 * time.Hour), Valid: true},
	}
	if settleStatus == "missing" {
		row.ClockInTime = sql.NullTime{}
		row.ClockInStatus = ""
		row.ClockOutTime = sql.NullTime{}
		row.ClockOutStatus = ""
	}
	if _, err := svcCtx.Model.HrmPunchRecord.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("seed punch row %s: %v", date, err)
	}
}

// seedOvertimeBill 落一张已审批通过的加班单（confirm_hours 直接快照）。
func seedOvertimeBill(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date string, confirmHours float64) {
	t.Helper()
	row := &hrm_overtime_bill.HrmOvertimeBill{
		Uuid:         uuid.New().String(),
		BillNo:       "HR217-TEST-" + uuid.New().String()[:8],
		Status:       "approved",
		EmployeeUuid: employeeUuid,
		EmployeeNo:   "E0001",
		EmployeeName: "张三",
		OvertimeDate: date,
		DayType:      "1",
		ConfirmHours: confirmHours,
	}
	if _, err := svcCtx.Model.HrmOvertimeBill.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("seed overtime bill: %v", err)
	}
}

// seedAnnualBalance 落一行年假余额。
func seedAnnualBalance(t *testing.T, svcCtx *svc.ServiceContext, year int64, employeeUuid string, granted, used, frozen float64) {
	t.Helper()
	row := &hrm_leave_balance.HrmLeaveBalance{
		Uuid:          uuid.New().String(),
		Year:          year,
		EmployeeUuid:  employeeUuid,
		LeaveTypeCode: "annual_leave",
		Granted:       granted,
		Used:          used,
		Frozen:        frozen,
	}
	if _, err := svcCtx.Model.HrmLeaveBalance.Insert(context.Background(), nil, row); err != nil {
		t.Fatalf("seed annual balance: %v", err)
	}
}
