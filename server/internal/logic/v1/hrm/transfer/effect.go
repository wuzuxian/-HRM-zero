package transfer

import (
	"context"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Effect is the business result of one 生效 of a 人事调动单: what should be
// written back onto the 员工档案 (HRM-03) when the bill takes effect.
type Effect struct {
	BillUuid     string  // 调动单 uuid
	EmployeeUuid string  // 员工档案 uuid（HRM-03）
	EmployeeNo   string  // 员工工号（快照，档案未接入时的兜底定位）
	EmployeeName string  // 员工姓名（快照）
	ChangeType   string  // 异动类型 post/org/salary/company/other
	NewCompany   string  // 新公司（组织 uuid）
	NewOrg       string  // 新部门（组织 uuid）
	NewPosition  string  // 新职位
	NewPost      string  // 新岗位
	NewSalary    float64 // 调整后薪资（调薪时填写，0=不涉及）
	EffectiveDate string // 生效日期 YYYY-MM-DD
}

// EffectApplier writes a transfer Effect onto the 员工档案.
// The call always runs inside the caller's transaction (审批通过事务 or
// 定时任务事务), so a failing write-back rolls back the whole action.
type EffectApplier func(ctx context.Context, session sqlx.Session, e Effect) error

var (
	applierMu sync.RWMutex
	// effectApplier is the HRM-03 seam (回写解耦接缝): HRM-03 员工档案
	// registers its implementation at init time via SetEffectApplier:
	//
	//	transfer.SetEffectApplier(func(ctx, session, e transfer.Effect) error {
	//	    return session-ctx update of hrm_employee (所属公司/部门/职位/岗位/薪资)
	//	})
	//
	// Until that lands (HRM-03 not merged yet), applier is nil and 生效
	// degrades gracefully: the action succeeds, the bill row keeps
	// applied=0, and the daily scan retries every day — so no effect is
	// lost once HRM-03 plugs in.
	effectApplier EffectApplier
)

// SetEffectApplier installs (or, with nil, clears) the employee-archive
// write-back implementation. Idempotent; later calls replace earlier ones.
func SetEffectApplier(f EffectApplier) {
	applierMu.Lock()
	defer applierMu.Unlock()
	effectApplier = f
}

func effectApplierSnapshot() EffectApplier {
	applierMu.RLock()
	defer applierMu.RUnlock()
	return effectApplier
}

// applyEffect runs the write-back. It returns applied=false (with nil error)
// when the HRM-03 seam is not wired yet — callers must then leave the bill's
// applied flag untouched so the daily scan retries after HRM-03 lands.
func applyEffect(ctx context.Context, session sqlx.Session, e Effect) (applied bool, err error) {
	applier := effectApplierSnapshot()
	if applier == nil {
		logx.WithContext(ctx).Infof(
			"transfer: 生效回写跳过（员工档案 HRM-03 接缝未接入，单据 %s 员工 %s/%s 保持待生效，接入后由定时任务自动补生效）",
			e.BillUuid, e.EmployeeNo, e.EmployeeName)
		return false, nil
	}
	if err := applier(ctx, session, e); err != nil {
		return false, err
	}
	return true, nil
}
