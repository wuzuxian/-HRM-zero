package taxcalc

import (
	"math"
	"testing"
)

// TestCumulativeWithholdingWorkedExamples 用独立算例（手工算好的期望值）验证公式，
// 不用实现本身推期望（防套套逻辑）。
func TestCumulativeWithholdingWorkedExamples(t *testing.T) {
	cases := []struct {
		name string
		// 累计收入/免税/专项扣除/专项附加/其他扣除/已缴/月数
		income, exempt, special, extra, other, paid float64
		months                                      int
		want                                        float64
	}{
		// 年中入职期初为 0，1 月: 收入 8000, 三险一金 1600, 专项附加 1500
		// 应纳税所得额 = 8000-5000-1600-1500 = -100 → 0
		{"firstMonthBelowThreshold", 8000, 0, 1600, 1500, 0, 0, 1, 0},
		// 3 月累计: 收入 24000, 免税 0, 三险一金 4800, 附加 4500 → 24000-15000-4800-4500=-300 → 0
		{"threeMonthsStillNegative", 24000, 0, 4800, 4500, 0, 0, 3, 0},
		// 8 月累计: 收入 64000, 三险 12800, 附加 12000 → 64000-40000-12800-12000 = -800 → 0
		{"eightMonthsStillNegative", 64000, 0, 12800, 12000, 0, 0, 8, 0},
		// 9 月累计: 收入 72000 → 72000-45000-14400-13500 = -900 → 0
		{"nineMonthsBreakEven", 72000, 0, 14400, 13500, 0, 0, 9, 0},
		// 10 月累计: 收入 80000 → 80000-50000-16000-15000 = -1000? 80000-50000=30000-16000=14000-15000=-1000 → 0
		{"tenMonthsNegative", 80000, 0, 16000, 15000, 0, 0, 10, 0},
		// 12 月累计: 收入 96000, 三险 19200, 附加 18000 → 96000-60000-19200-18000 = -1200 → 0
		{"yearEndStillNegative", 96000, 0, 19200, 18000, 0, 0, 12, 0},
		// 高薪 1 月: 收入 30000, 三险 3000, 附加 2000 → 30000-5000-3000-2000=20000 → 20000×3%=600
		{"firstMonthTaxable", 30000, 0, 3000, 2000, 0, 0, 1, 600},
		// 高薪 2 月: 累计收入 60000, 三险 6000, 附加 4000, 已缴 600 → 60000-10000-6000-4000=40000
		// 40000×10%-2520=1480; 1480-600=880
		{"secondMonthSecondBracket", 60000, 0, 6000, 4000, 0, 600, 2, 880},
		// 免税收入参与减除: 收入 30000, 免税 5000, 附加 2000 → 30000-5000-5000-2000=18000 → 540
		{"exemptIncomeDeducted", 30000, 5000, 0, 2000, 0, 0, 1, 540},
		// 其他扣除参与减除: 收入 30000, 其他扣除 4000, 三险 3000 → 30000-5000-3000-4000=18000 → 540
		{"otherDeductionDeducted", 30000, 0, 3000, 0, 4000, 0, 1, 540},
		// 跳档补税: 累计收入 360000, 三险 24000, 附加 24000, 已缴 4000（前几个月低档）
		// taxable=360000-60000-24000-24000=252000 → 20% 档: 252000×0.2-16920=33480; 33480-4000=29480
		{"bracketJump", 360000, 0, 24000, 24000, 0, 4000, 12, 29480},
		// 已缴超过应缴（补税方向为退）→ 0，不退
		{"overPaidClampToZero", 30000, 0, 3000, 2000, 0, 1000, 1, 0},
		// months<1 按 1 个月兜底
		{"monthsClampToOne", 30000, 0, 3000, 2000, 0, 0, 0, 600},
		// 最高档: 收入 1,200,000（12 个月），基本减除 60000 → taxable=1,140,000
		// → 1140000×0.45-181920=331080
		{"topBracket", 1200000, 0, 0, 0, 0, 0, 12, 331080},
	}
	for _, c := range cases {
		got := CumulativeWithholding(c.income, c.exempt, c.special, c.extra, c.other, c.paid, c.months)
		if math.Abs(got-c.want) > 0.005 {
			t.Errorf("%s: CumulativeWithholding = %.2f, want %.2f", c.name, got, c.want)
		}
	}
}

// TestTaxableIncome 累计应纳税所得额可为负（期初表落库口径），与预扣税额分开。
func TestTaxableIncome(t *testing.T) {
	if got := TaxableIncome(8000, 0, 1600, 1500, 0, 1); got != -100 {
		t.Errorf("TaxableIncome negative case = %.2f, want -100", got)
	}
	if got := TaxableIncome(30000, 0, 3000, 2000, 0, 1); got != 20000 {
		t.Errorf("TaxableIncome positive case = %.2f, want 20000", got)
	}
}

// TestSumSpecialExtra 按月汇总 1..months 的专项附加合计（HRM-37 试算口径）。
func TestSumSpecialExtra(t *testing.T) {
	ded := map[int64]float64{1: 1500, 2: 1500, 3: 2000}
	if got := SumSpecialExtra(ded, 3); got != 5000 {
		t.Errorf("SumSpecialExtra = %.2f, want 5000", got)
	}
	// 未维护月份按 0 计
	if got := SumSpecialExtra(ded, 6); got != 5000 {
		t.Errorf("SumSpecialExtra with missing months = %.2f, want 5000", got)
	}
}

// TestBrackets 税率表边界：级次切换点归属（≤ 上限含等号）。
func TestBrackets(t *testing.T) {
	cases := []struct {
		taxable float64
		level   int
	}{
		{36000, 1}, {36000.01, 2}, {144000, 2}, {300000, 3},
		{420000, 4}, {660000, 5}, {960000, 6}, {960000.01, 7},
	}
	for _, c := range cases {
		if b := bracketFor(c.taxable); b.Level != c.level {
			t.Errorf("bracketFor(%.2f).Level = %d, want %d", c.taxable, b.Level, c.level)
		}
	}
}
