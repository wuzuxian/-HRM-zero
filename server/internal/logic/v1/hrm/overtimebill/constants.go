package overtimebill

// 单据状态码引用（与共享单据底座 internal/logic/v1/hrm/bill 一致的本地别名，列表/守卫用）.
const (
	StatusUnsubmitted = "unsubmitted"
	StatusApproving   = "approving"
	StatusApproved    = "approved"
	StatusRejected    = "rejected"
	StatusRevoked     = "revoked"
)
