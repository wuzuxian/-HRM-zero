package salary_archive

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

// assertErrCode 断言业务错误码（与 scheduleboard_test 同款断言口径）.
func assertErrCode(t *testing.T, err error, wantCode int, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error %d, got nil", what, wantCode)
	}
	if sc := status.FromError(err); sc == nil || sc.Code() != status.Code(wantCode) {
		t.Fatalf("%s: error = %v, want code %d", what, err, wantCode)
	}
}

// seedOrg 直接插组织行（org_type 1=公司 2=部门）供名称快照解析.
func seedOrg(t *testing.T, svcCtx *svc.ServiceContext, name, orgType string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmOrg.Insert(t.Context(), nil, &hrm_org.HrmOrg{
		Uuid: u, OrgName: name, OrgType: orgType, OrderNum: 1, Status: "1",
	}); err != nil {
		t.Fatalf("seed org %s: %v", name, err)
	}
	return u
}

// seedEmployeeRow 直接插员工行（company/org uuid 由调用方决定，档案快照列从该行带出）.
func seedEmployeeRow(t *testing.T, svcCtx *svc.ServiceContext, name, employeeNo, companyUuid, orgUuid string) string {
	t.Helper()
	u := uuid.New().String()
	if err := svcCtx.Model.HrmEmployee.InsertV2(t.Context(), nil, &hrm_employee.HrmEmployee{
		Uuid: u, EmployeeNo: employeeNo, Name: name,
		CompanyUuid: companyUuid, OrgUuid: orgUuid, PersonStatus: "1",
	}); err != nil {
		t.Fatalf("seed employee %s: %v", name, err)
	}
	return u
}

// seedSubject 直接插薪酬主体行.
func seedSubject(t *testing.T, svcCtx *svc.ServiceContext, name string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalarySubject.Insert(t.Context(), nil, &hrm_salary_subject.HrmSalarySubject{
		Uuid: u, SubjectName: name, OrgUuid: uuid.New().String(), OrgName: "甲公司", Status: "1",
	}); err != nil {
		t.Fatalf("seed subject %s: %v", name, err)
	}
	return u
}

// seedScheme 插方案行.
func seedScheme(t *testing.T, svcCtx *svc.ServiceContext, name, subjectUuid string) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalaryScheme.Insert(t.Context(), nil, &hrm_salary_scheme.HrmSalaryScheme{
		Uuid: u, SchemeName: name, SubjectUuid: subjectUuid, SubjectName: name + "-主体",
		PersonStatus: "1", SettleCycle: "monthly", PayDay: 1, Currency: "CNY", Status: "1",
	}); err != nil {
		t.Fatalf("seed scheme %s: %v", name, err)
	}
	return u
}

// seedSchemeItem 插方案×项目关联行.
func seedSchemeItem(t *testing.T, svcCtx *svc.ServiceContext, schemeUuid, itemUuid string) {
	t.Helper()
	if err := svcCtx.Model.HrmSalarySchemeItem.InsertV2(t.Context(), nil, &hrm_salary_scheme_item.HrmSalarySchemeItem{
		Uuid: uuid.New().String(), SchemeUuid: schemeUuid, ItemUuid: itemUuid,
	}); err != nil {
		t.Fatalf("seed scheme item: %v", err)
	}
}

// seedSalaryItem 插薪资项目行（item_code 全局唯一，code 须避开迁移种子）.
func seedSalaryItem(t *testing.T, svcCtx *svc.ServiceContext, code, name, itemType, itemClass, valueMode string, calcOrder int64) string {
	t.Helper()
	u := uuid.New().String()
	if _, err := svcCtx.Model.HrmSalaryItem.Insert(t.Context(), nil, &hrm_salary_item.HrmSalaryItem{
		Uuid: u, ItemCode: code, ItemName: name, ItemType: itemType, ItemClass: itemClass,
		ValueMode: valueMode, CalcOrder: calcOrder, OrderNum: calcOrder, IsPreset: 1, Status: "1",
	}); err != nil {
		t.Fatalf("seed salary item %s: %v", code, err)
	}
	return u
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmSalaryArchive {
	t.Helper()
	req.Current = 1
	req.Size = 100
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list salary archives: %v", err)
	}
	return resp.Records
}

func findByName(records []types.HrmSalaryArchive, name string) *types.HrmSalaryArchive {
	for i := range records {
		if records[i].EmployeeName == name {
			return &records[i]
		}
	}
	return nil
}
