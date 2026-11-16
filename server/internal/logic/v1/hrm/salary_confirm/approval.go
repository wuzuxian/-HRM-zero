package salary_confirm

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_confirm_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// HR209 薪资确认审批（HRM-38，ADR-0001 单步通用审批 + HRM-06 单据公共组件）：
//   - 提交（SubmitSalaryConfirm）：批次须 已试算（trial_status='1'）且未确认
//     （confirm_status='0'，demo 8.4 批次选择器口径）。一批次一张审批单
//     （UNIQUE(batch_uuid)）：已存在 未提交/审批不通过/已撤销 单则更新快照后重新提交；
//     审批中/审批通过 → 拒绝。提交时快照 批次编号/名称/主体/年月/人数/实发合计。
//   - 审批通过事务内（OnApproved）：批次 confirm_status='1' + confirm 时间（本表
//     approve_time，幂等 applied：单据已 approved 时 ApproveBill 状态机直接拒绝；
//     confirm_status='1' 幂等写回）。驳回/撤销不推进批次确认标记。
//     HRM-39 发薪登记与工资条发布以 confirm_status='1' 为前置门槛（消费契约）。
//
// 嵌套事务纪律（hrm-42-notes gotcha 1）：bill.SubmitBill/ApproveBill 自带事务，
// 必须在本域事务之外调用，绝不嵌套。

// BillTypeCode is the 单据类型码 for 薪资确认审批（HR209，demo-inventory 8.4）。
const BillTypeCode = "209"

// svcCtxKey carries the *svc.ServiceContext into the approval OnApproved closure,
// which the bill base calls with only (ctx, session, billUuid) — same pattern as
// perf_result (HRM-45) / overtimebill (HRM-21).
type svcCtxKey struct{}

// WithSvcCtx attaches the ServiceContext used by 批次 confirm 写回.
func WithSvcCtx(ctx context.Context, svcCtx *svc.ServiceContext) context.Context {
	return context.WithValue(ctx, svcCtxKey{}, svcCtx)
}

func svcCtxOf(ctx context.Context) *svc.ServiceContext {
	if v, ok := ctx.Value(svcCtxKey{}).(*svc.ServiceContext); ok {
		return v
	}
	return nil
}

// submitAllowedStatuses lists the 单据状态 values that allow (re-)submission.
var submitAllowedStatuses = map[string]struct{}{
	bill.StatusUnsubmitted: {},
	bill.StatusRejected:    {},
	bill.StatusRevoked:     {},
}

// nowStr 当前时间字符串（YYYY-MM-DD HH:MM:SS，与 perf_result 快照时间同款口径）。
func nowStr() string {
	return time.Now().Format(time.DateTime)
}

// currentOperator resolves the JWT user to a bill Operator (测试路径匿名).
func currentOperator(ctx context.Context) bill.Operator {
	user := hrmuser.Current(ctx)
	return bill.Operator{UserUuid: user.Uuid, Name: user.Username}
}

// FindBillByBatch returns the existing 确认审批单 of one 批次;
// ErrNotFound（model 层）when none.
func FindBillByBatch(ctx context.Context, conn sqlx.SqlConn, batchUuid string) (*model.HrmSalaryConfirmBill, error) {
	return model.NewHrmSalaryConfirmBillModel(conn).FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(model.BatchUuid, batchUuid).Build()...)
}

// SubmitSalaryConfirm creates (or re-submits) the HR209 审批单 of one 批次
// and submits it through the shared approval base. Returns the 审批单 uuid.
func SubmitSalaryConfirm(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext,
	batchUuid, opinion string) (string, error) {
	batch, err := salary_batch.FindBatchByUuid(ctx, svcCtx, batchUuid)
	if err != nil {
		return "", err
	}
	// 守卫：批次须 已试算 且 未确认（demo 8.4 批次选择器口径）。
	if batch.TrialStatus != "1" {
		return "", statusInvalid("批次尚未试算，请先执行试算再发起薪资确认审批")
	}
	// confirm_status='' 与 '0' 同义（HRM-37 批次插入仅显式落 trial_status，
	// 其余标记列为 ''/NULL 语义即未推进）。
	if batch.ConfirmStatus != "" && batch.ConfirmStatus != "0" {
		return "", statusInvalid("批次已确认，不可重复发起薪资确认审批")
	}

	// 一批次一张审批单：已存在则校验状态（审批中/审批通过 拒绝）。
	existing, err := FindBillByBatch(ctx, conn, batchUuid)
	if err != nil && err != model.ErrNotFound {
		return "", err
	}
	if existing != nil {
		if _, ok := submitAllowedStatuses[existing.Status]; !ok {
			return "", statusInvalid(fmt.Sprintf("该批次已有薪资确认审批单（状态 %s）", existing.Status))
		}
	}

	op := currentOperator(ctx)

	// 事务一：快照聚合 + 审批单行创建/更新（编号与单据行原子）。
	var billUuid string
	txErr := conn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		if existing != nil {
			billUuid = existing.Uuid
			return model.NewHrmSalaryConfirmBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session, map[string]any{
					"batch_no":     batch.BatchNo,
					"batch_name":   batch.BatchName,
					"subject_uuid": batch.SubjectUuid,
					"subject_name": batch.SubjectName,
					"batch_year":   batch.BatchYear,
					"batch_month":  batch.BatchMonth,
					"headcount":    batch.Headcount,
					"net_total":    batch.NetTotal,
					"submit_by":    op.Name,
					"submit_time":  nowStr(),
					"opinion":      opinion,
				}, condition.NewChain().Equal(model.Uuid, existing.Uuid).Build()...)
		}

		billNo, berr := bill.GenerateBillNo(ctx, session, BillTypeCode, time.Now())
		if berr != nil {
			return berr
		}
		row := &model.HrmSalaryConfirmBill{
			Uuid:        uuid.New().String(),
			BillNo:      billNo,
			BatchUuid:   batchUuid,
			BatchNo:     batch.BatchNo,
			BatchName:   batch.BatchName,
			SubjectUuid: batch.SubjectUuid,
			SubjectName: batch.SubjectName,
			BatchYear:   batch.BatchYear,
			BatchMonth:  batch.BatchMonth,
			Headcount:   batch.Headcount,
			NetTotal:    batch.NetTotal,
			Status:      bill.StatusUnsubmitted,
			SubmitBy:    op.Name,
			SubmitTime:  nowStr(),
			Opinion:     opinion,
			CreateBy:    op.UserUuid,
			CreatorName: op.Name,
		}
		if ierr := model.NewHrmSalaryConfirmBillModel(sqlx.NewSqlConnFromSession(session)).
			InsertV2(ctx, session, row); ierr != nil {
			return ierr
		}
		billUuid = row.Uuid
		return nil
	})
	if txErr != nil {
		return "", txErr
	}

	// 事务二（底座自带）：单据生命周期 提交 → 审批中 + 审批记录。
	if err := bill.SubmitBill(WithSvcCtx(ctx, svcCtx), conn, BillTypeCode, billUuid, op, opinion); err != nil {
		return "", err
	}
	return billUuid, nil
}

// onApprovedConfirmBatch 审批通过事务内写回：批次 confirm_status='1'（幂等 applied）
// + 审批单通过信息。驳回/撤销不走本闭包。
func onApprovedConfirmBatch(ctx context.Context, session sqlx.Session, billUuid string) error {
	conn := sqlx.NewSqlConnFromSession(session)
	row, err := model.NewHrmSalaryConfirmBillModel(conn).FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	batchModel := hrm_salary_batch.NewHrmSalaryBatchModel(conn)
	batch, err := batchModel.FindOneByUuid(ctx, session, row.BatchUuid)
	if err != nil {
		return err
	}
	// 幂等 applied：confirm_status='1' 重复写回无害；已确认批次直接跳过。
	if batch.ConfirmStatus != "1" {
		if err := batchModel.UpdateFieldsByCondition(ctx, session, map[string]any{
			"confirm_status": "1",
		}, condition.NewChain().Equal(hrm_salary_batch.Uuid, batch.Uuid).Build()...); err != nil {
			return err
		}
	}
	op := currentOperator(ctx)
	return model.NewHrmSalaryConfirmBillModel(conn).UpdateFieldsByCondition(ctx, session, map[string]any{
		"approve_by":   op.Name,
		"approve_time": nowStr(),
	}, condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
}

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmSalaryConfirmBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, statusCode string) error {
			return model.NewHrmSalaryConfirmBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": statusCode},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApprovedConfirmBatch,
	})
}
