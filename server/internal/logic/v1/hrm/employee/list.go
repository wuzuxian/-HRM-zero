package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 员工档案分页列表, 支持按工号/姓名/部门/人员状态筛选 (spec user story 11).
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmEmployee.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			Like(hrm_employee.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == "")).
			Like(hrm_employee.Name, "%"+req.Name+"%", condition.WithSkip(req.Name == "")).
			Equal(hrm_employee.OrgUuid, req.OrgUuid, condition.WithSkip(req.OrgUuid == "")).
			Equal(hrm_employee.PersonStatus, req.PersonStatus, condition.WithSkip(req.PersonStatus == "")).
			OrderByDesc(hrm_employee.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmEmployee, 0, len(rows))
	for _, row := range rows {
		records = append(records, employeeToType(row))
	}
	// 登录账号列: 回填账号侧用户名 (工号变更不重命名账号, 二者可能不一致)
	fillLoginNames(l.ctx, l.svcCtx, records)

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
