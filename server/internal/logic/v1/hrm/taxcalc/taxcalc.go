// Package taxcalc 个税累计预扣法纯函数（HRM-36 个税管理三件套）。
//
// 消费契约（HRM-37 薪资试算）: 试算引擎用 hrm_tax_opening（员工+主体+年度 唯一）的
// 期初累计数 + 本月数调用 CumulativeWithholding 计算本月应预扣个税；专项附加按月数用
// hrm_tax_special_deduction（员工+年度+月份+扣除项 唯一）。
// 本包不依赖 svc/model —— 纯函数，可独立单测。
package taxcalc

import "math"

// Bracket 累计预扣法税率表（居民个人工资薪金，个人所得税预扣率表一）。
// Level 级次 / Upper 累计应纳税所得额上限（含）/ Rate 税率 / QuickDeduction 速算扣除数（元）。
type Bracket struct {
	Level          int
	Upper          float64 // 全年累计应纳税所得额上限（最后一档 math.MaxFloat64）
	Rate           float64 // 0.03 / 0.10 / 0.20 / 0.25 / 0.30 / 0.35 / 0.45
	QuickDeduction float64 // 速算扣除数
}

// MonthlyBrackets 居民个人累计预扣率表（按年累计口径）:
// ≤36000 3% 速算扣除 0；≤144000 10% 2520；≤300000 20% 16920；
// ≤420000 25% 31920；≤660000 30% 52920；≤960000 35% 85920；>960000 45% 181920。
var MonthlyBrackets = []Bracket{
	{Level: 1, Upper: 36000, Rate: 0.03, QuickDeduction: 0},
	{Level: 2, Upper: 144000, Rate: 0.10, QuickDeduction: 2520},
	{Level: 3, Upper: 300000, Rate: 0.20, QuickDeduction: 16920},
	{Level: 4, Upper: 420000, Rate: 0.25, QuickDeduction: 31920},
	{Level: 5, Upper: 660000, Rate: 0.30, QuickDeduction: 52920},
	{Level: 6, Upper: 960000, Rate: 0.35, QuickDeduction: 85920},
	{Level: 7, Upper: math.MaxFloat64, Rate: 0.45, QuickDeduction: 181920},
}

// StandardMonthlyDeduction 基本减除费用（5000 元/月）。
const StandardMonthlyDeduction = 5000

// bracketFor 取累计应纳税所得额所在级次.
func bracketFor(taxable float64) Bracket {
	for _, b := range MonthlyBrackets {
		if taxable <= b.Upper {
			return b
		}
	}
	return MonthlyBrackets[len(MonthlyBrackets)-1]
}

// CumulativeWithholding 累计预扣法（单步）:
//
//	累计应纳税所得额 = 累计收入 - 累计免税收入 - 5000×月数 - 累计专项扣除
//	                   - 累计专项附加扣除 - 累计其他扣除
//	本期应预扣预缴税额 = (累计应纳税所得额×税率 - 速算扣除数) - 累计已缴税额
//
// 负值（含累计应纳税所得额为负、或减去已缴后为负）按 0 处理（不退税，余量结转以后月份）。
// 入参均为截至本月的累计值（期初 + 本月，由调用方合并）；months 为累计月数（含本月）。
func CumulativeWithholding(income, exempt, specialDeduction, specialExtra, otherDeduction, taxPaid float64, months int) float64 {
	if months < 1 {
		months = 1
	}
	taxable := income - exempt - StandardMonthlyDeduction*float64(months) - specialDeduction - specialExtra - otherDeduction
	if taxable <= 0 {
		return 0
	}
	b := bracketFor(taxable)
	tax := taxable*b.Rate - b.QuickDeduction
	owed := tax - taxPaid
	if owed < 0 {
		return 0
	}
	return round2(owed)
}

// TaxableIncome 返回累计应纳税所得额（可为负，供期初表落库口径参考）。
func TaxableIncome(income, exempt, specialDeduction, specialExtra, otherDeduction float64, months int) float64 {
	if months < 1 {
		months = 1
	}
	return income - exempt - StandardMonthlyDeduction*float64(months) - specialDeduction - specialExtra - otherDeduction
}

// round2 保留两位小数（四舍五入）。
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// SumSpecialExtra 按月汇总专项附加: dedByMonth[月份] = 该月 7 类扣除金额合计。
// 试算时对未维护的月份可用档案 default_special_extra 兜底（由调用方决定）。
func SumSpecialExtra(dedByMonth map[int64]float64, months int) float64 {
	total := 0.0
	for m := int64(1); m <= int64(months); m++ {
		total += dedByMonth[m]
	}
	return round2(total)
}
