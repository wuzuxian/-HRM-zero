package perf_assessment

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 五阶段流转（前置守卫：前一节点完成才可下一节点；越权填报拒绝）：
//
//	1 目标确认（被考核人） → 2 自评填报（被考核人） → 3 上级评审（review 码）
//	→ 4 校准填报（calibrate 码） → 5 结果确认（被考核人；或发起申诉 → 申诉裁决闭环）。
//
// 每阶段完成后调用 SyncPeriodByTasks 联动批次状态（联动点见 appeal.go 头注释）。

// ConfirmPlanForTask 阶段1：目标确认（被考核人确认指标目标）。
func ConfirmPlanForTask(ctx context.Context, conn sqlx.SqlConn, assessmentUuid, comment string, actor *actorContext) error {
	row, err := FindAssessmentByUuid(ctx, conn, assessmentUuid)
	if err != nil {
		return err
	}
	if err := guardNode(row, NodePlanConfirm); err != nil {
		return err
	}
	// 仅被考核人本人可确认目标（空操作人=未绑定员工账号，同样拒绝）
	if actor.EmployeeUuid != row.EmployeeUuid {
		return ErrOperatorDenied
	}

	m := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn)
	row.CurrentNode = NodeSelfReview
	row.PlanTime = nowStamp()
	row.PlanBy = actor.EmployeeUuid
	return m.Update(ctx, nil, row)
}

// SubmitSelfReviewForTask 阶段2：自评填报（按模板快照 items 逐项评分，权重汇总）。
func SubmitSelfReviewForTask(ctx context.Context, svcCtx *svc.ServiceContext, assessmentUuid string,
	items []ScoreItem, comment string, actor *actorContext) error {
	row, err := FindAssessmentByUuid(ctx, svcCtx.SqlxConn, assessmentUuid)
	if err != nil {
		return err
	}
	if err := guardNode(row, NodeSelfReview); err != nil {
		return err
	}
	// 仅被考核人本人可自评（空操作人=未绑定员工账号，同样拒绝）
	if actor.EmployeeUuid != row.EmployeeUuid {
		return ErrOperatorDenied
	}
	return submitScoredStage(ctx, svcCtx, row, NodeLeaderReview, "self", items, comment, actor)
}

// SubmitLeaderReviewForTask 阶段3：上级评审（简化口径：持 hrm:perf_assessment:review 码者；
// 被考核人部门负责人语义由按钮码分配承担）。
func SubmitLeaderReviewForTask(ctx context.Context, svcCtx *svc.ServiceContext, assessmentUuid string,
	items []ScoreItem, comment string, actor *actorContext) error {
	row, err := FindAssessmentByUuid(ctx, svcCtx.SqlxConn, assessmentUuid)
	if err != nil {
		return err
	}
	if err := guardNode(row, NodeLeaderReview); err != nil {
		return err
	}
	if !hasCode(actor.Codes, PermAssessmentReview) {
		return ErrOperatorDenied
	}
	return submitScoredStage(ctx, svcCtx, row, NodeCalibration, "review", items, comment, actor)
}

// submitScoredStage 共享的自评/评审提交：覆盖校验 → 权重汇总 → 等级映射 → 落列 → 节点推进。
// stage: "self" 写 self_score/self_grade/self_items；"review" 写 review_*。
func submitScoredStage(ctx context.Context, svcCtx *svc.ServiceContext, row *hrm_perf_assessment.HrmPerfAssessment,
	nextNode, stage string, items []ScoreItem, comment string, actor *actorContext) error {
	snapshot, err := parseItemsSnapshot(row.ItemsSnapshot)
	if err != nil {
		return err
	}
	total, err := WeightedTotal(snapshot, items)
	if err != nil {
		return err
	}
	itemsJSON, err := encodeScoreItems(items)
	if err != nil {
		return err
	}

	period, err := findPeriod(ctx, svcCtx.SqlxConn, row.PeriodUuid)
	if err != nil {
		return err
	}
	scheme, err := findScheme(ctx, svcCtx.SqlxConn, period.SchemeUuid)
	if err != nil {
		return err
	}
	score, grade, err := NormalizeResult(scheme.ScoreMode, total, "")
	if err != nil {
		return err
	}

	by := actor.EmployeeUuid
	now := nowStamp()
	updates := map[string]any{
		stage + "_comment": comment,
		stage + "_time":    now,
		stage + "_by":      by,
	}
	switch stage {
	case "self":
		updates["self_score"] = score
		updates["self_grade"] = grade
		updates["self_items"] = itemsJSON
	case "review":
		updates["review_score"] = score
		updates["review_grade"] = grade
		updates["review_items"] = itemsJSON
	}

	err = svcCtx.SqlxConn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
		return applyStage(ctx, session, row, nextNode, updates)
	})
	if err != nil {
		return err
	}
	// 阶段完成 → 批次状态联动（幂等，失败不阻断阶段提交语义由 handler 决定）
	return SyncPeriodByTasks(ctx, svcCtx, row.PeriodUuid)
}

// SubmitCalibrationForTask 阶段4：校准填报（校准人调整最终得分；默认以评审得分为基础）。
func SubmitCalibrationForTask(ctx context.Context, svcCtx *svc.ServiceContext, assessmentUuid string,
	finalScore float64, finalGrade, comment string, actor *actorContext) error {
	conn := svcCtx.SqlxConn
	row, err := FindAssessmentByUuid(ctx, conn, assessmentUuid)
	if err != nil {
		return err
	}
	if err := guardNode(row, NodeCalibration); err != nil {
		return err
	}
	if !hasCode(actor.Codes, PermAssessmentCalibrate) {
		return ErrOperatorDenied
	}

	period, err := findPeriod(ctx, conn, row.PeriodUuid)
	if err != nil {
		return err
	}
	scheme, err := findScheme(ctx, conn, period.SchemeUuid)
	if err != nil {
		return err
	}
	score, grade, err := NormalizeResult(scheme.ScoreMode, finalScore, finalGrade)
	if err != nil {
		return err
	}

	m := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn)
	row.CurrentNode = NodeResultConfirm
	row.FinalScore = score
	row.FinalGrade = grade
	row.CalibComment = comment
	row.CalibTime = nowStamp()
	row.CalibBy = actor.EmployeeUuid
	if err := m.Update(ctx, nil, row); err != nil {
		return err
	}
	return SyncPeriodByTasks(ctx, svcCtx, row.PeriodUuid)
}

// ConfirmResultForTask 阶段5：结果确认（被考核人确认 → 任务完成 → 批次联动）。
func ConfirmResultForTask(ctx context.Context, svcCtx *svc.ServiceContext, assessmentUuid string, actor *actorContext) error {
	conn := svcCtx.SqlxConn
	row, err := FindAssessmentByUuid(ctx, conn, assessmentUuid)
	if err != nil {
		return err
	}
	if err := guardNode(row, NodeResultConfirm); err != nil {
		return err
	}
	// 仅被考核人本人可确认结果（空操作人=未绑定员工账号，同样拒绝）
	if actor.EmployeeUuid != row.EmployeeUuid {
		return ErrOperatorDenied
	}

	m := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn)
	row.TaskStatus = TaskStatusDone
	row.ResultTime = nowStamp()
	row.ResultBy = actor.EmployeeUuid
	if err := m.Update(ctx, nil, row); err != nil {
		return err
	}
	return SyncPeriodByTasks(ctx, svcCtx, row.PeriodUuid)
}

// ListTasks paginated 考核任务 list (五个填报列表页共用，按批次/员工/节点/状态过滤).
func ListTasks(ctx context.Context, svcCtx *svc.ServiceContext, periodUuid, employeeUuid, currentNode, taskStatus string,
	current, size int) ([]*hrm_perf_assessment.HrmPerfAssessment, int64, error) {
	chain := condition.NewChain().
		OrderByDesc(hrm_perf_assessment.CreateTime)
	if periodUuid != "" {
		chain = chain.Equal(hrm_perf_assessment.PeriodUuid, periodUuid)
	}
	if employeeUuid != "" {
		chain = chain.Equal(hrm_perf_assessment.EmployeeUuid, employeeUuid)
	}
	if currentNode != "" {
		chain = chain.Equal(hrm_perf_assessment.CurrentNode, currentNode)
	}
	if taskStatus != "" {
		chain = chain.Equal(hrm_perf_assessment.TaskStatus, taskStatus)
	}
	rows, total, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(svcCtx.SqlxConn).PageByCondition(ctx, nil,
		chain.Page(current, size).Build()...)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// guard used implicitly: perf_period imported for status constants.
var _ = perf_period.PeriodStatusOngoing
