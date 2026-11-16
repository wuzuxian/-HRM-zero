package salary_batch

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_bill_sequence"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_subject"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 批次类型（demo-inventory 8.4：批次类型仅 月度批次，预留扩展）。
const (
	BatchTypeMonthly = "monthly"
	// ItemCodeNetSalary / ItemCodeLaborCost 公式项种子编码（hrm-33 预置项，试算引擎特殊处理键）。
	ItemCodeNetSalary = "NET_SALARY"
	ItemCodeLaborCost = "LABOR_COST"
	// ItemCodeTax 预置个税项目编码（个税引擎消费键）。
	ItemCodeTax = "TAX"
	// ItemCodeMeritSalary 预置绩效工资项目编码（绩效联薪消费键）。
	ItemCodeMeritSalary = "MERIT_SALARY"
	// ItemCodeRewardBonus / ItemCodeRewardPenalty 奖惩取数键。
	ItemCodeRewardBonus    = "REWARD_BONUS"
	ItemCodeRewardPenalty  = "REWARD_PENALTY"
	ItemCodeAbsenceDeduction = "ABSENCE_DEDUCTION"
	// ItemCodeSiPersonal/HfPersonal/SiCompany/HfCompany 社保取数键（社保模块未建，兜底 0）。
	ItemCodeSiPersonal = "SI_PERSONAL"
	ItemCodeHfPersonal = "HF_PERSONAL"
	ItemCodeSiCompany  = "SI_COMPANY"
	ItemCodeHfCompany  = "HF_COMPANY"
)

// salaryBatchNoPrefix 批次编号前缀（GLOSSARY 单据编号：薪资批次独立编号体系 SA{yyyyMMdd}{4位流水}）。
const salaryBatchNoPrefix = "SA"

// salaryBatchNoSeqMax 4 位流水上限（1..9999）。
const salaryBatchNoSeqMax = 9999

// batchNoLayout 批次编号内日期段布局（批次月份首日）。
const batchNoLayout = "20060102"

// FormatBatchNo renders 批次编号 `SA{yyyyMMdd}{4位流水}` (demo-inventory 0 全局约定), e.g. SA202603010001.
// 纯函数：流水分配在 GenerateBatchNo。
func FormatBatchNo(day time.Time, seq int64) (string, error) {
	if seq < 1 || seq > salaryBatchNoSeqMax {
		return "", fmt.Errorf("批次编号流水超出范围: %d", seq)
	}
	return fmt.Sprintf("%s%s%04d", salaryBatchNoPrefix, day.Format(batchNoLayout), seq), nil
}

// GenerateBatchNo allocates the next 4-digit 流水 for the 批次月份首日 and returns the formatted
// 批次编号 SA+日期+流水。复用共享 hrm_bill_sequence（bill_type_code='SA', bill_date=批次月份首日
// YYYYMMDD），select-then-upsert 语义与 PERF 批次编号一致（per 月份首日流水，非 per 主体）。
//
// 必须在批次插入事务内调用（传入 tx session）使编号与批次行原子落库。
func GenerateBatchNo(ctx context.Context, session sqlx.Session, batchMonthFirst time.Time) (string, error) {
	ymd := batchMonthFirst.Format(batchNoLayout)
	seqModel := hrm_bill_sequence.NewHrmBillSequenceModel(sqlx.NewSqlConnFromSession(session))

	row, err := seqModel.FindOneByCondition(ctx, session, condition.NewChain().
		Equal(hrm_bill_sequence.BillTypeCode, salaryBatchNoPrefix).
		Equal(hrm_bill_sequence.BillDate, ymd).
		Build()...)
	var seq int64
	switch {
	case err == hrm_bill_sequence.ErrNotFound:
		seq = 1
		if err := seqModel.InsertV2(ctx, session, &hrm_bill_sequence.HrmBillSequence{
			Uuid:         uuid.New().String(),
			BillTypeCode: salaryBatchNoPrefix,
			BillDate:     ymd,
			LastSeq:      seq,
		}); err != nil {
			return "", err
		}
	case err != nil:
		return "", err
	default:
		seq = row.LastSeq + 1
		if err := seqModel.UpdateFieldsByCondition(ctx, session,
			map[string]any{"last_seq": seq},
			condition.NewChain().Equal(hrm_bill_sequence.Id, row.Id).Build()...); err != nil {
			return "", err
		}
	}
	return FormatBatchNo(batchMonthFirst, seq)
}

// validateBatchMonth 校验 年份/月份 并返回批次月份首日（YYYY-MM-01，编号用）与 YYYY-MM。
func validateBatchMonth(year, month int64) (time.Time, string, error) {
	if year < 2000 || year > 2999 {
		return time.Time{}, "", statusInvalid("年份须为 2000-2999")
	}
	if month < 1 || month > 12 {
		return time.Time{}, "", statusInvalid("月份须为 1-12")
	}
	first := time.Date(int(year), time.Month(month), 1, 0, 0, 0, 0, time.Local)
	return first, first.Format("2006-01"), nil
}

// findBatchByUuid 按 uuid 取批次行（不存在 → SalaryBatchNotFoundCode）。
func findBatchByUuid(ctx context.Context, svcCtx *svc.ServiceContext, batchUuid string) (*hrm_salary_batch.HrmSalaryBatch, error) {
	return FindBatchByUuid(ctx, svcCtx, batchUuid)
}

// FindBatchByUuid is the exported 按 uuid 取批次行（不存在 → SalaryBatchNotFoundCode）；
// 跨域消费入口（HRM-38 确认审批 / HRM-39 发薪登记按批次行四标记守卫）。
func FindBatchByUuid(ctx context.Context, svcCtx *svc.ServiceContext, batchUuid string) (*hrm_salary_batch.HrmSalaryBatch, error) {
	if batchUuid == "" {
		return nil, statusInvalid("batch_uuid 不能为空")
	}
	row, err := svcCtx.Model.HrmSalaryBatch.FindOneByUuid(ctx, nil, batchUuid)
	if err != nil {
		if err == hrm_salary_batch.ErrNotFound {
			return nil, statusNotFound("薪资批次不存在")
		}
		return nil, err
	}
	return row, nil
}

// findSubjectByUuid 校验薪酬主体存在并返回（不存在 → SalaryBatchSubjectNotFoundCode；
// 不做主体启停检查，HRM-33 契约 #1：停用主体允许存量批次只读）。
func findSubjectByUuid(ctx context.Context, svcCtx *svc.ServiceContext, subjectUuid string) (*hrm_salary_subject.HrmSalarySubject, error) {
	if subjectUuid == "" {
		return nil, statusInvalid("subject_uuid 不能为空")
	}
	row, err := svcCtx.Model.HrmSalarySubject.FindOneByUuid(ctx, nil, subjectUuid)
	if err != nil {
		if err == hrm_salary_subject.ErrNotFound {
			return nil, statusSubjectNotFound("薪酬主体不存在")
		}
		return nil, err
	}
	return row, nil
}

// round2 保留 2 位小数（金额精度口径，与 taxcalc/perf 同款）。
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

