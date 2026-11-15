// Package bill is the shared HRM 单据 approval foundation (ADR-0001):
// unified 单据状态 state machine, 单据编号 generator, shared approval record
// storage and the submit/approve/reject/revoke lifecycle. Business write-back
// after approval stays in each document's own logic (see Handler.OnApproved).
//
// Button-code convention (casbin API codes / bare frontend codes):
//	v1:hrm:bill:submit   hrm:bill:submit
//	v1:hrm:bill:approve  hrm:bill:approve
//	v1:hrm:bill:reject   hrm:bill:reject
//	v1:hrm:bill:revoke   hrm:bill:revoke
// Document page menus must list these API codes in their permissions JSON.
package bill

// 单据状态 codes stored in every document table's status column.
// UI labels (未提交/审批中/审批通过/审批不通过/已撤销) map 1:1, see GLOSSARY 单据状态.
const (
	StatusUnsubmitted = "unsubmitted" // 未提交（草稿/暂存）
	StatusApproving   = "approving"   // 审批中
	StatusApproved    = "approved"    // 审批通过
	StatusRejected    = "rejected"    // 审批不通过
	StatusRevoked     = "revoked"     // 已撤销
)

// Statuses lists all valid 单据状态 codes.
var Statuses = []string{
	StatusUnsubmitted,
	StatusApproving,
	StatusApproved,
	StatusRejected,
	StatusRevoked,
}

// Action is a lifecycle action applied to a 单据.
type Action string

const (
	ActionSubmit  Action = "submit"  // 提交
	ActionApprove Action = "approve" // 审批通过
	ActionReject  Action = "reject"  // 审批驳回
	ActionRevoke  Action = "revoke"  // 撤销
)

// transitions maps 单据状态 → legal actions → target 单据状态.
var transitions = map[string]map[Action]string{
	StatusUnsubmitted: {ActionSubmit: StatusApproving},
	StatusApproving: {
		ActionApprove: StatusApproved,
		ActionReject:  StatusRejected,
		ActionRevoke:  StatusRevoked,
	},
	StatusRejected: {ActionSubmit: StatusApproving},
	StatusRevoked:  {ActionSubmit: StatusApproving},
	// StatusApproved is terminal.
}

// NextStatus returns the target 单据状态 of applying action to from.
// ok is false when the transition is illegal for the shared state machine.
func NextStatus(from string, action Action) (target string, ok bool) {
	t, exists := transitions[from]
	if !exists {
		return "", false
	}
	target, exists = t[action]
	return target, exists
}

// CanTransit reports whether action is legal from 单据状态 from.
func CanTransit(from string, action Action) bool {
	_, ok := NextStatus(from, action)
	return ok
}

// IsValidStatus reports whether s is a known 单据状态 code.
func IsValidStatus(s string) bool {
	_, exists := transitions[s]
	return exists
}
