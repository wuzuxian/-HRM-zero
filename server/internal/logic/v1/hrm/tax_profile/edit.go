package tax_profile

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_profile"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_profile"
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
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 编辑员工个税档案: 可换关联员工（仍保持员工唯一），姓名/身份证快照随换随更。
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateProfileCommon(req.EmployeeUuid, req.ResidentType, req.Status); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmTaxProfile.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TaxProfileNotFoundCode, "个税档案不存在")
		}
		return nil, err
	}

	// 员工唯一守卫（排除自身）
	if req.EmployeeUuid != row.EmployeeUuid {
		employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return nil, status.ErrorMessage(errcodes.TaxProfileEmployeeNotFoundCode, "关联员工不存在")
			}
			return nil, err
		}
		dup, err := l.svcCtx.Model.HrmTaxProfile.FindOneByCondition(l.ctx, nil,
			taxProfileEmployeeCondition(req.EmployeeUuid)...)
		if err != nil && err != hrm_tax_profile.ErrNotFound {
			return nil, err
		}
		if dup != nil && dup.Uuid != row.Uuid {
			return nil, status.ErrorMessage(errcodes.TaxProfileDuplicateCode, "该员工已存在个税档案")
		}
		row.EmployeeUuid = req.EmployeeUuid
		row.EmployeeName = employee.Name
		row.IdCard = employee.IdCard
	}

	if err = l.svcCtx.Model.HrmTaxProfile.UpdateFieldsByCondition(l.ctx, nil, map[string]any{
		string(hrm_tax_profile.EmployeeUuid):        row.EmployeeUuid,
		string(hrm_tax_profile.EmployeeName):        row.EmployeeName,
		string(hrm_tax_profile.IdCard):              row.IdCard,
		string(hrm_tax_profile.ResidentType):        req.ResidentType,
		string(hrm_tax_profile.TaxCity):             req.TaxCity,
		string(hrm_tax_profile.DefaultSpecialExtra): req.DefaultSpecialExtra,
		string(hrm_tax_profile.Status):              req.Status,
		string(hrm_tax_profile.Remark):              req.Remark,
	}, taxProfileUuidCondition(req.Uuid)...); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
