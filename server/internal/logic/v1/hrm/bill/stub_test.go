package bill

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// The stub document type: a minimal 单据 backed by its own table (stub_bill),
// registered under a private 类型码. It exercises the same integration path
// real document logic (HR201 入职单 etc.) will take: Load / SetStatus /
// OnApproved closures over its own table.

const stubTypeCode = "901"

// failingTypeCode behaves like the stub but its OnApproved always fails —
// used to verify transaction atomicity of the approval flow.
const failingTypeCode = "903"

var (
	stubOnce sync.Once
	stubErr  error
)

func registerStubs(t *testing.T) {
	t.Helper()
	stubOnce.Do(func() {
		RegisterType(stubTypeCode, Handler{
			Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
				var row struct {
					BillNo string `db:"bill_no"`
					Status string `db:"status"`
				}
				if err := session.QueryRowCtx(ctx, &row,
					"SELECT bill_no, status FROM stub_bill WHERE uuid = ?", billUuid); err != nil {
					return "", "", err
				}
				return row.BillNo, row.Status, nil
			},
			SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
				_, err := session.ExecCtx(ctx,
					"UPDATE stub_bill SET status = ? WHERE uuid = ?", status, billUuid)
				return err
			},
			OnApproved: func(ctx context.Context, session sqlx.Session, billUuid string) error {
				_, err := session.ExecCtx(ctx,
					"UPDATE stub_bill SET write_back = 'done' WHERE uuid = ?", billUuid)
				return err
			},
		})

		RegisterType(failingTypeCode, Handler{
			Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
				var row struct {
					BillNo string `db:"bill_no"`
					Status string `db:"status"`
				}
				if err := session.QueryRowCtx(ctx, &row,
					"SELECT bill_no, status FROM stub_bill WHERE uuid = ?", billUuid); err != nil {
					return "", "", err
				}
				return row.BillNo, row.Status, nil
			},
			SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
				_, err := session.ExecCtx(ctx,
					"UPDATE stub_bill SET status = ? WHERE uuid = ?", status, billUuid)
				return err
			},
			OnApproved: func(ctx context.Context, session sqlx.Session, billUuid string) error {
				return errors.New("write-back failed")
			},
		})
	})
	if stubErr != nil {
		t.Fatalf("stub registration failed: %v", stubErr)
	}
}
