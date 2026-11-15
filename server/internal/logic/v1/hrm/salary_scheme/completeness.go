package salary_scheme

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

// 配置完整度口径（HRM-34，6 项检查；demo-inventory 8.1）:
//
//   - 项目✓  schemeItems   : 方案已配置 ≥1 个薪资项目（hrm_salary_scheme_item 有行）— 真实校验
//   - 计薪✓  schemeCalc    : 已配置计算规则 — 真实校验（口径：方案项目构成中存在
//     value_mode=formula 的项目，或任一项目配置了覆盖公式/覆盖金额）
//   - 个税✓  schemeTax     : 真实校验（口径：方案项目构成中包含预置个税项目 TAX）
//   - 岗位✓  schemePosition: 未接入，恒 0（岗位挂接属 HRM-37 薪资档案的适用岗位配置，待建）
//   - 社保✓  schemeSi      : 未接入，恒 0（社保配置模块待建；SI_PERSONAL/SI_COMPANY 数据源未接）
//   - 考勤✓  schemeAttendance: 未接入，恒 0（考勤取数口径 ABSENCE_DEDUCTION 待 HRM-37 打通）
//
// 未接入项在前端展示为灰态「未接入」而非 ✗。
func completenessFor(ctx context.Context, conn sqlx.SqlConn, schemeUuid string) (types.HrmSalarySchemeCompleteness, int64, error) {
	completeness := types.HrmSalarySchemeCompleteness{}

	itemRows, err := hrm_salary_scheme_item.NewHrmSalarySchemeItemModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_salary_scheme_item.SchemeUuid, schemeUuid).Build()...)
	if err != nil {
		return completeness, 0, err
	}
	itemCount := int64(len(itemRows))

	if itemCount > 0 {
		completeness.SchemeItems = 1
	}

	if itemCount > 0 {
		itemUuids := make([]string, 0, len(itemRows))
		hasOverride := false
		for _, row := range itemRows {
			itemUuids = append(itemUuids, row.ItemUuid)
			if row.OverrideFormula != "" || (row.OverrideAmount.Valid && row.OverrideAmount.Float64 != 0) {
				hasOverride = true
			}
		}
		items, err := hrm_salary_item.NewHrmSalaryItemModel(conn).FindByCondition(ctx, nil,
			condition.NewChain().In(hrm_salary_item.Uuid, itemUuids).Build()...)
		if err != nil {
			return completeness, itemCount, err
		}
		byUuid := make(map[string]*hrm_salary_item.HrmSalaryItem, len(items))
		for _, item := range items {
			byUuid[item.Uuid] = item
		}
		for _, row := range itemRows {
			item, ok := byUuid[row.ItemUuid]
			if !ok {
				continue
			}
			if item.ValueMode == "formula" {
				completeness.SchemeCalc = 1
			}
			if item.ItemCode == taxItemCode {
				completeness.SchemeTax = 1
			}
		}
		if hasOverride {
			completeness.SchemeCalc = 1
		}
	}

	return completeness, itemCount, nil
}

// taxItemCode 预置个税项目编码（hrm-33 预置种子，消费侧 join 键，勿改）.
const taxItemCode = "TAX"
