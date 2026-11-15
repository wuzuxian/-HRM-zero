package perf_indicator

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_indicator"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template_item"
)

// 指标库（HRM-40，演示站 9.1）领域枚举与 sentinel errors。
var (
	// ErrIndicatorCodeDuplicated: 指标编码已存在。
	ErrIndicatorCodeDuplicated = errors.New("指标编码已存在")
	// ErrIndicatorNotFound: 指标不存在。
	ErrIndicatorNotFound = errors.New("指标不存在")
	// ErrIndicatorInvalid: 指标字段非法（编码/名称为空、类型或规则枚举越界、权重越界）。
	ErrIndicatorInvalid = errors.New("指标字段非法")
	// ErrIndicatorInUse: 指标已被考核模板引用，不能删除。
	ErrIndicatorInUse = errors.New("指标已被考核模板引用，不能删除")
)

// 指标类型：结果指标(KPI) / 行为能力。
const (
	IndicatorTypeKpi      = "kpi"
	IndicatorTypeBehavior = "behavior"
)

// 评分规则：公式评分 / 手工评分。
const (
	ScoreRuleFormula = "formula"
	ScoreRuleManual  = "manual"
)

// 数据来源：系统取数 / 手工填报。
const (
	DataSourceSystem = "system"
	DataSourceManual = "manual"
)

// 通用启用状态（与 manage 域一致）：1=启用 2=停用。
const (
	StatusEnabled  = "1"
	StatusDisabled = "2"
)

// Indicator is the 指标库 aggregate (指标编码/名称、指标类型、计量单位、
// 评分规则、数据来源、默认权重).
type Indicator struct {
	Uuid          string
	Code          string
	Name          string
	IndicatorType string // kpi=结果指标(KPI) behavior=行为能力
	Unit          string // 计量单位
	ScoreRule     string // formula=公式评分 manual=手工评分
	DataSource    string // system=系统取数 manual=手工填报
	DefaultWeight float64
	Status        string
	Remark        string
}

func validateIndicator(ind Indicator) error {
	if ind.Code == "" || ind.Name == "" {
		return ErrIndicatorInvalid
	}
	switch ind.IndicatorType {
	case IndicatorTypeKpi, IndicatorTypeBehavior:
	default:
		return ErrIndicatorInvalid
	}
	switch ind.ScoreRule {
	case ScoreRuleFormula, ScoreRuleManual:
	default:
		return ErrIndicatorInvalid
	}
	switch ind.DataSource {
	case DataSourceSystem, DataSourceManual:
	default:
		return ErrIndicatorInvalid
	}
	if ind.DefaultWeight < 0 || ind.DefaultWeight > 100 {
		return ErrIndicatorInvalid
	}
	switch ind.Status {
	case StatusEnabled, StatusDisabled:
	default:
		return ErrIndicatorInvalid
	}
	return nil
}

// AddIndicator inserts one 指标库 row; 指标编码 must stay unique.
func AddIndicator(ctx context.Context, conn sqlx.SqlConn, ind Indicator) error {
	if err := validateIndicator(ind); err != nil {
		return err
	}
	m := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn)
	_, err := m.FindOneByCode(ctx, nil, ind.Code)
	if err == nil {
		return ErrIndicatorCodeDuplicated
	} else if !errors.Is(err, hrm_perf_indicator.ErrNotFound) {
		return err
	}
	return m.InsertV2(ctx, nil, &hrm_perf_indicator.HrmPerfIndicator{
		Uuid:          uuid.New().String(),
		Code:          ind.Code,
		Name:          ind.Name,
		IndicatorType: ind.IndicatorType,
		Unit:          ind.Unit,
		ScoreRule:     ind.ScoreRule,
		DataSource:    ind.DataSource,
		DefaultWeight: ind.DefaultWeight,
		Status:        ind.Status,
		Remark:        ind.Remark,
	})
}

// EditIndicator updates one 指标库 row by uuid (编码唯一性仍校验).
func EditIndicator(ctx context.Context, conn sqlx.SqlConn, ind Indicator) error {
	if err := validateIndicator(ind); err != nil {
		return err
	}
	m := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn)
	old, err := m.FindOneByUuid(ctx, nil, ind.Uuid)
	if err != nil {
		if errors.Is(err, hrm_perf_indicator.ErrNotFound) {
			return ErrIndicatorNotFound
		}
		return err
	}
	if old.Code != ind.Code {
		_, err := m.FindOneByCode(ctx, nil, ind.Code)
		if err == nil {
			return ErrIndicatorCodeDuplicated
		} else if !errors.Is(err, hrm_perf_indicator.ErrNotFound) {
			return err
		}
	}
	old.Code = ind.Code
	old.Name = ind.Name
	old.IndicatorType = ind.IndicatorType
	old.Unit = ind.Unit
	old.ScoreRule = ind.ScoreRule
	old.DataSource = ind.DataSource
	old.DefaultWeight = ind.DefaultWeight
	old.Status = ind.Status
	old.Remark = ind.Remark
	return m.Update(ctx, nil, old)
}

// DeleteIndicators removes 指标库 rows by uuids. An indicator referenced by a
// 考核模板 (hrm_perf_template_item) must not be deleted — nothing is removed
// when any uuid is still in use.
func DeleteIndicators(ctx context.Context, conn sqlx.SqlConn, uuids []string) error {
	itemModel := hrm_perf_template_item.NewHrmPerfTemplateItemModel(conn)
	count, err := itemModel.CountByCondition(ctx, nil, condition.NewChain().
		In(hrm_perf_template_item.IndicatorUuid, uuids).
		Build()...)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrIndicatorInUse
	}
	indicatorModel := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn)
	return indicatorModel.DeleteByCondition(ctx, nil, condition.NewChain().
		In(hrm_perf_indicator.Uuid, uuids).
		Build()...)
}

// FindIndicatorByCode returns one indicator row by its unique 指标编码
// (ErrIndicatorNotFound when missing).
func FindIndicatorByCode(ctx context.Context, conn sqlx.SqlConn, code string) (*hrm_perf_indicator.HrmPerfIndicator, error) {
	row, err := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn).FindOneByCode(ctx, nil, code)
	if err != nil {
		if errors.Is(err, hrm_perf_indicator.ErrNotFound) {
			return nil, ErrIndicatorNotFound
		}
		return nil, err
	}
	return row, nil
}

// FindIndicatorByUuid returns one indicator row by uuid
// (ErrIndicatorNotFound when missing).
func FindIndicatorByUuid(ctx context.Context, conn sqlx.SqlConn, uuid string) (*hrm_perf_indicator.HrmPerfIndicator, error) {
	row, err := hrm_perf_indicator.NewHrmPerfIndicatorModel(conn).FindOneByUuid(ctx, nil, uuid)
	if err != nil {
		if errors.Is(err, hrm_perf_indicator.ErrNotFound) {
			return nil, ErrIndicatorNotFound
		}
		return nil, err
	}
	return row, nil
}
