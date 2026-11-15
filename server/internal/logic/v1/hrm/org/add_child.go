package org

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

type AddChild struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAddChild(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *AddChild {
	return &AddChild{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// AddChild 新增下级组织 under an existing node.
func (l *AddChild) AddChild(req *types.AddChildRequest) (resp *types.AddChildResponse, err error) {
	if req.ParentUuid == "" {
		return nil, status.ErrorMessage(errcodes.OrgParentNotFoundCode, "上级组织不能为空")
	}
	if !validOrgType(req.OrgType) {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织类型必须是 公司 或 部门")
	}
	if !validOrgStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织状态必须是 开启 或 关闭")
	}
	if _, err = checkParent(l.ctx, l.svcCtx, req.ParentUuid); err != nil {
		return nil, err
	}

	if _, err = insertOne(l.ctx, l.svcCtx, req.ParentUuid, req.OrgName, req.OrgType, req.Leader, req.OrderNum, req.Status, req.Remark); err != nil {
		return nil, err
	}

	return &types.AddChildResponse{}, nil
}
