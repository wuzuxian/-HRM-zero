package scheduleboard

import (
	"context"

	"github.com/google/uuid"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// newUuid uuid v4.
func newUuid() string {
	return uuid.New().String()
}

// expandMembers 解析考勤组适用范围成员（部门含下级 BFS + 指定人员，去重），
// 与 attendancegroup.Members / expandGroupMembers 同口径（hrm-12-notes 契约）。
func expandMembers(ctx context.Context, svcCtx *svc.ServiceContext, group *hrm_attendance_group.HrmAttendanceGroup, seen map[string]bool) ([]string, error) {
	var members []string

	for _, orgUuid := range splitCsv(group.OrgUuids) {
		orgUuids, err := findAllSubOrgUuids(ctx, svcCtx, orgUuid)
		if err != nil {
			return nil, err
		}
		for _, u := range orgUuids {
			rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
				condition.NewChain().Equal(hrm_employee.OrgUuid, u).Build()...)
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				if !seen[row.Uuid] {
					seen[row.Uuid] = true
					members = append(members, row.Uuid)
				}
			}
		}
	}

	for _, memberUuid := range splitCsv(group.MemberUuids) {
		if !seen[memberUuid] {
			seen[memberUuid] = true
			members = append(members, memberUuid)
		}
	}
	return members, nil
}

// findAllSubOrgUuids 组织含下级展开（带环保护），与 attendancegroup.resolve.go 同语义。
func findAllSubOrgUuids(ctx context.Context, svcCtx *svc.ServiceContext, rootUuid string) ([]string, error) {
	visited := make(map[string]bool)
	queue := []string{rootUuid}
	out := []string{rootUuid}
	visited[rootUuid] = true
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		children, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil,
			condition.NewChain().Equal(hrm_org.ParentUuid, cur).Build()...)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			if visited[child.Uuid] {
				continue
			}
			visited[child.Uuid] = true
			out = append(out, child.Uuid)
			queue = append(queue, child.Uuid)
		}
	}
	return out, nil
}
