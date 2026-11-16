package perf_assessment

import (
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_assessment"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_period"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

// record builders shared by List / GetDetail.

func itemsSnapshotToTypes(items []ItemSnapshot) []types.AssessmentItemSnapshot {
	out := make([]types.AssessmentItemSnapshot, 0, len(items))
	for _, item := range items {
		out = append(out, types.AssessmentItemSnapshot{
			IndicatorUuid: item.IndicatorUuid,
			IndicatorCode: item.IndicatorCode,
			IndicatorName: item.IndicatorName,
			IndicatorType: item.IndicatorType,
			Unit:          item.Unit,
			ScoreRule:     item.ScoreRule,
			DataSource:    item.DataSource,
			Weight:        item.Weight,
		})
	}
	return out
}

func scoreItemsToTypes(items []ScoreItem) []types.ScoreItemRequest {
	out := make([]types.ScoreItemRequest, 0, len(items))
	for _, item := range items {
		out = append(out, types.ScoreItemRequest{
			IndicatorUuid: item.IndicatorUuid,
			Score:         item.Score,
		})
	}
	return out
}

// assessmentRecord maps the model row to the API record.
// period may be nil (batch lookup miss — keep display snapshots empty).
func assessmentRecord(row *model.HrmPerfAssessment, items []types.AssessmentItemSnapshot,
	period *hrm_perf_period.HrmPerfPeriod, employeeName, templateName, appealStatus string) types.HrmPerfAssessment {
	rec := types.HrmPerfAssessment{
		Uuid:         row.Uuid,
		AssessmentNo: row.AssessmentNo,
		PeriodUuid:   row.PeriodUuid,
		EmployeeUuid: row.EmployeeUuid,
		EmployeeName: employeeName,
		TemplateUuid: row.TemplateUuid,
		TemplateName: templateName,
		Items:        items,
		CurrentNode:  row.CurrentNode,
		TaskStatus:   row.TaskStatus,
		FinalScore:   row.FinalScore,
		FinalGrade:   row.FinalGrade,
		SelfScore:    row.SelfScore,
		SelfGrade:    row.SelfGrade,
		ReviewScore:  row.ReviewScore,
		ReviewGrade:  row.ReviewGrade,
		AppealStatus: appealStatus,
		CreateTime:   row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:   row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
	if period != nil {
		rec.PeriodNo = period.PeriodNo
		rec.PeriodName = period.Name
	}
	return rec
}
