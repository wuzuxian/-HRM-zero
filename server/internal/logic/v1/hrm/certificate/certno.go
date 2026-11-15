package certificate

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_bill_sequence"
)

// certNoSeqMax is the 6-digit 流水 upper bound (1..999999, matching the demo
// format JN-2026-000001 in 迁移 15).
const certNoSeqMax = 999999

// ErrCertNoSeqOutOfRange: 6 位流水超出 1..999999.
var ErrCertNoSeqOutOfRange = fmt.Errorf("证书号流水超出范围")

// FormatCertNo renders the 证书号 `{前缀}-{年}-{6位流水}` (GLOSSARY 单据编号),
// e.g. JN-2026-000001. Pure function: the sequence allocation itself lives in
// GenerateCertNo.
func FormatCertNo(prefix string, year int, seq int64) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("%w: empty prefix", ErrCertNoSeqOutOfRange)
	}
	if year < 1000 || year > 9999 {
		return "", fmt.Errorf("%w: %d", ErrCertNoSeqOutOfRange, year)
	}
	if seq < 1 || seq > certNoSeqMax {
		return "", fmt.Errorf("%w: %d", ErrCertNoSeqOutOfRange, seq)
	}
	return fmt.Sprintf("%s-%d-%06d", prefix, year, seq), nil
}

// GenerateCertNo allocates the next 6-digit 流水 for the template prefix + year
// and returns the formatted 证书号 前缀-年-流水. The sequence is per 前缀+年 (the
// template prefix is the 流水 segment; per-year reset mirrors PX 班次编号). It
// reuses the shared hrm_bill_sequence storage (迁移 11) with the type code =
// 编号前缀, so the allocation semantics (select-then-upsert) stay identical to
// GenerateSessionNo / GenerateBillNo.
//
// Call it inside the certificate insert transaction (pass the tx session) so
// 编号 and 证书 row are created atomically.
func GenerateCertNo(ctx context.Context, session sqlx.Session, prefix string, year int) (string, error) {
	yearStr := fmt.Sprintf("%d", year)
	seqModel := hrm_bill_sequence.NewHrmBillSequenceModel(sqlx.NewSqlConnFromSession(session))

	row, err := seqModel.FindOneByCondition(ctx, session, condition.NewChain().
		Equal(hrm_bill_sequence.BillTypeCode, prefix).
		Equal(hrm_bill_sequence.BillDate, yearStr).
		Build()...)
	var seq int64
	switch {
	case err == hrm_bill_sequence.ErrNotFound:
		seq = 1
		if err := seqModel.InsertV2(ctx, session, &hrm_bill_sequence.HrmBillSequence{
			Uuid:         uuid.New().String(),
			BillTypeCode: prefix,
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
	return FormatCertNo(prefix, year, seq)
}
