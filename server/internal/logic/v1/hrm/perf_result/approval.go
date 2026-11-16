package perf_result

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	perf_period "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result_approval"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// HR211 批次结果审批（HRM-45，ADR-0001 单步通用审批 + HRM-06 单据公共组件）：
//   - 提交（SubmitPerfResultApproval）：批次须 '4'结果已发布（HRM-43 任务链联动后的
//     待审批口径）。一批次一张审批单（UNIQUE(period_uuid)）：已存在 未提交/审批不通过/
//     已撤销 单则更新快照后重新提交；审批中/审批通过 → 拒绝。快照从 hrm_perf_assessment
//     聚合：人数 / 平均得分 / 绩效金额合计（系数映射口径）+ 公司/部门名称快照。
//   - 审批通过事务内（OnApproved）：生成 hrm_perf_result 结果行（每参与人一行，
//     UNIQUE(period_uuid,employee_uuid) 幂等）+ 批次 AdvancePeriod '4'→'5'已发布 +
//     结果行 published=1。驳回/撤销不产生结果行、批次不推进。
//
// 嵌套事务纪律（hrm-42-notes gotcha 1）：bill.SubmitBill/ApproveBill 自带事务，
// 必须在本域事务之外调用，绝不嵌套。

// svcCtxKey carries the *svc.ServiceContext into the approval OnApproved closure,
// which the bill base calls with only (ctx, session, billUuid) — same pattern as
// overtimebill (HRM-21).
type svcCtxKey struct{}

// WithSvcCtx attaches the ServiceContext used by 结果行生成.
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

// SubmitPerfResultApproval creates (or re-submits) the HR211 审批单 of one 批次
// and submits it through the shared approval base. Returns the 审批单 uuid.
func SubmitPerfResultApproval(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext,
	periodUuid, opinion string) (string, error) {
	period, err := hrm_perf_period.NewHrmPerfPeriodModel(conn).FindOneByUuid(ctx, nil, periodUuid)
	if err != nil {
		if err == hrm_perf_period.ErrNotFound {
			return "", fmt.Errorf("%w: 绩效批次不存在 %s", ErrResultApprovalInvalid, periodUuid)
		}
		return "", err
	}
	// 守卫：仅 '4'结果已发布（HRM-43 任务链全部结果确认后的联动状态）可提交结果审批。
	if period.Status != perf_period.PeriodStatusResultPublished {
		return "", fmt.Errorf("%w: 批次状态不符（须为 结果已发布，当前 %s）", ErrResultApprovalInvalid, period.Status)
	}

	// 一批次一张审批单：已存在则校验状态（审批中/审批通过 拒绝）。
	existing, err := FindApprovalByPeriod(ctx, conn, periodUuid)
	if err != nil && err != ErrResultApprovalNotFound {
		return "", err
	}
	if existing != nil {
		if _, ok := submitAllowedStatuses[existing.Status]; !ok {
			return "", fmt.Errorf("%w: 该批次已有结果审批单（状态 %s）", ErrResultApprovalInvalid, existing.Status)
		}
	}

	op := currentOperator(ctx)

	// 事务一：快照聚合 + 审批单行创建/更新（编号与单据行原子）。
	var approvalUuid string
	txErr := conn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		agg, aerr := aggregatePeriod(ctx, sqlx.NewSqlConnFromSession(session), period)
		if aerr != nil {
			return aerr
		}

		if existing != nil {
			approvalUuid = existing.Uuid
			m := model.NewHrmPerfResultApprovalModel(sqlx.NewSqlConnFromSession(session))
			return m.UpdateFieldsByCondition(ctx, session, map[string]any{
				"period_no":         period.PeriodNo,
				"period_name":       period.Name,
				"company_snapshot":  agg.companyJSON,
				"org_snapshot":      agg.orgJSON,
				"participant_count": agg.count,
				"avg_score":         agg.avgScore,
				"total_amount":      agg.totalAmount,
				"submit_by":         op.Name,
				"submit_time":       nowStr(),
				"opinion":           opinion,
			}, condition.NewChain().Equal(model.Uuid, existing.Uuid).Build()...)
		}

		billNo, berr := bill.GenerateBillNo(ctx, session, BillTypeCode, time.Now())
		if berr != nil {
			return berr
		}
		row := &model.HrmPerfResultApproval{
			BillNo:           billNo,
			PeriodUuid:       periodUuid,
			PeriodNo:         period.PeriodNo,
			PeriodName:       period.Name,
			CompanySnapshot:  agg.companyJSON,
			OrgSnapshot:      agg.orgJSON,
			ParticipantCount: agg.count,
			AvgScore:         agg.avgScore,
			TotalAmount:      agg.totalAmount,
			Status:           bill.StatusUnsubmitted,
			SubmitBy:         op.Name,
			SubmitTime:       nowStr(),
			Opinion:          opinion,
		}
		if ierr := insertApproval(ctx, session, row); ierr != nil {
			return ierr
		}
		approvalUuid = row.Uuid
		return nil
	})
	if txErr != nil {
		return "", txErr
	}

	// 事务二（底座自带）：单据生命周期 提交 → 审批中 + 审批记录。
	if err := bill.SubmitBill(ctx, conn, BillTypeCode, approvalUuid, op, opinion); err != nil {
		return "", err
	}
	return approvalUuid, nil
}

// currentOperator resolves the JWT user to a bill Operator (测试路径匿名).
func currentOperator(ctx context.Context) bill.Operator {
	user := hrmuser.Current(ctx)
	return bill.Operator{UserUuid: user.Uuid, Name: user.Username}
}

// periodAggregation is the 提交时快照 aggregate of one 批次.
type periodAggregation struct {
	count       int64
	avgScore    float64
	totalAmount float64
	companyJSON string
	orgJSON     string
}

// aggregatePeriod aggregates the 批次考核任务 into the HR211 快照 columns:
// 人数（任务行数）/ 平均得分（final_score 均值）/ 绩效金额合计（金额制 final_score 直取、
// 系数制 系数×基数(0, HRM-49 回填) 按行求和）+ 公司/部门名称去重快照。
func aggregatePeriod(ctx context.Context, conn sqlx.SqlConn,
	period *hrm_perf_period.HrmPerfPeriod) (*periodAggregation, error) {
	assessments, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_assessment.PeriodUuid, period.Uuid).Build()...)
	if err != nil {
		return nil, err
	}
	if len(assessments) == 0 {
		return nil, fmt.Errorf("%w: 批次无考核任务，无法提交结果审批", ErrResultApprovalInvalid)
	}

	employeeUuids := make([]string, 0, len(assessments))
	for _, a := range assessments {
		employeeUuids = append(employeeUuids, a.EmployeeUuid)
	}
	employees, err := employeeRowsByUuids(ctx, conn, employeeUuids)
	if err != nil {
		return nil, err
	}

	scheme, err := hrm_perf_scheme.NewHrmPerfSchemeModel(conn).FindOneByUuid(ctx, nil, period.SchemeUuid)
	if err != nil {
		if err == hrm_perf_scheme.ErrNotFound {
			return nil, fmt.Errorf("%w: 绩效方案不存在 %s", ErrResultApprovalInvalid, period.SchemeUuid)
		}
		return nil, err
	}

	var sumScore, sumAmount float64
	companySeen := map[string]struct{}{}
	orgSeen := map[string]struct{}{}
	companies := make([]string, 0, 4)
	orgs := make([]string, 0, 8)
	for _, a := range assessments {
		sumScore += a.FinalScore
		sumAmount += resultAmount(scheme.PayLinkMode, a)
		if emp, ok := employees[a.EmployeeUuid]; ok {
			if emp.CompanyUuid != "" {
				if _, dup := companySeen[emp.CompanyUuid]; !dup {
					companySeen[emp.CompanyUuid] = struct{}{}
					companies = append(companies, companyNameOf(ctx, conn, emp.CompanyUuid))
				}
			}
			if emp.OrgUuid != "" {
				if _, dup := orgSeen[emp.OrgUuid]; !dup {
					orgSeen[emp.OrgUuid] = struct{}{}
					name, oerr := orgNameByUuid(ctx, conn, emp.OrgUuid)
					if oerr != nil {
						return nil, oerr
					}
					orgs = append(orgs, name)
				}
			}
		}
	}

	companyJSON, err := encodeJSON(companies)
	if err != nil {
		return nil, err
	}
	orgJSON, err := encodeJSON(orgs)
	if err != nil {
		return nil, err
	}

	n := float64(len(assessments))
	return &periodAggregation{
		count:       int64(len(assessments)),
		avgScore:    round2(sumScore / n),
		totalAmount: round2(sumAmount),
		companyJSON: companyJSON,
		orgJSON:     orgJSON,
	}, nil
}

// resultAmount computes one row's 绩效金额 per the HRM-43 系数计算契约:
//   - pay_link_mode=amount（金额制）：绩效金额 = final_score 直取；
//   - pay_link_mode=coefficient（系数制）：绩效金额 = 系数 × 绩效工资基数；
//     基数取自薪资档案（HRM-49 联薪域），本域不可得 → 落 0，HRM-49 联薪时回填。
func resultAmount(payLinkMode string, a *hrm_perf_assessment.HrmPerfAssessment) float64 {
	if payLinkMode == "amount" {
		return a.FinalScore
	}
	return 0
}

// companyNameOf resolves company org uuid → org_name (missing → "").
func companyNameOf(ctx context.Context, conn sqlx.SqlConn, companyUuid string) string {
	name, err := orgNameByUuid(ctx, conn, companyUuid)
	if err != nil {
		return ""
	}
	return name
}

func encodeJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// onApprovedGenerateResults is the HR211 OnApproved write-back (审批通过事务内):
//   - 生成 hrm_perf_result 结果行（每考核任务一行：得分/等级/系数/绩效金额/结果状态；
//     UNIQUE(period_uuid,employee_uuid) 幂等——重复审批通过不重复插行）；
//   - 批次 AdvancePeriod '4'结果已发布 → '5'已发布；
//   - 结果行 published=1。
func onApprovedGenerateResults(ctx context.Context, session sqlx.Session, billUuid string) error {
	conn := sqlx.NewSqlConnFromSession(session)
	approval, err := model.NewHrmPerfResultApprovalModel(conn).FindOneByUuid(ctx, session, billUuid)
	if err != nil {
		return err
	}
	period, err := hrm_perf_period.NewHrmPerfPeriodModel(conn).FindOneByUuid(ctx, session, approval.PeriodUuid)
	if err != nil {
		return err
	}
	scheme, err := hrm_perf_scheme.NewHrmPerfSchemeModel(conn).FindOneByUuid(ctx, session, period.SchemeUuid)
	if err != nil {
		return err
	}

	assessments, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindByCondition(ctx, session,
		condition.NewChain().Equal(hrm_perf_assessment.PeriodUuid, approval.PeriodUuid).Build()...)
	if err != nil {
		return err
	}
	if len(assessments) == 0 {
		return fmt.Errorf("%w: 批次无考核任务，无法生成结果行", ErrResultApprovalInvalid)
	}

	employeeUuids := make([]string, 0, len(assessments))
	for _, a := range assessments {
		employeeUuids = append(employeeUuids, a.EmployeeUuid)
	}
	employees, err := employeeRowsByUuids(ctx, conn, employeeUuids)
	if err != nil {
		return err
	}

	resultModel := hrm_perf_result.NewHrmPerfResultModel(conn)
	for _, a := range assessments {
		// 幂等：UNIQUE(period_uuid, employee_uuid) — 已存在（重复审批）跳过生成。
		dup, derr := resultModel.FindOneByCondition(ctx, session,
			condition.NewChain().
				Equal(hrm_perf_result.PeriodUuid, approval.PeriodUuid).
				Equal(hrm_perf_result.EmployeeUuid, a.EmployeeUuid).Build()...)
		if derr != nil && derr != hrm_perf_result.ErrNotFound {
			return derr
		}
		if dup != nil {
			continue
		}

		row := &hrm_perf_result.HrmPerfResult{
			Uuid:           uuid.New().String(),
			ApprovalUuid:   approval.Uuid,
			AssessmentUuid: a.Uuid,
			PeriodUuid:     approval.PeriodUuid,
			PeriodNo:       approval.PeriodNo,
			EmployeeUuid:   a.EmployeeUuid,
			Score:          a.FinalScore,
			Grade:          a.FinalGrade,
			Coefficient:    GradeCoefficient(a.FinalGrade),
			PerfAmount:     resultAmount(scheme.PayLinkMode, a),
			ResultStatus:   ResultStatusNormal,
			Published:      1,
		}
		if emp := employees[a.EmployeeUuid]; emp != nil {
			row.EmployeeName = emp.Name
			row.CompanyUuid = emp.CompanyUuid
			row.OrgUuid = emp.OrgUuid
			row.CompanyName = companyNameOf(ctx, conn, emp.CompanyUuid)
			orgName, oerr := orgNameByUuid(ctx, conn, emp.OrgUuid)
			if oerr != nil {
				return oerr
			}
			row.OrgName = orgName
		}
		if err := resultModel.InsertV2(ctx, session, row); err != nil {
			return err
		}
	}

	// 批次联动 '4'→'5'（HRM-42 状态机，forward-only 守卫兜底）。
	if err := perf_period.AdvancePeriod(ctx, conn, approval.PeriodUuid); err != nil {
		return err
	}

	// 审批单回写通过信息（姓名快照取 JWT 操作人；测试路径为空）。
	op := currentOperator(ctx)
	return model.NewHrmPerfResultApprovalModel(conn).UpdateFieldsByCondition(ctx, session, map[string]any{
		"approve_by":   op.Name,
		"approve_time": nowStr(),
	}, condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
}

func init() {
	bill.RegisterType(BillTypeCode, bill.Handler{
		Load: func(ctx context.Context, session sqlx.Session, billUuid string) (string, string, error) {
			row, err := model.NewHrmPerfResultApprovalModel(sqlx.NewSqlConnFromSession(session)).
				FindOneByUuid(ctx, session, billUuid)
			if err != nil {
				return "", "", err
			}
			return row.BillNo, row.Status, nil
		},
		SetStatus: func(ctx context.Context, session sqlx.Session, billUuid, statusCode string) error {
			return model.NewHrmPerfResultApprovalModel(sqlx.NewSqlConnFromSession(session)).
				UpdateFieldsByCondition(ctx, session,
					map[string]any{"status": statusCode},
					condition.NewChain().Equal(model.Uuid, billUuid).Build()...)
		},
		OnApproved: onApprovedGenerateResults,
	})
}
