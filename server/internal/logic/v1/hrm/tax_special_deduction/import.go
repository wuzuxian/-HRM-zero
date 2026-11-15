package tax_special_deduction

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_special_deduction"
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

// importHeader 导入/导出共用模板列（姓名+身份证 定位员工，扣除项用中文名）。
var importHeader = []string{
	"姓名", "身份证号", "扣除年度", "月份", "扣除项", "扣除金额", "备注",
}

// deductTypeByName 扣除项中文名 → 编码（与 validate.go 的 7 类一一对应）。
var deductTypeByName = map[string]string{
	"子女教育":         DeductTypeChildrenEdu,
	"继续教育":         DeductTypeContinuingEdu,
	"大病医疗":         DeductTypeCriticalIllness,
	"住房贷款利息": DeductTypeHouseLoan,
	"住房租金":         DeductTypeHouseRent,
	"赡养老人":         DeductTypeSupportElderly,
	"3岁以下婴幼儿照护": DeductTypeInfantCare,
}

// deductTypeName 编码 → 中文名（导出用）。
func deductTypeName(code string) string {
	for name, c := range deductTypeByName {
		if c == code {
			return name
		}
	}
	return code
}

func cellAt(row []string, i int) string {
	if i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}

// Import 导入专项附加扣除: 按 姓名+身份证 定位员工，(员工,年度,月份,扣除项) 已存在则覆盖金额
// （source=import），不存在则插入；任一行失败整体回滚（同 HRM-01 员工导入事务约定）。
func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	raw, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode, "导入文件内容不是合法的 base64")
	}

	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode, "导入文件必须是 xlsx 格式")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode, "导入文件读取失败")
	}
	if len(rows) < 2 {
		return nil, status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode, "导入文件没有数据行")
	}

	// 姓名 → 候选员工（同名单靠身份证区分）
	employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	type empRef struct{ uuid, idCard string }
	empByName := make(map[string][]empRef)
	for _, emp := range employees {
		empByName[emp.Name] = append(empByName[emp.Name], empRef{uuid: emp.Uuid, idCard: emp.IdCard})
	}

	var successCount int64
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		successCount = 0
		for _, row := range rows {
			name := cellAt(row, 0)
			if name == "" || name == importHeader[0] {
				continue // blank line or header
			}
			idCard := cellAt(row, 1)
			year := cellAt(row, 2)
			monthStr := cellAt(row, 3)
			typeName := cellAt(row, 4)
			amountStr := cellAt(row, 5)

			// 定位员工: 优先 姓名+身份证，同名单证号缺失则要求唯一
			var empUuid string
			for _, cand := range empByName[name] {
				if idCard != "" && cand.idCard == idCard {
					empUuid = cand.uuid
					break
				}
			}
			if empUuid == "" {
				cands := empByName[name]
				if len(cands) == 1 {
					empUuid = cands[0].uuid
				} else {
					return status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode,
						"无法唯一定位员工: "+name+"（请填写身份证号）")
				}
			}
			employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, empUuid)
			if err != nil {
				return status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode, "员工不存在: "+name)
			}

			code, ok := deductTypeByName[typeName]
			if !ok {
				return status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode,
					"扣除项不合法（须为 子女教育/继续教育/大病医疗/住房贷款利息/住房租金/赡养老人/3岁以下婴幼儿照护）: "+typeName)
			}
			month, err := strconv.ParseInt(monthStr, 10, 64)
			if err != nil || !validMonth(month) {
				return status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode,
					"月份必须在 1-12 之间: "+monthStr)
			}
			amount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil || !validAmount(amount) {
				return status.ErrorMessage(errcodes.TaxSpecialDeductionImportInvalidCode,
					"扣除金额必须是非负数字: "+amountStr)
			}
			if err = validateDeductionCommon(empUuid, year, month, code, amount); err != nil {
				return err
			}

			existing, err := l.svcCtx.Model.HrmTaxSpecialDeduction.FindOneByCondition(ctx, session,
				deductionUniqueCondition(empUuid, year, month, code)...)
			if err != nil && err != hrm_tax_special_deduction.ErrNotFound {
				return err
			}
			if existing != nil {
				if err = l.svcCtx.Model.HrmTaxSpecialDeduction.UpdateFieldsByCondition(ctx, session, map[string]any{
					string(hrm_tax_special_deduction.Amount):       amount,
					string(hrm_tax_special_deduction.Source):       SourceImport,
					string(hrm_tax_special_deduction.EmployeeName): employee.Name,
				}, deductionUuidCondition(existing.Uuid)...); err != nil {
					return err
				}
			} else {
				if _, err = l.svcCtx.Model.HrmTaxSpecialDeduction.Insert(ctx, session, &hrm_tax_special_deduction.HrmTaxSpecialDeduction{
					Uuid:         uuid.New().String(),
					EmployeeUuid: empUuid,
					EmployeeName: employee.Name,
					DeductYear:   year,
					DeductMonth:  month,
					DeductType:   code,
					Amount:       amount,
					Source:       SourceImport,
				}); err != nil {
					return err
				}
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
