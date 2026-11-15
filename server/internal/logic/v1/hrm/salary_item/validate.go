package salary_item

import (
	"regexp"
	"time"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_item"
)

// 薪资项目（GLOSSARY「薪资项目」）枚举：
// 项目类型 item_type: fixed=固定项 data=取数项 formula=公式项
// 项目分类 item_class: earning=应发 deduction=应扣 cost=成本
// 取值方式 value_mode: manual=手工维护 system=系统取数 formula=公式计算
var (
	validItemTypes  = map[string]bool{"fixed": true, "data": true, "formula": true}
	validItemClasss = map[string]bool{"earning": true, "deduction": true, "cost": true}
	validValueModes = map[string]bool{"manual": true, "system": true, "formula": true}
)

// itemCodeRegexp 项目编码须为大写字母/数字/下划线（BASIC_SALARY/NET_SALARY…）.
var itemCodeRegexp = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

func validItemStatus(v string) bool {
	return v == "1" || v == "0"
}

func validateItemCode(code string) error {
	if !itemCodeRegexp.MatchString(code) {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "项目编码必须为大写字母/数字/下划线且以字母开头")
	}
	return nil
}

// validateItemCommon 新增/编辑共用的参数校验:
// 名称/编码/枚举（项目类型/项目分类/取值方式/启停）；公式项必须有公式表达式，
// 非公式项清空公式，取值方式 formula 要求项目类型为公式项（演示口径：value_mode=formula ⇒ item_type=formula）.
func validateItemCommon(itemCode, itemName, itemType, itemClass, valueMode, formula, statusStr string) error {
	if itemName == "" {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "项目名称不能为空")
	}
	if err := validateItemCode(itemCode); err != nil {
		return err
	}
	if !validItemTypes[itemType] {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "项目类型必须是 固定项/取数项/公式项")
	}
	if !validItemClasss[itemClass] {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "项目分类必须是 应发/应扣/成本")
	}
	if !validValueModes[valueMode] {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "取值方式必须是 手工维护/系统取数/公式计算")
	}
	if !validItemStatus(statusStr) {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "薪资项目状态必须是 启用 或 停用")
	}
	if valueMode == "formula" && itemType != "formula" {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "取值方式为公式计算时项目类型必须为公式项")
	}
	if itemType == "formula" && formula == "" {
		return status.ErrorMessage(errcodes.SalaryItemInvalidCode, "公式项必须填写公式表达式")
	}
	return nil
}

// normalizeFormula 非公式项强制清空公式表达式.
func normalizeFormula(itemType, formula string) string {
	if itemType != "formula" {
		return ""
	}
	return formula
}

// itemRowToRecord 薪资项目行 → API record（list/get_all 共用映射）.
func itemRowToRecord(row *hrm_salary_item.HrmSalaryItem) types.HrmSalaryItem {
	return types.HrmSalaryItem{
		Uuid:       row.Uuid,
		ItemCode:   row.ItemCode,
		ItemName:   row.ItemName,
		ItemType:   row.ItemType,
		ItemClass:  row.ItemClass,
		ValueMode:  row.ValueMode,
		CalcOrder:  row.CalcOrder,
		OrderNum:   row.OrderNum,
		Formula:    row.Formula,
		IsPreset:   row.IsPreset,
		Status:     row.Status,
		Remark:     row.Remark,
		CreateTime: row.CreateTime.Format(time.DateTime),
		UpdateTime: row.UpdateTime.Format(time.DateTime),
	}
}
