package signin

import (
	"database/sql"
	"time"
)

// formatNullTime renders sql.NullTime as YYYY-MM-DD HH:mm:ss (empty when NULL).
func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.DateTime)
}
