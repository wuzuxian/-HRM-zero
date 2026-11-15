package certificate

import (
	"database/sql"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
)

// 证书状态 (迁移 15 规范): 0=待发 1=已发 2=已撤销.
const (
	CertStatusPending = "0"
	CertStatusIssued  = "1"
	CertStatusRevoked = "2"
)

// 证书来源: 1=班次结业 2=外部导入.
const (
	CertSourceSession = "1"
	CertSourceImport  = "2"
)

// templateEnabled mirrors cert_template.TemplateStatusEnabled without an import cycle.
const templateEnabled = "1"

// conditionEq is the shared equal-condition helper (single field = value).
func conditionEq(field condition.Field, value string) []condition.Condition {
	return condition.NewChain().Equal(field, value).Build()
}

// conditionUuid is the shared uuid-condition helper.
func conditionUuid(uuid string) []condition.Condition {
	return conditionEq(hrm_training_certificate.Uuid, uuid)
}

// nullTime wraps a time into sql.NullTime (zero time → invalid).
func nullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

// nullString wraps a string into sql.NullString (empty → invalid).
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
