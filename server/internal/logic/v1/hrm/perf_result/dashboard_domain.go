package perf_result

import (
	"context"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	hrm_perf_assessment "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	hrm_perf_period "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	hrm_perf_result "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result"
	perf_assessment "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_assessment"
	perf_period "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/perf_period"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_result"
)

// 绩效驾驶舱（HRM-46，demo-inventory 9.3 /hrm/performance/dashboard）。
//
// ⚠️ 口径契约（pre-decided，详见 hrm-46-notes.md；快照列直读勿重算）：
//   - 范围（scope）：companyUuid/orgUuid/periodUuid 三过滤可组合；company/org
//     匹配结果行快照列（company_uuid/org_uuid）与考核任务无快照列 → 任务口径
//     借助结果行/参与人映射；本实现简化：任务级口径（完成率/逾期/申诉中）按
//     periodUuid + company/org 快照存在时通过该范围员工集合过滤。
//   - 进行中批次数 = hrm_perf_period.status='1'（PeriodStatusOngoing）且范围匹配。
//   - 考核完成率 = 范围内批次考核任务 task_status='1' 数 / 任务总数 ×100（round2；
//     任务总数 0 → 100）。
//   - 逾期待办数 = 任务级：任务所属批次 period_end < 今天 且 task_status='0'。
//   - 风险人数 = 范围内已发布（published=1）结果行 score < lowScore 的去重员工数。
//   - 平均得分 = 范围内已发布结果行 score 均值（round2；无行 → 0）。
//   - 申诉中数 = 范围内考核任务 task_status='2' 数。
//   - 低分明细 = 风险口径的结果行明细；逾期任务明细 = 逾期口径的任务明细；
//     申诉中名单 = task_status='2' 任务列（简化 demo「申诉积压」）。
//   - lowScore 缺省 60（阈值由前端筛选传入）。

// DefaultLowScore 驾驶舱低分阈值缺省值。
const DefaultLowScore = 60

// todayStr returns today as YYYY-MM-DD (period_end is varchar(10)).
func todayStr() string {
	return time.Now().Format("2006-01-02")
}

// scopeEmployeeUuids resolves the company/org snapshot filter to the set of
// employee uuids having results in that scope ("" filters → empty set = no
// restriction via this channel).
func scopeEmployeeUuids(ctx context.Context, conn sqlx.SqlConn, companyUuid, orgUuid string) ([]string, error) {
	if companyUuid == "" && orgUuid == "" {
		return nil, nil
	}
	chain := condition.NewChain()
	if companyUuid != "" {
		chain = chain.Equal(hrm_perf_result.CompanyUuid, companyUuid)
	}
	if orgUuid != "" {
		chain = chain.Equal(hrm_perf_result.OrgUuid, orgUuid)
	}
	rows, err := hrm_perf_result.NewHrmPerfResultModel(conn).FindByCondition(ctx, nil,
		chain.Build()...)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rows))
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if _, ok := seen[r.EmployeeUuid]; ok {
			continue
		}
		seen[r.EmployeeUuid] = struct{}{}
		out = append(out, r.EmployeeUuid)
	}
	return out, nil
}

// scopePeriodUuids resolves the period-side scope: explicit periodUuid, or all
// periods matching company/org scope via participant membership; "" channel
// when unrestricted.
func scopePeriodUuids(ctx context.Context, conn sqlx.SqlConn, periodUuid, companyUuid, orgUuid string, employeeUuids []string) ([]string, error) {
	if periodUuid != "" {
		return []string{periodUuid}, nil
	}
	if companyUuid == "" && orgUuid == "" {
		return nil, nil
	}
	// 参与人属期映射：员工 → 其参与的批次
	chain := condition.NewChain()
	if len(employeeUuids) > 0 {
		chain = chain.In(hrm_perf_assessment.EmployeeUuid, employeeUuids)
	} else {
		// scope 快照无结果行 → 无匹配批次
		return []string{}, nil
	}
	rows, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindByCondition(ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rows))
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if _, ok := seen[r.PeriodUuid]; ok {
			continue
		}
		seen[r.PeriodUuid] = struct{}{}
		out = append(out, r.PeriodUuid)
	}
	return out, nil
}

// BuildDashboard aggregates the 驾驶舱 payload (single endpoint, HRM-46).
func BuildDashboard(ctx context.Context, conn sqlx.SqlConn, req *types.HrmPerfDashboardRequest) (*types.DashboardResponse, error) {
	lowScore := req.LowScore
	if lowScore <= 0 {
		lowScore = DefaultLowScore
	}

	employeeUuids, err := scopeEmployeeUuids(ctx, conn, req.CompanyUuid, req.OrgUuid)
	if err != nil {
		return nil, err
	}
	periodUuids, err := scopePeriodUuids(ctx, conn, req.PeriodUuid, req.CompanyUuid, req.OrgUuid, employeeUuids)
	if err != nil {
		return nil, err
	}

	// ---- 批次集合（进行中计数 + period_end 逾期判定） ----
	periodChain := condition.NewChain()
	if periodUuids != nil {
		if len(periodUuids) == 0 {
			periodUuids = []string{"__none__"}
		}
		periodChain = periodChain.In(hrm_perf_period.Uuid, periodUuids)
	}
	periods, err := hrm_perf_period.NewHrmPerfPeriodModel(conn).FindByCondition(ctx, nil, periodChain.Build()...)
	if err != nil {
		return nil, err
	}
	periodEndByUuid := make(map[string]string, len(periods))
	periodNoByUuid := make(map[string]string, len(periods))
	ongoingCount := int64(0)
	for _, p := range periods {
		periodEndByUuid[p.Uuid] = p.PeriodEnd
		periodNoByUuid[p.Uuid] = p.PeriodNo
		if p.Status == perf_period.PeriodStatusOngoing {
			ongoingCount++
		}
	}

	// ---- 任务集合（完成率/逾期/申诉中） ----
	taskChain := condition.NewChain()
	if periodUuids != nil {
		if len(periodUuids) == 0 {
			periodUuids = []string{"__none__"}
		}
		taskChain = taskChain.In(hrm_perf_assessment.PeriodUuid, periodUuids)
	}
	tasks, err := hrm_perf_assessment.NewHrmPerfAssessmentModel(conn).FindByCondition(ctx, nil, taskChain.Build()...)
	if err != nil {
		return nil, err
	}
	today := todayStr()
	var doneCount, overdueCount, appealCount int64
	overdueTasks := make([]types.HrmPerfDashboardOverdueTask, 0)
	appeals := make([]types.HrmPerfDashboardAppeal, 0)
	employeeNameByUuid := make(map[string]string)
	for _, t := range tasks {
		if t.TaskStatus == perf_assessment.TaskStatusDone {
			doneCount++
		}
		if t.TaskStatus == perf_assessment.TaskStatusAppealed {
			appealCount++
			appeals = append(appeals, types.HrmPerfDashboardAppeal{
				PeriodUuid:   t.PeriodUuid,
				PeriodNo:     periodNoByUuid[t.PeriodUuid],
				EmployeeUuid: t.EmployeeUuid,
				AssessmentNo: t.AssessmentNo,
				CurrentNode:  t.CurrentNode,
			})
		}
		if t.TaskStatus != perf_assessment.TaskStatusDone && periodEndByUuid[t.PeriodUuid] != "" && periodEndByUuid[t.PeriodUuid] < today {
			overdueCount++
			overdueTasks = append(overdueTasks, types.HrmPerfDashboardOverdueTask{
				PeriodUuid:   t.PeriodUuid,
				PeriodNo:     periodNoByUuid[t.PeriodUuid],
				PeriodEnd:    periodEndByUuid[t.PeriodUuid],
				EmployeeUuid: t.EmployeeUuid,
				CurrentNode:  t.CurrentNode,
				TaskStatus:   t.TaskStatus,
				AssessmentNo: t.AssessmentNo,
			})
		}
		employeeNameByUuid[t.EmployeeUuid] = ""
	}
	completionRate := 100.0
	if len(tasks) > 0 {
		completionRate = round2(float64(doneCount) * 100 / float64(len(tasks)))
	}

	// ---- 结果集合（平均得分/风险人数/低分明细） ----
	resultChain := condition.NewChain().Equal(hrm_perf_result.Published, 1)
	if req.CompanyUuid != "" {
		resultChain = resultChain.Equal(hrm_perf_result.CompanyUuid, req.CompanyUuid)
	}
	if req.OrgUuid != "" {
		resultChain = resultChain.Equal(hrm_perf_result.OrgUuid, req.OrgUuid)
	}
	if req.PeriodUuid != "" {
		resultChain = resultChain.Equal(hrm_perf_result.PeriodUuid, req.PeriodUuid)
	}
	results, err := hrm_perf_result.NewHrmPerfResultModel(conn).FindByCondition(ctx, nil, resultChain.Build()...)
	if err != nil {
		return nil, err
	}
	var scoreSum float64
	riskUuids := make(map[string]struct{})
	lowScores := make([]types.HrmPerfDashboardLowScore, 0)
	for _, r := range results {
		scoreSum += r.Score
		if r.Score < lowScore {
			if _, ok := riskUuids[r.EmployeeUuid]; !ok {
				riskUuids[r.EmployeeUuid] = struct{}{}
			}
			lowScores = append(lowScores, types.HrmPerfDashboardLowScore{
				PeriodUuid:   r.PeriodUuid,
				PeriodNo:     r.PeriodNo,
				EmployeeUuid: r.EmployeeUuid,
				EmployeeName: r.EmployeeName,
				CompanyName:  r.CompanyName,
				OrgName:      r.OrgName,
				Score:        r.Score,
				Grade:        r.Grade,
				Coefficient:  r.Coefficient,
				PerfAmount:   r.PerfAmount,
				Signed:       r.Signed,
			})
		}
	}
	avgScore := 0.0
	if len(results) > 0 {
		avgScore = round2(scoreSum / float64(len(results)))
	}

	// ---- 名单员工姓名快照（appeals/overdueTasks 借结果行/员工表快照） ----
	if len(employeeNameByUuid) > 0 {
		// 优先结果行快照
		resByEmp := make(map[string]string, len(results))
		for _, r := range results {
			resByEmp[r.EmployeeUuid] = r.EmployeeName
		}
		missing := make([]string, 0)
		for uuid := range employeeNameByUuid {
			if _, ok := resByEmp[uuid]; !ok {
				missing = append(missing, uuid)
			}
		}
		if len(missing) > 0 {
			rows, rerr := employeeRowsByUuids(ctx, conn, missing)
			if rerr != nil {
				return nil, rerr
			}
			for uuid, row := range rows {
				resByEmp[uuid] = row.Name
			}
		}
		for i := range appeals {
			appeals[i].EmployeeName = resByEmp[appeals[i].EmployeeUuid]
		}
		for i := range overdueTasks {
			overdueTasks[i].EmployeeName = resByEmp[overdueTasks[i].EmployeeUuid]
		}
	}

	return &types.DashboardResponse{
		Metrics: types.HrmPerfDashboardMetric{
			OngoingPeriodCount: ongoingCount,
			CompletionRate:     completionRate,
			OverdueTaskCount:   overdueCount,
			RiskEmployeeCount:  int64(len(riskUuids)),
			AvgScore:           avgScore,
			AppealCount:        appealCount,
			LowScore:           lowScore,
		},
		LowScores:    lowScores,
		OverdueTasks: overdueTasks,
		Appeals:      appeals,
	}, nil
}
