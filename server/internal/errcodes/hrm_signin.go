// Package errcodes — HRM-29 培训签到业务错误码（1064x 段）。
// 开工时 grep 确认 1064x 空闲（1060x 加班 / 1061x 薪酬方案 / 1062x 薪资档案 / 1063x 看板已占用）。
package errcodes

import "github.com/jzero-io/jzero/core/status"

const (
	// SigninSessionNotFoundCode 班次不存在
	SigninSessionNotFoundCode = 10640
	// SigninTokenInvalidCode 课堂动态码无效（不存在/已轮换）
	SigninTokenInvalidCode = 10641
	// SigninTokenExpiredCode 课堂动态码已过期（短时效轮换）
	SigninTokenExpiredCode = 10642
	// SigninNotParticipantCode 签到人不在该班次学员名册中
	SigninNotParticipantCode = 10643
	// SigninSessionFinishedCode 班次已结束，不允许签到/补签
	SigninSessionFinishedCode = 10644
	// SigninParticipantNotFoundCode 签到明细行（participant row）不存在
	SigninParticipantNotFoundCode = 10645
)

func init() {
	status.Register(SigninSessionNotFoundCode)
	status.Register(SigninTokenInvalidCode)
	status.Register(SigninTokenExpiredCode)
	status.Register(SigninNotParticipantCode)
	status.Register(SigninSessionFinishedCode)
	status.Register(SigninParticipantNotFoundCode)
}
