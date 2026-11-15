package salary_archive

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
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

// Add 新增员工薪资档案（demo-inventory 8.2）:
// 员工（选择器带出 公司/部门/工号/姓名 快照）+ 主体/方案（名称快照）+ 固定薪资/固定补贴 +
// 生效区间（同员工同主体区间重叠拒绝；失效日期必须>生效日期）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	in := &archiveInput{
		EmployeeUuid:   req.EmployeeUuid,
		SubjectUuid:    req.SubjectUuid,
		SchemeUuid:     req.SchemeUuid,
		FixedSalary:    req.FixedSalary,
		FixedAllowance: req.FixedAllowance,
		Status:         req.Status,
		EffectiveDate:  req.EffectiveDate,
		ExpireDate:     req.ExpireDate,
	}
	normalizeArchiveInput(in)
	if err = validateArchiveInput(in); err != nil {
		return nil, err
	}

	emp, companyName, orgName, err := resolveEmployee(l.ctx, l.svcCtx.SqlxConn, in.EmployeeUuid)
	if err != nil {
		return nil, err
	}
	subject, err := resolveSubject(l.ctx, l.svcCtx.SqlxConn, in.SubjectUuid)
	if err != nil {
		return nil, err
	}
	scheme, err := resolveScheme(l.ctx, l.svcCtx.SqlxConn, in.SchemeUuid)
	if err != nil {
		return nil, err
	}
	if err = checkIntervalOverlap(l.ctx, l.svcCtx.SqlxConn, in.EmployeeUuid, in.SubjectUuid, in.EffectiveDate, in.ExpireDate, ""); err != nil {
		return nil, err
	}

	schemeUuid, schemeName := "", ""
	if scheme != nil {
		schemeUuid, schemeName = scheme.Uuid, scheme.SchemeName
	}

	if _, err = l.svcCtx.Model.HrmSalaryArchive.Insert(l.ctx, nil, &hrm_salary_archive.HrmSalaryArchive{
		Uuid:           uuid.New().String(),
		EmployeeUuid:   in.EmployeeUuid,
		EmployeeNo:     emp.EmployeeNo,
		EmployeeName:   emp.Name,
		CompanyUuid:    emp.CompanyUuid,
		CompanyName:    companyName,
		OrgUuid:        emp.OrgUuid,
		OrgName:        orgName,
		SubjectUuid:    in.SubjectUuid,
		SubjectName:    subject.SubjectName,
		SchemeUuid:     schemeUuid,
		SchemeName:     schemeName,
		FixedSalary:    in.FixedSalary,
		FixedAllowance: in.FixedAllowance,
		Status:         in.Status,
		EffectiveDate:  in.EffectiveDate,
		ExpireDate:     in.ExpireDate,
		Remark:         req.Remark,
	}); err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
