package org

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
)

// validOrgType / validOrgStatus are the HRM 组织 enums (GLOSSARY「组织」):
// 组织类型 1=公司 2=部门; 组织状态 1=开启 0=关闭.
const (
	orgTypeCompany    = "1"
	orgTypeDepartment = "2"

	orgStatusEnabled  = "1"
	orgStatusDisabled = "0"
)

func validOrgType(v string) bool {
	return v == orgTypeCompany || v == orgTypeDepartment
}

func validOrgStatus(v string) bool {
	return v == orgStatusEnabled || v == orgStatusDisabled
}

// checkParent returns the parent row when parentUuid is non-empty.
func checkParent(ctx context.Context, svcCtx *svc.ServiceContext, parentUuid string) (*hrm_org.HrmOrg, error) {
	if parentUuid == "" {
		return nil, nil
	}
	parent, err := svcCtx.Model.HrmOrg.FindOneByUuid(ctx, nil, parentUuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OrgParentNotFoundCode, "上级组织不存在")
		}
		return nil, err
	}
	return parent, nil
}

// insertOne inserts an 组织 row and returns the stored record.
func insertOne(ctx context.Context, svcCtx *svc.ServiceContext, parentUuid, orgName, orgType, leader string, orderNum int64, orgStatus, remark string) (*hrm_org.HrmOrg, error) {
	row := &hrm_org.HrmOrg{
		Uuid:       uuid.New().String(),
		ParentUuid: parentUuid,
		OrgName:    orgName,
		OrgType:    orgType,
		Leader:     leader,
		OrderNum:   orderNum,
		Status:     orgStatus,
		Remark:     remark,
	}
	if err := svcCtx.Model.HrmOrg.InsertV2(ctx, nil, row); err != nil {
		return nil, err
	}
	return svcCtx.Model.HrmOrg.FindOneByUuid(ctx, nil, row.Uuid)
}

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// Add 新增组织. Empty ParentUuid creates a root; with a ParentUuid it behaves
// the same as 新增下级 (kept as a separate endpoint for独立按钮权限码).
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
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

	return &types.AddResponse{}, nil
}
