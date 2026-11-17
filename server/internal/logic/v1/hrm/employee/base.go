package employee

import (
	"context"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_work"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

func uuidNew() string {
	return uuid.New().String()
}

// HRM 员工档案 enums (GLOSSARY「人员状态」/ demo-inventory 2.1):
// 性别 1=男 2=女; 婚姻状况 1=未婚 2=已婚 3=离异 4=丧偶;
// 人员状态 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休。
// 在册口径 = person_status NOT IN (离职, 退休) — 组织架构图 (HRM-02) 汇总人数使用。
// 用户状态由 user_uuid 派生: '1'=已生成系统账号, '0'=未生成 (#6 同步 seam)。
const (
	genderMale   = "1"
	genderFemale = "2"

	maritalUnmarried = "1"
	maritalMarried   = "2"
	maritalDivorced  = "3"
	maritalWidowed   = "4"

	// 人员状态 enum (GLOSSARY「人员状态」): 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休.
	// Exported so downstream document logics (入职单 HRM-07 自动建档等) reuse the
	// same vocabulary instead of re-declaring literals.
	PersonStatusRegular   = "1"
	PersonStatusProbation = "2"
	PersonStatusIntern    = "3"
	PersonStatusTemp      = "4"
	PersonStatusResigned  = "5"
	PersonStatusRetired   = "6"

	userStatusCreated = "1"
	userStatusNone    = "0"
)

func validGender(v string) bool {
	return v == genderMale || v == genderFemale
}

func validMaritalStatus(v string) bool {
	return v == maritalUnmarried || v == maritalMarried || v == maritalDivorced || v == maritalWidowed
}

func validPersonStatus(v string) bool {
	return v == PersonStatusRegular || v == PersonStatusProbation || v == PersonStatusIntern ||
		v == PersonStatusTemp || v == PersonStatusResigned || v == PersonStatusRetired
}

// employeeToType converts a model row into the API record (UserStatus 由
// user_uuid 派生: 是否已生成系统账号). loginName 单独通过 fillLoginNames 回填
// (需要 join manage_user, 列表/详情调用方按需调用).
func employeeToType(row *hrm_employee.HrmEmployee) types.HrmEmployee {
	out := types.HrmEmployee{
		Uuid:             row.Uuid,
		EmployeeNo:       row.EmployeeNo,
		Name:             row.Name,
		Gender:           row.Gender,
		Birthday:         row.Birthday,
		Nation:           row.Nation,
		PoliticalStatus:  row.PoliticalStatus,
		MaritalStatus:    row.MaritalStatus,
		NativePlace:      row.NativePlace,
		BloodType:        row.BloodType,
		IdCard:           row.IdCard,
		Height:           row.Height,
		Weight:           row.Weight,
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
		PersonStatus:     row.PersonStatus,
		HireDate:         row.HireDate,
		FormalDate:       row.FormalDate,
		UserStatus:       userStatusNone,
		UserUuid:         row.UserUuid,
		Remark:           row.Remark,
		CreateTime:       row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:       row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
	if row.UserUuid != "" {
		out.UserStatus = userStatusCreated
	}
	return out
}

// findEmployeeOrNotFound HRM-51 关联族端点 (link/unlink/createAndLink/开号接缝)
// 共用的档案读取错误收敛: 档案缺失 → EmployeeUserNotFoundCode ("员工不存在: <uuid>"),
// 其余错误原样透传。
func findEmployeeOrNotFound(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext,
	employeeUuid string) (*hrm_employee.HrmEmployee, error) {
	emp, err := svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, employeeUuid)
	if err == sqlx.ErrNotFound {
		return nil, status.ErrorMessage(errcodes.EmployeeUserNotFoundCode, "员工不存在: "+employeeUuid)
	}
	if err != nil {
		return nil, err
	}
	return emp, nil
}

// fillLoginNames 批量回填 records 的 LoginName (账号侧 manage_user.username;
// 工号变更不重命名账号, 二者可能不一致, 列表「登录账号」列展示该值).
// 未关联账号的记录保持空串。查询失败不阻断列表 (登录账号是展示字段), 仅记日志。
func fillLoginNames(ctx context.Context, svcCtx *svc.ServiceContext, records []types.HrmEmployee) {
	userUuids := make([]any, 0, len(records))
	seen := make(map[string]struct{}, len(records))
	for _, rec := range records {
		if rec.UserUuid == "" {
			continue
		}
		if _, ok := seen[rec.UserUuid]; ok {
			continue
		}
		seen[rec.UserUuid] = struct{}{}
		userUuids = append(userUuids, rec.UserUuid)
	}
	if len(userUuids) == 0 {
		return
	}
	users, err := svcCtx.Model.ManageUser.FindByCondition(ctx, nil, condition.Condition{
		Field:    manage_user.Uuid,
		Operator: condition.In,
		Value:    userUuids,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("hrm: 登录账号回填查询失败 (展示字段, 不阻断列表): %v", err)
		return
	}
	nameByUuid := make(map[string]string, len(users))
	for _, u := range users {
		nameByUuid[u.Uuid] = u.Username
	}
	for i := range records {
		if records[i].UserUuid != "" {
			records[i].LoginName = nameByUuid[records[i].UserUuid]
		}
	}
}

// fillLoginName 单条详情回填 LoginName。
func fillLoginName(ctx context.Context, svcCtx *svc.ServiceContext, record *types.HrmEmployee) {
	if record == nil || record.UserUuid == "" {
		return
	}
	user, err := svcCtx.Model.ManageUser.FindOneByUuid(ctx, nil, record.UserUuid)
	if err != nil {
		return
	}
	record.LoginName = user.Username
}

// generateEmployeeNo resolves the 工号 for a new employee inside a transaction:
// provided employee_no is duplicate-checked; blank employee_no is generated as
// 最大工号+1 (8 位补零). Concurrent adds are serialized by the transaction; the
// hrm_employee.employee_no UNIQUE index is the last line of defense.
func generateEmployeeNo(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, employeeNo string) (string, error) {
	if employeeNo != "" {
		_, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, session,
			condition.NewChain().Equal(hrm_employee.EmployeeNo, employeeNo).Build()...)
		if err == nil {
			return "", status.ErrorMessage(errcodes.EmployeeNoDuplicateCode, "工号已存在: "+employeeNo)
		}
		if err != sqlx.ErrNotFound {
			return "", err
		}
		return employeeNo, nil
	}

	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, session)
	if err != nil {
		return "", err
	}
	nos := make([]string, 0, len(rows))
	for _, row := range rows {
		nos = append(nos, row.EmployeeNo)
	}
	next, err := NextEmployeeNo(maxNumericEmployeeNo(nos))
	if err != nil {
		if errors.Is(err, ErrEmployeeNoExhausted) {
			return "", status.ErrorMessage(errcodes.EmployeeNoExhaustedCode, err.Error())
		}
		return "", status.ErrorMessage(errcodes.EmployeeNoDuplicateCode, err.Error())
	}
	return next, nil
}

// saveChildren 全删全插 the three sub-tables (教育经历/工作经历/家属) for an
// employee inside the caller's transaction, mirroring spec user story 9.
func saveChildren(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, employeeUuid string,
	educations []types.HrmEmployeeEducation, works []types.HrmEmployeeWork, families []types.HrmEmployeeFamily) error {
	chain := condition.NewChain().Equal(hrm_employee_education.EmployeeUuid, employeeUuid)
	if err := svcCtx.Model.HrmEmployeeEducation.DeleteByCondition(ctx, session, chain.Build()...); err != nil {
		return err
	}
	workChain := condition.NewChain().Equal(hrm_employee_work.EmployeeUuid, employeeUuid)
	if err := svcCtx.Model.HrmEmployeeWork.DeleteByCondition(ctx, session, workChain.Build()...); err != nil {
		return err
	}
	familyChain := condition.NewChain().Equal(hrm_employee_family.EmployeeUuid, employeeUuid)
	if err := svcCtx.Model.HrmEmployeeFamily.DeleteByCondition(ctx, session, familyChain.Build()...); err != nil {
		return err
	}

	if len(educations) > 0 {
		rows := make([]*hrm_employee_education.HrmEmployeeEducation, 0, len(educations))
		for _, item := range educations {
			rows = append(rows, &hrm_employee_education.HrmEmployeeEducation{
				Uuid:         uuidNew(),
				EmployeeUuid: employeeUuid,
				School:       item.School,
				Major:        item.Major,
				Education:    item.Education,
				StartDate:    item.StartDate,
				EndDate:      item.EndDate,
				Remark:       item.Remark,
			})
		}
		if err := svcCtx.Model.HrmEmployeeEducation.BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}

	if len(works) > 0 {
		rows := make([]*hrm_employee_work.HrmEmployeeWork, 0, len(works))
		for _, item := range works {
			rows = append(rows, &hrm_employee_work.HrmEmployeeWork{
				Uuid:         uuidNew(),
				EmployeeUuid: employeeUuid,
				Company:      item.Company,
				Position:     item.Position,
				StartDate:    item.StartDate,
				EndDate:      item.EndDate,
				Remark:       item.Remark,
			})
		}
		if err := svcCtx.Model.HrmEmployeeWork.BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}

	if len(families) > 0 {
		rows := make([]*hrm_employee_family.HrmEmployeeFamily, 0, len(families))
		for _, item := range families {
			rows = append(rows, &hrm_employee_family.HrmEmployeeFamily{
				Uuid:         uuidNew(),
				EmployeeUuid: employeeUuid,
				Name:         item.Name,
				Relation:     item.Relation,
				Phone:        item.Phone,
				WorkUnit:     item.WorkUnit,
				Remark:       item.Remark,
			})
		}
		if err := svcCtx.Model.HrmEmployeeFamily.BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}
	return nil
}

// GenerateEmployeeNo resolves the 工号 for a new employee inside the caller's
// transaction: provided employee_no is duplicate-checked; blank employee_no is
// generated as 最大工号+1 (8 位补零). Exported wrapper around generateEmployeeNo
// for downstream document logics (入职单 HRM-07 审批通过自动建档 reuses the same
// numbering instead of re-deriving it).
func GenerateEmployeeNo(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, employeeNo string) (string, error) {
	return generateEmployeeNo(ctx, session, svcCtx, employeeNo)
}

// InsertEmployeeRow writes a fully-built hrm_employee master row inside the
// caller's transaction. Exported for 入职单 HRM-07 自动建档, which snapshots the
// approved 入职单 into a new archive row and manages its own sub-tables
// (SaveEmployeeChildren below for the archive-side sub-table write).
func InsertEmployeeRow(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, row *hrm_employee.HrmEmployee) error {
	return svcCtx.Model.HrmEmployee.InsertV2(ctx, session, row)
}

// SaveEmployeeChildren 全删全插 the three archive sub-tables (教育经历/工作经历/家属)
// for an employee inside the caller's transaction. Exported wrapper around
// saveChildren for downstream document logics (入职单 HRM-07 三子表落档案).
func SaveEmployeeChildren(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, employeeUuid string,
	educations []types.HrmEmployeeEducation, works []types.HrmEmployeeWork, families []types.HrmEmployeeFamily) error {
	return saveChildren(ctx, session, svcCtx, employeeUuid, educations, works, families)
}

// insertEmployee writes the master row (+ any provided sub-table rows) inside
// the caller's transaction. Shared by Add and Import; returns the new archive
// uuid (开号接缝 ProvisionAccount 需要档案指针).
func insertEmployee(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, req *types.AddRequest) (string, error) {
	row := &hrm_employee.HrmEmployee{
		Uuid:             uuidNew(),
		EmployeeNo:       req.EmployeeNo,
		Name:             req.Name,
		Gender:           req.Gender,
		Birthday:         req.Birthday,
		Nation:           req.Nation,
		PoliticalStatus:  req.PoliticalStatus,
		MaritalStatus:    req.MaritalStatus,
		NativePlace:      req.NativePlace,
		BloodType:        req.BloodType,
		IdCard:           req.IdCard,
		Height:           req.Height,
		Weight:           req.Weight,
		HouseholdAddress: req.HouseholdAddress,
		Address:          req.Address,
		EmergencyContact: req.EmergencyContact,
		EmergencyPhone:   req.EmergencyPhone,
		Phone:            req.Phone,
		Email:            req.Email,
		Avatar:           req.Avatar,
		BankName:         req.BankName,
		BankAccount:      req.BankAccount,
		CompanyUuid:      req.CompanyUuid,
		OrgUuid:          req.OrgUuid,
		Position:         req.Position,
		Post:             req.Post,
		PersonStatus:     req.PersonStatus,
		HireDate:         req.HireDate,
		FormalDate:       req.FormalDate,
		Remark:           req.Remark,
	}
	if err := svcCtx.Model.HrmEmployee.InsertV2(ctx, session, row); err != nil {
		return "", err
	}
	if err := saveChildren(ctx, session, svcCtx, row.Uuid, req.Educations, req.Works, req.Families); err != nil {
		return "", err
	}
	return row.Uuid, nil
}

// findChildren loads the three sub-tables for a detail record.
func findChildren(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuid string) ([]types.HrmEmployeeEducation, []types.HrmEmployeeWork, []types.HrmEmployeeFamily, error) {
	educations, err := svcCtx.Model.HrmEmployeeEducation.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee_education.EmployeeUuid, employeeUuid).OrderByAsc(hrm_employee_education.Id).Build()...)
	if err != nil {
		return nil, nil, nil, err
	}
	works, err := svcCtx.Model.HrmEmployeeWork.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee_work.EmployeeUuid, employeeUuid).OrderByAsc(hrm_employee_work.Id).Build()...)
	if err != nil {
		return nil, nil, nil, err
	}
	families, err := svcCtx.Model.HrmEmployeeFamily.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee_family.EmployeeUuid, employeeUuid).OrderByAsc(hrm_employee_family.Id).Build()...)
	if err != nil {
		return nil, nil, nil, err
	}

	eduTypes := make([]types.HrmEmployeeEducation, 0, len(educations))
	for _, row := range educations {
		eduTypes = append(eduTypes, types.HrmEmployeeEducation{
			Uuid:      row.Uuid,
			School:    row.School,
			Major:     row.Major,
			Education: row.Education,
			StartDate: row.StartDate,
			EndDate:   row.EndDate,
			Remark:    row.Remark,
		})
	}
	workTypes := make([]types.HrmEmployeeWork, 0, len(works))
	for _, row := range works {
		workTypes = append(workTypes, types.HrmEmployeeWork{
			Uuid:      row.Uuid,
			Company:   row.Company,
			Position:  row.Position,
			StartDate: row.StartDate,
			EndDate:   row.EndDate,
			Remark:    row.Remark,
		})
	}
	familyTypes := make([]types.HrmEmployeeFamily, 0, len(families))
	for _, row := range families {
		familyTypes = append(familyTypes, types.HrmEmployeeFamily{
			Uuid:     row.Uuid,
			Name:     row.Name,
			Relation: row.Relation,
			Phone:    row.Phone,
			WorkUnit: row.WorkUnit,
			Remark:   row.Remark,
		})
	}
	return eduTypes, workTypes, familyTypes, nil
}
