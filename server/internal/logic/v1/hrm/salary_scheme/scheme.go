package salary_scheme

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_scheme_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_scheme"
)

// 薪酬方案 CRUD + 方案项目配置（HRM-34）。
// 关联表 hrm_salary_scheme_item 存 方案×薪资项目 + 每项覆盖值/公式（可选，继承项目默认）——
// 不复制 item_code/item_name 快照（预置项目可被编辑，快照会让试算口径漂移，见 hrm-33-notes）。

// validPersonStatus hrm_employee.person_status 枚举: 1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休.
func validPersonStatus(v string) bool {
	switch v {
	case "1", "2", "3", "4", "5", "6":
		return true
	}
	return false
}

// validSettleCycle 结算周期: monthly=月度（演示站口径仅月度）.
func validSettleCycle(v string) bool {
	return v == "monthly"
}

// normalizeSchemeDefaults 兜底默认值（HTTP 层 default 标签之外，测试/内部调用同口径）.
func normalizeSchemeDefaults(settleCycle string, payDay int64, currency string) (string, int64, string) {
	if settleCycle == "" {
		settleCycle = "monthly"
	}
	if payDay == 0 {
		payDay = 1
	}
	if currency == "" {
		currency = "CNY"
	}
	return settleCycle, payDay, currency
}

// validateSchemeCommon 方案公共字段校验: 名称/人员状态/结算周期/发薪日/状态.
func validateSchemeCommon(schemeName string, personStatus []string, settleCycle string, payDay int64, statusStr string) error {
	if schemeName == "" {
		return status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "方案名称不能为空")
	}
	if len(personStatus) == 0 {
		return status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "适用人员状态不能为空")
	}
	for _, v := range personStatus {
		if !validPersonStatus(v) {
			return status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "适用人员状态取值非法")
		}
	}
	if !validSettleCycle(settleCycle) {
		return status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "结算周期仅支持 月度")
	}
	if payDay < 1 || payDay > 31 {
		return status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "发薪日须为 1-31 号")
	}
	if statusStr != "1" && statusStr != "0" {
		return status.ErrorMessage(errcodes.SalarySchemeInvalidCode, "薪酬方案状态必须是 启用 或 停用")
	}
	return nil
}

// schemeItemPayloadsToDedup 把入参项目配置去重并校验 item_uuid 存在；
// 返回按入参顺序去重后的配置行。重复勾选 → SalarySchemeItemDuplicateCode。
func validateSchemeItems(ctx context.Context, conn sqlx.SqlConn, payloads []types.SchemeItemPayload) ([]types.SchemeItemPayload, error) {
	if len(payloads) == 0 {
		return nil, nil
	}
	seen := make(map[string]bool, len(payloads))
	deduped := make([]types.SchemeItemPayload, 0, len(payloads))
	itemUuids := make([]string, 0, len(payloads))
	for _, p := range payloads {
		if p.ItemUuid == "" {
			return nil, status.ErrorMessage(errcodes.SalarySchemeItemNotFoundCode, "方案项目配置缺少 item_uuid")
		}
		if seen[p.ItemUuid] {
			return nil, status.ErrorMessage(errcodes.SalarySchemeItemDuplicateCode, "同一方案内薪资项目重复勾选")
		}
		seen[p.ItemUuid] = true
		deduped = append(deduped, p)
		itemUuids = append(itemUuids, p.ItemUuid)
	}

	count, err := hrm_salary_item.NewHrmSalaryItemModel(conn).CountByCondition(ctx, nil,
		condition.NewChain().In(hrm_salary_item.Uuid, itemUuids).Build()...)
	if err != nil {
		return nil, err
	}
	if int(count) != len(itemUuids) {
		return nil, status.ErrorMessage(errcodes.SalarySchemeItemNotFoundCode, "方案引用的薪资项目不存在")
	}
	return deduped, nil
}

// resolveSubject 校验主体存在并返回（名称快照随编辑更新；引用主体时不做启停状态检查，
// 停用主体允许存量方案只读展示 —— hrm-33-notes seam contract #1）。
func resolveSubject(ctx context.Context, conn sqlx.SqlConn, subjectUuid string) (*hrm_salary_subject.HrmSalarySubject, error) {
	if subjectUuid == "" {
		return nil, status.ErrorMessage(errcodes.SalarySchemeSubjectNotFoundCode, "关联薪酬主体不能为空")
	}
	subject, err := hrm_salary_subject.NewHrmSalarySubjectModel(conn).FindOneByUuid(ctx, nil, subjectUuid)
	if err != nil {
		if errors.Is(err, hrm_salary_subject.ErrNotFound) {
			return nil, status.ErrorMessage(errcodes.SalarySchemeSubjectNotFoundCode, "关联薪酬主体不存在")
		}
		return nil, err
	}
	return subject, nil
}

// insertSchemeItems 在事务 session 内写入方案项目配置行.
func insertSchemeItems(ctx context.Context, conn sqlx.SqlConn, session sqlx.Session, schemeUuid string, payloads []types.SchemeItemPayload) error {
	m := hrm_salary_scheme_item.NewHrmSalarySchemeItemModel(conn)
	for _, p := range payloads {
		var overrideAmount sql.NullFloat64
		if p.OverrideAmount != 0 {
			overrideAmount = sql.NullFloat64{Float64: p.OverrideAmount, Valid: true}
		}
		if err := m.InsertV2(ctx, session, &hrm_salary_scheme_item.HrmSalarySchemeItem{
			Uuid:            uuid.New().String(),
			SchemeUuid:      schemeUuid,
			ItemUuid:        p.ItemUuid,
			OverrideAmount:  overrideAmount,
			OverrideFormula: p.OverrideFormula,
		}); err != nil {
			return err
		}
	}
	return nil
}

// joinPersonStatus 多选人员状态 → 逗号分隔串（落库口径 "1,2"；空串不应发生，已由 validateSchemeCommon 拦截）.
func joinPersonStatus(values []string) string {
	return strings.Join(values, ",")
}

// schemeRowToRecord 方案行 → API record（item_count/completeness 由调用方填充）.
func schemeRowToRecord(row *hrm_salary_scheme.HrmSalaryScheme) types.HrmSalaryScheme {
	return types.HrmSalaryScheme{
		Uuid:        row.Uuid,
		SchemeName:  row.SchemeName,
		SubjectUuid: row.SubjectUuid,
		SubjectName: row.SubjectName,
		// PersonStatus 后端存逗号分隔串（"1,2"），原样透传给前端做 tag 展示。
		PersonStatus: row.PersonStatus,
		SettleCycle:  row.SettleCycle,
		PayDay:       row.PayDay,
		Currency:     row.Currency,
		Status:       row.Status,
		Remark:       row.Remark,
		CreateTime:   row.CreateTime.Format(time.DateTime),
		UpdateTime:   row.UpdateTime.Format(time.DateTime),
	}
}
