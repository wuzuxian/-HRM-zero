package perf_assessment

import (
	"testing"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	v1hrmperf_assessment "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

// DB-backed logic tests for 考核任务工作台 (HRM-44, issue #46), run via svctest:
//
// Coverage (issue #46 acceptance criteria):
//   - 全周期待办汇总卡六卡数值（构造多任务多节点 → 按当前用户角色口径断言）
//   - 角色区分（被考核人本人 vs 评审码 vs 校准码 vs 无任何角色）
//   - 最近待办列表排序（更新时间倒序）+ 展示快照（批次/员工姓名）
//   - 申诉中的任务单列（appeals）+ 申诉卡计数

// runWorkbenchChain drives a task to the given node via the stage functions
// (复用 HRM-43 测试惯用步法：确认→自评→评审→校准 逐阶段推进；从当前节点增量推进).
func driveToNode(t *testing.T, svcCtx *svc.ServiceContext, env *svctest.Env, taskUuid, emp string, node string) {
	t.Helper()
	task, err := FindAssessmentByUuid(testCtx, svcCtx.SqlxConn, taskUuid)
	mustNoErr(t, err)
	snapshot, err := parseItemsSnapshot(task.ItemsSnapshot)
	mustNoErr(t, err)
	items := []ScoreItem{
		{IndicatorUuid: snapshot[0].IndicatorUuid, Score: 80},
		{IndicatorUuid: snapshot[1].IndicatorUuid, Score: 80},
	}
	reviewItems := []ScoreItem{
		{IndicatorUuid: snapshot[0].IndicatorUuid, Score: 90},
		{IndicatorUuid: snapshot[1].IndicatorUuid, Score: 78},
	}

	// 节点序：从当前节点增量推进到目标节点（前置守卫要求逐阶段）。
	nodeSeq := []string{NodePlanConfirm, NodeSelfReview, NodeLeaderReview, NodeCalibration, NodeResultConfirm}
	curIdx, wantIdx := -1, -1
	for i, n := range nodeSeq {
		if n == task.CurrentNode {
			curIdx = i
		}
		if n == node {
			wantIdx = i
		}
	}
	if curIdx < 0 || wantIdx < 0 || wantIdx < curIdx {
		t.Fatalf("driveToNode: bad node state cur=%s want=%s", task.CurrentNode, node)
	}
	for i := curIdx; i < wantIdx; i++ {
		switch i {
		case 0:
			mustNoErr(t, ConfirmPlanForTask(testCtx, svcCtx.SqlxConn, taskUuid, "", actorFor(emp)))
		case 1:
			mustNoErr(t, SubmitSelfReviewForTask(testCtx, svcCtx, taskUuid, items, "", actorFor(emp)))
		case 2:
			mustNoErr(t, SubmitLeaderReviewForTask(testCtx, svcCtx, taskUuid, reviewItems, "", actorFor("", reviewCodes...)))
		case 3:
			mustNoErr(t, SubmitCalibrationForTask(testCtx, svcCtx, taskUuid, 88, "", "", actorFor("", calibCodes...)))
		}
	}
}

// seedThreeTasks 批次含 3 个被考核人，生成 3 张任务（empA 推到节点2、empB 推到节点3、empC 停在节点1）。
func seedThreeTasks(t *testing.T, svcCtx *svc.ServiceContext, env *svctest.Env) (periodUuid string, empA, empB, empC string) {
	t.Helper()
	tpl := seedTemplate(t, env)
	scheme := seedScheme(t, env, tpl, "percent")
	empA = seedEmployee(t, env, "")
	empB = seedEmployee(t, env, "")
	empC = seedEmployee(t, env, "")
	periodUuid = seedOngoingPeriod(t, env, scheme, empA, empB, empC)
	mustGenerate(t, svcCtx, periodUuid)

	rows, _, err := ListTasks(testCtx, svcCtx, periodUuid, "", "", "", 1, 10)
	mustNoErr(t, err)
	assertLen(t, rows, 3)
	tasksByEmp := map[string]string{}
	for _, row := range rows {
		tasksByEmp[row.EmployeeUuid] = row.Uuid
	}
	driveToNode(t, svcCtx, env, tasksByEmp[empA], empA, NodeSelfReview)    // empA: 节点2 自评填报
	driveToNode(t, svcCtx, env, tasksByEmp[empB], empB, NodeLeaderReview) // empB: 节点3 上级评审
	return periodUuid, empA, empB, empC
}

// ---- 1. 被考核人视角：汇总卡数值 + 列表快照 ----

func TestWorkbenchOwnerCards(t *testing.T) {
	svcCtx, env := newSvc(t)
	_, empA, empB, empC := seedThreeTasks(t, svcCtx, env)

	// empC 视角（停留在节点1）：目标确认卡=1，其余本人卡=0
	cards, todos, appeals, err := MyAssessmentWorkbench(testCtx, svcCtx, actorFor(empC))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.PlanConfirm)
	assertEq(t, int64(0), cards.SelfReview)
	assertEq(t, int64(0), cards.LeaderReview)
	assertEq(t, int64(0), cards.Calibration)
	assertEq(t, int64(0), cards.ResultConfirm)
	assertEq(t, int64(0), cards.Appeal)
	assertLen(t, todos, 1)
	assertLen(t, appeals, 0)
	assertEq(t, WorkbenchRoleOwner, todos[0].myRole)
	assertEq(t, empC, todos[0].employeeUuid)
	assertEq(t, NodePlanConfirm, todos[0].currentNode)
	// 展示快照：批次名 + 员工名
	assertNotEmpty(t, todos[0].periodNo)
	assertNotEmpty(t, todos[0].periodName)
	assertNotEmpty(t, todos[0].employeeName)
	assertNotEmpty(t, todos[0].assessmentNo)

	// empA 视角（节点2 自评填报）：自评卡=1；节点1 任务不进本人待办
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor(empA))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.SelfReview)
	assertEq(t, int64(0), cards.PlanConfirm)
	assertLen(t, todos, 1)
	assertEq(t, NodeSelfReview, todos[0].currentNode)

	// 评审人（有 review 码，无本人任务）视角：上级评审卡=1（empB）
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", reviewCodes...))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.LeaderReview)
	assertEq(t, int64(0), cards.PlanConfirm)
	assertEq(t, int64(0), cards.SelfReview)
	assertLen(t, todos, 1)
	assertEq(t, WorkbenchRoleReviewer, todos[0].myRole)
	assertEq(t, empB, todos[0].employeeUuid)
	assertEq(t, NodeLeaderReview, todos[0].currentNode)

	// 校准人视角：无节点4 任务 → 全 0
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", calibCodes...))
	mustNoErr(t, err)
	assertEq(t, int64(0), cards.Calibration)
	assertLen(t, todos, 0)

	// 无角色（无码、非本人）视角：看不到任何待办
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("emp-stranger"))
	mustNoErr(t, err)
	assertEq(t, int64(0), cards.PlanConfirm+cards.SelfReview+cards.LeaderReview+cards.Calibration+cards.ResultConfirm+cards.Appeal)
	assertLen(t, todos, 0)
}

// ---- 2. 校准卡 + 申诉单列 ----

func TestWorkbenchCalibrationAndAppeal(t *testing.T) {
	svcCtx, env := newSvc(t)
	period, empA, _, _ := seedThreeTasks(t, svcCtx, env)

	// empA 推到节点4（校准填报）：校准人卡=1
	rows, _, err := ListTasks(testCtx, svcCtx, period, empA, "", "", 1, 10)
	mustNoErr(t, err)
	taskA := rows[0]
	driveToNode(t, svcCtx, env, taskA.Uuid, empA, NodeCalibration)

	cards, todos, _, err := MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", calibCodes...))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.Calibration)
	assertLen(t, todos, 1)
	assertEq(t, WorkbenchRoleCalibrator, todos[0].myRole)

	// 校准后推到节点5（结果确认）→ 校准卡归零、被考核人结果确认卡=1
	driveToNode(t, svcCtx, env, taskA.Uuid, empA, NodeResultConfirm)
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", calibCodes...))
	mustNoErr(t, err)
	assertEq(t, int64(0), cards.Calibration)
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor(empA))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.ResultConfirm)
	assertEq(t, NodeResultConfirm, todos[0].currentNode)

	// empA 发起申诉 → task_status='2'：被考核人申诉卡=1 且单列；校准人申诉卡=1
	mustNoErr(t, AppealForTask(testCtx, svcCtx.SqlxConn, taskA.Uuid, "对结果有异议", actorFor(empA)))

	cards, todos, appeals, err := MyAssessmentWorkbench(testCtx, svcCtx, actorFor(empA))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.Appeal)
	assertEq(t, int64(0), cards.ResultConfirm) // 申诉中不再是结果确认待办
	assertLen(t, appeals, 1)
	assertEq(t, TaskStatusAppealed, appeals[0].taskStatus)
	assertEq(t, AppealPending, appeals[0].appealStatus)
	assertEq(t, WorkbenchRoleOwner, appeals[0].myRole)

	cards, _, appeals, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", calibCodes...))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.Appeal)
	assertLen(t, appeals, 1)
	assertEq(t, WorkbenchRoleCalibrator, appeals[0].myRole)

	// 评审人（无 calibrate 码）看不到申诉卡
	cards, _, appeals, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", reviewCodes...))
	mustNoErr(t, err)
	assertEq(t, int64(0), cards.Appeal)
	assertLen(t, appeals, 0)
}

// ---- 3. 最近待办排序（更新时间倒序） + 逻辑端点 plumbing ----

func TestWorkbenchOrderAndLogic(t *testing.T) {
	svcCtx, env := newSvc(t)
	period, empA, empB, empC := seedThreeTasks(t, svcCtx, env)

	// 依次推进 empC（节点1→2）再 empA（节点2→3）；empA 的 update_time 最新。
	rows, _, err := ListTasks(testCtx, svcCtx, period, empC, "", "", 1, 10)
	mustNoErr(t, err)
	driveToNode(t, svcCtx, env, rows[0].Uuid, empC, NodeSelfReview)
	rows, _, err = ListTasks(testCtx, svcCtx, period, empA, "", "", 1, 10)
	mustNoErr(t, err)
	driveToNode(t, svcCtx, env, rows[0].Uuid, empA, NodeLeaderReview)

	// empA 视角：本人任务已到节点3（不再属于本人待办）+ 无其他 → 0 条
	cards, todos, _, err := MyAssessmentWorkbench(testCtx, svcCtx, actorFor(empA))
	mustNoErr(t, err)
	assertEq(t, int64(0), cards.PlanConfirm+cards.SelfReview+cards.ResultConfirm)
	assertLen(t, todos, 0)

	// empC 视角：本人节点2 自评待办 1 条
	cards, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor(empC))
	mustNoErr(t, err)
	assertEq(t, int64(1), cards.SelfReview)
	assertLen(t, todos, 1)

	// 评审人视角（empB 节点3 + empA 节点3 两条评审待办，empA updateTime 更新 → 排最前）
	_, todos, _, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor("", reviewCodes...))
	mustNoErr(t, err)
	assertLen(t, todos, 2)
	assertEq(t, empA, todos[0].employeeUuid)
	assertEq(t, empB, todos[1].employeeUuid)

	// 逻辑端点 plumbing：NewGetMyWorkbench 走通（测试 ctx 无登录态 → 空操作人全 0）
	ctx := WithCodes(testCtx, nil)
	resp, err := NewGetMyWorkbench(ctx, svcCtx, nil).GetMyWorkbench(&v1hrmperf_assessment.GetMyAssessmentWorkbenchRequest{})
	mustNoErr(t, err)
	assertLen(t, resp.Todos, 0)
	assertLen(t, resp.Appeals, 0)

	// 空操作人（未绑定员工）+ 无码：全 0
	var appeals []workbenchTodo
	cards, todos, appeals, err = MyAssessmentWorkbench(testCtx, svcCtx, actorFor(""))
	mustNoErr(t, err)
	assertEq(t, int64(0), cards.PlanConfirm+cards.SelfReview+cards.LeaderReview+cards.Calibration+cards.ResultConfirm+cards.Appeal)
	assertLen(t, todos, 0)
	assertLen(t, appeals, 0)
}
