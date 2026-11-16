package bill

// 共享单据导出 DB-backed 测试（final-review F5, spec 用户故事 23）：
// round-trip 一类单据（203 调动单）— mine 过滤口径 + xlsx base64 内容可解开且含表头/数据行。

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/jzero-io/jzero-admin/server/internal/model/manage_menu"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_role_menu"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_transfer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/bill"
)

func newExportSvc(t *testing.T) (*svc.ServiceContext, *svctest.Env) {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}, env
}

func newExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportBillList {
	return NewExportBillList(ctx, svcCtx, nil)
}

// seedTransfer 直插调动单行（不走 logic，聚焦导出行为）。
func seedTransfer(t *testing.T, svcCtx *svc.ServiceContext, creatorUuid, creatorName, billNo string) {
	t.Helper()
	if err := svcCtx.Model.HrmTransferBill.InsertV2(context.Background(), nil, &hrm_transfer_bill.HrmTransferBill{
		Uuid:         "tu-" + billNo,
		BillNo:       billNo,
		Status:       "unsubmitted",
		EmployeeNo:   "E0001",
		EmployeeName: "张三",
		CreateBy:     creatorUuid,
		CreatorName:  creatorName,
	}); err != nil {
		t.Fatalf("seed transfer: %v", err)
	}
}

func TestExportBillListTransfer(t *testing.T) {
	svcCtx, _ := newExportSvc(t)
	ctx := authExportCtx("user-owner") // mine 过滤：只看自己的
	seedTransfer(t, svcCtx, "user-owner", "李四", "HR203-2026-00001")
	seedTransfer(t, svcCtx, "user-other", "王五", "HR203-2026-00002")

	resp, err := newExportLogic(ctx, svcCtx).ExportBillList(&types.ExportBillListRequest{BillType: "203"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if resp.FileName == "" || resp.Content == "" {
		t.Fatal("expected non-empty fileName/content")
	}

	// round-trip: base64 → excelize → 断言行
	raw, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	sheet := f.GetSheetName(0)
	grid, err := f.GetRows(sheet)
	if err != nil {
		t.Fatal(err)
	}
	if len(grid) != 2 {
		t.Fatalf("expected header+1 mine-only row, got %d rows", len(grid))
	}
	if grid[0][0] != "单据编号" {
		t.Fatalf("unexpected header: %v", grid[0])
	}
	if grid[1][0] != "HR203-2026-00001" {
		t.Fatalf("expected own bill only, got %v", grid[1])
	}

	// 未注册类型 → 报错
	if _, err := newExportLogic(ctx, svcCtx).ExportBillList(&types.ExportBillListRequest{BillType: "999"}); err == nil {
		t.Fatal("expected error for unregistered billType")
	}
}

func TestExportBillListAdminSeesAll(t *testing.T) {
	svcCtx, _ := newExportSvc(t)
	// 管理员级角色（持共享审批码，F4 语义）→ 看全部
	seedApproveGrantForExport(t, svcCtx)
	ctx := authExportCtx("user-admin", "role-hr")
	seedTransfer(t, svcCtx, "user-owner", "李四", "HR203-2026-00001")
	seedTransfer(t, svcCtx, "user-other", "王五", "HR203-2026-00002")

	resp, err := newExportLogic(ctx, svcCtx).ExportBillList(&types.ExportBillListRequest{BillType: "203"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	grid, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(grid) != 3 {
		t.Fatalf("admin expected header+2 rows, got %d", len(grid))
	}
}

// mine=1 强制管理员也只看自己的。
func TestExportBillListForcedMine(t *testing.T) {
	svcCtx, _ := newExportSvc(t)
	seedApproveGrantForExport(t, svcCtx)
	ctx := authExportCtx("user-admin", "role-hr")
	seedTransfer(t, svcCtx, "user-owner", "李四", "HR203-2026-00001")
	seedTransfer(t, svcCtx, "user-admin", "管理员", "HR203-2026-00003")

	resp, err := newExportLogic(ctx, svcCtx).ExportBillList(&types.ExportBillListRequest{BillType: "203", Mine: 1})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	raw, _ := base64.StdEncoding.DecodeString(resp.Content)
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	grid, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(grid) != 2 {
		t.Fatalf("forced-mine expected header+1 row, got %d", len(grid))
	}
	if grid[1][0] != "HR203-2026-00003" {
		t.Fatalf("expected own bill, got %v", grid[1])
	}
}

// authExportCtx 与 authx 同形（keys uuid/username/role_uuids）。
func authExportCtx(userUuid string, roleUuids ...string) context.Context {
	roles := make([]any, len(roleUuids))
	for i, r := range roleUuids {
		roles[i] = r
	}
	return context.WithValue(context.WithValue(context.WithValue(context.Background(),
		"uuid", userUuid), "username", "exporter"), "role_uuids", roles)
}

// seedApproveGrantForExport 种子：角色 role-hr 持共享审批码 v1:hrm:bill:approve
// （F4 CanSeeAllBills 口径）。菜单 uuid 99xx（migrations < 42xx，测试局部唯一）。
func seedApproveGrantForExport(t *testing.T, svcCtx *svc.ServiceContext) {
	t.Helper()
	perms := `[{"code":"v1:hrm:bill:approve","desc":"审批通过"}]`
	if err := svcCtx.Model.ManageMenu.InsertV2(context.Background(), nil, &manage_menu.ManageMenu{
		Uuid: "aaa10000-1111-4000-8000-000000009921", Status: "1", MenuType: "3",
		MenuName: "审批通过", Permissions: perms,
	}); err != nil {
		t.Fatalf("seed menu: %v", err)
	}
	if err := svcCtx.Model.ManageRoleMenu.InsertV2(context.Background(), nil, &manage_role_menu.ManageRoleMenu{
		Uuid: "aaa20000-1111-4000-8000-000000009921", RoleUuid: "role-hr",
		MenuUuid: "aaa10000-1111-4000-8000-000000009921",
	}); err != nil {
		t.Fatalf("seed role menu: %v", err)
	}
}
