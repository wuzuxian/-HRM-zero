package user

import (
	"context"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/manage/user"
)

type LinkedEmployees struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLinkedEmployees(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *LinkedEmployees {
	return &LinkedEmployees{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// LinkedEmployees 批量按账号 uuid 反查关联员工 (HRM-51): 返回工号+姓名,
// 供用户管理列表只读「关联员工」列回填。只读展示, 不做任何写入 (ADR 0007:
// 账号↔档案的唯一写入入口在员工侧)。未关联/不存在的账号不出现在结果中。
func (l *LinkedEmployees) LinkedEmployees(req *types.LinkedEmployeesRequest) (resp *types.LinkedEmployeesResponse, err error) {
	uuids := make([]any, 0)
	for _, part := range strings.Split(req.Uuids, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			uuids = append(uuids, trimmed)
		}
	}
	if len(uuids) == 0 {
		return &types.LinkedEmployeesResponse{}, nil
	}

	emps, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
		condition.NewChain().In(hrm_employee.UserUuid, uuids).Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.LinkedEmployee, 0, len(emps))
	for _, emp := range emps {
		if emp.UserUuid == "" {
			continue
		}
		records = append(records, types.LinkedEmployee{
			UserUuid:     emp.UserUuid,
			EmployeeNo:   emp.EmployeeNo,
			EmployeeName: emp.Name,
		})
	}
	return &types.LinkedEmployeesResponse{Records: records}, nil
}
