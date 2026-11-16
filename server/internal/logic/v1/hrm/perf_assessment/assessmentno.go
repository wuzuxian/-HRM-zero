package perf_assessment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_bill_sequence"
)

// 考核单号 PA-年-流水（demo-inventory 9.2：考核单号；GLOSSARY 单据编号独立编号体系）。
// 复用共享 hrm_bill_sequence 存储（迁移 11，bill_type_code='PA'，bill_date=批次周期
// 开始年份 — 同 PERF 批次编号模式），select-then-upsert 语义与 PERF 一致。
// 调用方须在任务插入事务内调用（传入 session）。

const (
	// assessmentNoPrefix is the 考核单号 prefix.
	assessmentNoPrefix = "PA"
	// assessmentNoSeqMax is the 4-digit 流水 upper bound.
	assessmentNoSeqMax = 9999
)

// FormatAssessmentNo renders the 考核单号 `PA-{year}-{4位流水}` (pure function).
func FormatAssessmentNo(year int, seq int64) (string, error) {
	if year < 1000 || year > 9999 {
		return "", fmt.Errorf("%w: %d", ErrNoSeqOutOfRange, year)
	}
	if seq < 1 || seq > assessmentNoSeqMax {
		return "", fmt.Errorf("%w: %d", ErrNoSeqOutOfRange, seq)
	}
	return fmt.Sprintf("%s-%d-%04d", assessmentNoPrefix, year, seq), nil
}

// GenerateAssessmentNo allocates the next 4-digit 流水 for the year (bill_type_code='PA')
// and returns the formatted 考核单号. Must run inside the insert transaction.
func GenerateAssessmentNo(ctx context.Context, session sqlx.Session, year int) (string, error) {
	yearStr := fmt.Sprintf("%d", year)
	seqModel := hrm_bill_sequence.NewHrmBillSequenceModel(sqlx.NewSqlConnFromSession(session))

	row, err := seqModel.FindOneByCondition(ctx, session, condition.NewChain().
		Equal(hrm_bill_sequence.BillTypeCode, assessmentNoPrefix).
		Equal(hrm_bill_sequence.BillDate, yearStr).
		Build()...)
	var seq int64
	switch {
	case err == hrm_bill_sequence.ErrNotFound:
		seq = 1
		if err := seqModel.InsertV2(ctx, session, &hrm_bill_sequence.HrmBillSequence{
			Uuid:         uuid.New().String(),
			BillTypeCode: assessmentNoPrefix,
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
	return FormatAssessmentNo(year, seq)
}

// assessmentYear extracts the 编号年份 from a 周期开始 date (YYYY-MM-DD).
func assessmentYear(periodStart string) (int, error) {
	day, err := time.ParseInLocation("2006-01-02", periodStart, time.Local)
	if err != nil {
		return 0, fmt.Errorf("%w: 周期开始日期非法 %q", ErrAssessmentInvalid, periodStart)
	}
	return day.Year(), nil
}
