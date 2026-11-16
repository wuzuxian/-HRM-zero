// HRM-45 批次结果审批与结果发布/签收业务错误码。1069x 段
// （2026-11-16 grep errcodes/*.go 确认 1069x 空闲——已用 1060x 加班 / 1061x 薪酬方案 /
// 1062x 薪资档案 / 1063x 看板+考勤统计+绩效批次 / 1064x 培训签到 / 1065x offer /
// 1066x 考核任务）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// ResultApprovalNotFoundCode 批次结果审批单不存在
	ResultApprovalNotFoundCode = 10690
	// ResultApprovalInvalidCode 结果审批字段非法 / 批次状态不符（非 '4'结果已发布 不可提交）
	// / 一批次一单的重复提交被拒
	ResultApprovalInvalidCode = 10691
	// ResultNotFoundCode 绩效结果行不存在
	ResultNotFoundCode = 10692
	// ResultInvalidCode 绩效结果字段非法（已作废、未发布不可签收等）
	ResultInvalidCode = 10693
	// ResultSignDeniedCode 签收越权（仅员工本人可签收本人结果行）
	ResultSignDeniedCode = 10694
)

func init() {
	status.Register(ResultApprovalNotFoundCode)
	status.Register(ResultApprovalInvalidCode)
	status.Register(ResultNotFoundCode)
	status.Register(ResultInvalidCode)
	status.Register(ResultSignDeniedCode)
}
