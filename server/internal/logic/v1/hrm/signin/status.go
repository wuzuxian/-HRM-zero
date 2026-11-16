package signin

// 签到状态口径（与 hrm_training_session 状态机一致，GLOSSARY 培训班次）：
// 1=未开始 2=进行中 3=已结束。签到（自签/补签）允许在 未开始/进行中 进行，已结束锁死。
const (
	SessionStatusNotStarted = "1"
	SessionStatusOngoing    = "2"
	SessionStatusFinished   = "3"
)

// signinAllowed reports whether signin/补签 is open for the session.
func signinAllowed(status string) bool {
	return status == SessionStatusNotStarted || status == SessionStatusOngoing
}
