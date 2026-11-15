package perf_scheme

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
)

// DB-backed logic tests for 绩效方案 CRUD (HRM-41), run via svctest:
// sqlite by default; HRM_TEST_DB=mysql runs the same assertions against a
// disposable docker MySQL with the real migrations.
//
// Coverage (issue #43 acceptance criteria):
//   - 方案 CRUD (add/list/edit/delete)
//   - 模板引用有效性（不存在的模板拒绝）
//   - 枚举字段校验（周期类型/评估模式/评分模式/联薪方式/启停状态）
//   - 启停（状态默认启用、非法状态拒绝）

var testCtx = context.Background()

// mustAddTemplate seeds one 考核模板 row directly and returns its uuid.
func mustAddTemplate(t *testing.T, env *svctest.Env, name string) string {
	t.Helper()
	uuid := fmt.Sprintf("tpl-%s", name)
	err := hrm_perf_template.NewHrmPerfTemplateModel(env.SqlConn).InsertV2(testCtx, nil, &hrm_perf_template.HrmPerfTemplate{
		Uuid:   uuid,
		Name:   name,
		Status: "1",
		Remark: "",
	})
	if err != nil {
		t.Fatalf("seed template: %v", err)
	}
	return uuid
}

func allSchemes(t *testing.T, env *svctest.Env) []*hrm_perf_scheme.HrmPerfScheme {
	t.Helper()
	rows, err := hrm_perf_scheme.NewHrmPerfSchemeModel(env.SqlConn).FindByCondition(testCtx, nil, nil...)
	if err != nil {
		t.Fatalf("find schemes: %v", err)
	}
	return rows
}

func TestSchemeCrud(t *testing.T) {
	env := svctest.New(t)

	tplUuid := mustAddTemplate(t, env, "技术序列模板")

	// Add
	err := AddScheme(testCtx, env.SqlConn, "2026年度绩效方案", tplUuid, "year", "kpi_behavior", "percent", "coefficient", "1", "备注")
	if err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	rows := allSchemes(t, env)
	if len(rows) != 1 {
		t.Fatalf("want 1 scheme, got %d", len(rows))
	}
	row := rows[0]
	if row.Name != "2026年度绩效方案" || row.TemplateUuid != tplUuid ||
		row.PeriodType != "year" || row.AssessmentMode != "kpi_behavior" ||
		row.ScoreMode != "percent" || row.PayLinkMode != "coefficient" ||
		row.Status != "1" || row.Uuid == "" {
		t.Fatalf("unexpected scheme row: %+v", row)
	}

	// Edit — change enums + 停用
	if err := EditScheme(testCtx, env.SqlConn, row.Uuid, "2026季度绩效方案", tplUuid, "quarter", "kpi", "grade", "amount", "2", ""); err != nil {
		t.Fatalf("edit scheme: %v", err)
	}
	updated, err := hrm_perf_scheme.NewHrmPerfSchemeModel(env.SqlConn).FindOneByUuid(testCtx, nil, row.Uuid)
	if err != nil {
		t.Fatalf("find after edit: %v", err)
	}
	if updated.Name != "2026季度绩效方案" || updated.PeriodType != "quarter" ||
		updated.AssessmentMode != "kpi" || updated.ScoreMode != "grade" ||
		updated.PayLinkMode != "amount" || updated.Status != "2" {
		t.Fatalf("unexpected edited row: %+v", updated)
	}

	// Edit — not found
	err = EditScheme(testCtx, env.SqlConn, "missing-uuid", "n", tplUuid, "month", "kpi", "percent", "coefficient", "1", "")
	if !errors.Is(err, ErrSchemeNotFound) {
		t.Fatalf("want ErrSchemeNotFound, got %v", err)
	}

	// Delete
	if err := DeleteSchemes(testCtx, env.SqlConn, []string{row.Uuid}); err != nil {
		t.Fatalf("delete scheme: %v", err)
	}
	_, err = hrm_perf_scheme.NewHrmPerfSchemeModel(env.SqlConn).FindOneByUuid(testCtx, nil, row.Uuid)
	if !errors.Is(err, hrm_perf_scheme.ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}

func TestSchemeTemplateReference(t *testing.T) {
	env := svctest.New(t)

	// 不存在的模板拒绝保存（模板引用有效性）
	err := AddScheme(testCtx, env.SqlConn, "方案", "no-such-template", "month", "kpi", "percent", "coefficient", "1", "")
	if !errors.Is(err, ErrSchemeTemplateNotFound) {
		t.Fatalf("want ErrSchemeTemplateNotFound, got %v", err)
	}

	tplUuid := mustAddTemplate(t, env, "销售序列模板")
	if err := AddScheme(testCtx, env.SqlConn, "方案", tplUuid, "month", "kpi", "percent", "coefficient", "1", ""); err != nil {
		t.Fatalf("add with valid template: %v", err)
	}

	rows := allSchemes(t, env)
	if len(rows) != 1 {
		t.Fatalf("want 1 scheme, got %d", len(rows))
	}

	// 编辑时指向不存在的模板同样拒绝
	err = EditScheme(testCtx, env.SqlConn, rows[0].Uuid, "方案", "no-such-template", "month", "kpi", "percent", "coefficient", "1", "")
	if !errors.Is(err, ErrSchemeTemplateNotFound) {
		t.Fatalf("edit: want ErrSchemeTemplateNotFound, got %v", err)
	}

	// 空模板引用拒绝
	err = AddScheme(testCtx, env.SqlConn, "方案", "", "month", "kpi", "percent", "coefficient", "1", "")
	if !errors.Is(err, ErrSchemeTemplateRequired) {
		t.Fatalf("want ErrSchemeTemplateRequired, got %v", err)
	}
}

func TestSchemeEnumValidation(t *testing.T) {
	env := svctest.New(t)
	tplUuid := mustAddTemplate(t, env, "模板")

	cases := []struct {
		label  string
		mutate func(*schemeArgs)
	}{
		{"周期类型非法", func(a *schemeArgs) { a.periodType = "weekly" }},
		{"评估模式非法", func(a *schemeArgs) { a.assessmentMode = "360" }},
		{"评分模式非法", func(a *schemeArgs) { a.scoreMode = "stars" }},
		{"联薪方式非法", func(a *schemeArgs) { a.payLinkMode = "bonus" }},
		{"状态非法", func(a *schemeArgs) { a.status = "9" }},
		{"名称为空", func(a *schemeArgs) { a.name = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			a := validArgs(tplUuid)
			tc.mutate(&a)
			err := AddScheme(testCtx, env.SqlConn, a.name, a.templateUuid, a.periodType, a.assessmentMode, a.scoreMode, a.payLinkMode, a.status, "")
			if !errors.Is(err, ErrSchemeInvalid) && !errors.Is(err, ErrSchemeNameRequired) {
				t.Fatalf("%s: want scheme invalid error, got %v", tc.label, err)
			}
		})
	}

	// 每个合法枚举组合都能通过校验并持久化
	legal := [][]string{
		{"month", "kpi", "percent", "coefficient"},
		{"quarter", "kpi_behavior", "grade", "amount"},
		{"half_year", "kpi", "amount", "coefficient"},
		{"year", "kpi_behavior", "percent", "amount"},
	}
	for i, combo := range legal {
		if err := AddScheme(testCtx, env.SqlConn, fmt.Sprintf("方案%d", i), tplUuid, combo[0], combo[1], combo[2], combo[3], "1", ""); err != nil {
			t.Fatalf("combo %v should be legal: %v", combo, err)
		}
	}
}

func TestSchemeStatusDefault(t *testing.T) {
	env := svctest.New(t)
	tplUuid := mustAddTemplate(t, env, "模板")

	// 状态留空默认启用（'1'）—— default=1 标签只作用于 HTTP 解析，Add 逻辑兜底
	if err := AddScheme(testCtx, env.SqlConn, "默认启用方案", tplUuid, "month", "kpi", "percent", "coefficient", "", ""); err != nil {
		t.Fatalf("add with empty status: %v", err)
	}
	rows := allSchemes(t, env)
	if len(rows) != 1 {
		t.Fatalf("want 1 scheme, got %d", len(rows))
	}
	if rows[0].Status != "1" {
		t.Fatalf("want default status '1', got %q", rows[0].Status)
	}
}

// schemeArgs mirrors AddScheme's parameter list for table-driven mutation.
type schemeArgs struct {
	name, templateUuid, periodType, assessmentMode, scoreMode, payLinkMode, status string
}

func validArgs(templateUuid string) schemeArgs {
	return schemeArgs{
		name:           "绩效方案",
		templateUuid:   templateUuid,
		periodType:     "year",
		assessmentMode: "kpi_behavior",
		scoreMode:      "percent",
		payLinkMode:    "coefficient",
		status:         "1",
	}
}
