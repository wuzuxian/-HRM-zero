package salary_archive

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
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

// Delete 批量删除薪资档案（DeleteByCondition 按 uuid IN）；
// 删除无消费侧守卫（HRM-37 批次若需拦截「被批次引用的档案不可删」，在批次侧查自己的表，HRM-24/26 约定）。
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return nil, status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "请选择要删除的薪资档案")
	}
	if err = l.svcCtx.Model.HrmSalaryArchive.DeleteByCondition(l.ctx, nil,
		condition.NewChain().In(hrm_salary_archive.Uuid, req.Uuids).Build()...); err != nil {
		return nil, err
	}
	return &types.DeleteResponse{}, nil
}
