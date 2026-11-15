package salary_archive

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
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

// Edit 编辑员工薪资档案: 员工/主体/方案/金额/区间均可改；
// 员工或主体变更时快照列随之重取，区间变更时按新区间重新做同员工同主体重叠检查（排除自身）。
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
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

	row, err := l.svcCtx.Model.HrmSalaryArchive.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SalaryArchiveNotFoundCode, "薪资档案不存在")
		}
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
	if err = checkIntervalOverlap(l.ctx, l.svcCtx.SqlxConn, in.EmployeeUuid, in.SubjectUuid, in.EffectiveDate, in.ExpireDate, row.Uuid); err != nil {
		return nil, err
	}

	schemeUuid, schemeName := "", ""
	if scheme != nil {
		schemeUuid, schemeName = scheme.Uuid, scheme.SchemeName
	}

	row.EmployeeUuid = in.EmployeeUuid
	row.EmployeeNo = emp.EmployeeNo
	row.EmployeeName = emp.Name
	row.CompanyUuid = emp.CompanyUuid
	row.CompanyName = companyName
	row.OrgUuid = emp.OrgUuid
	row.OrgName = orgName
	row.SubjectUuid = in.SubjectUuid
	row.SubjectName = subject.SubjectName
	row.SchemeUuid = schemeUuid
	row.SchemeName = schemeName
	row.FixedSalary = in.FixedSalary
	row.FixedAllowance = in.FixedAllowance
	row.Status = in.Status
	row.EffectiveDate = in.EffectiveDate
	row.ExpireDate = in.ExpireDate
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmSalaryArchive.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
