package tax_profile

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_profile"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_profile"
)

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
		svcCtx: svcCtx,
		r:      r,
	}
}

// taxProfileEmployeeCondition 个税档案按员工查询的条件（员工唯一守卫复用）.
func taxProfileEmployeeCondition(employeeUuid string) []condition.Condition {
	return []condition.Condition{{
		Field:    hrm_tax_profile.EmployeeUuid,
		Operator: condition.Equal,
		Value:    employeeUuid,
	}}
}

// taxProfileUuidCondition 个税档案按 uuid 查询的条件.
func taxProfileUuidCondition(uuid string) []condition.Condition {
	return []condition.Condition{{
		Field:    hrm_tax_profile.Uuid,
		Operator: condition.Equal,
		Value:    uuid,
	}}
}

// Add 新增员工个税档案（demo-inventory 8.3）: 一个员工只有一份个税档案（员工唯一）；
// 姓名/身份证以快照方式落库（同 HRM-33 org_name 快照约定，编辑时随换随更）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.Status == "" {
		req.Status = "1"
	}
	if err = validateProfileCommon(req.EmployeeUuid, req.ResidentType, req.Status); err != nil {
		return nil, err
	}

	employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TaxProfileEmployeeNotFoundCode, "关联员工不存在")
		}
		return nil, err
	}

	// 员工唯一守卫
	dup, err := l.svcCtx.Model.HrmTaxProfile.FindOneByCondition(l.ctx, nil,
		taxProfileEmployeeCondition(req.EmployeeUuid)...)
	if err != nil && err != hrm_tax_profile.ErrNotFound {
		return nil, err
	}
	if dup != nil {
		return nil, status.ErrorMessage(errcodes.TaxProfileDuplicateCode, "该员工已存在个税档案")
	}

	if _, err = l.svcCtx.Model.HrmTaxProfile.Insert(l.ctx, nil, &hrm_tax_profile.HrmTaxProfile{
		Uuid:                uuid.New().String(),
		EmployeeUuid:        req.EmployeeUuid,
		EmployeeName:        employee.Name,
		IdCard:              employee.IdCard,
		ResidentType:        req.ResidentType,
		TaxCity:             req.TaxCity,
		DefaultSpecialExtra: req.DefaultSpecialExtra,
		Status:              req.Status,
		Remark:              req.Remark,
	}); err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
