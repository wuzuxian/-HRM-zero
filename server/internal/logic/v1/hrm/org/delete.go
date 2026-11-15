package org

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
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
		svcCtx: svcCtx, r: r,
	}
}

// Delete 删除/批量删除组织. A node that still has 下级组织 must not be deleted —
// callers delete bottom-up (same rule as the demo site tree table).
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	children, err := l.svcCtx.Model.HrmOrg.CountByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_org.ParentUuid,
		Operator: condition.In,
		Value:    req.Uuids,
	})
	if err != nil {
		return nil, err
	}
	if children > 0 {
		return nil, status.ErrorMessage(errcodes.OrgHasChildrenCode, "存在下级组织，请先删除下级组织")
	}

	if err = l.svcCtx.Model.HrmOrg.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_org.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	}); err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
