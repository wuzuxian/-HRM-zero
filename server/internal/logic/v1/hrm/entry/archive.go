package entry

import (
	"context"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_work"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_work"
)

// createEmployeeArchive 自动建档 (demo-inventory 3.1): inside the caller's
// (approval) transaction —
//   - 工号自动生成: 复用 HRM-03 纯函数 seam NextEmployeeNo + MaxNumericEmployeeNo
//     (最大工号+1, 8 位补零), against the same tx session;
//   - 人员状态: 有试用期(>0 月) → 试用期(PersonStatusProbation), else 正式;
//   - 预计转正: expect_formal_date (空则 入职日期+试用期月, computed at save time);
//   - 三子表 (工作经历/教育经历/家庭成员) 从入职单快照落 hrm_employee_* 档案表;
//
// returns the new archive uuid for the employee_uuid 回写 (idempotency marker).
func createEmployeeArchive(ctx context.Context, session sqlx.Session, row *hrm_entry_bill.HrmEntryBill) (string, error) {
	conn := sqlx.NewSqlConnFromSession(session)

	// 工号自动生成 (复用 HRM-03 seam): read existing numbers through the tx so
	// the generated 工号 is consistent with the archive insert (all-or-nothing).
	employees, err := hrm_employee.NewHrmEmployeeModel(conn).FindByCondition(ctx, session)
	if err != nil {
		return "", err
	}
	nos := make([]string, 0, len(employees))
	for _, e := range employees {
		nos = append(nos, e.EmployeeNo)
	}
	employeeNo, err := employee.NextEmployeeNo(employee.MaxNumericEmployeeNo(nos))
	if err != nil {
		return "", err
	}

	personStatus := employee.PersonStatusRegular
	if row.ProbationMonths > 0 {
		personStatus = employee.PersonStatusProbation
	}

	employeeUuid := uuid.New().String()
	emp := &hrm_employee.HrmEmployee{
		Uuid:             employeeUuid,
		EmployeeNo:       employeeNo,
		Name:             row.Name,
		Gender:           row.Gender,
		Birthday:         row.Birthday,
		Nation:           row.Nation,
		PoliticalStatus:  row.PoliticalStatus,
		MaritalStatus:    row.MaritalStatus,
		NativePlace:      row.NativePlace,
		HouseholdAddress: row.HouseholdAddress,
		Address:          row.Address,
		EmergencyContact: row.EmergencyContact,
		EmergencyPhone:   row.EmergencyPhone,
		Phone:            row.Phone,
		Email:            row.Email,
		Avatar:           row.Avatar,
		BankName:         row.BankName,
		BankAccount:      row.BankAccount,
		CompanyUuid:      row.CompanyUuid,
		OrgUuid:          row.OrgUuid,
		Position:         row.Position,
		Post:             row.Post,
		PersonStatus:     personStatus,
		HireDate:         row.HireDate,
		FormalDate:       row.ExpectFormalDate,
		Remark:           row.Remark,
	}
	if err := hrm_employee.NewHrmEmployeeModel(conn).InsertV2(ctx, session, emp); err != nil {
		return "", err
	}

	// 三子表落档案: snapshot rows from the 入职单 sub-tables (same tx; the
	// archive is brand new so no delete-then-insert is needed).
	if err := copyChildrenToArchive(ctx, session, conn, employeeUuid, row.Uuid); err != nil {
		return "", err
	}
	return employeeUuid, nil
}

// copyChildrenToArchive reads the 入职单 snapshot sub-tables and writes them
// into the employee archive sub-tables.
func copyChildrenToArchive(ctx context.Context, session sqlx.Session, conn sqlx.SqlConn, employeeUuid, entryUuid string) error {
	works, err := hrm_entry_bill_work.NewHrmEntryBillWorkModel(conn).FindByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_entry_bill_work.EntryUuid, entryUuid).
			OrderByAsc(hrm_entry_bill_work.Id).
			Build()...)
	if err != nil {
		return err
	}
	educations, err := hrm_entry_bill_education.NewHrmEntryBillEducationModel(conn).FindByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_entry_bill_education.EntryUuid, entryUuid).
			OrderByAsc(hrm_entry_bill_education.Id).
			Build()...)
	if err != nil {
		return err
	}
	families, err := hrm_entry_bill_family.NewHrmEntryBillFamilyModel(conn).FindByCondition(ctx, session,
		condition.NewChain().
			Equal(hrm_entry_bill_family.EntryUuid, entryUuid).
			OrderByAsc(hrm_entry_bill_family.Id).
			Build()...)
	if err != nil {
		return err
	}

	if len(works) > 0 {
		rows := make([]*hrm_employee_work.HrmEmployeeWork, 0, len(works))
		for _, w := range works {
			rows = append(rows, &hrm_employee_work.HrmEmployeeWork{
				Uuid:         uuid.New().String(),
				EmployeeUuid: employeeUuid,
				Company:      w.Company,
				Position:     w.Position,
				StartDate:    w.StartDate,
				EndDate:      w.EndDate,
				Remark:       w.Remark,
			})
		}
		if err := hrm_employee_work.NewHrmEmployeeWorkModel(conn).BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}

	if len(educations) > 0 {
		rows := make([]*hrm_employee_education.HrmEmployeeEducation, 0, len(educations))
		for _, e := range educations {
			rows = append(rows, &hrm_employee_education.HrmEmployeeEducation{
				Uuid:         uuid.New().String(),
				EmployeeUuid: employeeUuid,
				School:       e.School,
				Major:        e.Major,
				Education:    e.Education,
				StartDate:    e.StartDate,
				EndDate:      e.EndDate,
				Remark:       e.Remark,
			})
		}
		if err := hrm_employee_education.NewHrmEmployeeEducationModel(conn).BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}

	if len(families) > 0 {
		rows := make([]*hrm_employee_family.HrmEmployeeFamily, 0, len(families))
		for _, f := range families {
			rows = append(rows, &hrm_employee_family.HrmEmployeeFamily{
				Uuid:         uuid.New().String(),
				EmployeeUuid: employeeUuid,
				Name:         f.Name,
				Relation:     f.Relation,
				Phone:        f.Phone,
				WorkUnit:     f.WorkUnit,
				Remark:       f.Remark,
			})
		}
		if err := hrm_employee_family.NewHrmEmployeeFamilyModel(conn).BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}
	return nil
}
