package org

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// countEmployeesByOrg counts employees per 所属组织 by grouping hrm_employee on
// org_uuid (员工档案, HRM-03), 在册口径: person_status NOT IN ('5','6') —
// 即不含 离职('5')/退休('6') (见 hrm-03-notes.md 的人员状态枚举).
func countEmployeesByOrg(ctx context.Context, conn sqlx.SqlConn, orgUuids []string) (map[string]int64, error) {
	out := make(map[string]int64, len(orgUuids))
	if len(orgUuids) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(orgUuids))
	args := make([]any, 0, len(orgUuids))
	for i, uuid := range orgUuids {
		placeholders[i] = "?"
		args = append(args, uuid)
	}
	query := "SELECT org_uuid, COUNT(*) AS cnt FROM hrm_employee WHERE person_status NOT IN ('5','6') AND org_uuid IN (" +
		strings.Join(placeholders, ",") + ") GROUP BY org_uuid"

	var rows []struct {
		OrgUuid string
		Cnt     int64
	}
	if err := conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if isMissingTable(err) {
			// 员工档案未落地：按 0 人处理，组织架构图仍可用
			return out, nil
		}
		return nil, err
	}
	for _, row := range rows {
		out[row.OrgUuid] = row.Cnt
	}
	return out, nil
}

// isMissingTable matches the "table not found" error text of the three
// supported dialects (sqlite / mysql / pgx) without importing their drivers.
func isMissingTable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "no such table") || // sqlite
		strings.Contains(msg, "doesn't exist") || // mysql 1146
		strings.Contains(msg, "does not exist") // pgx 42P01
}
