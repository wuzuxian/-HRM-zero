package bill

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_approval_record"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_bill_sequence"
)

// Sentinel errors returned by the shared lifecycle. HTTP handlers map them to
// business error codes (see internal/errcodes/hrm.go).
var (
	// ErrBillTypeNotRegistered: 单据类型未注册审批处理器。
	ErrBillTypeNotRegistered = errors.New("单据类型未注册审批处理器")
	// ErrBillNotFound: 单据不存在。
	ErrBillNotFound = errors.New("单据不存在")
	// ErrInvalidTransition: 当前单据状态不允许该操作（单据状态机拒绝）。
	ErrInvalidTransition = errors.New("当前单据状态不允许该操作")
)

// Operator is who performed a lifecycle action; the name is snapshotted into
// the approval record for audit.
type Operator struct {
	UserUuid string
	Name     string
}

// Handler is how a document type (单据类型) plugs into the shared approval
// lifecycle. Document logic packages register one under their 类型码 (e.g. "201")
// in an init function; all closures receive the transaction session so the
// document row and the approval record change atomically.
type Handler struct {
	// Load returns the document's 单据编号 and current 单据状态.
	// Return sqlx.ErrNotFound when the document does not exist.
	Load func(ctx context.Context, session sqlx.Session, billUuid string) (billNo, status string, err error)
	// SetStatus persists the new 单据状态 on the document row.
	SetStatus func(ctx context.Context, session sqlx.Session, billUuid, status string) error
	// OnApproved, when set, runs inside the same transaction right after an
	// approval is recorded — this is where the document's own business
	// write-back lives (入职建档、转正、假期余额扣减…, ADR-0001).
	OnApproved func(ctx context.Context, session sqlx.Session, billUuid string) error
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Handler)
)

// RegisterType registers a document type under its 3-digit 类型码 (e.g. "201").
// Panics on duplicate registration — a programming error caught at startup.
func RegisterType(typeCode string, h Handler) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[typeCode]; dup {
		panic(fmt.Sprintf("hrm/bill: duplicate registration for type code %q", typeCode))
	}
	registry[typeCode] = h
}

func lookupHandler(typeCode string) (Handler, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	h, ok := registry[typeCode]
	return h, ok
}

// Record is the audit trail entry of one lifecycle action (审批记录).
type Record struct {
	BillTypeCode     string    `json:"billTypeCode"`
	BillUuid         string    `json:"billUuid"`
	BillNo           string    `json:"billNo"`
	Action           Action    `json:"action"`
	FromStatus       string    `json:"fromStatus"`
	ToStatus         string    `json:"toStatus"`
	OperatorUserUuid string    `json:"operatorUserUuid"`
	OperatorName     string    `json:"operatorName"`
	Opinion          string    `json:"opinion"`
	CreateTime       time.Time `json:"createTime"`
}

// Records returns the approval trail of one 单据, oldest first.
func Records(ctx context.Context, conn sqlx.SqlConn, typeCode, billUuid string) ([]*Record, error) {
	rows, err := hrm_approval_record.NewHrmApprovalRecordModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_approval_record.BillTypeCode, typeCode).
			Equal(hrm_approval_record.BillUuid, billUuid).
			OrderByAsc(hrm_approval_record.Id).
			Build()...)
	if err != nil {
		return nil, err
	}
	records := make([]*Record, 0, len(rows))
	for _, row := range rows {
		records = append(records, &Record{
			BillTypeCode:     row.BillTypeCode,
			BillUuid:         row.BillUuid,
			BillNo:           row.BillNo,
			Action:           Action(row.Action),
			FromStatus:       row.FromStatus,
			ToStatus:         row.ToStatus,
			OperatorUserUuid: row.OperatorUserUuid,
			OperatorName:     row.OperatorName,
			Opinion:          row.Opinion,
			CreateTime:       row.CreateTime,
		})
	}
	return records, nil
}

// SubmitBill 提交单据: 未提交/审批不通过/已撤销 → 审批中.
func SubmitBill(ctx context.Context, conn sqlx.SqlConn, typeCode, billUuid string, op Operator, opinion string) error {
	return doAction(ctx, conn, ActionSubmit, typeCode, billUuid, op, opinion)
}

// ApproveBill 审批通过: 审批中 → 审批通过, then runs the document's own
// OnApproved write-back in the same transaction.
func ApproveBill(ctx context.Context, conn sqlx.SqlConn, typeCode, billUuid string, op Operator, opinion string) error {
	return doAction(ctx, conn, ActionApprove, typeCode, billUuid, op, opinion)
}

// RejectBill 审批驳回: 审批中 → 审批不通过.
func RejectBill(ctx context.Context, conn sqlx.SqlConn, typeCode, billUuid string, op Operator, opinion string) error {
	return doAction(ctx, conn, ActionReject, typeCode, billUuid, op, opinion)
}

// RevokeBill 撤销: 审批中 → 已撤销.
func RevokeBill(ctx context.Context, conn sqlx.SqlConn, typeCode, billUuid string, op Operator, opinion string) error {
	return doAction(ctx, conn, ActionRevoke, typeCode, billUuid, op, opinion)
}

// doAction runs one lifecycle action as a single transaction:
// load → state-machine check → SetStatus → (write-back on approve) → record.
func doAction(ctx context.Context, conn sqlx.SqlConn, action Action, typeCode, billUuid string, op Operator, opinion string) error {
	h, ok := lookupHandler(typeCode)
	if !ok {
		return fmt.Errorf("%w: %s", ErrBillTypeNotRegistered, typeCode)
	}
	return conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		billNo, from, err := h.Load(ctx, session, billUuid)
		if err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				return ErrBillNotFound
			}
			return err
		}
		to, ok := NextStatus(from, action)
		if !ok {
			return fmt.Errorf("%w: %s + %s", ErrInvalidTransition, from, action)
		}
		if err := h.SetStatus(ctx, session, billUuid, to); err != nil {
			return err
		}
		if action == ActionApprove && h.OnApproved != nil {
			if err := h.OnApproved(ctx, session, billUuid); err != nil {
				return err
			}
		}
		return insertRecord(ctx, session, typeCode, billUuid, billNo, action, from, to, op, opinion)
	})
}

func insertRecord(ctx context.Context, session sqlx.Session, typeCode, billUuid, billNo string, action Action, from, to string, op Operator, opinion string) error {
	record := &hrm_approval_record.HrmApprovalRecord{
		Uuid:             uuid.New().String(),
		BillTypeCode:     typeCode,
		BillUuid:         billUuid,
		BillNo:           billNo,
		Action:           string(action),
		FromStatus:       from,
		ToStatus:         to,
		OperatorUserUuid: op.UserUuid,
		OperatorName:     op.Name,
		Opinion:          opinion,
	}
	return hrm_approval_record.NewHrmApprovalRecordModel(
		sqlx.NewSqlConnFromSession(session)).InsertV2(ctx, session, record)
}

// GenerateBillNo allocates the next 5-digit daily 流水 for typeCode and returns
// the formatted 单据编号 `HR{类型码}-{YYYYMMDD}{5位流水}`.
//
// Call it inside the document's own insert transaction (pass the tx session)
// so 编号 and 单据 row are created atomically. The sequence is per
// (类型码, 日期) and resets every day.
func GenerateBillNo(ctx context.Context, session sqlx.Session, typeCode string, day time.Time) (string, error) {
	if !IsValidTypeCode(typeCode) {
		return "", fmt.Errorf("%w: %q", ErrInvalidTypeCode, typeCode)
	}
	ymd := day.Format(BillNoLayout)
	seqModel := hrm_bill_sequence.NewHrmBillSequenceModel(sqlx.NewSqlConnFromSession(session))

	row, err := seqModel.FindOneByCondition(ctx, session, condition.NewChain().
		Equal(hrm_bill_sequence.BillTypeCode, typeCode).
		Equal(hrm_bill_sequence.BillDate, ymd).
		Build()...)
	var seq int64
	switch {
	case errors.Is(err, hrm_bill_sequence.ErrNotFound):
		seq = 1
		if err := seqModel.InsertV2(ctx, session, &hrm_bill_sequence.HrmBillSequence{
			Uuid:         uuid.New().String(),
			BillTypeCode: typeCode,
			BillDate:     ymd,
			LastSeq:      seq,
		}); err != nil {
			return "", err
		}
	case err != nil:
		return "", err
	default:
		seq = row.LastSeq + 1
		if err := seqModel.UpdateFieldsByCondition(ctx, session,
			map[string]any{"last_seq": seq},
			condition.NewChain().Equal(hrm_bill_sequence.Id, row.Id).Build()...); err != nil {
			return "", err
		}
	}
	return FormatBillNo(typeCode, day, seq)
}
