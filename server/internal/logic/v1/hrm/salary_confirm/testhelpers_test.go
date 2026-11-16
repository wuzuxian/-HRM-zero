package salary_confirm

import (
	"context"
	"testing"

	"github.com/jzero-io/jzero/core/status"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_confirm_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	sc_types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_confirm"
)

// DB-backed logic tests for 薪资确认审批 HR209 (HRM-38, issue #40), run via
// svctest: sqlite by default; HRM_TEST_DB=mysql runs the same assertions against
// docker MySQL.
//
// Coverage (issue #40 acceptance criteria + ticket 票面):
//   - 守卫：批次未试算拒绝（10685）；批次已确认拒绝（10711）
//   - 审批通过事务内：批次 confirm_status='1' + approve_time（幂等 applied）
//   - 幂等/单据状态机：一批次一张审批单，审批中/审批通过重复提交拒绝
//   - 快照：提交时 批次编号/名称/主体/年月/人数/实发合计 固化（批次编辑不影响单据）
//   - mine 口径：非管理员只看本人创建单据

var testCtx = context.Background()

// svcCtx package-level slot so 布景 helpers read like the salary_batch ones.
var svcCtx *svc.ServiceContext

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	svcCtx = &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
	return svcCtx
}

func assertErrCode(t *testing.T, err error, wantCode int, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error %d, got nil", what, wantCode)
	}
	if sc := status.FromError(err); sc == nil || sc.Code() != status.Code(wantCode) {
		t.Fatalf("%s: error = %v, want code %d", what, err, wantCode)
	}
}

func mustSubmit(t *testing.T, svcCtx *svc.ServiceContext, batchUuid string) string {
	t.Helper()
	resp, err := NewSubmit(testCtx, svcCtx, nil).Submit(&sc_types.SubmitRequest{BatchUuid: batchUuid})
	if err != nil {
		t.Fatalf("submit salary confirm for %s: %v", batchUuid, err)
	}
	return resp.Uuid
}

func mustApprove(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) {
	t.Helper()
	if err := bill.ApproveBill(WithSvcCtx(testCtx, svcCtx), svcCtx.SqlxConn, BillTypeCode, billUuid,
		bill.Operator{Name: "审批人"}, "同意"); err != nil {
		t.Fatalf("approve bill %s: %v", billUuid, err)
	}
}

func batchOf(t *testing.T, svcCtx *svc.ServiceContext, batchUuid string) *hrm_salary_batch.HrmSalaryBatch {
	t.Helper()
	row, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(testCtx, nil, batchUuid)
	if err != nil {
		t.Fatalf("find batch %s: %v", batchUuid, err)
	}
	return row
}

func billOf(t *testing.T, svcCtx *svc.ServiceContext, billUuid string) *hrm_salary_confirm_bill.HrmSalaryConfirmBill {
	t.Helper()
	row, err := svcCtx.Model.HrmSalaryConfirmBill.FindOneByUuid(testCtx, nil, billUuid)
	if err != nil {
		t.Fatalf("find confirm bill %s: %v", billUuid, err)
	}
	return row
}
