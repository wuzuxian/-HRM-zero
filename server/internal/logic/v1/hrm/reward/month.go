package reward

import (
	"strings"
	"time"
)

// monthOf derives the 入账薪资年月 (YYYY-MM) from a 发生日期 (YYYY-MM-DD).
// Invalid input yields "" (caller keeps the bill addressable via a later fix).
func monthOf(date string) string {
	if date == "" {
		return ""
	}
	if t, err := time.Parse(time.DateOnly, date); err == nil {
		return t.Format("2006-01")
	}
	// tolerate a already-YYYY-MM value or unexpected shape without panicking
	if len(date) >= 7 && date[4] == '-' {
		return date[:7]
	}
	return ""
}

// normalizeSalaryMonth defaults an empty 入账薪资年月 to the 发生日期's month
// at creation time so the list shows the effective month immediately; the
// approval write-back re-derives it as a safety net.
func normalizeSalaryMonth(salaryMonth, occurDate string) string {
	if salaryMonth != "" {
		return salaryMonth
	}
	return monthOf(strings.TrimSpace(occurDate))
}
