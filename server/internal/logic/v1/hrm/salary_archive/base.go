package salary_archive

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
)

// dateLayout 档案日期口径: YYYY-MM-DD（与 hrm_employee_contract 同款 varchar(10) 日期列）。
const dateLayout = "2006-01-02"

// validArchiveStatus 启停枚举: 1=启用 0=停用（HRM status 约定）.
func validArchiveStatus(v string) bool {
	return v == "1" || v == "0"
}

// validDate 校验 YYYY-MM-DD 格式（严格解析，拒绝 2026-13-01 之类的越界值）.
func validDate(v string) bool {
	if v == "" {
		return false
	}
	_, err := time.ParseInLocation(dateLayout, v, time.Local)
	return err == nil
}

// normalizeArchiveInput 兜底默认值（HTTP default 标签之外，测试/内部调用同口径）.
func normalizeArchiveInput(req *archiveInput) {
	if req.Status == "" {
		req.Status = "1"
	}
	req.ExpireDate = trimSpace(req.ExpireDate)
}

// archiveInput Add/Edit 共同字段（normalizeArchiveInput/validateArchiveInput 的入参抽象）.
type archiveInput struct {
	EmployeeUuid  string
	SubjectUuid   string
	SchemeUuid    string
	FixedSalary   float64
	FixedAllowance float64
	Status        string
	EffectiveDate string
	ExpireDate    string
}

func trimSpace(s string) string {
	// 等价 strings.TrimSpace，但避免多余 import 噪音；档案场景 trim 足够。
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// validateArchiveInput 公共字段校验（不含存在性检查与区间重叠检查）.
func validateArchiveInput(in *archiveInput) error {
	if in.EmployeeUuid == "" {
		return status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "员工不能为空")
	}
	if in.SubjectUuid == "" {
		return status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "薪酬主体不能为空")
	}
	if !validArchiveStatus(in.Status) {
		return status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "薪资档案状态必须是 启用 或 停用")
	}
	if in.FixedSalary < 0 || in.FixedAllowance < 0 {
		return status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "固定薪资/固定补贴不能为负数")
	}
	if !validDate(in.EffectiveDate) {
		return status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "生效日期格式必须是 YYYY-MM-DD")
	}
	// 失效日期为空=长期有效；非空必须>生效日期（业务守卫 #2）
	if in.ExpireDate != "" {
		if !validDate(in.ExpireDate) {
			return status.ErrorMessage(errcodes.SalaryArchiveInvalidCode, "失效日期格式必须是 YYYY-MM-DD")
		}
		eff, _ := time.ParseInLocation(dateLayout, in.EffectiveDate, time.Local)
		exp, _ := time.ParseInLocation(dateLayout, in.ExpireDate, time.Local)
		if !exp.After(eff) {
			return status.ErrorMessage(errcodes.SalaryArchiveDateInvalidCode, "失效日期必须大于生效日期")
		}
	}
	return nil
}

// resolveEmployee 校验员工存在，返回 员工行 + 公司/部门名称（选员工带出 公司/部门 快照口径）。
// 员工行只读展示，公司/部门名称实时查 hrm_org（不落员工名快照于组织表）。
func resolveEmployee(ctx context.Context, conn sqlx.SqlConn, employeeUuid string) (*hrm_employee.HrmEmployee, string, string, error) {
	emp, err := hrm_employee.NewHrmEmployeeModel(conn).FindOneByUuid(ctx, nil, employeeUuid)
	if err != nil {
		if errors.Is(err, hrm_employee.ErrNotFound) {
			return nil, "", "", status.ErrorMessage(errcodes.SalaryArchiveEmployeeNotFoundCode, "员工不存在")
		}
		return nil, "", "", err
	}
	companyName := orgNameByUuid(ctx, conn, emp.CompanyUuid)
	orgName := orgNameByUuid(ctx, conn, emp.OrgUuid)
	return emp, companyName, orgName, nil
}

// orgNameByUuid 查组织名称快照（查不到返回空串，容忍历史数据）.
func orgNameByUuid(ctx context.Context, conn sqlx.SqlConn, orgUuid string) string {
	if orgUuid == "" {
		return ""
	}
	org, err := hrm_org.NewHrmOrgModel(conn).FindOneByUuid(ctx, nil, orgUuid)
	if err != nil {
		return ""
	}
	return org.OrgName
}

// resolveSubject 校验主体存在并返回行（引用主体不做启停检查，hrm-33 契约 #1 同款约定）.
func resolveSubject(ctx context.Context, conn sqlx.SqlConn, subjectUuid string) (*hrm_salary_subject.HrmSalarySubject, error) {
	subject, err := hrm_salary_subject.NewHrmSalarySubjectModel(conn).FindOneByUuid(ctx, nil, subjectUuid)
	if err != nil {
		if errors.Is(err, hrm_salary_subject.ErrNotFound) {
			return nil, status.ErrorMessage(errcodes.SalaryArchiveSubjectNotFoundCode, "薪酬主体不存在")
		}
		return nil, err
	}
	return subject, nil
}

// resolveScheme 校验方案存在并返回行（可为空=未绑定方案；引用方案不做启停检查，与主体约定一致）.
func resolveScheme(ctx context.Context, conn sqlx.SqlConn, schemeUuid string) (*hrm_salary_scheme.HrmSalaryScheme, error) {
	if schemeUuid == "" {
		return nil, nil
	}
	scheme, err := hrm_salary_scheme.NewHrmSalarySchemeModel(conn).FindOneByUuid(ctx, nil, schemeUuid)
	if err != nil {
		if errors.Is(err, hrm_salary_scheme.ErrNotFound) {
			return nil, status.ErrorMessage(errcodes.SalaryArchiveSchemeNotFoundCode, "薪酬方案不存在")
		}
		return nil, err
	}
	return scheme, nil
}

// checkIntervalOverlap 同员工同主体同期间重叠拒绝（业务守卫 #1）:
// 新区间 [effective, expire] 与既有任一区间交叉即拒绝（expire 为空=+∞）。
// excludeUuid 非空时排除自身（Edit 场景）。
// 注意：仅对 启用(status=1) 的既有档案做重叠检查——停用档案视作历史归档不占区间。
func checkIntervalOverlap(ctx context.Context, conn sqlx.SqlConn, employeeUuid, subjectUuid, effectiveDate, expireDate, excludeUuid string) error {
	rows, err := hrm_salary_archive.NewHrmSalaryArchiveModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_salary_archive.EmployeeUuid, employeeUuid).
			Equal(hrm_salary_archive.SubjectUuid, subjectUuid).
			Equal(hrm_salary_archive.Status, "1").
			Build()...)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if excludeUuid != "" && row.Uuid == excludeUuid {
			continue
		}
		// 交叉判定（日期闭区间，含端点；expire 空=+∞）:
		//   无交叉 ⇔ newEff > oldExp 或 newExp < oldEff；其余情形共用至少一天 → 拒绝
		if row.ExpireDate != "" && effectiveDate > row.ExpireDate {
			continue
		}
		if expireDate != "" && expireDate < row.EffectiveDate {
			continue
		}
		return status.ErrorMessage(errcodes.SalaryArchiveOverlapCode,
			"该员工在此薪酬主体下已存在生效区间重叠的薪资档案（"+row.EffectiveDate+" ~ "+displayExpire(row.ExpireDate)+"）")
	}
	return nil
}

func displayExpire(expireDate string) string {
	if expireDate == "" {
		return "长期"
	}
	return expireDate
}

// archiveRowToRecord 档案行 → API record.
func archiveRowToRecord(row *hrm_salary_archive.HrmSalaryArchive) types.HrmSalaryArchive {
	return types.HrmSalaryArchive{
		Uuid:           row.Uuid,
		EmployeeUuid:   row.EmployeeUuid,
		EmployeeNo:     row.EmployeeNo,
		EmployeeName:   row.EmployeeName,
		CompanyUuid:    row.CompanyUuid,
		CompanyName:    row.CompanyName,
		OrgUuid:        row.OrgUuid,
		OrgName:        row.OrgName,
		SubjectUuid:    row.SubjectUuid,
		SubjectName:    row.SubjectName,
		SchemeUuid:     row.SchemeUuid,
		SchemeName:     row.SchemeName,
		FixedSalary:    row.FixedSalary,
		FixedAllowance: row.FixedAllowance,
		Status:         row.Status,
		EffectiveDate:  row.EffectiveDate,
		ExpireDate:     row.ExpireDate,
		Remark:         row.Remark,
		CreateTime:     row.CreateTime.Format(time.DateTime),
		UpdateTime:     row.UpdateTime.Format(time.DateTime),
	}
}

// nullFloatOf decimal 列可空转换（本表 fixed 列 NOT NULL，保留给测试/后续扩展）.
func nullFloatOf(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: true}
}
