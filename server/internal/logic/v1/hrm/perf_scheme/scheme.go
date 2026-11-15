package perf_scheme

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_template"
)

// 绩效方案 CRUD（HRM-41）：方案配置 = 关联模板 + 周期类型 + 评估模式 + 评分模式 + 联薪方式 + 启停。
// 模板引用有效性：template_uuid 必须指向 hrm_perf_template 中已存在的模板（HRM-40），否则拒绝保存。

// schemeTemplateExists reports whether templateUuid points at an existing 考核模板.
func schemeTemplateExists(ctx context.Context, conn sqlx.SqlConn, templateUuid string) (bool, error) {
	_, err := hrm_perf_template.NewHrmPerfTemplateModel(conn).FindOneByUuid(ctx, nil, templateUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_template.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// normalizeStatus defaults an empty status to 启用 ('1').
// Note: `default=1` api tags only apply on HTTP param parsing; logic tests and
// internal callers constructing AddRequest literals may pass "".
func normalizeStatus(status string) string {
	if status == "" {
		return "1"
	}
	return status
}

// AddScheme persists one 绩效方案 after validating enums + 模板引用有效性.
func AddScheme(ctx context.Context, conn sqlx.SqlConn, name, templateUuid, periodType, assessmentMode, scoreMode, payLinkMode, status, remark string) error {
	if err := ValidateScheme(name, templateUuid, periodType, assessmentMode, scoreMode, payLinkMode, status); err != nil {
		return err
	}
	status = normalizeStatus(status)
	ok, err := schemeTemplateExists(ctx, conn, templateUuid)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSchemeTemplateNotFound
	}
	return hrm_perf_scheme.NewHrmPerfSchemeModel(conn).InsertV2(ctx, nil, &hrm_perf_scheme.HrmPerfScheme{
		Uuid:           uuid.New().String(),
		Name:           name,
		TemplateUuid:   templateUuid,
		PeriodType:     periodType,
		AssessmentMode: assessmentMode,
		ScoreMode:      scoreMode,
		PayLinkMode:    payLinkMode,
		Status:         status,
		Remark:         remark,
	})
}

// EditScheme updates one 绩效方案 (enums + 模板引用 re-validated).
func EditScheme(ctx context.Context, conn sqlx.SqlConn, schemeUuid, name, templateUuid, periodType, assessmentMode, scoreMode, payLinkMode, status, remark string) error {
	if err := ValidateScheme(name, templateUuid, periodType, assessmentMode, scoreMode, payLinkMode, status); err != nil {
		return err
	}
	if status == "" {
		// 编辑时状态必填（枚举校验已拒绝非法值，此处兜底保持原值语义）
		return ErrSchemeInvalid
	}
	m := hrm_perf_scheme.NewHrmPerfSchemeModel(conn)
	row, err := m.FindOneByUuid(ctx, nil, schemeUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_scheme.ErrNotFound) {
			return ErrSchemeNotFound
		}
		return err
	}
	ok, err := schemeTemplateExists(ctx, conn, templateUuid)
	if err != nil {
		return err
	}
	if !ok {
		return ErrSchemeTemplateNotFound
	}
	row.Name = name
	row.TemplateUuid = templateUuid
	row.PeriodType = periodType
	row.AssessmentMode = assessmentMode
	row.ScoreMode = scoreMode
	row.PayLinkMode = payLinkMode
	row.Status = status
	row.Remark = remark
	return m.Update(ctx, nil, row)
}

// DeleteSchemes removes 绩效方案 rows by uuids.
func DeleteSchemes(ctx context.Context, conn sqlx.SqlConn, uuids []string) error {
	return hrm_perf_scheme.NewHrmPerfSchemeModel(conn).DeleteByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_scheme.Uuid, uuids).Build()...)
}

// GetSchemesByUuids loads 绩效方案 rows by uuids (helper for sibling domains, e.g. 考核批次).
func GetSchemesByUuids(ctx context.Context, conn sqlx.SqlConn, uuids []string) ([]*hrm_perf_scheme.HrmPerfScheme, error) {
	if len(uuids) == 0 {
		return nil, nil
	}
	return hrm_perf_scheme.NewHrmPerfSchemeModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_scheme.Uuid, uuids).Build()...)
}
