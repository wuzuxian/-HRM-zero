package salary_scheme

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
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

// Delete 删除/批量删除薪酬方案 + 其项目配置（事务内）。
// HRM-37 薪资档案/批次 seam: 若需拦截「被档案/批次引用的方案不可删」，
// 守卫落在消费侧（按 scheme_uuid 查自己的表），本侧不回查。
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(_ context.Context, session sqlx.Session) error {
		if err := l.svcCtx.Model.HrmSalaryScheme.DeleteByCondition(l.ctx, session,
			condition.NewChain().In(hrm_salary_scheme.Uuid, req.Uuids).Build()...); err != nil {
			return err
		}
		return l.svcCtx.Model.HrmSalarySchemeItem.DeleteByCondition(l.ctx, session,
			condition.NewChain().In(hrm_salary_scheme_item.SchemeUuid, req.Uuids).Build()...)
	})
	if err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
