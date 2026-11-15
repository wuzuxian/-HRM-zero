package perf_template

import (
	"errors"
	"math"
)

// 考核模板-指标权重校验（HRM-40）。
// 规则（演示站 9.1 模板管理）：模板内配置指标与权重之和必须等于 100%；
// 每条配置必须指向一个指标库指标且权重大于 0，同一指标在一个模板内只能出现一次。
var (
	// ErrTemplateItemWeightSum: 模板指标权重之和必须等于 100%。
	ErrTemplateItemWeightSum = errors.New("模板指标权重之和必须等于100%")
	// ErrTemplateItemInvalid: 模板指标配置非法（指标为空/权重不大于0/指标重复/指标不存在）。
	ErrTemplateItemInvalid = errors.New("模板指标配置非法")
	// ErrTemplateNotFound: 考核模板不存在。
	ErrTemplateNotFound = errors.New("考核模板不存在")
)

// weightSumTolerance covers float rounding of two-decimal weights
// (33.33 + 33.33 + 33.34 等场景按两位小数取整后比较).
const weightSumTolerance = 0.005

// TemplateItem is one 考核模板 indicator configuration row (模板-指标-权重).
type TemplateItem struct {
	IndicatorUuid string
	Weight        float64 // 权重（百分比，如 40 表示 40%）
}

// ValidateTemplateItems validates a template's 指标权重配置.
// An empty item list is allowed (模板可先保存草稿、后进入编辑配置指标)；
// once items are present the weights must sum to exactly 100%.
// Pure function — persistence lives in AddTemplateItems/ReplaceTemplateItems.
func ValidateTemplateItems(items []TemplateItem) error {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	var sum float64
	for _, item := range items {
		if item.IndicatorUuid == "" || item.Weight <= 0 || item.Weight > 100 {
			return ErrTemplateItemInvalid
		}
		if _, dup := seen[item.IndicatorUuid]; dup {
			return ErrTemplateItemInvalid
		}
		seen[item.IndicatorUuid] = struct{}{}
		sum += item.Weight
	}
	if math.Abs(sum-100) > weightSumTolerance {
		return ErrTemplateItemWeightSum
	}
	return nil
}
