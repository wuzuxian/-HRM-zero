package perf_assessment

import (
	"fmt"
	"math"
)

// 权重汇总计分 + 等级映射（HRM-43 issue #45 验收：权重汇总计分单测；纯函数）。

// WeightedTotal computes Σ(score_i × weight_i / 100) over the 模板快照.
// Returns an error when any snapshot item has no matching score（评分按模板 items
// 逐项、逐项得分须覆盖全部指标）。
func WeightedTotal(snapshot []ItemSnapshot, scores []ScoreItem) (float64, error) {
	byUuid := make(map[string]float64, len(scores))
	for _, s := range scores {
		byUuid[s.IndicatorUuid] = s.Score
	}
	total := 0.0
	for _, item := range snapshot {
		score, ok := byUuid[item.IndicatorUuid]
		if !ok {
			return 0, fmt.Errorf("%w: 缺少指标得分 %s(%s)", ErrAssessmentInvalid, item.IndicatorName, item.IndicatorCode)
		}
		if score < 0 {
			return 0, fmt.Errorf("%w: 指标得分不能为负 %s", ErrAssessmentInvalid, item.IndicatorCode)
		}
		total += score * item.Weight / 100
	}
	return round2(total), nil
}

// round2 rounds to 2 decimals (权重百分比 decimal(5,2) 精度口径).
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// PercentGrade maps 百分制最终得分 to 等级 A/B/C/D:
//
//	A: >=90  优秀
//	B: >=75  良好
//	C: >=60  合格
//	D: <60   待改进
//
// 等级阈值写死在本函数（demo 演示口径）；HRM-45/46/49 消费最终等级直接读 final_grade。
func PercentGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	default:
		return "D"
	}
}

// NormalizeResult returns (score, grade) for a 方案评分模式:
//
//	percent: score=权重汇总百分制，grade=PercentGrade(score)（显式传入的 grade 优先），
//	grade:   score=0，grade=提交原值直存（等级制不折算分数），
//	amount:  score=提交原值（绩效金额），grade 保持空。
//
// scoreMode is hrm_perf_scheme.score_mode（percent|grade|amount）。
func NormalizeResult(scoreMode string, score float64, grade string) (float64, string, error) {
	switch scoreMode {
	case "percent":
		if score < 0 {
			return 0, "", fmt.Errorf("%w: 百分制得分不能为负", ErrAssessmentInvalid)
		}
		if grade == "" {
			grade = PercentGrade(score)
		}
		return round2(score), grade, nil
	case "grade":
		if grade == "" {
			return 0, "", fmt.Errorf("%w: 等级制须提交等级", ErrAssessmentInvalid)
		}
		return 0, grade, nil
	case "amount":
		if score < 0 {
			return 0, "", fmt.Errorf("%w: 金额制得分不能为负", ErrAssessmentInvalid)
		}
		return round2(score), "", nil
	default:
		return 0, "", fmt.Errorf("%w: 方案评分模式非法 %s", ErrAssessmentInvalid, scoreMode)
	}
}
