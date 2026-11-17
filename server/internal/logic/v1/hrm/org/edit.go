package org

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Edit 编辑组织 (名称/类型/负责人/显示顺序/状态/备注).
// req.ParentUuid 非空且与原值不同时表示移动节点（换父组织），
// 目标父不能是自己或自己的子孙（会成环）。
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if !validOrgType(req.OrgType) {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织类型必须是 公司 或 部门")
	}
	if !validOrgStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织状态必须是 开启 或 关闭")
	}

	self, err := l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OrgNotFoundCode, "组织不存在")
		}
		return nil, err
	}

	newParent := req.ParentUuid
	if newParent != self.ParentUuid {
		if newParent == req.Uuid {
			return nil, status.ErrorMessage(errcodes.OrgMoveCycleCode, "上级组织不能是自己")
		}
		if _, err = checkParent(l.ctx, l.svcCtx, newParent); err != nil {
			return nil, err
		}
		// 沿目标父向上走，遇到自己即成环（把自己挂到自己的子孙下）
		cursor := newParent
		for cursor != "" {
			if cursor == req.Uuid {
				return nil, status.ErrorMessage(errcodes.OrgMoveCycleCode, "不能移动到自己的下级组织下")
			}
			row, err := l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, cursor)
			if err != nil {
				if err == sqlx.ErrNotFound {
					break
				}
				return nil, err
			}
			cursor = row.ParentUuid
		}
	}

	fields := map[string]any{
		string(hrm_org.OrgName):  req.OrgName,
		string(hrm_org.OrgType):  req.OrgType,
		string(hrm_org.Leader):   req.Leader,
		string(hrm_org.OrderNum): req.OrderNum,
		string(hrm_org.Status):   req.Status,
		string(hrm_org.Remark):   req.Remark,
	}
	if req.ParentUuid != "" && req.ParentUuid != self.ParentUuid {
		fields[string(hrm_org.ParentUuid)] = req.ParentUuid
	}

	err = l.svcCtx.Model.HrmOrg.UpdateFieldsByCondition(l.ctx, nil, fields, condition.Condition{
		Field:    hrm_org.Uuid,
		Operator: condition.Equal,
		Value:    req.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
