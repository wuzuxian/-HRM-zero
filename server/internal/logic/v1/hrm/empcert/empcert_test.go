package empcert

import (
	"context"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/empcert"
	emptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}
}

func addEmployee(t *testing.T, svcCtx *svc.ServiceContext, name string) emptypes.HrmEmployee {
	t.Helper()
	if _, err := employee.NewAdd(context.Background(), svcCtx, nil).Add(&emptypes.AddRequest{Name: name, PersonStatus: "1"}); err != nil {
		t.Fatalf("add employee %s: %v", name, err)
	}
	list, err := employee.NewList(context.Background(), svcCtx, nil).List(&emptypes.ListRequest{PageRequest: emptypes.PageRequest{Current: 1, Size: 10}, Name: name})
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	if len(list.Records) != 1 {
		t.Fatalf("expected 1 employee %q, got %d", name, len(list.Records))
	}
	return list.Records[0]
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) types.ListResponse {
	t.Helper()
	req.Current = 1
	if req.Size == 0 {
		req.Size = 10
	}
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list certificates: %v", err)
	}
	return *resp
}

// TestEmployeeCertificateAddAndMask 覆盖新增证照 + 列表证号脱敏（spec 用户故事 16/17）：
// 出参证号保留前 2 后 2；姓名/工号/证书类型筛选可用。
func TestEmployeeCertificateAddAndMask(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "钱七")

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid,
		CertType:     "身份证",
		CertNo:       "440823199001011234",
		IssueDate:    "2010-01-01",
		ExpiryDate:   "2030-01-01",
	}); err != nil {
		t.Fatalf("add certificate: %v", err)
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid,
		CertType:     "职业资格证",
		CertNo:       "PMP-2026-8888",
		ExpiryDate:   "2026-11-20",
	}); err != nil {
		t.Fatalf("add certificate 2: %v", err)
	}

	list := mustList(t, svcCtx, types.ListRequest{})
	if list.Total != 2 {
		t.Fatalf("total=%d, want 2", list.Total)
	}
	for _, rec := range list.Records {
		if rec.CertNo == "440823199001011234" || rec.CertNo == "PMP-2026-8888" {
			t.Fatalf("certNo leaked plaintext: %q", rec.CertNo)
		}
	}
	if rec := list.Records[0]; rec.CertType == "身份证" {
		if rec.CertNo != "44**************34" {
			t.Fatalf("id-card mask = %q, want 44**************34", rec.CertNo)
		}
		if rec.DaysToExpiry <= 90 {
			t.Fatalf("2030 expiry days=%d, want >90 (normal)", rec.DaysToExpiry)
		}
		if rec.ExpiryState != "normal" {
			t.Fatalf("state=%q, want normal", rec.ExpiryState)
		}
	} else {
		if rec.CertNo != "PM*********88" {
			t.Fatalf("pmp mask = %q, want PM*********88", rec.CertNo)
		}
		if rec.ExpiryState != "warning" {
			t.Fatalf("state=%q, want warning", rec.ExpiryState)
		}
	}

	// 证书类型模糊筛选
	byType := mustList(t, svcCtx, types.ListRequest{CertType: "身份"})
	if byType.Total != 1 || byType.Records[0].CertType != "身份证" {
		t.Fatalf("certType filter = %+v, want only 身份证", byType.Records)
	}
	// 姓名/工号筛选
	byName := mustList(t, svcCtx, types.ListRequest{EmployeeName: "钱七"})
	if byName.Total != 2 {
		t.Fatalf("name filter total=%d, want 2", byName.Total)
	}

	// 必填校验
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{EmployeeUuid: emp.Uuid, CertType: "身份证"}); err == nil {
		t.Fatal("missing certNo should fail")
	}
	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{EmployeeUuid: "no-such", CertType: "身份证", CertNo: "123"}); err == nil {
		t.Fatal("unknown employee should fail")
	}
}

// TestEmployeeCertificateEditKeepsMaskedNo 覆盖编辑：回显脱敏值原样提交不落库（不把星号当新证号），
// 新明文正常更新；编辑后列表仍脱敏。
func TestEmployeeCertificateEditKeepsMaskedNo(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "孙八")

	if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
		EmployeeUuid: emp.Uuid, CertType: "学历证书", CertNo: "1043012026001234", ExpiryDate: "2027-01-01",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	rec := mustList(t, svcCtx, types.ListRequest{}).Records[0]

	// 前端把脱敏值原样提交 → 不覆盖明文
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: rec.Uuid, CertType: "学历证书", CertNo: rec.CertNo, ExpiryDate: "2027-06-30",
	}); err != nil {
		t.Fatalf("edit with masked no: %v", err)
	}
	out := mustList(t, svcCtx, types.ListRequest{}).Records[0]
	if out.ExpiryDate != "2027-06-30" {
		t.Fatalf("expiry after edit = %q, want 2027-06-30", out.ExpiryDate)
	}
	// 明文未被星号覆盖：脱敏值应保持与首次一致
	if out.CertNo != rec.CertNo {
		t.Fatalf("certNo changed after masked resubmit: %q -> %q", rec.CertNo, out.CertNo)
	}

	// 提交新的明文 → 正常更新
	if _, err := NewEdit(ctx, svcCtx, nil).Edit(&types.EditRequest{
		Uuid: rec.Uuid, CertType: "学历证书", CertNo: "99990000", ExpiryDate: "2027-06-30",
	}); err != nil {
		t.Fatalf("edit with new plaintext: %v", err)
	}
	out = mustList(t, svcCtx, types.ListRequest{}).Records[0]
	if out.CertNo != "99****00" {
		t.Fatalf("after plaintext edit mask = %q, want 99****00", out.CertNo)
	}
}

// TestEmployeeCertificateDelete 覆盖单删/批删。
func TestEmployeeCertificateDelete(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()
	emp := addEmployee(t, svcCtx, "周九")

	add := func(no string) string {
		if _, err := NewAdd(ctx, svcCtx, nil).Add(&types.AddRequest{
			EmployeeUuid: emp.Uuid, CertType: no, CertNo: no + "-NO",
		}); err != nil {
			t.Fatalf("add cert %s: %v", no, err)
		}
		list := mustList(t, svcCtx, types.ListRequest{CertType: no})
		if len(list.Records) != 1 {
			t.Fatalf("cert %s not found", no)
		}
		return list.Records[0].Uuid
	}
	_ = add("健康证")
	u2 := add("特种作业证")

	if _, err := NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{u2}}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list := mustList(t, svcCtx, types.ListRequest{})
	if list.Total != 1 || list.Records[0].CertType != "健康证" {
		t.Fatalf("after delete total=%d, want only 健康证", list.Total)
	}
}
