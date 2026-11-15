package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// HRM 培训管理：培训计划单（HRM-28，HR216）业务错误码。1040x 段
	// （1039x 为绩效重编号修复预留，integration-warnings.md #4）。
	// TrainingPlanNotFoundCode 培训计划单不存在
	TrainingPlanNotFoundCode = 10400
	// TrainingPlanInvalidCode 培训计划单参数非法（名称/年度/类型/日期区间等）
	TrainingPlanInvalidCode = 10401
	// TrainingPlanNotEditableCode 单据当前状态不允许修改/删除（仅 未提交/审批不通过/已撤销 可改）
	TrainingPlanNotEditableCode = 10402
	// TrainingPlanPublishedGuardCode 已发布的计划不允许修改/删除
	TrainingPlanPublishedGuardCode = 10403

	// HRM 培训管理：培训班次（HRM-28）业务错误码。
	// TrainingSessionNotFoundCode 班次不存在
	TrainingSessionNotFoundCode = 10410
	// TrainingSessionInvalidCode 班次参数非法（名称/时间区间等）
	TrainingSessionInvalidCode = 10411
	// TrainingSessionPlanNotPublishedCode 关联培训计划未发布（仅已发布计划可建班次）
	TrainingSessionPlanNotPublishedCode = 10412
	// TrainingSessionStatusInvalidCode 班次状态流转非法（结束操作要求非已结束状态）
	TrainingSessionStatusInvalidCode = 10413
)

func init() {
	status.Register(TrainingPlanNotFoundCode)
	status.Register(TrainingPlanInvalidCode)
	status.Register(TrainingPlanNotEditableCode)
	status.Register(TrainingPlanPublishedGuardCode)

	status.Register(TrainingSessionNotFoundCode)
	status.Register(TrainingSessionInvalidCode)
	status.Register(TrainingSessionPlanNotPublishedCode)
	status.Register(TrainingSessionStatusInvalidCode)
}
