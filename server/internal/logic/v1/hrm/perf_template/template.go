package perf_template

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_indicator"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template_item"
)

// 考核模板 CRUD（HRM-40）：模板 + 模板内指标权重配置。
// 权重之和校验（=100%）见 template_validate.go；本文件负责持久化，
// 配置的写入在事务中整体替换（编辑时旧配置全部作废）。

// ItemWithIndicator is one 模板-指标 weight configuration enriched with the
// indicator snapshot from 指标库 (编码/名称/类型等，用于模板详情与考核单展示).
type ItemWithIndicator struct {
	Uuid          string
	IndicatorUuid string
	Weight        float64
	IndicatorCode string
	IndicatorName string
	IndicatorType string
	Unit          string
	ScoreRule     string
	DataSource    string
}

// TemplateDetail is one 考核模板 row plus its 指标权重配置.
type TemplateDetail struct {
	*hrm_perf_template.HrmPerfTemplate
	Items []ItemWithIndicator
}

// templateItemExists reports whether every configured 指标 exists in 指标库 —
// a dangling indicator uuid is a configuration error.
func templateItemExists(ctx context.Context, conn sqlx.SqlConn, items []TemplateItem) (bool, error) {
	uuids := make([]string, 0, len(items))
	for _, item := range items {
		uuids = append(uuids, item.IndicatorUuid)
	}
	count, err := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn).CountByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_indicator.Uuid, uuids).Build()...)
	if err != nil {
		return false, err
	}
	return int(count) == len(items), nil
}

// insertTemplateItems writes the 指标权重配置 rows inside a transaction session
// (model is built on conn; the session parameter carries the transaction).
func insertTemplateItems(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, templateUuid string, items []TemplateItem) error {
	m := hrm_perf_template_item.NewHrmPerfTemplateItemModel(conn)
	for _, item := range items {
		if err := m.InsertV2(ctx, session, &hrm_perf_template_item.HrmPerfTemplateItem{
			Uuid:          uuid.New().String(),
			TemplateUuid:  templateUuid,
			IndicatorUuid: item.IndicatorUuid,
			Weight:        item.Weight,
		}); err != nil {
			return err
		}
	}
	return nil
}

// AddTemplate persists one 考核模板 with its 指标权重配置 atomically.
// An empty item list is allowed (先存模板、后配置指标)；一旦配置了指标，
// 权重之和必须等于 100%（模板权重校验，见 template_validate.go）。
func AddTemplate(ctx context.Context, conn sqlx.SqlConn, name, status, remark string, items []TemplateItem) error {
	if err := ValidateTemplateItems(items); err != nil {
		return err
	}
	if len(items) > 0 {
		if ok, err := templateItemExists(ctx, conn, items); err != nil {
			return err
		} else if !ok {
			return ErrTemplateItemInvalid
		}
	}
	templateUuid := uuid.New().String()
	if err := conn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		if err := hrm_perf_template.NewHrmPerfTemplateModel(conn).InsertV2(ctx, session, &hrm_perf_template.HrmPerfTemplate{
			Uuid:   templateUuid,
			Name:   name,
			Status: status,
			Remark: remark,
		}); err != nil {
			return err
		}
		return insertTemplateItems(ctx, conn, session, templateUuid, items)
	}); err != nil {
		return err
	}
	return nil
}

// EditTemplate updates the template row and replaces its 指标权重配置
// atomically (旧配置整体作废). Weight rules identical to AddTemplate.
func EditTemplate(ctx context.Context, conn sqlx.SqlConn, templateUuid, name, status, remark string, items []TemplateItem) error {
	if err := ValidateTemplateItems(items); err != nil {
		return err
	}
	m := hrm_perf_template.NewHrmPerfTemplateModel(conn)
	row, err := m.FindOneByUuid(ctx, nil, templateUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_template.ErrNotFound) {
			return ErrTemplateNotFound
		}
		return err
	}
	if len(items) > 0 {
		if ok, err := templateItemExists(ctx, conn, items); err != nil {
			return err
		} else if !ok {
			return ErrTemplateItemInvalid
		}
	}
	row.Name = name
	row.Status = status
	row.Remark = remark
	return conn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		if err := hrm_perf_template.NewHrmPerfTemplateModel(conn).Update(ctx, session, row); err != nil {
			return err
		}
		if err := hrm_perf_template_item.NewHrmPerfTemplateItemModel(conn).DeleteByCondition(ctx, session,
			condition.NewChain().Equal(hrm_perf_template_item.TemplateUuid, templateUuid).Build()...); err != nil {
			return err
		}
		return insertTemplateItems(ctx, conn, session, templateUuid, items)
	})
}

// DeleteTemplates removes 考核模板 rows together with their 指标权重配置.
func DeleteTemplates(ctx context.Context, conn sqlx.SqlConn, uuids []string) error {
	return conn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		if err := hrm_perf_template.NewHrmPerfTemplateModel(conn).DeleteByCondition(ctx, session,
			condition.NewChain().In(hrm_perf_template.Uuid, uuids).Build()...); err != nil {
			return err
		}
		return hrm_perf_template_item.NewHrmPerfTemplateItemModel(conn).DeleteByCondition(ctx, session,
			condition.NewChain().In(hrm_perf_template_item.TemplateUuid, uuids).Build()...)
	})
}

// GetTemplateDetail returns one 考核模板 with its 指标权重配置
// (ErrTemplateNotFound when missing).
func GetTemplateDetail(ctx context.Context, conn sqlx.SqlConn, templateUuid string) (*TemplateDetail, error) {
	row, err := hrm_perf_template.NewHrmPerfTemplateModel(conn).FindOneByUuid(ctx, nil, templateUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_template.ErrNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	itemRows, err := hrm_perf_template_item.NewHrmPerfTemplateItemModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_template_item.TemplateUuid, templateUuid).
			OrderByAsc(hrm_perf_template_item.Id).Build()...)
	if err != nil {
		return nil, err
	}

	detail := &TemplateDetail{HrmPerfTemplate: row, Items: []ItemWithIndicator{}}
	if len(itemRows) == 0 {
		return detail, nil
	}

	indicatorUuids := make([]string, 0, len(itemRows))
	for _, item := range itemRows {
		indicatorUuids = append(indicatorUuids, item.IndicatorUuid)
	}
	indicators, err := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_indicator.Uuid, indicatorUuids).Build()...)
	if err != nil {
		return nil, err
	}
	byUuid := make(map[string]*hrm_perf_indicator.HrmPerfIndicator, len(indicators))
	for _, ind := range indicators {
		byUuid[ind.Uuid] = ind
	}

	for _, item := range itemRows {
		entry := ItemWithIndicator{
			Uuid:          item.Uuid,
			IndicatorUuid: item.IndicatorUuid,
			Weight:        item.Weight,
		}
		if ind, ok := byUuid[item.IndicatorUuid]; ok {
			entry.IndicatorCode = ind.Code
			entry.IndicatorName = ind.Name
			entry.IndicatorType = ind.IndicatorType
			entry.Unit = ind.Unit
			entry.ScoreRule = ind.ScoreRule
			entry.DataSource = ind.DataSource
		}
		detail.Items = append(detail.Items, entry)
	}
	return detail, nil
}

// FindTemplateItemIndicatorUuids exposes the 指标 uuids configured in one
// 模板 (helper for sibling domains, e.g. 考核任务展开).
func FindTemplateItemIndicatorUuids(ctx context.Context, conn sqlx.SqlConn, templateUuid string) ([]string, error) {
	itemRows, err := hrm_perf_template_item.NewHrmPerfTemplateItemModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_template_item.TemplateUuid, templateUuid).Build()...)
	if err != nil {
		return nil, err
	}
	uuids := make([]string, 0, len(itemRows))
	for _, item := range itemRows {
		uuids = append(uuids, item.IndicatorUuid)
	}
	return uuids, nil
}
