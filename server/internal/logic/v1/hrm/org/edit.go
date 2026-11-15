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

// Edit 编辑组织 (名称/类型/负责人/显示顺序/状态/备注). Parent stays fixed —
// moving nodes is not part of the demo spec.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if !validOrgType(req.OrgType) {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织类型必须是 公司 或 部门")
	}
	if !validOrgStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.OrgImportInvalidCode, "组织状态必须是 开启 或 关闭")
	}

	if _, err = l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, req.Uuid); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OrgNotFoundCode, "组织不存在")
		}
		return nil, err
	}

	err = l.svcCtx.Model.HrmOrg.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_org.OrgName):  req.OrgName,
		string(hrm_org.OrgType):  req.OrgType,
		string(hrm_org.Leader):   req.Leader,
		string(hrm_org.OrderNum): req.OrderNum,
		string(hrm_org.Status):   req.Status,
		string(hrm_org.Remark):   req.Remark,
	}, condition.Condition{
		Field:    hrm_org.Uuid,
		Operator: condition.Equal,
		Value:    req.Uuid,
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
