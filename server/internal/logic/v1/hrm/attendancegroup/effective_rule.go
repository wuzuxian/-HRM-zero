package attendancegroup

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
)

type EffectiveRule struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEffectiveRule(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *EffectiveRule {
	return &EffectiveRule{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// EffectiveRule 生效规则查询: 按 员工uuid/工号 定位员工后走三级回退
// （考勤组 > 本部门规则 > 最近上级组织规则 > 无规则兜底, 见 resolve.go）.
func (l *EffectiveRule) EffectiveRule(req *types.GetEffectiveRuleRequest) (resp *types.GetEffectiveRuleResponse, err error) {
	if req.EmployeeUuid == "" && req.EmployeeNo == "" {
		return nil, status.ErrorMessage(errcodes.AttendanceGroupInvalidCode, "必须提供员工 uuid 或工号")
	}

	chain := condition.NewChain()
	if req.EmployeeUuid != "" {
		chain = chain.Equal(hrm_employee.Uuid, req.EmployeeUuid)
	} else {
		chain = chain.Equal(hrm_employee.EmployeeNo, req.EmployeeNo)
	}
	row, err := l.svcCtx.Model.HrmEmployee.FindOneByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		if err == hrm_employee.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EmployeeNotFoundCode, "员工不存在")
		}
		return nil, err
	}

	resolved, err := ResolveEffectiveRule(l.ctx, l.svcCtx, row)
	if err != nil {
		return nil, err
	}

	resp = &types.GetEffectiveRuleResponse{
		Found:           resolved.Found,
		Source:          resolved.Source,
		SourceUuid:      resolved.SourceUuid,
		SourceName:      resolved.SourceName,
		GroupType:       resolved.GroupType,
		StartTime:       resolved.StartTime,
		EndTime:         resolved.EndTime,
		WorkDays:        resolved.WorkDays,
		FlexibleMinutes: resolved.FlexibleMinutes,
	}
	if !resolved.Found {
		// 兜底语义: 全空返回 found=false 且不报错（页面显示 未设置规则）
		return resp, nil
	}
	return resp, nil
}
