package perf_assessment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_indicator"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template_item"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 模板快照 + 权重汇总计分（HRM-43 issue #45 验收：权重汇总计分单测）。
// 快照在任务生成时固化（items_snapshot JSON），后续评分只依赖快照 — 模板后改不影响在途任务。

// ItemSnapshot is one 模板快照 item（items_snapshot JSON 数组元素）。
type ItemSnapshot struct {
	IndicatorUuid string  `json:"indicatorUuid"`
	IndicatorCode string  `json:"indicatorCode"`
	IndicatorName string  `json:"indicatorName"`
	IndicatorType string  `json:"indicatorType"`
	Unit          string  `json:"unit"`
	ScoreRule     string  `json:"scoreRule"`
	DataSource    string  `json:"dataSource"`
	Weight        float64 `json:"weight"`
}

// ScoreItem is one 逐项得分（自评/评审提交）。
type ScoreItem struct {
	IndicatorUuid string  `json:"indicatorUuid"`
	Score         float64 `json:"score"`
}

// parseItemsSnapshot decodes items_snapshot JSON.
func parseItemsSnapshot(raw string) ([]ItemSnapshot, error) {
	var items []ItemSnapshot
	if raw == "" {
		return items, nil
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("%w: items_snapshot %v", ErrAssessmentInvalid, err)
	}
	return items, nil
}

// encodeItemsSnapshot serialises the 模板快照.
func encodeItemsSnapshot(items []ItemSnapshot) (string, error) {
	buf, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// parseScoreItems decodes 逐项得分 JSON（self_items / review_items）。
func parseScoreItems(raw string) ([]ScoreItem, error) {
	var items []ScoreItem
	if raw == "" {
		return items, nil
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, fmt.Errorf("%w: score items %v", ErrAssessmentInvalid, err)
	}
	return items, nil
}

// encodeScoreItems serialises 逐项得分.
func encodeScoreItems(items []ScoreItem) (string, error) {
	buf, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// SnapshotFromTemplate loads the 考核模板 items+weights (hrm_perf_template_item joined
// with 指标库 hrm_perf_indicator) and freezes them into []ItemSnapshot.
// Same 取数口径 as perf_template.GetTemplateDetail (template items ordered by id,
// indicator rows by one IN query).
func SnapshotFromTemplate(ctx context.Context, svcCtx *svc.ServiceContext, templateUuid string) ([]ItemSnapshot, error) {
	itemRows, err := hrm_perf_template_item.NewHrmPerfTemplateItemModel(svcCtx.SqlxConn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_template_item.TemplateUuid, templateUuid).
			OrderByAsc(hrm_perf_template_item.Id).Build()...)
	if err != nil {
		return nil, err
	}
	if len(itemRows) == 0 {
		return nil, fmt.Errorf("%w: 模板无指标配置 %s", ErrAssessmentInvalid, templateUuid)
	}

	indicatorUuids := make([]string, 0, len(itemRows))
	for _, item := range itemRows {
		indicatorUuids = append(indicatorUuids, item.IndicatorUuid)
	}
	indicators, err := hrm_perf_indicator.NewHrmPerfIndicatorModel(svcCtx.SqlxConn).FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_indicator.Uuid, indicatorUuids).Build()...)
	if err != nil {
		return nil, err
	}
	byUuid := make(map[string]*hrm_perf_indicator.HrmPerfIndicator, len(indicators))
	for _, ind := range indicators {
		byUuid[ind.Uuid] = ind
	}

	snapshots := make([]ItemSnapshot, 0, len(itemRows))
	for _, item := range itemRows {
		snap := ItemSnapshot{
			IndicatorUuid: item.IndicatorUuid,
			Weight:        item.Weight,
		}
		if ind, ok := byUuid[item.IndicatorUuid]; ok {
			snap.IndicatorCode = ind.Code
			snap.IndicatorName = ind.Name
			snap.IndicatorType = ind.IndicatorType
			snap.Unit = ind.Unit
			snap.ScoreRule = ind.ScoreRule
			snap.DataSource = ind.DataSource
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, nil
}
