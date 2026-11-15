package contract

import (
	"errors"
	"testing"
	"time"
)

// TestDaysUntilDate 距到期天数边界：今天到期=0、明天=1、昨天=-1、跨月/跨年。
func TestDaysUntilDate(t *testing.T) {
	today := time.Date(2026, 11, 16, 15, 30, 0, 0, time.Local) // 带时分，验证按日截断
	cases := []struct {
		date string
		want int64
	}{
		{"2026-11-16", 0},  // 今天到期
		{"2026-11-17", 1},  // 明天
		{"2026-11-15", -1}, // 昨天已过期
		{"2026-12-31", 45},
		{"2027-11-16", 365},
		{"2026-02-28", -261}, // 今年年初（闰年 2026? 2026 非闰年，11-16 回到 02-28 = -261）
	}
	for _, c := range cases {
		got, err := DaysUntilDate(today, c.date)
		if err != nil {
			t.Fatalf("DaysUntilDate(%s): %v", c.date, err)
		}
		if got != c.want {
			t.Errorf("DaysUntilDate(%s) = %d, want %d", c.date, got, c.want)
		}
	}

	if _, err := DaysUntilDate(today, "2026/11/16"); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("bad format should return ErrInvalidDate, got %v", err)
	}
	if _, err := DaysUntilDate(today, ""); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("empty date should return ErrInvalidDate, got %v", err)
	}
}

// TestExpiryStateByDays 预警口径边界：warning 段含 0 与 90，>90 为 normal，<0 为 expired。
func TestExpiryStateByDays(t *testing.T) {
	cases := []struct {
		days int64
		want string
	}{
		{-1, ExpiryStateExpired},
		{-365, ExpiryStateExpired},
		{0, ExpiryStateWarning},
		{1, ExpiryStateWarning},
		{30, ExpiryStateWarning},
		{60, ExpiryStateWarning},
		{90, ExpiryStateWarning},
		{91, ExpiryStateNormal},
		{3650, ExpiryStateNormal},
	}
	for _, c := range cases {
		if got := ExpiryStateByDays(c.days); got != c.want {
			t.Errorf("ExpiryStateByDays(%d) = %q, want %q", c.days, got, c.want)
		}
	}
}

// TestContractExpiryState 已解除合同恒为 terminated，其余按天数映射。
func TestContractExpiryState(t *testing.T) {
	if got := ContractExpiryState(ContractStatusTerminated, 10); got != ExpiryStateTerminated {
		t.Errorf("terminated contract state = %q, want terminated", got)
	}
	if got := ContractExpiryState(ContractStatusActive, -1); got != ExpiryStateExpired {
		t.Errorf("active expired state = %q, want expired", got)
	}
}

// TestNextExpiryDate 续签新到期日：缺省=原到期日+1 年；指定则原样；闰日回退。
func TestNextExpiryDate(t *testing.T) {
	got, err := NextExpiryDate("2026-11-16", "")
	if err != nil || got != "2027-11-16" {
		t.Errorf("default renew = %q err=%v, want 2027-11-16", got, err)
	}
	got, err = NextExpiryDate("2026-11-16", "2029-06-30")
	if err != nil || got != "2029-06-30" {
		t.Errorf("specified renew = %q err=%v, want 2029-06-30", got, err)
	}
	got, err = NextExpiryDate("2028-02-29", "") // 闰日 + 1 年 → 2029-03-01
	if err != nil || got != "2029-03-01" {
		t.Errorf("leap-day renew = %q err=%v, want 2029-03-01", got, err)
	}
	if _, err := NextExpiryDate("bad", ""); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("bad current end date should error, got %v", err)
	}
	if _, err := NextExpiryDate("2026-11-16", "06-30"); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("bad specified date should error, got %v", err)
	}
}

// TestMaskCertNo 证号脱敏：保留前 2 后 2，等长打码；短证号只留首 1 位。
func TestMaskCertNo(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"123456", "1*****"},
		{"1234567", "12***67"},
		{"440823199001011234", "44**************34"},
		{"AB12", "A***"},
	}
	for _, c := range cases {
		if got := MaskCertNo(c.in); got != c.want {
			t.Errorf("MaskCertNo(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// 打码后长度不变
	if got := MaskCertNo("440823199001011234"); len([]rune(got)) != 18 {
		t.Errorf("masked length = %d, want 18", len([]rune(got)))
	}
}
