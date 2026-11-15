package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_work"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Delete 批量删除员工: 三个从表随主表级联删除, 全部在同一事务内.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, uuid := range req.Uuids {
			if err := l.svcCtx.Model.HrmEmployeeEducation.DeleteByCondition(ctx, session,
				condition.NewChain().Equal(hrm_employee_education.EmployeeUuid, uuid).Build()...); err != nil {
				return err
			}
			if err := l.svcCtx.Model.HrmEmployeeWork.DeleteByCondition(ctx, session,
				condition.NewChain().Equal(hrm_employee_work.EmployeeUuid, uuid).Build()...); err != nil {
				return err
			}
			if err := l.svcCtx.Model.HrmEmployeeFamily.DeleteByCondition(ctx, session,
				condition.NewChain().Equal(hrm_employee_family.EmployeeUuid, uuid).Build()...); err != nil {
				return err
			}
			if err := l.svcCtx.Model.HrmEmployee.DeleteByCondition(ctx, session,
				condition.NewChain().Equal(hrm_employee.Uuid, uuid).Build()...); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
