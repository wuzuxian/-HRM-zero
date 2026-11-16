package errcodes

import "github.com/jzero-io/jzero/core/status"

// HRM-43 考核任务链业务错误码（1064x 段；1063x 已被看板/考勤统计/绩效批次占用，
// 本票取 10640-10647，grep errcodes/*.go 于 2026-11-16 确认 1064x 空闲）。
const (
	// HrmPerfAssessmentNotFoundCode 考核任务不存在
	HrmPerfAssessmentNotFoundCode = 10660
	// HrmPerfAssessmentInvalidCode 考核任务字段非法（逐项得分缺失/申诉原因空等）
	HrmPerfAssessmentInvalidCode = 10661
	// HrmPerfAssessmentNodeInvalidCode 五阶段前置守卫：当前节点不允许该操作（前一节点未完成）
	HrmPerfAssessmentNodeInvalidCode = 10662
	// HrmPerfAssessmentOperatorDeniedCode 越权填报拒绝（仅被考核人可确认/自评；仅评审码者可评审）
	HrmPerfAssessmentOperatorDeniedCode = 10663
	// HrmPerfAssessmentPeriodNotOngoingCode 批次未发布（仅进行中批次可生成任务）
	HrmPerfAssessmentPeriodNotOngoingCode = 10664
	// HrmPerfAssessmentNoParticipantsCode 生成守卫：批次无参与人
	HrmPerfAssessmentNoParticipantsCode = 10665
	// HrmPerfAssessmentAppealNotFoundCode 申诉不存在
	HrmPerfAssessmentAppealNotFoundCode = 10666
	// HrmPerfAssessmentAppealExistsCode 该任务已有待处理申诉，不可重复发起
	HrmPerfAssessmentAppealExistsCode = 10667
)

func init() {
	status.Register(HrmPerfAssessmentNotFoundCode)
	status.Register(HrmPerfAssessmentInvalidCode)
	status.Register(HrmPerfAssessmentNodeInvalidCode)
	status.Register(HrmPerfAssessmentOperatorDeniedCode)
	status.Register(HrmPerfAssessmentPeriodNotOngoingCode)
	status.Register(HrmPerfAssessmentNoParticipantsCode)
	status.Register(HrmPerfAssessmentAppealNotFoundCode)
	status.Register(HrmPerfAssessmentAppealExistsCode)
}
