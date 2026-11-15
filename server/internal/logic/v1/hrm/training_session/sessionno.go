package training_session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_bill_sequence"
)

// sessionNoPrefix is the 班次编号 prefix (GLOSSARY 单据编号: PX 独立编号体系).
const sessionNoPrefix = "PX"

// sessionNoLayout is the year segment layout of the PX 编号 (PX-年-流水).
const sessionNoLayout = "2006"

// sessionNoSeqMax is the 4-digit 流水 upper bound (1..9999).
const sessionNoSeqMax = 9999

// ErrSessionNoSeqOutOfRange: 4 位流水超出 1..9999.
var ErrSessionNoSeqOutOfRange = fmt.Errorf("班次编号流水超出范围")

// FormatSessionNo renders the 班次编号 `PX-{year}-{4位流水}` (GLOSSARY 单据编号),
// e.g. PX-2026-0001. Pure function: the sequence allocation itself lives in
// GenerateSessionNo.
func FormatSessionNo(year int, seq int64) (string, error) {
	if year < 1000 || year > 9999 {
		return "", fmt.Errorf("%w: %d", ErrSessionNoSeqOutOfRange, year)
	}
	if seq < 1 || seq > sessionNoSeqMax {
		return "", fmt.Errorf("%w: %d", ErrSessionNoSeqOutOfRange, seq)
	}
	return fmt.Sprintf("%s-%d-%04d", sessionNoPrefix, year, seq), nil
}

// GenerateSessionNo allocates the next 4-digit 流水 for the year and returns
// the formatted 班次编号 PX-年-流水. The sequence is per year (not per day —
// 班次编号 are scarce, unlike 单据编号) and resets every year. It reuses the
// shared hrm_bill_sequence storage (迁移 11) with the type code "PX" so the
// allocation semantics (select-then-upsert) stay identical to GenerateBillNo.
//
// Call it inside the session's insert transaction (pass the tx session) so
// 编号 and 班次 row are created atomically.
func GenerateSessionNo(ctx context.Context, session sqlx.Session, year int, day time.Time) (string, error) {
	_ = day // year carries the segment; keep the day param for call-site symmetry with bill.GenerateBillNo
	yearStr := fmt.Sprintf("%d", year)
	seqModel := hrm_bill_sequence.NewHrmBillSequenceModel(sqlx.NewSqlConnFromSession(session))

	row, err := seqModel.FindOneByCondition(ctx, session, condition.NewChain().
		Equal(hrm_bill_sequence.BillTypeCode, sessionNoPrefix).
		Equal(hrm_bill_sequence.BillDate, yearStr).
		Build()...)
	var seq int64
	switch {
	case err == hrm_bill_sequence.ErrNotFound:
		seq = 1
		if err := seqModel.InsertV2(ctx, session, &hrm_bill_sequence.HrmBillSequence{
			Uuid:         uuid.New().String(),
			BillTypeCode: sessionNoPrefix,
			BillDate:     yearStr,
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
	return FormatSessionNo(year, seq)
}
