package perf_indicator

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	_ "modernc.org/sqlite"
)

// DB-backed logic tests for 指标库 CRUD, run against a temp sqlite database
// with the migration-12 tables.

var testCtx = context.Background()

func newTestConn(t *testing.T) sqlx.SqlConn {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "perf_indicator_test.db"))
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
	}
	for _, s := range stmts {
		if _, err := conn.ExecCtx(testCtx, s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	return conn
}

func mustAdd(t *testing.T, conn sqlx.SqlConn, code string) string {
	t.Helper()
	err := AddIndicator(testCtx, conn, Indicator{
		Code:          code,
		Name:          "指标-" + code,
		IndicatorType: IndicatorTypeKpi,
		ScoreRule:     ScoreRuleManual,
		DataSource:    DataSourceManual,
		DefaultWeight: 20,
		Status:        StatusEnabled,
	})
	if err != nil {
		t.Fatalf("AddIndicator %s: %v", code, err)
	}
	row, err := FindIndicatorByCode(testCtx, conn, code)
	if err != nil {
		t.Fatalf("FindIndicatorByCode %s: %v", code, err)
	}
	return row.Uuid
}

func TestAddIndicatorDuplicateCodeRejected(t *testing.T) {
	conn := newTestConn(t)
	mustAdd(t, conn, "KPI01")
	err := AddIndicator(testCtx, conn, Indicator{
		Code:          "KPI01",
		Name:          "另一个指标",
		IndicatorType: IndicatorTypeBehavior,
		ScoreRule:     ScoreRuleFormula,
		DataSource:    DataSourceSystem,
		Status:        StatusEnabled,
	})
	if !errors.Is(err, ErrIndicatorCodeDuplicated) {
		t.Fatalf("AddIndicator = %v, want ErrIndicatorCodeDuplicated", err)
	}
}

func TestAddIndicatorInvalidEnumRejected(t *testing.T) {
	conn := newTestConn(t)
	err := AddIndicator(testCtx, conn, Indicator{
		Code:          "KPI99",
		Name:          "非法类型",
		IndicatorType: "bad",
		ScoreRule:     ScoreRuleManual,
		DataSource:    DataSourceManual,
		Status:        StatusEnabled,
	})
	if !errors.Is(err, ErrIndicatorInvalid) {
		t.Fatalf("AddIndicator = %v, want ErrIndicatorInvalid", err)
	}
	err = AddIndicator(testCtx, conn, Indicator{
		Code:          "",
		Name:          "空编码",
		IndicatorType: IndicatorTypeKpi,
		ScoreRule:     ScoreRuleManual,
		DataSource:    DataSourceManual,
		Status:        StatusEnabled,
	})
	if !errors.Is(err, ErrIndicatorInvalid) {
		t.Fatalf("AddIndicator empty code = %v, want ErrIndicatorInvalid", err)
	}
}

func TestEditIndicator(t *testing.T) {
	conn := newTestConn(t)
	uuid := mustAdd(t, conn, "KPI01")

	err := EditIndicator(testCtx, conn, Indicator{
		Uuid:          uuid,
		Code:          "KPI01",
		Name:          "改名",
		IndicatorType: IndicatorTypeBehavior,
		ScoreRule:     ScoreRuleFormula,
		DataSource:    DataSourceSystem,
		DefaultWeight: 30,
		Status:        StatusDisabled,
		Remark:        "备注",
	})
	if err != nil {
		t.Fatalf("EditIndicator: %v", err)
	}
	row, err := FindIndicatorByUuid(testCtx, conn, uuid)
	if err != nil {
		t.Fatalf("FindIndicatorByUuid: %v", err)
	}
	if row.Name != "改名" || row.IndicatorType != IndicatorTypeBehavior ||
		row.ScoreRule != ScoreRuleFormula || row.DataSource != DataSourceSystem ||
		row.DefaultWeight != 30 || row.Status != StatusDisabled || row.Remark != "备注" {
		t.Fatalf("edit not applied: %+v", row)
	}

	// code stays unique across uuids
	other := mustAdd(t, conn, "KPI02")
	err = EditIndicator(testCtx, conn, Indicator{
		Uuid:          other,
		Code:          "KPI01",
		Name:          "抢占编码",
		IndicatorType: IndicatorTypeKpi,
		ScoreRule:     ScoreRuleManual,
		DataSource:    DataSourceManual,
		Status:        StatusEnabled,
	})
	if !errors.Is(err, ErrIndicatorCodeDuplicated) {
		t.Fatalf("EditIndicator dup code = %v, want ErrIndicatorCodeDuplicated", err)
	}
}

func TestEditIndicatorNotFound(t *testing.T) {
	conn := newTestConn(t)
	err := EditIndicator(testCtx, conn, Indicator{
		Uuid:          "missing-uuid",
		Code:          "KPI01",
		Name:          "不存在",
		IndicatorType: IndicatorTypeKpi,
		ScoreRule:     ScoreRuleManual,
		DataSource:    DataSourceManual,
		Status:        StatusEnabled,
	})
	if !errors.Is(err, ErrIndicatorNotFound) {
		t.Fatalf("EditIndicator = %v, want ErrIndicatorNotFound", err)
	}
}

func TestDeleteIndicatorInUseRejected(t *testing.T) {
	conn := newTestConn(t)
	free := mustAdd(t, conn, "KPI_FREE")
	used := mustAdd(t, conn, "KPI_USED")

	// simulate a template referencing the indicator (template logic lives in
	// the sibling perf_template package)
	_, err := conn.ExecCtx(testCtx,
		"INSERT INTO hrm_perf_template (uuid, name, status) VALUES ('tpl-1', '模板', '1')")
	if err != nil {
		t.Fatalf("seed template: %v", err)
	}
	_, err = conn.ExecCtx(testCtx,
		"INSERT INTO hrm_perf_template_item (uuid, template_uuid, indicator_uuid, weight) VALUES ('item-1', 'tpl-1', ?, 100)", used)
	if err != nil {
		t.Fatalf("seed item: %v", err)
	}

	err = DeleteIndicators(testCtx, conn, []string{free, used})
	if !errors.Is(err, ErrIndicatorInUse) {
		t.Fatalf("DeleteIndicators = %v, want ErrIndicatorInUse", err)
	}
	// nothing must be deleted when one indicator is still referenced
	if _, err := FindIndicatorByUuid(testCtx, conn, free); err != nil {
		t.Fatalf("free indicator must survive: %v", err)
	}

	if err := DeleteIndicators(testCtx, conn, []string{free}); err != nil {
		t.Fatalf("DeleteIndicators free: %v", err)
	}
	if _, err := FindIndicatorByUuid(testCtx, conn, free); !errors.Is(err, ErrIndicatorNotFound) {
		t.Fatalf("free indicator must be gone: %v", err)
	}
}
