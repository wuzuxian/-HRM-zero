package empcert

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/empcert"
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

// Add 新增员工证照（spec 用户故事 17）：员工档案 uuid / 证书类型 / 证号必填；
// 证号明文存储，列表/预警出参统一脱敏。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	emp, err := validateCertInput(l.ctx, l.svcCtx, req.EmployeeUuid, req.CertType, req.CertNo)
	if err != nil {
		return nil, err
	}
	if req.IssueDate != "" && !validDate(req.IssueDate) {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "发证日期格式须为 YYYY-MM-DD")
	}
	if req.ExpiryDate != "" && !validDate(req.ExpiryDate) {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "到期日格式须为 YYYY-MM-DD")
	}

	_, err = l.svcCtx.Model.HrmEmployeeCertificate.Insert(l.ctx, nil, &hrm_employee_certificate.HrmEmployeeCertificate{
		Uuid:         uuidNew(),
		EmployeeUuid: req.EmployeeUuid,
		EmployeeNo:   emp.EmployeeNo,
		EmployeeName: emp.Name,
		OrgUuid:      emp.OrgUuid,
		CertType:     req.CertType,
		CertNo:       req.CertNo,
		IssueDate:    req.IssueDate,
		ExpiryDate:   req.ExpiryDate,
		Remark:       req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}

// validateCertInput 证书必填项 + 员工档案定位（add/edit 共用）。
func validateCertInput(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid, certType, certNo string) (*hrm_employee.HrmEmployee, error) {
	if employeeUuid == "" {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "员工档案 uuid 不能为空")
	}
	if certType == "" {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "证书类型不能为空")
	}
	if certNo == "" {
		return nil, status.ErrorMessage(errcodes.EmployeeCertificateInvalidCode, "证号不能为空")
	}
	emp, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.Uuid, employeeUuid).Build()...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.EmployeeNotFoundCode, "员工不存在")
		}
		return nil, err
	}
	return emp, nil
}
