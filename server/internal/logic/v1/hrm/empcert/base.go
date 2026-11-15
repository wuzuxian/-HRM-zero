package empcert

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// findCertByUuid 按主键 uuid 定位证照行；不存在 → EmployeeCertificateNotFoundCode。
func findCertByUuid(ctx context.Context, svcCtx *svc.ServiceContext, certUuid string) (*hrm_employee_certificate.HrmEmployeeCertificate, error) {
	row, err := svcCtx.Model.HrmEmployeeCertificate.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee_certificate.Uuid, certUuid).Build()...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EmployeeCertificateNotFoundCode, "证照不存在")
		}
		return nil, err
	}
	return row, nil
}

// HRM-05 员工证照纯函数/常量。预警口径与劳动合同一致：
// expired(<0) / warning(0..90 天) / normal(>90 天)。证号出参统一脱敏。
const (
	WarningWindowDays = 90

	ExpiryStateNormal  = "normal"
	ExpiryStateWarning = "warning"
	ExpiryStateExpired = "expired"

	dateFormat = "2006-01-02"
)

// ErrInvalidDate 日期格式非法（YYYY-MM-DD）。
var ErrInvalidDate = errors.New("invalid date, expect YYYY-MM-DD")

// uuidNew 证照行 uuid。
func uuidNew() string {
	return uuid.New().String()
}

// DaysUntilDate 计算今天到目标日期的天数（目标-今天；负数=已过期）。
func DaysUntilDate(today time.Time, date string) (int64, error) {
	target, err := time.ParseInLocation(dateFormat, date, time.Local)
	if err != nil {
		return 0, ErrInvalidDate
	}
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	return int64(target.Sub(today).Hours() / 24), nil
}

// ExpiryStateByDays 将距到期天数映射为预警状态。
func ExpiryStateByDays(days int64) string {
	switch {
	case days < 0:
		return ExpiryStateExpired
	case days <= WarningWindowDays:
		return ExpiryStateWarning
	default:
		return ExpiryStateNormal
	}
}

// MaskCertNo 证号脱敏：保留前 2 位与后 2 位，中间以等长 * 打码；
// 长度 <= 6 的短证号只保留首 1 位；空串返回空串。
func MaskCertNo(no string) string {
	runes := []rune(no)
	n := len(runes)
	if n == 0 {
		return ""
	}
	if n <= 6 {
		out := []rune(no)
		for i := 1; i < n; i++ {
			out[i] = '*'
		}
		return string(out)
	}
	out := make([]rune, n)
	for i := 2; i < n-2; i++ {
		out[i] = '*'
	}
	out[0], out[1] = runes[0], runes[1]
	out[n-2], out[n-1] = runes[n-2], runes[n-1]
	return string(out)
}

// validDate 校验 YYYY-MM-DD 格式。
func validDate(v string) bool {
	_, err := time.ParseInLocation(dateFormat, v, time.Local)
	return err == nil
}
