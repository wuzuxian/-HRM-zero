package resignation

import (
	"context"
	"testing"
	"time"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// HRM-53 账号状态真实开关 (ADR 0007 生命周期一致性): 离职日定时任务把员工置
// 离职('5')的同一事务内, 关联账号 manage_user.status 同步置 '2'(禁用)。
// 档案未关联账号 / 档案不存在 的容错语义与档案回写一致 (幂等跳过)。

// TestScanDisablesLinkedAccount 离职生效回写同时禁用关联账号:
// 建档即开号后的落库形态 (manage_user 行 + 档案 user_uuid 指针)。
func TestScanDisablesLinkedAccount(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	ctx := testCtxFor("user-1", "张三HR", "some-role")

	seedEmployee(t, svcCtx, "emp-uuid-1", "EMP001", "张三", "2")
	if err := svcCtx.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid: "user-uuid-emp001", Username: "EMP001", Password: "x", Nickname: "张三", Status: "1",
	}); err != nil {
		t.Fatalf("seed linked account: %v", err)
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee: %v", err)
	}
	emp.UserUuid = "user-uuid-emp001"
	if err = svcCtx.Model.HrmEmployee.Update(context.Background(), nil, emp); err != nil {
		t.Fatalf("link account pointer: %v", err)
	}

	billUuid, _ := addResignation(t, svcCtx, ctx, nil)
	op := bill.Operator{UserUuid: "approver-1", Name: "李审批"}
	if err := bill.SubmitBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := bill.ApproveBill(context.Background(), env.SqlConn, BillTypeCode, billUuid, op, ""); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}

	// 离职日注入为扫描日 (与既有 scan 测试同一手法, 不依赖墙钟)
	now := time.Date(2026, 11, 16, 8, 0, 0, 0, time.Local)
	if _, err := env.SqlConn.ExecCtx(context.Background(),
		"UPDATE hrm_resignation_bill SET resign_date = ? WHERE uuid = ?",
		"2026-11-16", billUuid); err != nil {
		t.Fatalf("set resign_date: %v", err)
	}
	count, err := ScanPendingResigned(context.Background(), env.SqlConn, now)
	if err != nil {
		t.Fatalf("ScanPendingResigned: %v", err)
	}
	if count != 1 {
		t.Fatalf("due bill not applied: count=%d", count)
	}

	emp, err = svcCtx.Model.HrmEmployee.FindOneByUuid(context.Background(), nil, "emp-uuid-1")
	if err != nil {
		t.Fatalf("find employee after scan: %v", err)
	}
	if emp.PersonStatus != "5" {
		t.Fatalf("person_status = %q, want 5 (离职)", emp.PersonStatus)
	}
	user, err := svcCtx.Model.ManageUser.FindOneByUuid(context.Background(), nil, "user-uuid-emp001")
	if err != nil {
		t.Fatalf("find account after scan: %v", err)
	}
	if user.Status != "2" {
		t.Fatalf("account status = %q, want 2 (离职置禁用, 同一事务)", user.Status)
	}
}
