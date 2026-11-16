package perf_assessment

import (
	"errors"

	"github.com/jzero-io/jzero/core/status"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
)

// 考核任务域常量（HRM-43，GLOSSARY「考核任务」：五阶段任务链 + 申诉）。

// 五阶段当前节点（hrm_perf_assessment.current_node）。
const (
	NodePlanConfirm   = "1" // 目标确认（被考核人）
	NodeSelfReview    = "2" // 自评填报（被考核人）
	NodeLeaderReview  = "3" // 上级评审（有 hrm:perf_assessment:review 码者）
	NodeCalibration   = "4" // 校准填报（校准人，有 hrm:perf_assessment:calibrate 码者）
	NodeResultConfirm = "5" // 结果确认（被考核人确认或发起申诉）
)

// 任务状态（hrm_perf_assessment.task_status）。
const (
	TaskStatusOngoing  = "0" // 进行中
	TaskStatusDone     = "1" // 已完成
	TaskStatusAppealed = "2" // 申诉中
)

// 申诉状态（hrm_perf_appeal.status）。
const (
	AppealPending  = "0" // 待处理
	AppealUpheld   = "1" // 已裁决-维持原结果
	AppealAdjusted = "2" // 已裁决-调整得分
)

// API 权限码（须与 route2code.go 生成映射及菜单种子一致；
// review/calibrate 是前端按钮码命名，API 码以端点动作为准）。
const (
	PermAssessmentReview    = "v1:hrm:perf_assessment:submitLeaderReview" // 上级评审
	PermAssessmentCalibrate = "v1:hrm:perf_assessment:submitCalibration"  // 校准填报 + 申诉裁决
)

// AssessmentDoneNode is the 当前节点 value once the whole chain is done
// (结果确认完成 — node 5 has no successor; the row keeps node 5 + status 1).
const AssessmentDoneNode = NodeResultConfirm

// sentinel errors（wrapBizErr 映射到注册错误码 10660-10667）。
var (
	// ErrAssessmentNotFound: 考核任务不存在。
	ErrAssessmentNotFound = errors.New("考核任务不存在")
	// ErrAssessmentInvalid: 字段非法（逐项得分缺失、申诉原因为空、分数越界等）。
	ErrAssessmentInvalid = errors.New("考核任务字段非法")
	// ErrNodeInvalid: 五阶段前置守卫失败（当前节点不允许该操作）。
	ErrNodeInvalid = errors.New("当前节点不允许该操作")
	// ErrOperatorDenied: 越权填报拒绝。
	ErrOperatorDenied = errors.New("无权执行该操作")
	// ErrPeriodNotOngoing: 批次未发布（仅进行中批次可生成任务）。
	ErrPeriodNotOngoing = errors.New("批次未发布，无法生成考核任务")
	// ErrNoParticipants: 批次无参与人。
	ErrNoParticipants = errors.New("批次无参与人")
	// ErrAppealNotFound: 申诉不存在。
	ErrAppealNotFound = errors.New("申诉不存在")
	// ErrAppealExists: 已有待处理申诉。
	ErrAppealExists = errors.New("该考核任务已有待处理申诉")
	// ErrNoSeqOutOfRange: PA 考核单号流水超出 1..9999。
	ErrNoSeqOutOfRange = errors.New("考核单号流水超出范围")
)

// wrapBizErr maps sentinel errors to registered business error codes.
func wrapBizErr(err error) error {
	switch {
	case errors.Is(err, ErrAssessmentNotFound):
		return status.Wrap(errcodes.HrmPerfAssessmentNotFoundCode, err)
	case errors.Is(err, ErrNodeInvalid):
		return status.Wrap(errcodes.HrmPerfAssessmentNodeInvalidCode, err)
	case errors.Is(err, ErrOperatorDenied):
		return status.Wrap(errcodes.HrmPerfAssessmentOperatorDeniedCode, err)
	case errors.Is(err, ErrPeriodNotOngoing):
		return status.Wrap(errcodes.HrmPerfAssessmentPeriodNotOngoingCode, err)
	case errors.Is(err, ErrNoParticipants):
		return status.Wrap(errcodes.HrmPerfAssessmentNoParticipantsCode, err)
	case errors.Is(err, ErrAppealNotFound):
		return status.Wrap(errcodes.HrmPerfAssessmentAppealNotFoundCode, err)
	case errors.Is(err, ErrAppealExists):
		return status.Wrap(errcodes.HrmPerfAssessmentAppealExistsCode, err)
	case errors.Is(err, ErrAssessmentInvalid), errors.Is(err, ErrNoSeqOutOfRange):
		return status.Wrap(errcodes.HrmPerfAssessmentInvalidCode, err)
	default:
		return err
	}
}
