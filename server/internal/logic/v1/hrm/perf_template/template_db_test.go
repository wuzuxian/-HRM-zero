package perf_template

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	_ "modernc.org/sqlite"
)

// DB-backed logic tests for 考核模板 CRUD, run against a temp sqlite database
// with the migration-12 tables (hrm_perf_indicator / hrm_perf_template /
// hrm_perf_template_item). Indicators are seeded with raw SQL — the 指标库
// logic lives in the sibling perf_indicator package.

var testCtx = context.Background()

func newTestConn(t *testing.T) sqlx.SqlConn {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "perf_template_test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	conn := sqlx.NewSqlConnFromDB(db)

	stmts := []string{
		// mirrors desc/sql_migration/sqlite/12_perf_indicator_template.up.sql
		`CREATE TABLE hrm_perf_indicator (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			indicator_type TEXT NOT NULL,
			unit TEXT NOT NULL DEFAULT '',
			score_rule TEXT NOT NULL,
			data_source TEXT NOT NULL,
			default_weight REAL NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT '1',
			remark TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE hrm_perf_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT '1',
			remark TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE hrm_perf_template_item (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			template_uuid TEXT NOT NULL,
			indicator_uuid TEXT NOT NULL,
			weight REAL NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX idx_hrm_perf_template_item_template ON hrm_perf_template_item (template_uuid)`,
	}
	for _, s := range stmts {
		if _, err := conn.ExecCtx(testCtx, s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	return conn
}

// mustAddIndicator seeds one 指标库 indicator row and returns its uuid.
func mustAddIndicator(t *testing.T, conn sqlx.SqlConn, code string) string {
	t.Helper()
	_, err := conn.ExecCtx(testCtx,
		"INSERT INTO hrm_perf_indicator (uuid, code, name, indicator_type, unit, score_rule, data_source, default_weight, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fmt.Sprintf("ind-%s", code), code, "指标-"+code, "kpi", "", "manual", "manual", 20, "1")
	if err != nil {
		t.Fatalf("seed indicator %s: %v", code, err)
	}
	var row struct {
		Uuid string `db:"uuid"`
	}
	if err := conn.QueryRowCtx(testCtx, &row,
		"SELECT uuid FROM hrm_perf_indicator WHERE code = ?", code); err != nil {
		t.Fatalf("read indicator %s: %v", code, err)
	}
	return row.Uuid
}

func mustTemplateCount(t *testing.T, conn sqlx.SqlConn) int64 {
	t.Helper()
	n, err := hrm_perf_template.NewHrmPerfTemplateModel(conn).CountByCondition(testCtx, nil)
	if err != nil {
		t.Fatalf("count templates: %v", err)
	}
	return n
}

func mustFirstTemplateUuid(t *testing.T, conn sqlx.SqlConn) string {
	t.Helper()
	var row struct {
		Uuid string `db:"uuid"`
	}
	if err := conn.QueryRowCtx(testCtx, &row,
		"SELECT uuid FROM hrm_perf_template ORDER BY id LIMIT 1"); err != nil {
		t.Fatalf("read template uuid: %v", err)
	}
	return row.Uuid
}

func TestAddTemplateWithItemsAndDetail(t *testing.T) {
	conn := newTestConn(t)
	ind1 := mustAddIndicator(t, conn, "KPI01")
	ind2 := mustAddIndicator(t, conn, "BEH01")

	items := []TemplateItem{
		{IndicatorUuid: ind1, Weight: 60},
		{IndicatorUuid: ind2, Weight: 40},
	}
	if err := AddTemplate(testCtx, conn, "管理岗季度模板", "1", "备注A", items); err != nil {
		t.Fatalf("AddTemplate: %v", err)
	}

	detail, err := GetTemplateDetail(testCtx, conn, mustFirstTemplateUuid(t, conn))
	if err != nil {
		t.Fatalf("GetTemplateDetail: %v", err)
	}
	if detail.Name != "管理岗季度模板" || detail.Status != "1" || detail.Remark != "备注A" {
		t.Fatalf("unexpected template row: %+v", detail.HrmPerfTemplate)
	}
	if len(detail.Items) != 2 {
		t.Fatalf("want 2 items, got %d: %+v", len(detail.Items), detail.Items)
	}
	byIndicator := map[string]ItemWithIndicator{}
	for _, it := range detail.Items {
		byIndicator[it.IndicatorUuid] = it
	}
	got1, ok := byIndicator[ind1]
	if !ok {
		t.Fatalf("item for %s missing: %+v", ind1, detail.Items)
	}
	if got1.Weight != 60 || got1.IndicatorCode != "KPI01" || got1.IndicatorName != "指标-KPI01" || got1.IndicatorType != "kpi" {
		t.Fatalf("unexpected item snapshot: %+v", got1)
	}
	if byIndicator[ind2].Weight != 40 {
		t.Fatalf("unexpected weight for %s: %+v", ind2, byIndicator[ind2])
	}
}

func TestAddTemplateWeightSumRejected(t *testing.T) {
	conn := newTestConn(t)
	ind1 := mustAddIndicator(t, conn, "KPI01")
	ind2 := mustAddIndicator(t, conn, "KPI02")

	err := AddTemplate(testCtx, conn, "模板", "1", "", []TemplateItem{
		{IndicatorUuid: ind1, Weight: 50},
		{IndicatorUuid: ind2, Weight: 40},
	})
	if !errors.Is(err, ErrTemplateItemWeightSum) {
		t.Fatalf("AddTemplate = %v, want ErrTemplateItemWeightSum", err)
	}
	if n := mustTemplateCount(t, conn); n != 0 {
		t.Fatalf("rejected template must not persist, got %d rows", n)
	}
}

func TestAddTemplateUnknownIndicatorRejected(t *testing.T) {
	conn := newTestConn(t)
	err := AddTemplate(testCtx, conn, "模板", "1", "", []TemplateItem{
		{IndicatorUuid: "no-such-uuid", Weight: 100},
	})
	if !errors.Is(err, ErrTemplateItemInvalid) {
		t.Fatalf("AddTemplate = %v, want ErrTemplateItemInvalid", err)
	}
}

func TestEditTemplateReplacesItems(t *testing.T) {
	conn := newTestConn(t)
	ind1 := mustAddIndicator(t, conn, "KPI01")
	ind2 := mustAddIndicator(t, conn, "KPI02")
	ind3 := mustAddIndicator(t, conn, "KPI03")

	if err := AddTemplate(testCtx, conn, "模板", "1", "", []TemplateItem{
		{IndicatorUuid: ind1, Weight: 60},
		{IndicatorUuid: ind2, Weight: 40},
	}); err != nil {
		t.Fatalf("AddTemplate: %v", err)
	}
	templateUuid := mustFirstTemplateUuid(t, conn)

	if err := EditTemplate(testCtx, conn, templateUuid, "模板-改", "0", "备注B", []TemplateItem{
		{IndicatorUuid: ind3, Weight: 30},
		{IndicatorUuid: ind2, Weight: 70},
	}); err != nil {
		t.Fatalf("EditTemplate: %v", err)
	}

	detail, err := GetTemplateDetail(testCtx, conn, templateUuid)
	if err != nil {
		t.Fatalf("GetTemplateDetail: %v", err)
	}
	if detail.Name != "模板-改" || detail.Status != "0" || detail.Remark != "备注B" {
		t.Fatalf("template row not updated: %+v", detail.HrmPerfTemplate)
	}
	if len(detail.Items) != 2 {
		t.Fatalf("want 2 items after replace, got %d", len(detail.Items))
	}
	for _, it := range detail.Items {
		if it.IndicatorUuid == ind1 {
			t.Fatalf("old item %s should have been replaced", ind1)
		}
	}
}

func TestEditTemplateWeightSumRejected(t *testing.T) {
	conn := newTestConn(t)
	ind1 := mustAddIndicator(t, conn, "KPI01")
	ind2 := mustAddIndicator(t, conn, "KPI02")
	if err := AddTemplate(testCtx, conn, "模板", "1", "", []TemplateItem{
		{IndicatorUuid: ind1, Weight: 60},
		{IndicatorUuid: ind2, Weight: 40},
	}); err != nil {
		t.Fatalf("AddTemplate: %v", err)
	}
	templateUuid := mustFirstTemplateUuid(t, conn)

	err := EditTemplate(testCtx, conn, templateUuid, "模板", "1", "", []TemplateItem{
		{IndicatorUuid: ind1, Weight: 10},
		{IndicatorUuid: ind2, Weight: 20},
	})
	if !errors.Is(err, ErrTemplateItemWeightSum) {
		t.Fatalf("EditTemplate = %v, want ErrTemplateItemWeightSum", err)
	}
	detail, err := GetTemplateDetail(testCtx, conn, templateUuid)
	if err != nil {
		t.Fatalf("GetTemplateDetail: %v", err)
	}
	if len(detail.Items) != 2 {
		t.Fatalf("rejected edit must keep old items, got %d", len(detail.Items))
	}
}

func TestEditTemplateNotFound(t *testing.T) {
	conn := newTestConn(t)
	err := EditTemplate(testCtx, conn, "missing-uuid", "模板", "1", "", nil)
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("EditTemplate = %v, want ErrTemplateNotFound", err)
	}
}

func TestDeleteTemplatesRemovesItems(t *testing.T) {
	conn := newTestConn(t)
	ind1 := mustAddIndicator(t, conn, "KPI01")
	ind2 := mustAddIndicator(t, conn, "KPI02")
	if err := AddTemplate(testCtx, conn, "模板", "1", "", []TemplateItem{
		{IndicatorUuid: ind1, Weight: 60},
		{IndicatorUuid: ind2, Weight: 40},
	}); err != nil {
		t.Fatalf("AddTemplate: %v", err)
	}
	templateUuid := mustFirstTemplateUuid(t, conn)

	if err := DeleteTemplates(testCtx, conn, []string{templateUuid}); err != nil {
		t.Fatalf("DeleteTemplates: %v", err)
	}
	if n := mustTemplateCount(t, conn); n != 0 {
		t.Fatalf("template not deleted, %d rows left", n)
	}
	if _, err := GetTemplateDetail(testCtx, conn, templateUuid); !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("detail after delete = %v, want ErrTemplateNotFound", err)
	}
}
