package perf_template

import (
	"errors"
	"testing"
)

// 权重之和校验（验收标准）：模板内配置指标与权重之和必须等于 100%。
func TestValidateTemplateItemsWeightSum(t *testing.T) {
	cases := []struct {
		name    string
		items   []TemplateItem
		wantErr error
	}{
		{
			name:    "权重之和等于100_通过",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 60}, {IndicatorUuid: "ind-2", Weight: 40}},
			wantErr: nil,
		},
		{
			name:    "小数权重之和等于100_通过",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 33.33}, {IndicatorUuid: "ind-2", Weight: 33.33}, {IndicatorUuid: "ind-3", Weight: 33.34}},
			wantErr: nil,
		},
		{
			name:    "权重之和大于100_拒绝",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 70}, {IndicatorUuid: "ind-2", Weight: 40}},
			wantErr: ErrTemplateItemWeightSum,
		},
		{
			name:    "权重之和小于100_拒绝",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 50}, {IndicatorUuid: "ind-2", Weight: 30}},
			wantErr: ErrTemplateItemWeightSum,
		},
		{
			name:    "浮点误差99.99999999_视为等于100_通过",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 33.333333}, {IndicatorUuid: "ind-2", Weight: 33.333333}, {IndicatorUuid: "ind-3", Weight: 33.333334}},
			wantErr: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTemplateItems(tc.items)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateTemplateItems(%v) = %v, want %v", tc.items, err, tc.wantErr)
			}
		})
	}
}

func TestValidateTemplateItemsItemRules(t *testing.T) {
	cases := []struct {
		name    string
		items   []TemplateItem
		wantErr error
	}{
		{
			name:    "空指标列表_允许保存草稿",
			items:   nil,
			wantErr: nil,
		},
		{
			name:    "指标uuid为空_拒绝",
			items:   []TemplateItem{{IndicatorUuid: "", Weight: 100}},
			wantErr: ErrTemplateItemInvalid,
		},
		{
			name:    "权重为0_拒绝",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 0}, {IndicatorUuid: "ind-2", Weight: 100}},
			wantErr: ErrTemplateItemInvalid,
		},
		{
			name:    "权重为负数_拒绝",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: -10}, {IndicatorUuid: "ind-2", Weight: 110}},
			wantErr: ErrTemplateItemInvalid,
		},
		{
			name:    "同一指标重复配置_拒绝",
			items:   []TemplateItem{{IndicatorUuid: "ind-1", Weight: 50}, {IndicatorUuid: "ind-1", Weight: 50}},
			wantErr: ErrTemplateItemInvalid,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTemplateItems(tc.items)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateTemplateItems(%v) = %v, want %v", tc.items, err, tc.wantErr)
			}
		})
	}
}
