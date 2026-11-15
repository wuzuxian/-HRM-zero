package bill

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	_ "modernc.org/sqlite"
)

// End-to-end logic tests for the shared 单据 approval lifecycle, run against
// sqlite with the migration tables for hrm_approval_record / hrm_bill_sequence
// and a stub document type (see stub_test.go).

var (
	testCtx   = context.Background()
	stubOp    = Operator{UserUuid: "user-uuid-1", Name: "张三"}
	stubDay   = time.Date(2026, 11, 12, 15, 4, 5, 0, time.Local)
	stubYM    = "20261112"
	stubBillNo = "HR901-2026111200001"
)

func newTestConn(t *testing.T) sqlx.SqlConn {
	t.Helper()
	registerStubs(t)

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "bill_test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	conn := sqlx.NewSqlConnFromDB(db)

	stmts := []string{
		// stub document table (test-only; a real 单据 table belongs to its own ticket)
		`CREATE TABLE stub_bill (
			uuid TEXT PRIMARY KEY,
			bill_no TEXT NOT NULL,
			status TEXT NOT NULL,
			write_back TEXT NOT NULL DEFAULT ''
		)`,
		// base tables — mirrors desc/sql_migration/sqlite/11_hrm_bill_approval_base.up.sql
		`CREATE TABLE hrm_approval_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			bill_type_code TEXT NOT NULL,
			bill_uuid TEXT NOT NULL,
			bill_no TEXT NOT NULL,
			"action" TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			operator_user_uuid TEXT NOT NULL,
			operator_name TEXT NOT NULL DEFAULT '',
			opinion TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX idx_hrm_approval_record_bill ON hrm_approval_record (bill_type_code, bill_uuid)`,
		`CREATE TABLE hrm_bill_sequence (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL UNIQUE,
			create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			bill_type_code TEXT NOT NULL,
			bill_date TEXT NOT NULL,
			last_seq INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE UNIQUE INDEX uni_hrm_bill_sequence_type_date ON hrm_bill_sequence (bill_type_code, bill_date)`,
	}
	for _, s := range stmts {
		if _, err := conn.ExecCtx(testCtx, s); err != nil {
			t.Fatalf("exec %q: %v", s, err)
		}
	}
	return conn
}

// mustCreateStubBill seeds a 未提交 stub document.
func mustCreateStubBill(t *testing.T, conn sqlx.SqlConn, uuid, billNo string) {
	t.Helper()
	_, err := conn.ExecCtx(testCtx,
		"INSERT INTO stub_bill (uuid, bill_no, status) VALUES (?, ?, ?)", uuid, billNo, StatusUnsubmitted)
	if err != nil {
		t.Fatalf("create stub bill: %v", err)
	}
}

func stubStatus(t *testing.T, conn sqlx.SqlConn, uuid string) (status, writeBack string) {
	t.Helper()
	var row struct {
		Status    string `db:"status"`
		WriteBack string `db:"write_back"`
	}
	if err := conn.QueryRowCtx(testCtx, &row,
		"SELECT status, write_back FROM stub_bill WHERE uuid = ?", uuid); err != nil {
		t.Fatalf("read stub bill: %v", err)
	}
	return row.Status, row.WriteBack
}

func mustRecords(t *testing.T, conn sqlx.SqlConn, typeCode, billUuid string) []*Record {
	t.Helper()
	records, err := Records(testCtx, conn, typeCode, billUuid)
	if err != nil {
		t.Fatalf("Records: %v", err)
	}
	return records
}

func recordSummary(records []*Record) []string {
	var out []string
	for _, r := range records {
		out = append(out, string(r.Action)+":"+r.FromStatus+"->"+r.ToStatus)
	}
	return out
}

func TestGenerateBillNo(t *testing.T) {
	conn := newTestConn(t)

	no1, err := GenerateBillNo(testCtx, conn, stubTypeCode, stubDay)
	if err != nil {
		t.Fatalf("GenerateBillNo: %v", err)
	}
	if no1 != "HR901-2026111200001" {
		t.Errorf("first no = %q, want HR901-2026111200001", no1)
	}

	no2, err := GenerateBillNo(testCtx, conn, stubTypeCode, stubDay)
	if err != nil {
		t.Fatalf("GenerateBillNo second: %v", err)
	}
	if no2 != "HR901-2026111200002" {
		t.Errorf("second no = %q, want HR901-2026111200002", no2)
	}

	// sequence is per type code
	no3, err := GenerateBillNo(testCtx, conn, "902", stubDay)
	if err != nil {
		t.Fatalf("GenerateBillNo other type: %v", err)
	}
	if no3 != "HR902-2026111200001" {
		t.Errorf("other type no = %q, want HR902-2026111200001", no3)
	}

	// sequence resets per day
	nextDay := stubDay.AddDate(0, 0, 1)
	no4, err := GenerateBillNo(testCtx, conn, stubTypeCode, nextDay)
	if err != nil {
		t.Fatalf("GenerateBillNo next day: %v", err)
	}
	if no4 != "HR901-2026111300001" {
		t.Errorf("next day no = %q, want HR901-2026111300001", no4)
	}

	// sequence keeps counting within the same day after a day switch
	no5, err := GenerateBillNo(testCtx, conn, stubTypeCode, stubDay)
	if err != nil {
		t.Fatalf("GenerateBillNo back to first day: %v", err)
	}
	if no5 != "HR901-2026111200003" {
		t.Errorf("back to first day no = %q, want HR901-2026111200003", no5)
	}

	// invalid type code is rejected
	if _, err := GenerateBillNo(testCtx, conn, "20", stubDay); !errors.Is(err, ErrInvalidTypeCode) {
		t.Errorf("invalid type code error = %v, want ErrInvalidTypeCode", err)
	}
}

func TestSubmitApproveLifecycle(t *testing.T) {
	conn := newTestConn(t)
	const billUuid = "bill-uuid-1"
	mustCreateStubBill(t, conn, billUuid, stubBillNo)

	// submit
	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	status, writeBack := stubStatus(t, conn, billUuid)
	if status != StatusApproving {
		t.Errorf("after submit status = %q, want %q", status, StatusApproving)
	}
	if writeBack != "" {
		t.Errorf("write-back must not run on submit, got %q", writeBack)
	}
	records := mustRecords(t, conn, stubTypeCode, billUuid)
	if len(records) != 1 {
		t.Fatalf("records after submit = %d, want 1", len(records))
	}
	r := records[0]
	if r.Action != ActionSubmit || r.FromStatus != StatusUnsubmitted || r.ToStatus != StatusApproving {
		t.Errorf("submit record = %+v, want submit unsubmitted->approving", r)
	}
	if r.BillNo != stubBillNo || r.OperatorUserUuid != stubOp.UserUuid || r.OperatorName != stubOp.Name {
		t.Errorf("submit record audit fields = %+v", r)
	}

	// approve (with opinion)
	if err := ApproveBill(testCtx, conn, stubTypeCode, billUuid, stubOp, "同意"); err != nil {
		t.Fatalf("ApproveBill: %v", err)
	}
	status, writeBack = stubStatus(t, conn, billUuid)
	if status != StatusApproved {
		t.Errorf("after approve status = %q, want %q", status, StatusApproved)
	}
	if writeBack != "done" {
		t.Errorf("OnApproved write-back = %q, want done", writeBack)
	}
	records = mustRecords(t, conn, stubTypeCode, billUuid)
	if len(records) != 2 {
		t.Fatalf("records after approve = %d, want 2", len(records))
	}
	last := records[1]
	if last.Action != ActionApprove || last.FromStatus != StatusApproving || last.ToStatus != StatusApproved {
		t.Errorf("approve record = %+v, want approve approving->approved", last)
	}
	if last.Opinion != "同意" {
		t.Errorf("approve opinion = %q, want 同意", last.Opinion)
	}
}

func TestRejectAndResubmit(t *testing.T) {
	conn := newTestConn(t)
	const billUuid = "bill-uuid-2"
	mustCreateStubBill(t, conn, billUuid, stubBillNo)

	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := RejectBill(testCtx, conn, stubTypeCode, billUuid, stubOp, "材料不全"); err != nil {
		t.Fatalf("RejectBill: %v", err)
	}
	status, _ := stubStatus(t, conn, billUuid)
	if status != StatusRejected {
		t.Errorf("after reject status = %q, want %q", status, StatusRejected)
	}

	// 驳回后可修改再提交
	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("resubmit after reject: %v", err)
	}
	status, _ = stubStatus(t, conn, billUuid)
	if status != StatusApproving {
		t.Errorf("after resubmit status = %q, want %q", status, StatusApproving)
	}

	want := []string{"submit:unsubmitted->approving", "reject:approving->rejected", "submit:rejected->approving"}
	if got := recordSummary(mustRecords(t, conn, stubTypeCode, billUuid)); !equalStrings(got, want) {
		t.Errorf("records = %v, want %v", got, want)
	}
}

func TestRevokeAndResubmit(t *testing.T) {
	conn := newTestConn(t)
	const billUuid = "bill-uuid-3"
	mustCreateStubBill(t, conn, billUuid, stubBillNo)

	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	if err := RevokeBill(testCtx, conn, stubTypeCode, billUuid, stubOp, "不想申请了"); err != nil {
		t.Fatalf("RevokeBill: %v", err)
	}
	status, _ := stubStatus(t, conn, billUuid)
	if status != StatusRevoked {
		t.Errorf("after revoke status = %q, want %q", status, StatusRevoked)
	}
	// 已撤销后可重新提交
	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("resubmit after revoke: %v", err)
	}
	status, _ = stubStatus(t, conn, billUuid)
	if status != StatusApproving {
		t.Errorf("after resubmit status = %q, want %q", status, StatusApproving)
	}
}

func TestIllegalTransitions(t *testing.T) {
	conn := newTestConn(t)
	const billUuid = "bill-uuid-4"
	mustCreateStubBill(t, conn, billUuid, stubBillNo)

	// approve before submit
	if err := ApproveBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("approve on unsubmitted error = %v, want ErrInvalidTransition", err)
	}
	// double submit
	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if err := SubmitBill(testCtx, conn, stubTypeCode, billUuid, stubOp, ""); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("double submit error = %v, want ErrInvalidTransition", err)
	}

	// revoke on a separate 未提交 bill
	const draftUuid = "bill-uuid-4b"
	mustCreateStubBill(t, conn, draftUuid, stubBillNo)
	if err := RevokeBill(testCtx, conn, stubTypeCode, draftUuid, stubOp, ""); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("revoke on unsubmitted error = %v, want ErrInvalidTransition", err)
	}

	// failed actions must not change state nor leave records behind
	status, _ := stubStatus(t, conn, billUuid)
	if status != StatusApproving {
		t.Errorf("status after illegal actions = %q, want %q", status, StatusApproving)
	}
	records := mustRecords(t, conn, stubTypeCode, billUuid)
	if len(records) != 1 {
		t.Errorf("records after illegal actions = %d, want 1 (submit only)", len(records))
	}
	records = mustRecords(t, conn, stubTypeCode, draftUuid)
	if len(records) != 0 {
		t.Errorf("records after illegal revoke = %d, want 0", len(records))
	}
}

// TestApproveRollbackOnWriteBackError: the whole lifecycle action is one
// transaction — if the document's own OnApproved write-back fails, the status
// change and the approval record roll back too.
func TestApproveRollbackOnWriteBackError(t *testing.T) {
	conn := newTestConn(t)
	const billUuid = "bill-uuid-5"
	mustCreateStubBill(t, conn, billUuid, stubBillNo)

	if err := SubmitBill(testCtx, conn, failingTypeCode, billUuid, stubOp, ""); err != nil {
		t.Fatalf("SubmitBill: %v", err)
	}
	err := ApproveBill(testCtx, conn, failingTypeCode, billUuid, stubOp, "")
	if err == nil || errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("ApproveBill error = %v, want the write-back failure", err)
	}

	status, writeBack := stubStatus(t, conn, billUuid)
	if status != StatusApproving {
		t.Errorf("status after failed approve = %q, want rolled back %q", status, StatusApproving)
	}
	if writeBack != "" {
		t.Errorf("write-back after failed approve = %q, want empty", writeBack)
	}
	records := mustRecords(t, conn, failingTypeCode, billUuid)
	if len(records) != 1 {
		t.Errorf("records after failed approve = %d, want 1 (submit only)", len(records))
	}
}

func TestBillNotFound(t *testing.T) {
	conn := newTestConn(t)
	err := SubmitBill(testCtx, conn, stubTypeCode, "missing-uuid", stubOp, "")
	if !errors.Is(err, ErrBillNotFound) {
		t.Errorf("SubmitBill unknown uuid error = %v, want ErrBillNotFound", err)
	}
}

func TestTypeNotRegistered(t *testing.T) {
	conn := newTestConn(t)
	err := SubmitBill(testCtx, conn, "999", "whatever", stubOp, "")
	if !errors.Is(err, ErrBillTypeNotRegistered) {
		t.Errorf("SubmitBill unknown type error = %v, want ErrBillTypeNotRegistered", err)
	}
}

func TestRegisterTypeDuplicatePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("duplicate RegisterType did not panic")
		}
	}()
	RegisterType(stubTypeCode, Handler{})
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
