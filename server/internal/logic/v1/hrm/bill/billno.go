package bill

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidTypeCode: 类型码必须是 3 位数字（如 201）。
	ErrInvalidTypeCode = errors.New("单据类型码必须是 3 位数字")
	// ErrBillNoSeqOutOfRange: 5 位流水超出 1..99999。
	ErrBillNoSeqOutOfRange = errors.New("单据编号流水超出范围")
)

// billNoTypeCodeLen is the fixed 类型码 width (HR201 ~ HR217).
const billNoTypeCodeLen = 3

// BillNoLayout is the date layout inside 单据编号 (YYYYMMDD).
const BillNoLayout = "20060102"

// FormatBillNo renders the shared 单据编号 format `HR{类型码}-{YYYYMMDD}{5位流水}`
// (GLOSSARY 单据编号), e.g. HR201-2026111200001.
//
// Pure function: the sequence allocation itself lives in GenerateBillNo.
// 绩效批次（PERF-）、薪资批次（SA）、培训班次（PX）有各自独立的编号体系，
// 不走本格式。
func FormatBillNo(typeCode string, day time.Time, seq int64) (string, error) {
	if len(typeCode) != billNoTypeCodeLen {
		return "", fmt.Errorf("%w: %q", ErrInvalidTypeCode, typeCode)
	}
	for _, r := range typeCode {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("%w: %q", ErrInvalidTypeCode, typeCode)
		}
	}
	if seq < 1 || seq > 99999 {
		return "", fmt.Errorf("%w: %d", ErrBillNoSeqOutOfRange, seq)
	}
	return fmt.Sprintf("HR%s-%s%05d", typeCode, day.Format(BillNoLayout), seq), nil
}

// IsValidTypeCode reports whether s is a valid 3-digit 类型码.
func IsValidTypeCode(s string) bool {
	_, err := FormatBillNo(s, time.Unix(0, 0).UTC(), 1)
	return !errors.Is(err, ErrInvalidTypeCode)
}
