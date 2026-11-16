package salary_batch

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_archive"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/taxcalc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_punch_record"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch_item"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_opening"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_tax_special_deduction"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	archiveTypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_archive"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
)

// 试算明细行取数来源标记（source 列取值）。
const (
	SourceArchive       = "archive"         // 固定项=档案值（fixed_salary/fixed_allowance）
	SourceSchemeOverrid = "scheme_override" // 方案覆盖金额（OverrideAmount>0）
	SourceReward        = "reward"          // 奖惩单（synced=1 + approved + salary_month）
	SourceAttendance    = "attendance"      // 考勤月报/打卡日结兜底
	SourcePerf          = "perf"            // 绩效结果（final_score/final_grade 联薪）
	SourceTax           = "tax"             // 个税累计预扣引擎
	SourceSiMissing     = "si_missing"      // 社保配置未建，兜底 0
	SourceFormula       = "formula"         // 预置简单公式求值器
	SourceZero          = "zero"            // 手工维护项无档案值/金额=0
)

// perfCoefficientGradeMap 系数映射（hrm-43 契约：A→1.2 / B→1.0 / C→0.8 / D→0.6；
// 等级制非 A/B/C/D 值 → 1.0 默认，与 perf_assessment.PercentGrade 口径对齐）。
var perfCoefficientGradeMap = map[string]float64{
	"A": 1.2,
	"B": 1.0,
	"C": 0.8,
	"D": 0.6,
}

// monthReportSource 可选考勤月报数据源（HRM-23 in-flight：月报包合入后由 merger 通过
// SetMonthReportSource 注入；未注入时缺勤扣款按打卡日结 settle_status='missing' 天数兜底）。
type monthReportSource func(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string, orgUuids []string) ([]MonthReportRowLike, bool, error)

// MonthReportRowLike 考勤月报行适配接口（合入后适配 monthreport.MonthReportRow 即可）。
type MonthReportRowLike interface {
	EmployeeUuid() string
	AbsentDays() int64
}

// monthReportSrc 包级数据源槽位（默认 nil = 月报未接入，走打卡日结兜底）。
var monthReportSrc monthReportSource

// SetMonthReportSource 注入考勤月报数据源（merger 在 monthreport 包合入后于 routes 装配处调用一次；
// 见 hrm-37-notes 集成说明）。
func SetMonthReportSource(src monthReportSource) {
	monthReportSrc = src
}

// RunPayrollTrial 薪资试算引擎（HRM-37 核心；HRM-38 确认审批 / HRM-39 发薪登记与工资条
// 直接消费本函数产出的 hrm_salary_batch_item 明细行）。
//
// 批次员工集 = 全量在职员工（有部门归属）中 GetEffectiveArchiveForEmployee(yearMonth) 命中者
//（HRM-35：停用档案不返回=不参与试算；无档案员工跳过不入批）。
//
// 逐人按 calc_order ASC 消费项目清单（hrm-33 契约 #1）：
//   - fixed/manual 项：OverrideAmount>0 用覆盖值，否则 BASIC_SALARY/POST_SALARY←档案 fixed_salary、
//     *ALLOWANCE←档案 fixed_allowance，其他兜底 0；
//   - data/system 项按 item_code 取数：REWARD_BONUS/REWARD_PENALTY←奖惩单、
//     ABSENCE_DEDUCTION←考勤月报（未接入→打卡日结缺卡天兜底）、MERIT_SALARY←绩效联薪、
//     SI/HF←社保配置（未建，兜底 0 + 告警）、TAX←个税累计预扣引擎；
//   - formula 项：预置简单求值器（仅 NET_SALARY=应发合计−应扣合计 与
//     LABOR_COST=应发合计+公司社保+公司公积金；不支持任意表达式，见 notes）。
//
// 产出：每人明细行（项目/金额/来源标记）→ 批次汇总（人数/应发/应扣/实发）+ 告警文本。
// 幂等：重算整体覆盖旧明细（事务内全删全插），批次 version+1、trial_time=now、试算标记置 1
//（确认/发薪/工资条进度归 HRM-38/39，本函数不动）。
func RunPayrollTrial(ctx context.Context, svcCtx *svc.ServiceContext, batchUuid string) (*types.TrialResponse, error) {
	batch, err := findBatchByUuid(ctx, svcCtx, batchUuid)
	if err != nil {
		return nil, err
	}
	if _, err := findSubjectByUuid(ctx, svcCtx, batch.SubjectUuid); err != nil {
		return nil, err
	}
	yearMonth := fmt.Sprintf("%04d-%02d", batch.BatchYear, batch.BatchMonth)

	// ---- 1. 圈定批次员工集 ----
	employees, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
		condition.NewChain().NotEqual(hrm_employee.OrgUuid, "").Build()...)
	if err != nil {
		return nil, err
	}
	sort.Slice(employees, func(i, j int) bool {
		if employees[i].EmployeeNo != employees[j].EmployeeNo {
			return employees[i].EmployeeNo < employees[j].EmployeeNo
		}
		return employees[i].Uuid < employees[j].Uuid
	})

	// 跨域数据源一次性加载（按人索引）
	rewards, err := loadRewardSums(ctx, svcCtx, yearMonth)
	if err != nil {
		return nil, err
	}
	absentDays, monthLocked, err := loadAbsentDays(ctx, svcCtx, yearMonth)
	if err != nil {
		return nil, err
	}
	perf, err := loadPerfResults(ctx, svcCtx, yearMonth)
	if err != nil {
		return nil, err
	}
	taxData, err := loadTaxData(ctx, svcCtx, yearMonth)
	if err != nil {
		return nil, err
	}
	orgNames, err := loadOrgNames(ctx, svcCtx)
	if err != nil {
		return nil, err
	}

	warnings := newWarningCollector()

	// ---- 2. 逐人试算 ----
	type personResult struct {
		emp   *hrm_employee.HrmEmployee
		items []*types.HrmSalaryBatchItem
	}
	var persons []personResult
	for _, emp := range employees {
		eff, err := salary_archive.GetEffectiveArchiveForEmployee(ctx, svcCtx, emp.Uuid, yearMonth)
		if err != nil {
			if isArchiveNotFound(err) {
				continue // 无有效档案 → 不入批（HRM-35 契约）
			}
			return nil, err
		}
		// 主体隔离：HRM-35 取档按员工维度（不筛主体，跨主体可并行档案），批次员工集
		// 只收 生效档案主体=批次主体 的员工（票面「主体下 GetEffectiveArchiveForEmployee 命中」）。
		if eff.Archive.SubjectUuid != batch.SubjectUuid {
			continue
		}
		items, perr := trialPerson(emp, eff, yearMonth, trialInputs{
			rewards:     rewards,
			absentDays:  absentDays,
			monthLocked: monthLocked,
			perf:        perf,
			tax:         taxData,
			warnings:    warnings,
		})
		if perr != nil {
			return nil, perr
		}
		persons = append(persons, personResult{emp: emp, items: items})
	}

	if len(persons) == 0 {
		return nil, statusTrialFailed("试算失败：该主体下没有任何员工命中有效薪资档案（先在员工薪资档案建档）")
	}

	// ---- 3. 汇总 ----
	var earningTotal, deductionTotal, netTotal float64
	var allItems []*hrm_salary_batch_item.HrmSalaryBatchItem
	for _, p := range persons {
		for _, it := range p.items {
			switch it.ItemClass {
			case "earning":
				earningTotal += it.Amount
			case "deduction":
				deductionTotal += it.Amount
			case "cost":
				// 成本类不进应发/应扣（LABOR_COST 是管理口径）
			}
			if it.ItemCode == ItemCodeNetSalary {
				netTotal += it.Amount
			}
			allItems = append(allItems, &hrm_salary_batch_item.HrmSalaryBatchItem{
				Uuid:         uuid.New().String(),
				BatchUuid:    batch.Uuid,
				EmployeeUuid: p.emp.Uuid,
				EmployeeNo:   p.emp.EmployeeNo,
				EmployeeName: p.emp.Name,
				OrgName:      orgNames[p.emp.OrgUuid],
				ItemUuid:     it.ItemUuid,
				ItemCode:     it.ItemCode,
				ItemName:     it.ItemName,
				ItemClass:    it.ItemClass,
				ItemType:     it.ItemType,
				Amount:       it.Amount,
				Source:       it.Source,
				SourceDetail: it.SourceDetail,
				CalcOrder:    it.CalcOrder,
			})
		}
	}
	warnings.addIf(!monthLocked, "考勤月报未锁定：缺勤扣款按实时打卡日结兜底取数，锁定月报后建议重新试算")
	warningText := warnings.text()

	// ---- 4. 事务落库：汇总行更新 + 明细全删全插（幂等重算覆盖） ----
	now := time.Now()
	newVersion := batch.Version + 1
	txErr := svcCtx.SqlxConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		batchModel := hrm_salary_batch.NewHrmSalaryBatchModel(sqlx.NewSqlConnFromSession(session))
		if err := batchModel.UpdateFieldsByCondition(ctx, session, map[string]any{
			"headcount":       len(persons),
			"earning_total":   round2(earningTotal),
			"deduction_total": round2(deductionTotal),
			"net_total":       round2(netTotal),
			"trial_status":    "1",
			"trial_time":      now,
			"version":         newVersion,
			"warnings":        warningText,
		}, condition.NewChain().Equal(hrm_salary_batch.Uuid, batch.Uuid).Build()...); err != nil {
			return err
		}
		itemModel := hrm_salary_batch_item.NewHrmSalaryBatchItemModel(sqlx.NewSqlConnFromSession(session))
		if err := itemModel.DeleteByCondition(ctx, session,
			condition.NewChain().Equal(hrm_salary_batch_item.BatchUuid, batch.Uuid).Build()...); err != nil {
			return err
		}
		if len(allItems) > 0 {
			if err := itemModel.BulkInsert(ctx, session, allItems); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	logx.WithContext(ctx).Infof("payroll trial batch=%s version=%d headcount=%d earning=%.2f deduction=%.2f net=%.2f",
		batch.BatchNo, newVersion, len(persons), earningTotal, deductionTotal, netTotal)

	return &types.TrialResponse{
		BatchUuid:      batch.Uuid,
		BatchNo:        batch.BatchNo,
		Headcount:      int64(len(persons)),
		EarningTotal:   round2(earningTotal),
		DeductionTotal: round2(deductionTotal),
		NetTotal:       round2(netTotal),
		Version:        newVersion,
		Warnings:       warningText,
	}, nil
}

// trialInputs 逐人试算的跨域数据源输入（一次性加载、按人复用）。
type trialInputs struct {
	rewards     map[string]rewardSum
	absentDays  map[string]float64
	monthLocked bool
	perf        map[string]perfResult
	tax         map[string]*taxMonthData
	warnings    *warningCollector
}

// trialPerson 单人试算：按 calc_order ASC 消费项目清单，产出明细行。
func trialPerson(emp *hrm_employee.HrmEmployee, eff *archiveTypes.GetEffectiveArchiveResponse,
	yearMonth string, in trialInputs) ([]*types.HrmSalaryBatchItem, error) {
	items := make([]*types.HrmSalaryBatchItem, 0, len(eff.Items))

	var earningSum, deductionSum float64
	var siCompany, hfCompany float64 // 公司社保/公积金（社保未建恒 0）

	for idx := range eff.Items {
		it := eff.Items[idx]
		if it.Status == "0" {
			// 停用项目容忍（hrm-33 契约 #4）：跳过计薪并记录告警
			in.warnings.addDedup(fmt.Sprintf("员工 %s：薪资项目 %s 已停用，试算按 0 处理", emp.EmployeeNo, it.ItemName))
			continue
		}

		amount, source, detail := resolveItemAmount(emp, &it, eff.Archive, yearMonth, in, &earningSum, &deductionSum, &siCompany, &hfCompany)		// 公式项：TAX 是预置公式型个税项目（formula/deduction），其取数走个税引擎而非表达式；
		// NET_SALARY/LABOR_COST 走预置求值器。
		if it.ItemType == "formula" || it.ValueMode == "formula" {
			if it.ItemCode == ItemCodeTax {
				amount, source, detail = resolveTax(emp, in, &earningSum, &deductionSum)
			} else {
				amount, detail = evalFormulaItem(it.ItemCode, earningSum, deductionSum, siCompany, hfCompany, it.OverrideFormula)
				source = SourceFormula
			}
		}

		amount = round2(amount)
		switch it.ItemClass {
		case "earning":
			earningSum += amount
		case "deduction":
			deductionSum += amount
		case "cost":
			// 成本类不计入应发/应扣
		}
		items = append(items, &types.HrmSalaryBatchItem{
			EmployeeUuid: emp.Uuid,
			ItemUuid:     it.ItemUuid,
			ItemCode:     it.ItemCode,
			ItemName:     it.ItemName,
			ItemClass:    it.ItemClass,
			ItemType:     it.ItemType,
			Amount:       amount,
			Source:       source,
			SourceDetail: detail,
			CalcOrder:    it.CalcOrder,
		})
	}
	return items, nil
}

// resolveItemAmount 单项取数（fixed/manual/data 项；formula 项由调用方后置处理）。
func resolveItemAmount(emp *hrm_employee.HrmEmployee, it *archiveTypes.HrmEffectiveItem,
	archive archiveTypes.HrmSalaryArchive, yearMonth string, in trialInputs,
	earningSum, deductionSum, siCompany, hfCompany *float64) (float64, string, string) {

	// 方案覆盖金额优先（HRM-34 契约：fixed/manual 项 OverrideAmount>0 覆盖）
	if it.OverrideAmount > 0 {
		return it.OverrideAmount, SourceSchemeOverrid, "方案覆盖金额"
	}

	switch {
	case it.ItemType == "fixed":
		// 固定项：BASIC_SALARY/POST_SALARY←档案固定薪资；*ALLOWANCE←档案固定补贴；其余兜底 0
		if it.ItemCode == "BASIC_SALARY" || it.ItemCode == "POST_SALARY" {
			return archive.FixedSalary, SourceArchive, "档案固定薪资"
		}
		if strings.Contains(it.ItemCode, "ALLOWANCE") {
			return archive.FixedAllowance, SourceArchive, "档案固定补贴"
		}
		return 0, SourceZero, "固定项无档案值兜底 0"

	case it.ValueMode == "system":
		return resolveSystemItem(emp, it, yearMonth, in, archive.FixedSalary, earningSum, deductionSum, siCompany, hfCompany)
	}

	// manual 取数项（无覆盖金额）：兜底 0
	return 0, SourceZero, "手工维护项无覆盖金额兜底 0"
}

// resolveSystemItem system 取数项按 item_code 数据源约定取数（hrm-33 契约 #2）。
func resolveSystemItem(emp *hrm_employee.HrmEmployee, it *archiveTypes.HrmEffectiveItem,
	yearMonth string, in trialInputs, meritBase float64,
	earningSum, deductionSum, siCompany, hfCompany *float64) (float64, string, string) {

	switch it.ItemCode {
	case ItemCodeRewardBonus:
		if r, ok := in.rewards[emp.Uuid]; ok && r.bonus > 0 {
			return r.bonus, SourceReward, fmt.Sprintf("奖惩单 %d 笔合计", r.bonusCount)
		}
		return 0, SourceReward, "当月无奖励单"

	case ItemCodeRewardPenalty:
		if r, ok := in.rewards[emp.Uuid]; ok && r.penalty > 0 {
			return r.penalty, SourceReward, fmt.Sprintf("奖惩单 %d 笔合计", r.penaltyCount)
		}
		return 0, SourceReward, "当月无处分单"

	case ItemCodeAbsenceDeduction:
		days := in.absentDays[emp.Uuid]
		if days <= 0 {
			return 0, SourceAttendance, "当月无缺勤"
		}
		base := *earningSum
		if base <= 0 {
			return 0, SourceAttendance, "缺勤但应发基数≤0，按 0 兜底"
		}
		totalDays := daysInMonthOf(yearMonth)
		if !in.monthLocked {
			in.warnings.addDedup(fmt.Sprintf("员工 %s：考勤月报未锁定，缺勤扣款按实时打卡兜底（缺勤 %.0f 天）", emp.EmployeeNo, days))
		}
		perDay := base / float64(totalDays)
		return perDay * days, SourceAttendance,
			fmt.Sprintf("缺勤 %.0f 天 × 日薪 %.2f（应发基数 %.2f / 当月 %d 天）", days, perDay, base, totalDays)

	case ItemCodeMeritSalary:
		// 绩效基数 = 档案固定薪资（HRM-43 notes「HRM-49 取薪资档案基数」同款口径）；
		// 方案覆盖金额优先在 resolveItemAmount 已处理，此处 it.OverrideAmount 兜底同样生效。
		return resolveMeritSalary(emp, it, in, meritBase)

	case ItemCodeSiPersonal, ItemCodeHfPersonal, ItemCodeSiCompany, ItemCodeHfCompany:
		// 社保模块未建：兜底 0 + 去重告警
		in.warnings.addDedup("社保模块未接入：" + it.ItemName + " 按 0 兜底")
		return 0, SourceSiMissing, "社保未配置兜底 0"

	case ItemCodeTax:
		return resolveTax(emp, in, earningSum, deductionSum)
	}

	return 0, SourceZero, "未知取数项编码兜底 0"
}

// resolveMeritSalary 绩效联薪（hrm-43 契约）：
//   - pay_link_mode=amount：绩效金额 = final_score 直取；
//   - pay_link_mode=coefficient（及未知模式兜底）：绩效金额 = 系数(A1.2/B1.0/C0.8/D0.6) × 绩效基数；
//     基数 = 本行 OverrideAmount>0 用覆盖金额，否则=档案固定薪资（HRM-43 notes 同款口径）。
//   - 批次月份无已完成考核任务：兜底 0 + 告警。
func resolveMeritSalary(emp *hrm_employee.HrmEmployee, it *archiveTypes.HrmEffectiveItem, in trialInputs, meritBase float64) (float64, string, string) {
	pr, ok := in.perf[emp.Uuid]
	if !ok {
		in.warnings.addDedup("员工 " + emp.EmployeeNo + "：批次月份无已确认的绩效考核结果，绩效工资按 0 兜底")
		return 0, SourcePerf, "无绩效结果兜底 0"
	}
	switch pr.payLinkMode {
	case "amount":
		return pr.finalScore, SourcePerf, fmt.Sprintf("绩效金额制直取 final_score=%.2f（%s）", pr.finalScore, pr.assessmentNo)
	default:
		base := it.OverrideAmount
		if base <= 0 {
			base = meritBase
		}
		coefficient := perfCoefficientGradeMap[pr.finalGrade]
		if coefficient == 0 {
			coefficient = 1.0
		}
		return coefficient * base, SourcePerf,
			fmt.Sprintf("等级 %s 系数 %.1f × 绩效基数 %.2f（%s）", pr.finalGrade, coefficient, base, pr.assessmentNo)
	}
}

// resolveTax 个税取数：累计预扣法（hrm-36 契约 #1）。
// 期初 = hrm_tax_opening（员工+年度唯一；截至月 < 试算月 → 直接并入累计；≥ 试算月按 0）；
// 本月收入 = 当前累计应发（earningSum）+ 期初累计收入；
// 专项附加 = 期初累计 + 试算月当月维护值合计。
func resolveTax(emp *hrm_employee.HrmEmployee, in trialInputs, earningSum, deductionSum *float64) (float64, string, string) {
	td, ok := in.tax[emp.Uuid]
	if !ok {
		in.warnings.addDedup("员工 " + emp.EmployeeNo + "：无个税期初数据，个税按 0 兜底")
		return 0, SourceTax, "无个税期初兜底 0"
	}
	income := *earningSum + td.incomeOpening
	special := td.specialDeductionOpening + td.specialExtraOpening
	other := td.otherDeductionOpening
	tax := taxcalc.CumulativeWithholding(income, 0, special, td.specialExtraThisMonth, other, td.taxPaidOpening, td.months+1)
	if tax <= 0 {
		return 0, SourceTax, fmt.Sprintf("累计预扣本期应缴 0（累计收入 %.2f 期初已缴 %.2f）", income, td.taxPaidOpening)
	}
	return tax, SourceTax, fmt.Sprintf("累计预扣法（含期初 %d 月）", td.months)
}

// ---- 跨域数据源加载 ----

type rewardSum struct {
	bonus        float64
	bonusCount   int
	penalty      float64
	penaltyCount int
}

// loadRewardSums 奖惩数据源（hrm-07 契约）：synced=1 + status='approved' + salary_month=试算月。
func loadRewardSums(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (map[string]rewardSum, error) {
	rows, err := svcCtx.Model.HrmRewardBill.FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_reward_bill.Synced, 1).
			Equal(hrm_reward_bill.Status, "approved").
			Equal(hrm_reward_bill.SalaryMonth, yearMonth).
			Build()...)
	if err != nil {
		return nil, err
	}
	out := make(map[string]rewardSum, len(rows))
	for _, r := range rows {
		s := out[r.EmployeeUuid]
		if r.RewardType == "reward" {
			s.bonus += r.Amount
			s.bonusCount++
		} else if r.RewardType == "punishment" {
			s.penalty += r.Amount // 处分金额以正数存，应扣取正数
			s.penaltyCount++
		}
		out[r.EmployeeUuid] = s
	}
	return out, nil
}

// loadAbsentDays 缺勤天数（按人）：HRM-23 月报数据源注入时走月报（锁定月=快照，未锁定=实时聚合）；
// 未注入（本票合入时点月报 in-flight）→ 打卡日结 settle_status='missing' 天数兜底（hrm-16 契约）。
func loadAbsentDays(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (map[string]float64, bool, error) {
	if monthReportSrc != nil {
		rows, locked, err := monthReportSrc(ctx, svcCtx, yearMonth, nil)
		if err == nil {
			out := make(map[string]float64, len(rows))
			for _, r := range rows {
				out[r.EmployeeUuid()] = float64(r.AbsentDays())
			}
			return out, locked, nil
		}
		// 月报源出错降级为打卡兜底（保持未锁定语义）
		logx.WithContext(ctx).Errorf("month report source fallback to punch settle: %v", err)
	}

	start := yearMonth + "-01"
	end := monthEndOf(yearMonth)
	records, err := svcCtx.Model.HrmPunchRecord.FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_punch_record.SettleStatus, "missing").
			GreaterEqualThan(hrm_punch_record.PunchDate, start).
			LessEqualThan(hrm_punch_record.PunchDate, end).
			Build()...)
	if err != nil {
		return nil, false, err
	}
	out := make(map[string]float64, len(records))
	for _, r := range records {
		out[r.EmployeeUuid]++
	}
	return out, false, nil // 兜底口径视为未锁定（提示重新试算）
}

type perfResult struct {
	assessmentNo string
	finalScore   float64
	finalGrade   string
	payLinkMode  string
	meritBase    float64
}

// loadPerfResults 绩效结果（HRM-43/46 契约）：final_score/final_grade 在 assessment 行上直读。
// 批次月份 = 绩效批次 period 覆盖该月（period_start ≤ 月末 AND period_end ≥ 月初）且任务已完成
//（task_status='1'）。同人多任务取 final_score 最大的一条。
func loadPerfResults(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (map[string]perfResult, error) {
	periods, err := svcCtx.Model.HrmPerfPeriod.FindByCondition(ctx, nil, condition.NewChain().Build()...)
	if err != nil {
		return nil, err
	}
	monthFirst := yearMonth + "-01"
	monthEnd := monthEndOf(yearMonth)
	var matched []*hrm_perf_period.HrmPerfPeriod
	schemeUuids := make([]string, 0)
	for _, p := range periods {
		if p.PeriodStart <= monthEnd && p.PeriodEnd >= monthFirst {
			matched = append(matched, p)
			schemeUuids = append(schemeUuids, p.SchemeUuid)
		}
	}
	if len(matched) == 0 {
		return map[string]perfResult{}, nil
	}
	periodUuids := make([]string, 0, len(matched))
	for _, p := range matched {
		periodUuids = append(periodUuids, p.Uuid)
	}
	assessments, err := svcCtx.Model.HrmPerfAssessment.FindByCondition(ctx, nil,
		condition.NewChain().
			In(hrm_perf_assessment.PeriodUuid, periodUuids).
			Equal(hrm_perf_assessment.TaskStatus, "1").
			Build()...)
	if err != nil {
		return nil, err
	}
	schemes, err := svcCtx.Model.HrmPerfScheme.FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_scheme.Uuid, dedupOf(schemeUuids)).Build()...)
	if err != nil {
		return nil, err
	}
	modeByScheme := make(map[string]string, len(schemes))
	for _, s := range schemes {
		modeByScheme[s.Uuid] = s.PayLinkMode
	}
	periodScheme := make(map[string]string, len(matched))
	for _, p := range matched {
		periodScheme[p.Uuid] = p.SchemeUuid
	}

	out := make(map[string]perfResult, len(assessments))
	for _, a := range assessments {
		pr := perfResult{
			assessmentNo: a.AssessmentNo,
			finalScore:   a.FinalScore,
			finalGrade:   a.FinalGrade,
			payLinkMode:  modeByScheme[periodScheme[a.PeriodUuid]],
		}
		if old, ok := out[a.EmployeeUuid]; !ok || a.FinalScore > old.finalScore {
			out[a.EmployeeUuid] = pr
		}
	}
	return out, nil
}

type taxMonthData struct {
	months                  int     // 期初快照累计月数（截至月）
	incomeOpening           float64 // 期初累计收入
	specialDeductionOpening float64 // 期初累计专项扣除
	specialExtraOpening     float64 // 期初累计专项附加
	specialExtraThisMonth   float64 // 试算月当月专项附加（维护值合计）
	otherDeductionOpening   float64
	taxPaidOpening          float64
	subjectPreferred        bool // 主体行优先于不限主体行
}

// loadTaxData 个税数据源（hrm-36 契约 #1）：期初（员工+年度唯一，subject_uuid=''=不限主体；
// 同员工多行时优先主体行）+ 试算月专项附加（员工+年度+月+扣除项唯一，多类型合计）。
func loadTaxData(ctx context.Context, svcCtx *svc.ServiceContext, yearMonth string) (map[string]*taxMonthData, error) {
	year := yearMonth[:4]
	monthInt := int64(atoiOf(yearMonth[5:7]))
	openings, err := svcCtx.Model.HrmTaxOpening.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_tax_opening.TaxYear, year).Build()...)
	if err != nil {
		return nil, err
	}
	out := make(map[string]*taxMonthData, len(openings))
	for _, o := range openings {
		if existing, ok := out[o.EmployeeUuid]; ok && existing.subjectPreferred {
			continue
		}
		if o.UntilMonth >= monthInt {
			continue // 期初快照已含/超过试算月（未来数据），按 0（hrm-36 notes #2 口径）
		}
		out[o.EmployeeUuid] = &taxMonthData{
			months:                  int(o.UntilMonth),
			incomeOpening:           o.IncomeTotal,
			specialDeductionOpening: o.SpecialDeductionTotal,
			specialExtraOpening:     o.SpecialExtraTotal,
			otherDeductionOpening:   o.OtherDeductionTotal,
			taxPaidOpening:          o.TaxPaidTotal,
			subjectPreferred:        o.SubjectUuid != "",
		}
	}
	// 试算月专项附加
	deductions, err := svcCtx.Model.HrmTaxSpecialDeduction.FindByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_tax_special_deduction.DeductYear, year).
			Equal(hrm_tax_special_deduction.DeductMonth, monthInt).
			Build()...)
	if err != nil {
		return nil, err
	}
	for _, d := range deductions {
		td, ok := out[d.EmployeeUuid]
		if !ok {
			td = &taxMonthData{}
			out[d.EmployeeUuid] = td
		}
		td.specialExtraThisMonth += d.Amount
	}
	return out, nil
}

// ---- 公式求值器（预置种子，不支持任意表达式 — hrm-33 契约 #1 公式项种子仅两个） ----

// evalFormulaItem 预置简单求值器：仅识别 NET_SALARY=应发合计−应扣合计 与
// LABOR_COST=应发合计+公司社保+公司公积金。自定义覆盖公式/未知公式编码按 0 兜底并给说明
//（demo 口径：自定义公式项不在本票求值范围，写入 hrm-37-notes）。
func evalFormulaItem(itemCode string, earningSum, deductionSum, siCompany, hfCompany float64, overrideFormula string) (float64, string) {
	if overrideFormula != "" && itemCode != ItemCodeNetSalary && itemCode != ItemCodeLaborCost {
		return 0, "自定义覆盖公式不支持任意表达式求值，按 0 兜底（仅 NET_SALARY/LABOR_COST 预置公式生效）"
	}
	switch itemCode {
	case ItemCodeNetSalary:
		return earningSum - deductionSum, fmt.Sprintf("公式：应发合计 %.2f − 应扣合计 %.2f", earningSum, deductionSum)
	case ItemCodeLaborCost:
		return earningSum + siCompany + hfCompany, fmt.Sprintf("公式：应发合计 %.2f + 公司社保 %.2f + 公司公积金 %.2f", earningSum, siCompany, hfCompany)
	}
	return 0, "未知公式项按 0 兜底"
}

// ---- 小工具 ----

func dedupOf(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func daysInMonthOf(yearMonth string) int {
	first, err := time.ParseInLocation("2006-01", yearMonth, time.Local)
	if err != nil {
		return 30
	}
	return first.AddDate(0, 1, -1).Day()
}

func monthEndOf(yearMonth string) string {
	first, _ := time.ParseInLocation("2006-01", yearMonth, time.Local)
	return first.AddDate(0, 1, -1).Format("2006-01-02")
}

func atoiOf(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// warningCollector 告警文本收集（去重 + 条件追加）。
type warningCollector struct {
	set   map[string]bool
	order []string
}

func newWarningCollector() *warningCollector {
	return &warningCollector{set: map[string]bool{}}
}

func (w *warningCollector) addIf(cond bool, msg string) {
	if !cond || w.set[msg] {
		return
	}
	w.set[msg] = true
	w.order = append(w.order, msg)
}

func (w *warningCollector) addDedup(msg string) { w.addIf(true, msg) }

func (w *warningCollector) text() string { return strings.Join(w.order, "\n") }
