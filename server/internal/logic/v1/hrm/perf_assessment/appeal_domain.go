package perf_assessment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_appeal"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 申诉（HRM-43）：结果确认阶段被考核人发起（原因）→ 校准人裁决（维持/调整得分）→
// 关闭；申诉期间 task_status='2'，裁决后任务回到已完成（final_score 按裁决刷新）。

// newAppealUuid returns a fresh v4 uuid for hrm_perf_appeal rows.
func newAppealUuid() string {
	return uuid.New().String()
}

// actorName renders the 裁决人 name stamp (fallback to username-like identifier).
func actorName(actor *actorContext) string {
	if actor.EmployeeUuid != "" {
		return actor.EmployeeUuid
	}
	return ""
}

// AppealForTask 发起申诉（仅结果确认阶段、任务已完成且无待处理申诉；仅被考核人本人）。
func AppealForTask(ctx context.Context, conn sqlx.SqlConn, assessmentUuid, reason string, actor *actorContext) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("%w: 申诉原因不能为空", ErrAssessmentInvalid)
	}
	row, err := FindAssessmentByUuid(ctx, conn, assessmentUuid)
	if err != nil {
		return err
	}
	// 仅被考核人本人可发起申诉（空操作人=未绑定员工账号，同样拒绝）
	if actor.EmployeeUuid != row.EmployeeUuid {
		return ErrOperatorDenied
	}
	if row.CurrentNode != NodeResultConfirm {
		return fmt.Errorf("%w: 仅结果确认阶段可发起申诉（当前节点 %s）", ErrNodeInvalid, row.CurrentNode)
	}

	// 已有待处理申诉 → 拒绝重复发起
	pending, err := hrm_perf_appeal.NewHrmPerfAppealModel(conn).FindOneByCondition(ctx, nil,
		condition.NewChain().
			Equal(hrm_perf_appeal.AssessmentUuid, assessmentUuid).
			Equal(hrm_perf_appeal.Status, AppealPending).
			Build()...)
	if err != nil && !errors.Is(err, hrm_perf_appeal.ErrNotFound) {
		return err
	}
	if pending != nil {
		return ErrAppealExists
	}

	// 任务行标记申诉中
	m := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn)
	row.TaskStatus = TaskStatusAppealed
	if err := m.Update(ctx, nil, row); err != nil {
		return err
	}

	return hrm_perf_appeal.NewHrmPerfAppealModel(conn).InsertV2(ctx, nil, &hrm_perf_appeal.HrmPerfAppeal{
		Uuid:           newAppealUuid(),
		AssessmentUuid: assessmentUuid,
		PeriodUuid:     row.PeriodUuid,
		EmployeeUuid:   row.EmployeeUuid,
		Reason:         reason,
		Status:         AppealPending,
	})
}

// HandleAppealForTask 校准人裁决：'1'=维持原结果（关闭申诉）；'2'=调整得分（回写 final_score/
// final_grade 后关闭）。仅 calibrate 码持有者；仅待处理申诉可裁决。
// 入参 assessmentUuid 为考核任务 uuid（UI 上下文是任务行）：定位其最新待处理申诉。
func HandleAppealForTask(ctx context.Context, conn sqlx.SqlConn, assessmentUuid, decision, comment string,
	adjustedScore float64, adjustedGrade string, actor *actorContext) error {
	if !hasCode(actor.Codes, PermAssessmentCalibrate) {
		return ErrOperatorDenied
	}
	switch decision {
	case AppealUpheld, AppealAdjusted:
	default:
		return fmt.Errorf("%w: 裁决结果非法 %s", ErrAssessmentInvalid, decision)
	}

	m := hrm_perf_appeal.NewHrmPerfAppealModel(conn)
	appeals, err := m.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_appeal.AssessmentUuid, assessmentUuid).
			Equal(hrm_perf_appeal.Status, AppealPending).
			OrderByDesc(hrm_perf_appeal.Id).Build()...)
	if err != nil {
		return err
	}
	if len(appeals) == 0 {
		return ErrAppealNotFound
	}
	appeal := appeals[0]

	row, err := FindAssessmentByUuid(ctx, conn, appeal.AssessmentUuid)
	if err != nil {
		return err
	}
	period, err := findPeriod(ctx, conn, row.PeriodUuid)
	if err != nil {
		return err
	}
	scheme, err := findScheme(ctx, conn, period.SchemeUuid)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"status":           decision,
		"decision_comment": comment,
		"handled_by":       actorName(actor),
		"handled_time":     nowStamp(),
	}
	if decision == AppealAdjusted {
		score, grade, err := NormalizeResult(scheme.ScoreMode, adjustedScore, adjustedGrade)
		if err != nil {
			return err
		}
		updates["adjusted_score"] = score
		updates["adjusted_grade"] = grade
		// 调整得分回写考核任务最终结果
		row.FinalScore = score
		row.FinalGrade = grade
		if err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).Update(ctx, nil, row); err != nil {
			return err
		}
	}

	// 关闭申诉 + 任务回到已完成
	appeal.Status = decision
	if err := m.Update(ctx, nil, appeal); err != nil {
		return err
	}
	row.TaskStatus = TaskStatusDone
	return hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).Update(ctx, nil, row)
}

// latestAppealStatus returns the 申诉状态 of the latest appeal for the task (”=无申诉).
func latestAppealStatus(ctx context.Context, conn sqlx.SqlConn, assessmentUuid string) (string, error) {
	rows, err := hrm_perf_appeal.NewHrmPerfAppealModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_appeal.AssessmentUuid, assessmentUuid).
			OrderByDesc(hrm_perf_appeal.Id).Build()...)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0].Status, nil
}

// ---- 批次状态联动（联动点契约，供 HRM-45/46 消费口径参考） ----

// SyncPeriodByTasks 依据任务链进度联动批次状态（调用时机：每次任务阶段提交后）：
//
// 全部任务 current_node >= '4'（到达校准）且批次为 '1'进行中 → advance '2'待确认 → '3'待校准；
// 全部任务 task_status = '1'（结果确认完成）且批次为 '3'待校准 → advance '4'结果已发布。
//
// 联动是幂等的（批次状态机 forward-only 守卫保证重复调用无副作用）。
func SyncPeriodByTasks(ctx context.Context, svcCtx *svc.ServiceContext, periodUuid string) error {
	conn := svcCtx.SqlxConn
	period, err := findPeriod(ctx, conn, periodUuid)
	if err != nil {
		return err
	}
	if period.Status != perf_period.PeriodStatusOngoing && period.Status != perf_period.PeriodStatusCalibrating {
		return nil // 非进行中/待校准批次不联动
	}

	rows, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_assessment.PeriodUuid, periodUuid).Build()...)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	allAtCalibration := true
	allConfirmed := true
	for _, row := range rows {
		if row.CurrentNode < NodeCalibration {
			allAtCalibration = false
		}
		if row.TaskStatus != TaskStatusDone {
			allConfirmed = false
		}
	}

	switch {
	case allConfirmed && period.Status == perf_period.PeriodStatusCalibrating:
		// 全部结果确认 → '4'结果已发布
		return perf_period.AdvancePeriod(ctx, conn, periodUuid)
	case allAtCalibration && period.Status == perf_period.PeriodStatusOngoing:
		// 全部到达校准 → '2'待确认（无独立停留）→ '3'待校准
		if err := perf_period.AdvancePeriod(ctx, conn, periodUuid); err != nil {
			return err
		}
		return perf_period.AdvancePeriod(ctx, conn, periodUuid)
	default:
		return nil
	}
}
