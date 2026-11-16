package perf_assessment

import (
	"context"
	"sort"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_appeal"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

// 考核任务工作台聚合（HRM-44，demo-inventory 9.2 /hrm/performance/assessment-task/workbench）：
// 当前登录人视角的只读聚合端点 —
//   - 全周期待办汇总卡（六卡）：统计当前用户在各节点的实际待办任务数
//     （被考核人阶段 1/2/5 按 employee_uuid=我；评审节点3 持 review 码；
//     校准节点4 + 申诉裁决 持 calibrate 码 — 与 HRM-43 五阶段守卫同口径）。
//   - 最近待办列表（类型/批次/员工/当前节点/当前处理人，按更新时间倒序，默认前 20）。
//   - 申诉中的任务单列（task_status='2'，携带最新申诉状态）。
// 与 HRM-43 共享操作人解析 resolveActor（JWT user uuid → hrm_employee.user_uuid 反查）
// 和权限码 ctx WithCodes 约定。

// 我的工作台待办数上限（聚合页面固定取最近 N 条，不分页）。
const WorkbenchTodoLimit = 20

// 我的角色（AssessmentWorkbenchTodo.MyRole）。
const (
	WorkbenchRoleOwner      = "owner"      // 被考核人
	WorkbenchRoleReviewer   = "reviewer"   // 考核人（上级评审）
	WorkbenchRoleCalibrator = "calibrator" // 校准人（校准填报 + 申诉裁决）
)

// workbenchTodo is the normalized 待办 row (shared by Todos / Appeals).
type workbenchTodo struct {
	uuid         string
	assessmentNo string
	periodUuid   string
	periodNo     string
	periodName   string
	employeeUuid string
	employeeName string
	myRole       string
	currentNode  string
	taskStatus   string
	appealStatus string
	updateTime   string
}

// myOwnerTodoNodes are the nodes where the 被考核人 themselves act.
var myOwnerTodoNodes = []string{NodePlanConfirm, NodeSelfReview, NodeResultConfirm}

// loadPeriods 一次 IN 查询批次展示快照（编号/名称）。
func loadPeriods(ctx context.Context, svcCtx *svc.ServiceContext, periodUuids []string) map[string]*hrm_perf_period.HrmPerfPeriod {
	out := map[string]*hrm_perf_period.HrmPerfPeriod{}
	if len(periodUuids) == 0 {
		return out
	}
	rows, err := svcCtx.Model.HrmPerfPeriod.FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_period.Uuid, periodUuids).Build()...)
	if err != nil {
		return out
	}
	for _, row := range rows {
		out[row.Uuid] = row
	}
	return out
}

// loadEmployeeNames 一次 IN 查询被考核人姓名快照。
func loadEmployeeNames(ctx context.Context, svcCtx *svc.ServiceContext, employeeUuids []string) map[string]string {
	out := map[string]string{}
	if len(employeeUuids) == 0 {
		return out
	}
	rows, err := svcCtx.Model.HrmEmployee.FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_employee.Uuid, employeeUuids).Build()...)
	if err != nil {
		return out
	}
	for _, row := range rows {
		out[row.Uuid] = row.Name
	}
	return out
}

// loadLatestAppealStatus 按任务分组取最新申诉状态（id 倒序第一条）。
func loadLatestAppealStatus(ctx context.Context, svcCtx *svc.ServiceContext, assessmentUuids []string) map[string]string {
	out := map[string]string{}
	if len(assessmentUuids) == 0 {
		return out
	}
	rows, err := hrm_perf_appeal.NewHrmPerfAppealModel(svcCtx.SqlxConn).FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_perf_appeal.AssessmentUuid, assessmentUuids).
			OrderByDesc(hrm_perf_appeal.Id).Build()...)
	if err != nil {
		return out
	}
	for _, row := range rows {
		if _, ok := out[row.AssessmentUuid]; !ok {
			out[row.AssessmentUuid] = row.Status
		}
	}
	return out
}

// buildTodo converts a task row + snapshots into a workbenchTodo.
func buildTodo(row *hrm_perf_assessment.HrmPerfAssessment, myRole, appealStatus string) workbenchTodo {
	return workbenchTodo{
		uuid:         row.Uuid,
		assessmentNo: row.AssessmentNo,
		periodUuid:   row.PeriodUuid,
		employeeUuid: row.EmployeeUuid,
		myRole:       myRole,
		currentNode:  row.CurrentNode,
		taskStatus:   row.TaskStatus,
		appealStatus: appealStatus,
		updateTime:   row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}

// MyAssessmentWorkbench returns the aggregated workbench payload for the
// current actor (汇总卡 + 最近待办 + 申诉单列).
//
// 角色判定（与 HRM-43 五阶段守卫同口径）：
//   - 被考核人：actor.EmployeeUuid == row.EmployeeUuid（空操作人=未绑定，不算本人）。
//   - 评审：codes 含 v1:hrm:perf_assessment:review。
//   - 校准/申诉裁决：codes 含 v1:hrm:perf_assessment:calibrate。
// 一行任务可能同时命中多角色（如校准人自己也是被考核人），各角色口径分别计数。
func MyAssessmentWorkbench(ctx context.Context, svcCtx *svc.ServiceContext, actor *actorContext) (
	types.AssessmentWorkbenchCards, []workbenchTodo, []workbenchTodo, error) {
	var cards types.AssessmentWorkbenchCards
	isOwner := actor.EmployeeUuid != ""
	isReviewer := hasCode(actor.Codes, PermAssessmentReview)
	isCalibrator := hasCode(actor.Codes, PermAssessmentCalibrate)

	// ---- 1. 拉取未完成（或申诉中）任务候选，按更新时间倒序 ----
	allRows, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(svcCtx.SqlxConn).FindByCondition(ctx, nil,
		condition.NewChain().
			In(hrm_perf_assessment.TaskStatus, []string{TaskStatusOngoing, TaskStatusAppealed}).
			OrderByDesc(hrm_perf_assessment.UpdateTime).
			Build()...)
	if err != nil {
		return cards, nil, nil, err
	}

	// ---- 2. 逐行角色判定 + 汇总卡计数 ----
	collected := make([]workbenchTodo, 0, len(allRows))
	for _, row := range allRows {
		ownerHit := isOwner && row.EmployeeUuid == actor.EmployeeUuid
		reviewerHit := isReviewer && row.TaskStatus == TaskStatusOngoing && row.CurrentNode == NodeLeaderReview
		calibratorHit := isCalibrator && row.TaskStatus == TaskStatusOngoing && row.CurrentNode == NodeCalibration
		appealHit := row.TaskStatus == TaskStatusAppealed && (ownerHit || isCalibrator)

		if ownerHit && row.TaskStatus == TaskStatusOngoing {
			// 被考核人三卡（节点 1/2/5）；节点3/4（评审/校准中）不属于本人待办
			counted := true
			switch row.CurrentNode {
			case NodePlanConfirm:
				cards.PlanConfirm++
			case NodeSelfReview:
				cards.SelfReview++
			case NodeResultConfirm:
				cards.ResultConfirm++
			default:
				counted = false
			}
			if counted {
				collected = append(collected, buildTodo(row, WorkbenchRoleOwner, ""))
			}
			continue
		}
		if reviewerHit {
			cards.LeaderReview++
			collected = append(collected, buildTodo(row, WorkbenchRoleReviewer, ""))
			continue
		}
		if calibratorHit {
			cards.Calibration++
			collected = append(collected, buildTodo(row, WorkbenchRoleCalibrator, ""))
			continue
		}
		if appealHit {
			// 申诉卡（校准人裁决 / 被考核人等待裁决结果）
			cards.Appeal++
			collected = append(collected, buildTodo(row, mapRole(ownerHit, isCalibrator), ""))
		}
	}

	// ---- 3. 展示快照（批次编号/名称、被考核人姓名、最新申诉状态） ----
	periodUuids := make([]string, 0, len(collected))
	employeeUuids := make([]string, 0, len(collected))
	assessmentUuids := make([]string, 0, len(collected))
	seenPeriod := map[string]struct{}{}
	seenEmployee := map[string]struct{}{}
	for _, t := range collected {
		if _, ok := seenPeriod[t.periodUuid]; !ok {
			seenPeriod[t.periodUuid] = struct{}{}
			periodUuids = append(periodUuids, t.periodUuid)
		}
		if _, ok := seenEmployee[t.employeeUuid]; !ok {
			seenEmployee[t.employeeUuid] = struct{}{}
			employeeUuids = append(employeeUuids, t.employeeUuid)
		}
		assessmentUuids = append(assessmentUuids, t.uuid)
	}
	periods := loadPeriods(ctx, svcCtx, periodUuids)
	employeeNames := loadEmployeeNames(ctx, svcCtx, employeeUuids)
	appealStatuses := loadLatestAppealStatus(ctx, svcCtx, assessmentUuids)

	todos := make([]workbenchTodo, 0, len(collected))
	for _, t := range collected {
		if p := periods[t.periodUuid]; p != nil {
			t.periodNo = p.PeriodNo
			t.periodName = p.Name
		}
		t.employeeName = employeeNames[t.employeeUuid]
		t.appealStatus = appealStatuses[t.uuid]
		todos = append(todos, t)
	}
	// 最近待办：按更新时间倒序，默认前 N
	sort.Slice(todos, func(i, j int) bool { return todos[i].updateTime > todos[j].updateTime })
	if len(todos) > WorkbenchTodoLimit {
		todos = todos[:WorkbenchTodoLimit]
	}

	// ---- 4. 申诉中的任务单列 ----
	appeals := make([]workbenchTodo, 0)
	for _, t := range todos {
		if t.taskStatus == TaskStatusAppealed {
			appeals = append(appeals, t)
		}
	}

	return cards, todos, appeals, nil
}

// mapRole picks the primary role label for an appealed task (校准人优先展示裁决视角).
func mapRole(ownerHit, isCalibrator bool) string {
	if isCalibrator {
		return WorkbenchRoleCalibrator
	}
	if ownerHit {
		return WorkbenchRoleOwner
	}
	return WorkbenchRoleOwner
}
