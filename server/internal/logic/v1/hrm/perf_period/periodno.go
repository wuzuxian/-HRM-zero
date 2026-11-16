package perf_period

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_bill_sequence"
)

// periodNoPrefix is the 批次编号 prefix (GLOSSARY 单据编号: PERF 独立编号体系).
const periodNoPrefix = "PERF"

// periodNoSeqMax is the 4-digit 流水 upper bound (1..9999).
const periodNoSeqMax = 9999

// FormatPeriodNo renders the 批次编号 `PERF-{year}-{4位流水}` (GLOSSARY 单据编号),
// e.g. PERF-2026-0001. Pure function: the sequence allocation itself lives in
// GeneratePeriodNo.
func FormatPeriodNo(year int, seq int64) (string, error) {
	if year < 1000 || year > 9999 {
		return "", fmt.Errorf("%w: %d", ErrPeriodNoSeqOutOfRange, year)
	}
	if seq < 1 || seq > periodNoSeqMax {
		return "", fmt.Errorf("%w: %d", ErrPeriodNoSeqOutOfRange, seq)
	}
	return fmt.Sprintf("%s-%d-%04d", periodNoPrefix, year, seq), nil
}

// GeneratePeriodNo allocates the next 4-digit 流水 for the year and returns the
// formatted 批次编号 PERF-年-流水. The sequence is per year (日期=年份, not per
// day — 批次编号 are scarce) and resets every year. It reuses the shared
// hrm_bill_sequence storage (迁移 11) with bill_type_code='PERF' so the
// allocation semantics (select-then-upsert) stay identical to the PX 班次编号.
//
// Call it inside the period's insert transaction (pass the tx session) so 编号
// and 批次 row are created atomically.
func GeneratePeriodNo(ctx context.Context, session sqlx.Session, year int) (string, error) {
	yearStr := fmt.Sprintf("%d", year)
	seqModel := hrm_bill_sequence.NewHrmBillSequenceModel(sqlx.NewSqlConnFromSession(session))

	row, err := seqModel.FindOneByCondition(ctx, session, condition.NewChain().
		Equal(hrm_bill_sequence.BillTypeCode, periodNoPrefix).
		Equal(hrm_bill_sequence.BillDate, yearStr).
		Build()...)
	var seq int64
	switch {
	case err == hrm_bill_sequence.ErrNotFound:
		seq = 1
		if err := seqModel.InsertV2(ctx, session, &hrm_bill_sequence.HrmBillSequence{
			Uuid:         uuid.New().String(),
			BillTypeCode: periodNoPrefix,
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
	return FormatPeriodNo(year, seq)
}

// periodStartYear extracts the 编号年份 from the 周期开始 date (YYYY-MM-DD).
func periodStartYear(periodStart string) (int, error) {
	day, err := time.ParseInLocation(dateLayout, periodStart, time.Local)
	if err != nil {
		return 0, err
	}
	return day.Year(), nil
}
