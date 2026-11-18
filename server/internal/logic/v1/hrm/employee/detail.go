package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type Detail struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDetail(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Detail {
	return &Detail{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Detail 员工档案详情: 主表 + 三个从表全量 (编辑弹窗回显用).
func (l *Detail) Detail(req *types.DetailRequest) (resp *types.DetailResponse, err error) {
	row, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EmployeeNotFoundCode, "员工不存在")
		}
		return nil, err
	}

	record := employeeToType(row)
	fillLoginName(l.ctx, l.svcCtx, &record)
	educations, works, families, err := findChildren(l.ctx, l.svcCtx, row.Uuid)
	if err != nil {
		return nil, err
	}
	record.Educations = educations
	record.Works = works
	record.Families = families

	return &types.DetailResponse{Employee: record}, nil
}
