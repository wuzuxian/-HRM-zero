package employee

import (
	"context"
	"testing"

	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

// HRM-53 账号状态真实开关 (ADR 0007 生命周期一致性): 人员状态在编辑中被置为
// 离职('5')时, 同一事务内把关联账号 manage_user.status 置 '2'(禁用)。
// 行为都从外部入口 (Add/Edit) 驱动, 账号状态经 model 读真值列。

// TestEditResignDisablesAccount 编辑人员状态翻成离职 → 关联账号同步禁用,
// 资料同步 (昵称=姓名) 不因离职失效。
func TestEditResignDisablesAccount(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)
	ctx := context.Background()

	addEmployee(t, svcCtx, &types.AddRequest{Name: "离职工", EmployeeNo: "10000041", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000041")
	if emp.UserUuid == "" {
		t.Fatal("precondition: 建档即开号 should have provisioned an account")
	}

	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: emp.Uuid, EmployeeNo: "10000041", Name: "离职改名", PersonStatus: PersonStatusResigned,
	}); err != nil {
		t.Fatalf("edit to resigned: %v", err)
	}

	user := mustFindUserByUsername(t, svcCtx, "10000041")
	if user.Status != "2" {
		t.Fatalf("resigned employee account status = %q, want 2 (离职置禁用)", user.Status)
	}
	if user.Nickname != "离职改名" {
		t.Fatalf("nickname = %q, want 离职改名 (离职编辑仍同步资料)", user.Nickname)
	}
}

// TestEditResignIdempotentAlreadyResigned 已是离职的档案再次全量编辑 (回传 '5')
// 不报错, 账号保持禁用。
func TestEditResignIdempotentAlreadyResigned(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)
	ctx := context.Background()

	addEmployee(t, svcCtx, &types.AddRequest{Name: "再离", EmployeeNo: "10000042", PersonStatus: PersonStatusRegular})
	emp := mustFindEmployeeByNo(t, svcCtx, "10000042")
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: emp.Uuid, Name: "再离", PersonStatus: PersonStatusResigned,
	}); err != nil {
		t.Fatalf("first resign edit: %v", err)
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: emp.Uuid, Name: "再离2", PersonStatus: PersonStatusResigned,
	}); err != nil {
		t.Fatalf("second resign edit (idempotent): %v", err)
	}
	if user := mustFindUserByUsername(t, svcCtx, "10000042"); user.Status != "2" {
		t.Fatalf("account status = %q, want 2 (保持禁用)", user.Status)
	}
}

// TestEditResignWithoutAccountNoError 未关联账号的档案编辑成离职不报错 (无账号可禁用).
func TestEditResignWithoutAccountNoError(t *testing.T) {
	svcCtx, _ := newSyncCtx(t)
	ctx := context.Background()

	addEmployeeWithoutAccount(t, svcCtx, &types.AddRequest{Name: "无号离职", PersonStatus: PersonStatusRegular})
	var noAccountUuid string
	for _, r := range mustList(t, svcCtx, types.ListRequest{}).Records {
		if r.UserStatus == "0" {
			noAccountUuid = r.Uuid
		}
	}
	if noAccountUuid == "" {
		t.Fatal("precondition: no employee without account found")
	}
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: noAccountUuid, Name: "无号离职", PersonStatus: PersonStatusResigned,
	}); err != nil {
		t.Fatalf("resign edit without account: %v", err)
	}
}
