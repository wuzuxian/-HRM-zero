package perf_period

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 绩效批次核心逻辑（HRM-42）：CRUD + 适用范围圈人 + 生命周期状态机。
// 参与人表 hrm_perf_period_participant 是 HRM-43 考核任务链的取数契约
// （GetPeriodParticipants，见 hrm-notes/hrm-42-notes.md）。

// dateLayout is the 周期开始/结束/结果发布日期 format (YYYY-MM-DD).
const dateLayout = "2006-01-02"

// 批次类型（GLOSSARY/演示站 9.2：月度/季度）。
var PeriodTypes = map[string]struct{}{
	"month":   {}, // 月度
	"quarter": {}, // 季度
}

// 人员状态取值复用员工档案域（1=正式 2=试用期 3=实习生 4=临时工 5=离职 6=退休），
// 在册口径 = person_status NOT IN ('5','6')。
var validPersonStatuses = map[string]struct{}{
	"1": {}, "2": {}, "3": {}, "4": {}, "5": {}, "6": {},
}

// Scope is the 适用范围多维筛选条件（存 hrm_perf_period.scope JSON）。
// 空 slice 表示该维度不设限（全员口径下仅保留在册过滤）。
type Scope struct {
	CompanyUuids   []string `json:"companyUuids"`             // 公司（hrm_org org_type=1）
	OrgUuids       []string `json:"orgUuids"`                 // 部门（含全部下级组织）
	Positions      []string `json:"positions"`                // 职位（hrm_employee.position）
	Posts          []string `json:"posts"`                    // 职务（hrm_employee.post）
	PersonStatuses []string `json:"personStatuses"`           // 人员状态
}

// parseScope validates + decodes the scope JSON.
func parseScope(raw string) (*Scope, error) {
	var s Scope
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPeriodScopeInvalid, err)
	}
	for _, ps := range s.PersonStatuses {
		if _, ok := validPersonStatuses[ps]; !ok {
			return nil, fmt.Errorf("%w: 人员状态非法 %q", ErrPeriodScopeInvalid, ps)
		}
	}
	return &s, nil
}

// orgTreeDescendants resolves the given 部门 uuids to themselves plus ALL
// descendant org uuids（部门含下级）. One full-table scan of hrm_org + in-memory
// tree walk — same approach as the org-chart (HRM-02).
func orgTreeDescendants(ctx context.Context, conn sqlx.SqlConn, orgUuids []string) ([]string, error) {
	if len(orgUuids) == 0 {
		return nil, nil
	}
	rows, err := hrm_org.NewHrmOrgModel(conn).FindByCondition(ctx, nil, nil...)
	if err != nil {
		return nil, err
	}
	children := make(map[string][]string, len(rows))
	byUuid := make(map[string]*hrm_org.HrmOrg, len(rows))
	for _, row := range rows {
		children[row.ParentUuid] = append(children[row.ParentUuid], row.Uuid)
		byUuid[row.Uuid] = row
	}
	seen := make(map[string]struct{}, len(orgUuids))
	var walk func(uuid string)
	walk = func(uuid string) {
		if _, ok := seen[uuid]; ok {
			return
		}
		seen[uuid] = struct{}{}
		for _, child := range children[uuid] {
			walk(child)
		}
	}
	for _, u := range orgUuids {
		if _, exists := byUuid[u]; exists {
			walk(u)
		}
	}
	out := make([]string, 0, len(seen))
	for u := range seen {
		out = append(out, u)
	}
	return out, nil
}

// ScopeEmployeeUuids evaluates the 适用范围 against hrm_employee 在册员工
// (person_status NOT IN ('5','6')) with the multi-dim AND semantics:
// company_uuid IN companyUuids (if set) AND org_uuid IN orgTreeDescendants
// (if set) AND position IN positions (if set) AND post IN posts (if set) AND
// person_status IN personStatuses (if set).
func ScopeEmployeeUuids(ctx context.Context, conn sqlx.SqlConn, s *Scope) ([]string, error) {
	chain := condition.NewChain().
		// 在册口径: 排除离职('5')/退休('6')
		NotIn(hrm_employee.PersonStatus, []string{"5", "6"})

	if len(s.CompanyUuids) > 0 {
		chain = chain.In(hrm_employee.CompanyUuid, s.CompanyUuids)
	}
	if len(s.PersonStatuses) > 0 {
		chain = chain.In(hrm_employee.PersonStatus, s.PersonStatuses)
	}
	if len(s.Positions) > 0 {
		chain = chain.In(hrm_employee.Position, s.Positions)
	}
	if len(s.Posts) > 0 {
		chain = chain.In(hrm_employee.Post, s.Posts)
	}
	if len(s.OrgUuids) > 0 {
		orgs, err := orgTreeDescendants(ctx, conn, s.OrgUuids)
		if err != nil {
			return nil, err
		}
		chain = chain.In(hrm_employee.OrgUuid, orgs)
	}

	rows, err := hrm_employee.NewHrmEmployeeModel(conn).FindByCondition(ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	uuids := make([]string, 0, len(rows))
	for _, row := range rows {
		uuids = append(uuids, row.Uuid)
	}
	return uuids, nil
}

// RefreshParticipants 圈定参与人（按适用范围条件筛 hrm_employee 在册员工 →
// 批次参与人表全量覆盖，删全插幂等）。仅草稿期可执行。
func RefreshParticipants(ctx context.Context, conn sqlx.SqlConn, periodUuid string) (int64, error) {
	period, err := FindPeriodByUuid(ctx, conn, periodUuid)
	if err != nil {
		return 0, err
	}
	if period.Status != PeriodStatusDraft {
		return 0, ErrPeriodNotEditable
	}
	scope, err := parseScope(period.Scope)
	if err != nil {
		return 0, err
	}
	employeeUuids, err := ScopeEmployeeUuids(ctx, conn, scope)
	if err != nil {
		return 0, err
	}
	var replaced int64
	txErr := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var err error
		replaced, err = replaceParticipants(ctx, session, periodUuid, employeeUuids)
		return err
	})
	if txErr != nil {
		return 0, txErr
	}
	return replaced, nil
}

// replaceParticipants does the 删全插 binding replace (idempotent). Runs inside
// the caller's transaction — never opens its own.
func replaceParticipants(ctx context.Context, session sqlx.Session, periodUuid string, employeeUuids []string) (int64, error) {
	participantModel := hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(sqlx.NewSqlConnFromSession(session))

	// dedup while preserving order
	seen := make(map[string]struct{}, len(employeeUuids))
	deduped := make([]string, 0, len(employeeUuids))
	for _, u := range employeeUuids {
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		deduped = append(deduped, u)
	}

	if err := participantModel.DeleteByCondition(ctx, session,
		condition.NewChain().Equal(hrm_perf_period_participant.PeriodUuid, periodUuid).Build()...); err != nil {
		return 0, err
	}
	for _, u := range deduped {
		if err := participantModel.InsertV2(ctx, session, &hrm_perf_period_participant.HrmPerfPeriodParticipant{
			Uuid:         uuid.New().String(),
			PeriodUuid:   periodUuid,
			EmployeeUuid: u,
		}); err != nil {
			return 0, err
		}
	}
	return int64(len(deduped)), nil
}

// ---- 生命周期状态机 ----

// 生命周期（GLOSSARY「绩效批次」+ 演示站 9.2）：
// 0=草稿 → 1=进行中(发布) → 2=待确认 → 3=待校准 → 4=结果已发布 → 5=已发布。
// forward-only，终态不可再推进。
const (
	PeriodStatusDraft           = "0" // 草稿
	PeriodStatusOngoing         = "1" // 进行中
	PeriodStatusConfirming      = "2" // 待确认
	PeriodStatusCalibrating     = "3" // 待校准
	PeriodStatusResultPublished = "4" // 结果已发布
	PeriodStatusPublished       = "5" // 已发布
)

// nextLifecycleStatus maps 状态 → 下一状态（forward-only 状态机）。
var nextLifecycleStatus = map[string]string{
	PeriodStatusDraft:           PeriodStatusOngoing,
	PeriodStatusOngoing:         PeriodStatusConfirming,
	PeriodStatusConfirming:      PeriodStatusCalibrating,
	PeriodStatusCalibrating:     PeriodStatusResultPublished,
	PeriodStatusResultPublished: PeriodStatusPublished,
	// PeriodStatusPublished 终态：无后继
}

// FindPeriodByUuid loads one 批次; ErrPeriodNotFound when missing.
func FindPeriodByUuid(ctx context.Context, conn sqlx.SqlConn, uuid string) (*hrm_perf_period.HrmPerfPeriod, error) {
	row, err := hrm_perf_period.NewHrmPerfPeriodModel(conn).FindOneByUuid(ctx, nil, uuid)
	if err != nil {
		if errors.Is(err, hrm_perf_period.ErrNotFound) {
			return nil, ErrPeriodNotFound
		}
		return nil, err
	}
	return row, nil
}

// schemeExists reports whether schemeUuid points at an existing 绩效方案.
func schemeExists(ctx context.Context, conn sqlx.SqlConn, schemeUuid string) (bool, error) {
	_, err := hrm_perf_scheme.NewHrmPerfSchemeModel(conn).FindOneByUuid(ctx, nil, schemeUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_scheme.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// validatePeriod validates the shared Add/Edit body: 名称/批次类型/方案/周期日期/scope.
func validatePeriod(name, periodType, schemeUuid, scope, periodStart, periodEnd string) error {
	if name == "" {
		return ErrPeriodNameRequired
	}
	if schemeUuid == "" {
		return ErrPeriodSchemeRequired
	}
	if _, ok := PeriodTypes[periodType]; !ok {
		return fmt.Errorf("%w: 批次类型非法 %q", ErrPeriodInvalid, periodType)
	}
	if _, err := time.ParseInLocation(dateLayout, periodStart, time.Local); err != nil {
		return fmt.Errorf("%w: 周期开始格式非法（须为 YYYY-MM-DD）: %q", ErrPeriodInvalid, periodStart)
	}
	end, err := time.ParseInLocation(dateLayout, periodEnd, time.Local)
	if err != nil {
		return fmt.Errorf("%w: 周期结束格式非法（须为 YYYY-MM-DD）: %q", ErrPeriodInvalid, periodEnd)
	}
	start, _ := time.ParseInLocation(dateLayout, periodStart, time.Local)
	if !end.After(start) {
		return fmt.Errorf("%w: 周期结束必须晚于周期开始", ErrPeriodInvalid)
	}
	if _, err := parseScope(scope); err != nil {
		return err
	}
	return nil
}

// AddPeriod persists one 批次 in a transaction: 编号 PERF-年-流水 (bill_type_code='PERF',
// 日期=周期开始年份) + 草稿状态 + 按适用范围圈定初始参与人。
func AddPeriod(ctx context.Context, conn sqlx.SqlConn, name, periodType, schemeUuid, scope, periodStart, periodEnd, resultPublishDate, remark string) (string, error) {
	if err := validatePeriod(name, periodType, schemeUuid, scope, periodStart, periodEnd); err != nil {
		return "", err
	}
	ok, err := schemeExists(ctx, conn, schemeUuid)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrPeriodSchemeNotFound
	}
	if resultPublishDate != "" {
		if _, err := time.ParseInLocation(dateLayout, resultPublishDate, time.Local); err != nil {
			return "", fmt.Errorf("%w: 结果发布日期格式非法（须为 YYYY-MM-DD）: %q", ErrPeriodInvalid, resultPublishDate)
		}
	}

	var periodUuid string
	txErr := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		year, err := periodStartYear(periodStart)
		if err != nil {
			return err
		}
		periodNo, err := GeneratePeriodNo(ctx, session, year)
		if err != nil {
			return err
		}
		periodUuid = uuid.New().String()
		if err := hrm_perf_period.NewHrmPerfPeriodModel(sqlx.NewSqlConnFromSession(session)).InsertV2(ctx, session,
			&hrm_perf_period.HrmPerfPeriod{
				Uuid:              periodUuid,
				PeriodNo:          periodNo,
				Name:              name,
				PeriodType:        periodType,
				SchemeUuid:        schemeUuid,
				Scope:             scope,
				PeriodStart:       periodStart,
				PeriodEnd:         periodEnd,
				ResultPublishDate: resultPublishDate,
				Status:            PeriodStatusDraft,
				Remark:            remark,
			}); err != nil {
			return err
		}
		// 初始圈人（草稿期）；0 参与人不阻断新增（发布守卫兜底）
		s, perr := parseScope(scope)
		if perr != nil {
			return perr
		}
		employeeUuids, perr := ScopeEmployeeUuids(ctx, sqlx.NewSqlConnFromSession(session), s)
		if perr != nil {
			return perr
		}
		_, perr = replaceParticipants(ctx, session, periodUuid, employeeUuids)
		return perr
	})
	if txErr != nil {
		return "", txErr
	}
	return periodUuid, nil
}

// EditPeriod updates one 批次 (草稿期 only). Re-evaluates 参与人 after scope change.
func EditPeriod(ctx context.Context, conn sqlx.SqlConn, periodUuid, name, periodType, schemeUuid, scope, periodStart, periodEnd, resultPublishDate, remark string) error {
	if err := validatePeriod(name, periodType, schemeUuid, scope, periodStart, periodEnd); err != nil {
		return err
	}
	period, err := FindPeriodByUuid(ctx, conn, periodUuid)
	if err != nil {
		return err
	}
	if period.Status != PeriodStatusDraft {
		return ErrPeriodNotEditable
	}
	ok, err := schemeExists(ctx, conn, schemeUuid)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPeriodSchemeNotFound
	}
	period.Name = name
	period.PeriodType = periodType
	period.SchemeUuid = schemeUuid
	period.Scope = scope
	period.PeriodStart = periodStart
	period.PeriodEnd = periodEnd
	period.ResultPublishDate = resultPublishDate
	period.Remark = remark
	m := hrm_perf_period.NewHrmPerfPeriodModel(conn)
	if err := m.Update(ctx, nil, period); err != nil {
		return err
	}
	_, err = RefreshParticipants(ctx, conn, periodUuid)
	return err
}

// DeletePeriods removes 批次 rows by uuids (草稿期 only) together with their
// participant bindings.
func DeletePeriods(ctx context.Context, conn sqlx.SqlConn, uuids []string) error {
	if len(uuids) == 0 {
		return nil
	}
	m := hrm_perf_period.NewHrmPerfPeriodModel(conn)
	rows, err := m.FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_period.Uuid, uuids).Build()...)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Status != PeriodStatusDraft {
			return ErrPeriodNotEditable
		}
	}
	txErr := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := m.DeleteByCondition(ctx, session,
			condition.NewChain().In(hrm_perf_period.Uuid, uuids).Build()...); err != nil {
			return err
		}
		return hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(sqlx.NewSqlConnFromSession(session)).
			DeleteByCondition(ctx, session,
				condition.NewChain().In(hrm_perf_period_participant.PeriodUuid, uuids).Build()...)
	})
	return txErr
}

// PublishPeriod 发布批次（草稿→进行中）。守卫：发布前须已圈定参与人（>0）。
func PublishPeriod(ctx context.Context, conn sqlx.SqlConn, periodUuid string) error {
	return transitionPeriod(ctx, conn, periodUuid, PeriodStatusDraft, PeriodStatusOngoing)
}

// AdvancePeriod 生命周期推进一步（forward-only；终态已发布不可推进）。
func AdvancePeriod(ctx context.Context, conn sqlx.SqlConn, periodUuid string) error {
	period, err := FindPeriodByUuid(ctx, conn, periodUuid)
	if err != nil {
		return err
	}
	next, ok := nextLifecycleStatus[period.Status]
	if !ok {
		return ErrPeriodTransitionInvalid
	}
	return transitionPeriod(ctx, conn, periodUuid, period.Status, next)
}

// transitionPeriod guards + applies one lifecycle edge; participant guard
// applies only to the 发布 edge (草稿→进行中).
func transitionPeriod(ctx context.Context, conn sqlx.SqlConn, periodUuid, from, to string) error {
	m := hrm_perf_period.NewHrmPerfPeriodModel(conn)
	period, err := FindPeriodByUuid(ctx, conn, periodUuid)
	if err != nil {
		return err
	}
	if period.Status != from {
		return ErrPeriodTransitionInvalid
	}
	if from == PeriodStatusDraft {
		count, err := hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(conn).CountByCondition(ctx, nil,
			condition.NewChain().Equal(hrm_perf_period_participant.PeriodUuid, periodUuid).Build()...)
		if err != nil {
			return err
		}
		if count == 0 {
			return ErrPeriodNoParticipants
		}
	}
	period.Status = to
	return m.Update(ctx, nil, period)
}

// GetPeriodParticipants loads the 批次参与人 rows by 批次 uuid.
//
// 消费契约（HRM-43 考核任务链）：按批次参与人生成考核任务——以本函数返回的
// EmployeeUuid 列表为准（每行对应 hrm_employee.uuid，仅已圈定入批次的在册员工）。
// svcCtx 仅用到 Model/SqlxConn；rows 按插入顺序返回。
func GetPeriodParticipants(ctx context.Context, svcCtx *svc.ServiceContext, periodUuid string) ([]*hrm_perf_period_participant.HrmPerfPeriodParticipant, error) {
	conn := svcCtx.SqlxConn
	return hrm_perf_period_participant.NewHrmPerfPeriodParticipantModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_period_participant.PeriodUuid, periodUuid).Build()...)
}
