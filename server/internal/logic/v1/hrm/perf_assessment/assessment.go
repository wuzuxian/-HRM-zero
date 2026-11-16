package perf_assessment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_scheme"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// 考核任务核心逻辑（HRM-43）：任务生成（幂等）+ 五阶段流转（前置守卫）+ 申诉 + 批次联动。
//
// 批次状态联动契约（联动点，HRM-42 状态机/代码零改动，走 perf_period.AdvancePeriod）：
//   - 全部任务 current_node >= '4'（到达校准填报）→ 批次 '1'进行中 → '3'待校准
//     （'2'待确认 无独立停留：任务链内联目标确认/自评/评审，批次直接两步 advance）。
//   - 全部任务结果确认完成（task_status='1'）→ 批次 '3'待校准 → '4'结果已发布。
//   - HRM-45 审批通过后推 '4'→'5'已发布（消费方，不改本域）。

// FindAssessmentByUuid loads one 考核任务.
func FindAssessmentByUuid(ctx context.Context, conn sqlx.SqlConn, uuid string) (*hrm_perf_assessment.HrmPerfAssessment, error) {
	row, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindOneByUuid(ctx, nil, uuid)
	if err != nil {
		if errors.Is(err, hrm_perf_assessment.ErrNotFound) {
			return nil, ErrAssessmentNotFound
		}
		return nil, err
	}
	return row, nil
}

// findPeriod loads the 绩效批次 row.
func findPeriod(ctx context.Context, conn sqlx.SqlConn, periodUuid string) (*hrm_perf_period.HrmPerfPeriod, error) {
	row, err := hrm_perf_period.NewHrmPerfPeriodModel(conn).FindOneByUuid(ctx, nil, periodUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_period.ErrNotFound) {
			return nil, fmt.Errorf("%w: 绩效批次不存在 %s", ErrAssessmentInvalid, periodUuid)
		}
		return nil, err
	}
	return row, nil
}

// findScheme loads the 绩效方案 row (评分模式来源).
func findScheme(ctx context.Context, conn sqlx.SqlConn, schemeUuid string) (*hrm_perf_scheme.HrmPerfScheme, error) {
	row, err := hrm_perf_scheme.NewHrmPerfSchemeModel(conn).FindOneByUuid(ctx, nil, schemeUuid)
	if err != nil {
		if errors.Is(err, hrm_perf_scheme.ErrNotFound) {
			return nil, fmt.Errorf("%w: 绩效方案不存在 %s", ErrAssessmentInvalid, schemeUuid)
		}
		return nil, err
	}
	return row, nil
}

// GenerateAssessments 批次发布后为每个参与人生成考核任务（幂等：UNIQUE(period_uuid,
// employee_uuid)，已存在者跳过；补发新增参与人的任务并返回本次实际插入数）。
// 每任务固化：考核单号 PA-年-流水 + 模板快照 items+权重 + 方案评分模式写入任务行。
func GenerateAssessments(ctx context.Context, conn sqlx.SqlConn, svcCtx *svc.ServiceContext, periodUuid string) (int64, error) {
	period, err := findPeriod(ctx, conn, periodUuid)
	if err != nil {
		return 0, err
	}
	if period.Status != perf_period.PeriodStatusOngoing {
		return 0, ErrPeriodNotOngoing
	}

	// 消费 HRM-42 契约：GetPeriodParticipants（发布后锁定稳定，见 hrm-42-notes）。
	participants, err := perf_period.GetPeriodParticipants(ctx, &svc.ServiceContext{
		ServiceContext: svcCtx.ServiceContext,
		Model:          svcCtx.Model,
	}, periodUuid)
	if err != nil {
		return 0, err
	}
	if len(participants) == 0 {
		return 0, ErrNoParticipants
	}

	scheme, err := findScheme(ctx, conn, period.SchemeUuid)
	if err != nil {
		return 0, err
	}
	year, err := assessmentYear(period.PeriodStart)
	if err != nil {
		return 0, err
	}

	existing, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_perf_assessment.PeriodUuid, periodUuid).Build()...)
	if err != nil {
		return 0, err
	}
	have := make(map[string]struct{}, len(existing))
	for _, row := range existing {
		have[row.EmployeeUuid] = struct{}{}
	}

	var inserted int64
	for _, p := range participants {
		if _, ok := have[p.EmployeeUuid]; ok {
			continue // 幂等：已存在跳过
		}
		snapshot, err := SnapshotFromTemplate(ctx, svcCtx, scheme.TemplateUuid)
		if err != nil {
			return inserted, err
		}
		snapshotJSON, err := encodeItemsSnapshot(snapshot)
		if err != nil {
			return inserted, err
		}

		var insertedCount int64
		err = svcCtx.SqlxConn.TransactCtx(ctx, func(_ context.Context, session sqlx.Session) error {
			no, err := GenerateAssessmentNo(ctx, session, year)
			if err != nil {
				return err
			}
			if err := hrm_perf_assessment.NewHrmPerfAssessmentModel(sqlx.NewSqlConnFromSession(session)).InsertV2(ctx, session,
				&hrm_perf_assessment.HrmPerfAssessment{
					Uuid:          uuid.New().String(),
					AssessmentNo:  no,
					PeriodUuid:    periodUuid,
					EmployeeUuid:  p.EmployeeUuid,
					TemplateUuid:  scheme.TemplateUuid,
					ItemsSnapshot: snapshotJSON,
					CurrentNode:   NodePlanConfirm,
					TaskStatus:    TaskStatusOngoing,
				}); err != nil {
				return err
			}
			insertedCount++
			return nil
		})
		if err != nil {
			return inserted, err
		}
		inserted += insertedCount
	}
	return inserted, nil
}

// nodeRequirement describes who may act on a node.
//   - Plan/Self/Result: 仅被考核人本人（employee_uuid == JWT user uuid 经
//     hrm_employee.user_uuid 反查；逻辑测试注入 fake user uuid）。
//   - LeaderReview: 持有 hrm:perf_assessment:review 码（简化：批次创建人语义由
//     casbin 按钮码承担；服务端以 codes 含 review 判定）。
//   - Calibration/申诉裁决: 持有 hrm:perf_assessment:calibrate 码。

// actorContext is the resolved 操作人身份（服务端从 JWT 取；测试注入）。
type actorContext struct {
	// EmployeeUuid 是当前登录账号对应的员工 uuid（user_uuid 反查）；空=未绑定。
	EmployeeUuid string
	// Codes 是当前账号的按钮/接口权限码（含 'v1:' 前缀，与 route2code 一致）。
	Codes []string
}

// resolveActor resolves the JWT user to an actorContext: JWT user uuid →
// hrm_employee.user_uuid → employee uuid（被考核人身份校验用）+ permission codes
// (withCodes 注入 — handler 从 authx 提供的按钮码取，测试直接构造)。
func resolveActor(ctx context.Context, conn sqlx.SqlConn) (*actorContext, error) {
	user := hrmuser.Current(ctx)
	actor := &actorContext{Codes: codesFromCtx(ctx)}
	if user.Uuid == "" {
		return actor, nil // 无登录态（逻辑测试路径）
	}
	rows, err := hrm_employee.NewHrmEmployeeModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_employee.UserUuid, user.Uuid).Build()...)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		actor.EmployeeUuid = rows[0].Uuid
	}
	return actor, nil
}

func hasCode(codes []string, code string) bool {
	for _, c := range codes {
		if c == code {
			return true
		}
	}
	return false
}

// guardNode is the 五阶段前置守卫：任务必须处于期望节点且进行中。
func guardNode(row *hrm_perf_assessment.HrmPerfAssessment, wantNode string) error {
	if row.TaskStatus != TaskStatusOngoing {
		return fmt.Errorf("%w: 任务状态不允许填报（%s）", ErrNodeInvalid, row.TaskStatus)
	}
	if row.CurrentNode != wantNode {
		return fmt.Errorf("%w: 当前节点 %s，操作要求节点 %s（前一节点未完成）", ErrNodeInvalid, row.CurrentNode, wantNode)
	}
	return nil
}

// stageResult is the shared output of a stage submission.
type stageResult struct {
	score float64
	grade string
}

// applyStage writes the stage columns + advances the node (事务内调用).
func applyStage(ctx context.Context, session sqlx.Session, row *hrm_perf_assessment.HrmPerfAssessment,
	nextNode string, updates map[string]any) error {
	updates["current_node"] = nextNode
	return hrm_perf_assessment.NewHrmPerfAssessmentModel(sqlx.NewSqlConnFromSession(session)).UpdateFieldsByCondition(ctx, session,
		updates, condition.NewChain().Equal(hrm_perf_assessment.Id, row.Id).Build()...)
}

// nowStamp renders the 各阶段时间 (YYYY-MM-DD HH:MM:SS).
func nowStamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
