package attendancegroup

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_group"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
)

type Members struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewMembers(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Members {
	return &Members{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Members 生效人数查询: 解析考勤组适用范围（部门含下级 + 指定人员, 去重）,
// 指定 uuid 时查单个考勤组, 缺省时返回所有启用考勤组的并集.
func (l *Members) Members(req *types.MemberRequest) (resp *types.MemberResponse, err error) {
	groups, err := l.svcCtx.Model.HrmAttendanceGroup.FindByCondition(l.ctx, nil,
		newEnabledGroupChain(req.Uuid).Build()...)
	if err != nil {
		return nil, err
	}
	if req.Uuid != "" && len(groups) == 0 {
		return nil, status.ErrorMessage(errcodes.AttendanceGroupNotFoundCode, "考勤组不存在")
	}

	seen := make(map[string]bool)
	records := make([]types.EffectiveGroupMember, 0, 16)
	for _, group := range groups {
		memberUuids, err := expandGroupMembers(l.ctx, l.svcCtx, group)
		if err != nil {
			return nil, err
		}
		for _, memberUuid := range memberUuids {
			if seen[memberUuid] {
				continue
			}
			seen[memberUuid] = true
			employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, memberUuid)
			if err != nil {
				if err == sqlx.ErrNotFound {
					continue // 指定人员已被删除, 跳过
				}
				return nil, err
			}
			orgName := ""
			if employee.OrgUuid != "" {
				org, err := l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, employee.OrgUuid)
				if err != nil {
					if err == hrm_org.ErrNotFound {
						org = nil
					} else {
						return nil, err
					}
				}
				if org != nil {
					orgName = org.OrgName
				}
			}
			records = append(records, types.EffectiveGroupMember{
				Uuid:       employee.Uuid,
				EmployeeNo: employee.EmployeeNo,
				Name:       employee.Name,
				OrgName:    orgName,
			})
		}
	}

	return &types.MemberResponse{Total: int64(len(records)), Members: records}, nil
}

// newEnabledGroupChain builds the condition for Members: enabled groups,
// optionally scoped to one uuid.
func newEnabledGroupChain(uuid string) condition.Chain {
	chain := condition.NewChain().Equal(hrm_attendance_group.Status, statusEnabled)
	if uuid != "" {
		chain = chain.Equal(hrm_attendance_group.Uuid, uuid)
	}
	return chain.OrderByAsc(hrm_attendance_group.Id)
}
