package employee

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/jzero-io/jzero/core/status"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type Import struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewImport(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Import {
	return &Import{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// importHeader is the excelize template shared by Import and Export, mirroring
// the demo-inventory 2.1 master-table fields.
var importHeader = []string{
	"工号", "姓名", "性别", "手机号", "邮箱",
	"所属部门", "所属单位", "职位", "职务",
	"人员状态", "入职日期", "转正日期",
	"生日", "民族", "政治面貌", "婚姻状况", "籍贯", "身份证号",
	"户口地址", "现住址", "紧急联系人", "紧急联系人电话",
	"工资开户行", "工资卡账户", "备注",
}

// Import 导入员工 (spec user story 13): reads the shared xlsx template (one row
// per employee, first column must be 姓名); blank 工号 rows get auto-generated
// numbers, 所属部门/所属单位 resolved by name against hrm_org.
func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	raw, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.EmployeeImportInvalidCode, "导入文件内容不是合法的 base64")
	}

	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.EmployeeImportInvalidCode, "导入文件必须是 xlsx 格式")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.EmployeeImportInvalidCode, "导入文件读取失败")
	}
	if len(rows) < 2 {
		return nil, status.ErrorMessage(errcodes.EmployeeImportInvalidCode, "导入文件没有数据行")
	}

	// 组织名 → uuid (部门与单位共用一个组织树)
	orgs, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	orgNameToUuid := make(map[string]string, len(orgs))
	for _, org := range orgs {
		if _, ok := orgNameToUuid[org.OrgName]; !ok {
			orgNameToUuid[org.OrgName] = org.Uuid
		}
	}

	var successCount int64
	// 整个导入一个事务: 任一行失败即整体回滚, 避免"导入一半"的脏数据。
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, row := range rows {
			if strings.TrimSpace(cellAt(row, 1)) == "" || strings.TrimSpace(cellAt(row, 0)) == importHeader[0] {
				continue // blank line or header
			}
			name := strings.TrimSpace(cellAt(row, 1))

			gender, err := parseGender(cellAt(row, 2))
			if err != nil {
				return err
			}
			personStatus, err := parsePersonStatus(cellAt(row, 9))
			if err != nil {
				return err
			}
			maritalStatus, err := parseMaritalStatus(cellAt(row, 15))
			if err != nil {
				return err
			}
			orgUuid, err := resolveOrgUuid(orgNameToUuid, cellAt(row, 5), "所属部门")
			if err != nil {
				return err
			}
			companyUuid, err := resolveOrgUuid(orgNameToUuid, cellAt(row, 6), "所属单位")
			if err != nil {
				return err
			}

			addReq := &types.AddRequest{
				Name:             name,
				Gender:           gender,
				PersonStatus:     personStatus,
				MaritalStatus:    maritalStatus,
				Phone:            strings.TrimSpace(cellAt(row, 3)),
				Email:            strings.TrimSpace(cellAt(row, 4)),
				OrgUuid:          orgUuid,
				CompanyUuid:      companyUuid,
				Position:         strings.TrimSpace(cellAt(row, 7)),
				Post:             strings.TrimSpace(cellAt(row, 8)),
				HireDate:         strings.TrimSpace(cellAt(row, 10)),
				FormalDate:       strings.TrimSpace(cellAt(row, 11)),
				Birthday:         strings.TrimSpace(cellAt(row, 12)),
				Nation:           strings.TrimSpace(cellAt(row, 13)),
				PoliticalStatus:  strings.TrimSpace(cellAt(row, 14)),
				NativePlace:      strings.TrimSpace(cellAt(row, 16)),
				IdCard:           strings.TrimSpace(cellAt(row, 17)),
				HouseholdAddress: strings.TrimSpace(cellAt(row, 18)),
				Address:          strings.TrimSpace(cellAt(row, 19)),
				EmergencyContact: strings.TrimSpace(cellAt(row, 20)),
				EmergencyPhone:   strings.TrimSpace(cellAt(row, 21)),
				BankName:         strings.TrimSpace(cellAt(row, 22)),
				BankAccount:      strings.TrimSpace(cellAt(row, 23)),
				Remark:           strings.TrimSpace(cellAt(row, 24)),
			}
			if employeeNo := strings.TrimSpace(cellAt(row, 0)); employeeNo != "" {
				addReq.EmployeeNo = employeeNo
			}

			if addReq.PersonStatus == "" {
				addReq.PersonStatus = personStatusRegular
			}
			if addReq.Gender == "" {
				addReq.Gender = genderMale
			}

			employeeNo, err := generateEmployeeNo(ctx, session, l.svcCtx, addReq.EmployeeNo)
			if err != nil {
				return err
			}
			addReq.EmployeeNo = employeeNo
			if err = insertEmployee(ctx, session, l.svcCtx, addReq); err != nil {
				return err
			}
			successCount++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.ImportResponse{SuccessCount: successCount}, nil
}

func resolveOrgUuid(orgNameToUuid map[string]string, orgName, field string) (string, error) {
	orgName = strings.TrimSpace(orgName)
	if orgName == "" {
		return "", nil
	}
	u, ok := orgNameToUuid[orgName]
	if !ok {
		return "", status.ErrorMessage(errcodes.EmployeeImportInvalidCode, field+"不存在: "+orgName)
	}
	return u, nil
}

func cellAt(row []string, i int) string {
	if i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}

func parseGender(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "男":
		return genderMale, nil
	case "女":
		return genderFemale, nil
	case "":
		return genderMale, nil
	case genderMale, genderFemale:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.EmployeeImportInvalidCode, "性别必须是 男 或 女")
	}
}

func parsePersonStatus(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "正式":
		return personStatusRegular, nil
	case "试用期":
		return personStatusProbation, nil
	case "实习生":
		return personStatusIntern, nil
	case "临时工":
		return personStatusTemp, nil
	case "离职":
		return personStatusResigned, nil
	case "退休":
		return personStatusRetired, nil
	case "":
		return personStatusRegular, nil
	case personStatusRegular, personStatusProbation, personStatusIntern,
		personStatusTemp, personStatusResigned, personStatusRetired:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.EmployeeImportInvalidCode,
			"人员状态必须是 正式/试用期/实习生/临时工/离职/退休")
	}
}

func parseMaritalStatus(v string) (string, error) {
	switch strings.TrimSpace(v) {
	case "未婚":
		return maritalUnmarried, nil
	case "已婚":
		return maritalMarried, nil
	case "离异":
		return maritalDivorced, nil
	case "丧偶":
		return maritalWidowed, nil
	case "":
		return "", nil
	case maritalUnmarried, maritalMarried, maritalDivorced, maritalWidowed:
		return strings.TrimSpace(v), nil
	default:
		return "", status.ErrorMessage(errcodes.EmployeeImportInvalidCode, "婚姻状况必须是 未婚/已婚/离异/丧偶")
	}
}

// personStatusLabel / genderLabel / maritalStatusLabel are the label → code
// inverses used by Export.
func personStatusLabel(v string) string {
	switch v {
	case personStatusRegular:
		return "正式"
	case personStatusProbation:
		return "试用期"
	case personStatusIntern:
		return "实习生"
	case personStatusTemp:
		return "临时工"
	case personStatusResigned:
		return "离职"
	case personStatusRetired:
		return "退休"
	default:
		return ""
	}
}

func genderLabel(v string) string {
	if v == genderFemale {
		return "女"
	}
	return "男"
}

func maritalStatusLabel(v string) string {
	switch v {
	case maritalUnmarried:
		return "未婚"
	case maritalMarried:
		return "已婚"
	case maritalDivorced:
		return "离异"
	case maritalWidowed:
		return "丧偶"
	default:
		return ""
	}
}
