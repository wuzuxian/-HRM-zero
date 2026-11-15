package cert_template

import (
	"context"
	"testing"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cert_template"
)

// Seams under test (HRM-30 证书模板): CRUD + 前缀唯一 + 启用停用.
func testSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func addTemplate(t *testing.T, svcCtx *svc.ServiceContext, name, prefix, source, status string) string {
	t.Helper()
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		TemplateName: name,
		CertType:     "结业证",
		Source:       source,
		NoPrefix:     prefix,
	}); err != nil {
		t.Fatalf("Add template %s: %v", name, err)
	}
	resp, err := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest:  types.PageRequest{Current: 1, Size: 10},
		TemplateName: name,
	})
	if err != nil {
		t.Fatalf("List template: %v", err)
	}
	if len(resp.Records) != 1 {
		t.Fatalf("list records = %d, want 1", len(resp.Records))
	}
	rec := resp.Records[0]
	if rec.Status != status {
		t.Fatalf("template status = %q, want %q", rec.Status, status)
	}
	return rec.Uuid
}

func TestAddTemplateDefaultEnabledAndPrefixUnique(t *testing.T) {
	svcCtx := testSvcCtx(t, svctest.New(t))

	// 默认启用
	addTemplate(t, svcCtx, "测试结业模板", "ZZ", "1", "1")

	// 前缀唯一：同前缀新增被拒（大小写归一）
	if _, err := NewAdd(context.Background(), svcCtx, nil).Add(&types.AddRequest{
		TemplateName: "另一个模板", CertType: "合格证", Source: "1", NoPrefix: "zz",
	}); err == nil {
		t.Fatal("duplicate prefix (case-insensitive) should be rejected")
	}

	// 前缀大小写归一为大写存储
	resp, _ := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, TemplateName: "测试结业模板",
	})
	if resp.Records[0].NoPrefix != "ZZ" {
		t.Fatalf("prefix = %q, want ZZ (normalized upper)", resp.Records[0].NoPrefix)
	}
}

func TestAddValidation(t *testing.T) {
	svcCtx := testSvcCtx(t, svctest.New(t))

	cases := []struct {
		name   string
		mutate func(*types.AddRequest)
	}{
		{"missing name", func(r *types.AddRequest) { r.TemplateName = "" }},
		{"missing certType", func(r *types.AddRequest) { r.CertType = "" }},
		{"bad source", func(r *types.AddRequest) { r.Source = "3" }},
		{"missing prefix", func(r *types.AddRequest) { r.NoPrefix = "" }},
		{"bad prefix chars", func(r *types.AddRequest) { r.NoPrefix = "J-N" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{TemplateName: "模板", CertType: "结业证", Source: "1", NoPrefix: "XX"}
			tc.mutate(req)
			if _, err := NewAdd(context.Background(), svcCtx, nil).Add(req); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}

func TestEditPrefixConflictAndToggleStatus(t *testing.T) {
	svcCtx := testSvcCtx(t, svctest.New(t))
	tn := addTemplate(t, svcCtx, "测试结业模板", "ZZ", "1", "1")
	aq := addTemplate(t, svcCtx, "安全模板", "YY", "1", "1")

	// 编辑改成别人已占用的前缀 → 拒绝
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{
		Uuid: tn, TemplateName: "测试结业模板", CertType: "结业证", Source: "1", NoPrefix: "YY",
	}); err == nil {
		t.Fatal("edit to a taken prefix should be rejected")
	}

	// 编辑成自己的前缀（不冲突）→ 允许；停用→启用切换
	if _, err := NewEdit(context.Background(), svcCtx, nil).Edit(&types.EditRequest{
		Uuid: tn, TemplateName: "测试结业模板", CertType: "结业证", Source: "2", NoPrefix: "ZZ",
	}); err != nil {
		t.Fatalf("edit with own prefix: %v", err)
	}

	// 停用
	if _, err := NewToggleStatus(context.Background(), svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{
		Uuid: tn, Status: "2",
	}); err != nil {
		t.Fatalf("ToggleStatus disable: %v", err)
	}
	resp, _ := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, TemplateName: "测试结业模板",
	})
	if resp.Records[0].Status != "2" {
		t.Fatalf("status after disable = %q, want 2", resp.Records[0].Status)
	}

	// 非法状态值被拒
	if _, err := NewToggleStatus(context.Background(), svcCtx, nil).ToggleStatus(&types.ToggleStatusRequest{
		Uuid: aq, Status: "9",
	}); err == nil {
		t.Fatal("invalid status should be rejected")
	}

	// 删除
	if _, err := NewDelete(context.Background(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{aq}}); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	resp, _ = NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, TemplateName: "安全模板",
	})
	if len(resp.Records) != 0 {
		t.Fatalf("deleted template still listed: %d", len(resp.Records))
	}
}
