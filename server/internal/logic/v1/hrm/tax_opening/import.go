package tax_opening

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
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/tax_opening"
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

// importHeader 导入/导出共用模板列（期初 7 个累计数 + 年度/截至月 + 主体）。
var importHeader = []string{
	"姓名", "身份证号", "薪酬主体", "年度", "截至月",
	"累计收入", "累计专项扣除", "累计专项附加扣除", "累计免税收入", "累计其他扣除",
	"累计应纳税所得额", "累计已缴税额", "备注",
}

func cellAt(row []string, i int) string {
	if i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}

func parseAmount(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// Import 导入个税累计期初: 按 姓名+身份证 定位员工、主体按名称定位（可空）；
// (员工,主体,年度) 已存在则整行覆盖，不存在则插入；任一行失败整体回滚。
func (l *Import) Import(req *types.ImportRequest) (resp *types.ImportResponse, err error) {
	raw, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "导入文件内容不是合法的 base64")
	}

	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "导入文件必须是 xlsx 格式")
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "导入文件读取失败")
	}
	if len(rows) < 2 {
		return nil, status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "导入文件没有数据行")
	}

	employees, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	type empRef struct{ uuid, idCard string }
	empByName := make(map[string][]empRef)
	for _, emp := range employees {
		empByName[emp.Name] = append(empByName[emp.Name], empRef{uuid: emp.Uuid, idCard: emp.IdCard})
	}

	subjects, err := l.svcCtx.Model.HrmSalarySubject.FindByCondition(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	subjectByName := make(map[string]string, len(subjects))
	for _, s := range subjects {
		subjectByName[s.SubjectName] = s.Uuid
	}

	var successCount int64
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		successCount = 0
		for _, row := range rows {
			name := cellAt(row, 0)
			if name == "" || name == importHeader[0] {
				continue
			}
			idCard := cellAt(row, 1)
			subjectName := cellAt(row, 2)
			year := cellAt(row, 3)
			untilMonthStr := cellAt(row, 4)

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
					return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode,
						"无法唯一定位员工: "+name+"（请填写身份证号）")
				}
			}
			employee, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, empUuid)
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "员工不存在: "+name)
			}

			subjectUuid := ""
			if subjectName != "" {
				su, ok := subjectByName[subjectName]
				if !ok {
					return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "薪酬主体不存在: "+subjectName)
				}
				subjectUuid = su
			}

			untilMonth, err := strconv.ParseInt(untilMonthStr, 10, 64)
			if err != nil || !validMonth(untilMonth) {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode,
					"截至月必须在 1-12 之间: "+untilMonthStr)
			}
			income, err := parseAmount(cellAt(row, 5))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计收入不是数字: "+cellAt(row, 5))
			}
			special, err := parseAmount(cellAt(row, 6))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计专项扣除不是数字: "+cellAt(row, 6))
			}
			extra, err := parseAmount(cellAt(row, 7))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计专项附加扣除不是数字: "+cellAt(row, 7))
			}
			exempt, err := parseAmount(cellAt(row, 8))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计免税收入不是数字: "+cellAt(row, 8))
			}
			other, err := parseAmount(cellAt(row, 9))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计其他扣除不是数字: "+cellAt(row, 9))
			}
			taxable, err := parseAmount(cellAt(row, 10))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计应纳税所得额不是数字: "+cellAt(row, 10))
			}
			paid, err := parseAmount(cellAt(row, 11))
			if err != nil {
				return status.ErrorMessage(errcodes.TaxOpeningImportInvalidCode, "累计已缴税额不是数字: "+cellAt(row, 11))
			}
			remark := cellAt(row, 12)
			if err = validateOpeningCommon(empUuid, year, untilMonth,
				income, special, extra, exempt, other, taxable, paid); err != nil {
				return err
			}

			existing, err := l.svcCtx.Model.HrmTaxOpening.FindOneByCondition(ctx, session,
				openingUniqueCondition(empUuid, subjectUuid, year)...)
			if err != nil && err != hrm_tax_opening.ErrNotFound {
				return err
			}
			fields := map[string]any{
				string(hrm_tax_opening.EmployeeUuid):          empUuid,
				string(hrm_tax_opening.EmployeeName):          employee.Name,
				string(hrm_tax_opening.SubjectUuid):           subjectUuid,
				string(hrm_tax_opening.SubjectName):           subjectName,
				string(hrm_tax_opening.TaxYear):               year,
				string(hrm_tax_opening.UntilMonth):            untilMonth,
				string(hrm_tax_opening.IncomeTotal):           income,
				string(hrm_tax_opening.SpecialDeductionTotal): special,
				string(hrm_tax_opening.SpecialExtraTotal):     extra,
				string(hrm_tax_opening.ExemptTotal):           exempt,
				string(hrm_tax_opening.OtherDeductionTotal):   other,
				string(hrm_tax_opening.TaxableIncomeTotal):    taxable,
				string(hrm_tax_opening.TaxPaidTotal):          paid,
				string(hrm_tax_opening.Remark):                remark,
			}
			if existing != nil {
				if err = l.svcCtx.Model.HrmTaxOpening.UpdateFieldsByCondition(ctx, session, fields,
					openingUuidCondition(existing.Uuid)...); err != nil {
					return err
				}
			} else {
				if _, err = l.svcCtx.Model.HrmTaxOpening.Insert(ctx, session, &hrm_tax_opening.HrmTaxOpening{
					Uuid:                  uuid.New().String(),
					EmployeeUuid:          fields[string(hrm_tax_opening.EmployeeUuid)].(string),
					EmployeeName:          employee.Name,
					SubjectUuid:           subjectUuid,
					SubjectName:           subjectName,
					TaxYear:               year,
					UntilMonth:            untilMonth,
					IncomeTotal:           income,
					SpecialDeductionTotal: special,
					SpecialExtraTotal:     extra,
					ExemptTotal:           exempt,
					OtherDeductionTotal:   other,
					TaxableIncomeTotal:    taxable,
					TaxPaidTotal:          paid,
					Remark:                remark,
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
