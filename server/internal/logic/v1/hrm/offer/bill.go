package offer

import (
	"context"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	candidate "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/candidate"
	entry "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/entry"
	position "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/position"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_offer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// BillTypeCode lives in validate.go; this file registers the 单据 handler
// against the shared approval base (HRM-06).
//
// OnApproved（审批通过事务内三联动，demo-inventory §5 录用审批 → 衔接员工入职）：
//  1) position.IncreaseHiredCountTx — 职位已录用+1（每单恰好一次，applied 标记幂等，
//     达到 headcount 自动置 status='full'）；
//  2) 候选人状态推进至 hired（HRM-26 契约：offer→hired 合法流转，hired 终态）；
//  3) 创建草稿 HR201 入职单（姓名=候选人姓名、计划入职日期=plan_entry_date、
//     试用期/部门职位按录用单快照回填；HR 仍可补全入职详情后走入职单自身审批）。
// 任一联动失败整体回滚（OnApproved failure rolls back the approval tx）。
// 幂等：applied=1 直接 return nil（重放审批不重复三联动）。
func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := hrm_offer_bill.NewHrmOfferBillModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, status string) error {
			return hrm_offer_bill.NewHrmOfferBillModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": status},
					condition.NewChain().Equal(hrm_offer_bill.Uuid, billUuid).Build()...)
		},
		OnApproved: onApproved,
	})
}

// onApproved is the 审批通过 三联动 write-back (ADR-0001), inside the approval
// transaction. Idempotent via the applied 回写 marker (重放审批不重复三联动).
func onApproved(ctx context.Context, session sqlx.Session, billUuid string) error {
	txSvcCtx := TxBoundSvcCtx(ctx, session, svcCtxOf(ctx))

	m := hrm_offer_bill.NewHrmOfferBillModel(sqlx.NewSqlConnFromSession(session))
	row, err := m.FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	if row.Applied == 1 {
		return nil // idempotent: 三联动 already executed for this 录用单
	}

	// 1) 职位已录用+1（IncreaseHiredCountTx：校验 closed/超招，自动 full）
	if err = position.IncreaseHiredCountTx(ctx, txSvcCtx, session, row.PositionUuid); err != nil {
		return err
	}

	// 2) 候选人状态推进至 hired（offer→hired 合法流转；hired 终态）
	if err = candidateAdvanceToHired(ctx, session, row.CandidateUuid); err != nil {
		return err
	}

	// 3) 创建草稿 HR201 入职单（衔接员工入职，HR 仍可补全入职详情）
	entryBillUuid, err := createDraftEntryBill(ctx, session, txSvcCtx, row)
	if err != nil {
		return err
	}

	// 回写 applied + entry_bill_uuid inside the same tx — 幂等标记.
	return m.UpdateFieldsByCondition(ctx, session,
		map[string]any{"applied": 1, "entry_bill_uuid": entryBillUuid},
		condition.NewChain().Equal(hrm_offer_bill.Uuid, row.Uuid).Build()...)
}

// candidateAdvanceToHired 推进候选人至 hired：逐级调用状态推进（HRM-26 状态机
// 每次只能 +1 阶段，skip-level 被 10442 拒绝，故按 forwardOrder 逐级写），
// 当前已是 hired 直接返回（幂等）.
func candidateAdvanceToHired(ctx context.Context, session sqlx.Session, candidateUuid string) error {
	m := hrm_recruitment_candidate.NewHrmRecruitmentCandidateModel(sqlx.NewSqlConnFromSession(session))
	row, err := m.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_recruitment_candidate.Uuid, candidateUuid).Build()...)
	if err != nil {
		return err
	}
	if row.Status == candidate.CandidateStatusHired {
		return nil
	}
	order := []string{
		candidate.CandidateStatusDelivered,
		candidate.CandidateStatusInterviewing,
		candidate.CandidateStatusOffer,
		candidate.CandidateStatusHired,
	}
	started := false
	for _, next := range order {
		if !started {
			if next == row.Status {
				started = true
			}
			continue
		}
		row.Status = next
		if err = m.Update(ctx, session, row); err != nil {
			return err
		}
	}
	return nil
}

// createDraftEntryBill 创建草稿 HR201 入职单（未提交状态，HR 补全详情后走入职单
// 自身的提交/审批流转）：姓名=候选人姓名、入职日期=计划入职日期、试用期按录用单、
// 部门/职位按录用单快照回填；手机号/邮箱从候选人档案补充；预计转正由 entry 建档
// 口径归一化（hire_date + 试用期月）；三子表快照为空.
func createDraftEntryBill(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, offerRow *hrm_offer_bill.HrmOfferBill) (string, error) {
	cand, err := svcCtx.Model.HrmRecruitmentCandidate.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_recruitment_candidate.Uuid, offerRow.CandidateUuid).Build()...)
	if err != nil {
		return "", err
	}

	// 草稿入职单编号：复用共享单据编号生成器（HR201 段，同日流水顺延）
	entryBillNo, err := bill.GenerateBillNo(ctx, session, entry.BillTypeCode, timeNow())
	if err != nil {
		return "", err
	}

	entryUuid := uuid.New().String()
	m := hrm_entry_bill.NewHrmEntryBillModel(sqlx.NewSqlConnFromSession(session))
	if err = m.InsertV2(ctx, session, &hrm_entry_bill.HrmEntryBill{
		Uuid:             entryUuid,
		BillNo:           entryBillNo,
		Status:           bill.StatusUnsubmitted,
		Name:             offerRow.CandidateName,
		Phone:            cand.Phone,
		Email:            cand.Email,
		Position:         offerRow.PostName,
		ProbationMonths:  offerRow.ProbationMonths,
		OrgUuid:          offerRow.OrgUuid,
		HireDate:         offerRow.PlanEntryDate,
		ExpectFormalDate: "", // entry 建档时按 hire_date+试用期 归一化
		EmployeeUuid:     "",
		CreateBy:         offerRow.CreateBy,
		CreatorName:      offerRow.CreatorName,
	}); err != nil {
		return "", err
	}
	return entryUuid, nil
}
