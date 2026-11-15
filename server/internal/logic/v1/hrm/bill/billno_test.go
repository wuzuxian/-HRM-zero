package bill

import (
	"errors"
	"testing"
	"time"
)

// TestFormatBillNo specifies the 单据编号 format HR{类型码}-{YYYYMMDD}{5位流水}
// (GLOSSARY 单据编号), e.g. HR201-2026111200001.
func TestFormatBillNo(t *testing.T) {
	day := time.Date(2026, 11, 12, 15, 4, 5, 0, time.Local)

	tests := []struct {
		name     string
		typeCode string
		day      time.Time
		seq      int64
		want     string
		wantErr  bool
	}{
		{"spec example", "201", day, 1, "HR201-2026111200001", false},
		{"mid sequence is zero padded", "215", day, 42, "HR215-2026111200042", false},
		{"max sequence", "217", day, 99999, "HR217-2026111299999", false},
		{"other date", "203", time.Date(2027, 1, 2, 0, 0, 0, 0, time.Local), 7, "HR203-2027010200007", false},
		// seq bounds
		{"seq zero is invalid", "201", day, 0, "", true},
		{"seq negative is invalid", "201", day, -1, "", true},
		{"seq overflow is invalid", "201", day, 100000, "", true},
		// type code bounds
		{"two digit type code is invalid", "20", day, 1, "", true},
		{"four digit type code is invalid", "2011", day, 1, "", true},
		{"alphabetic type code is invalid", "HR2", day, 1, "", true},
		{"mixed type code is invalid", "2a1", day, 1, "", true},
		{"empty type code is invalid", "", day, 1, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatBillNo(tt.typeCode, tt.day, tt.seq)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("FormatBillNo(%q, %v, %d) error = nil, want error", tt.typeCode, tt.day, tt.seq)
				}
				if !errors.Is(err, ErrInvalidTypeCode) && !errors.Is(err, ErrBillNoSeqOutOfRange) {
					t.Fatalf("FormatBillNo error = %v, want ErrInvalidTypeCode or ErrBillNoSeqOutOfRange", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("FormatBillNo(%q, %v, %d) unexpected error: %v", tt.typeCode, tt.day, tt.seq, err)
			}
			if got != tt.want {
				t.Errorf("FormatBillNo(%q, %v, %d) = %q, want %q", tt.typeCode, tt.day, tt.seq, got, tt.want)
			}
		})
	}
}

// TestFormatBillNoSeqErrorKind distinguishes which sentinel error is returned.
func TestFormatBillNoSeqErrorKind(t *testing.T) {
	day := time.Date(2026, 11, 15, 0, 0, 0, 0, time.Local)

	if _, err := FormatBillNo("201", day, 100000); !errors.Is(err, ErrBillNoSeqOutOfRange) {
		t.Errorf("seq 100000 error = %v, want ErrBillNoSeqOutOfRange", err)
	}
	if _, err := FormatBillNo("20", day, 1); !errors.Is(err, ErrInvalidTypeCode) {
		t.Errorf("type code 20 error = %v, want ErrInvalidTypeCode", err)
	}
}
