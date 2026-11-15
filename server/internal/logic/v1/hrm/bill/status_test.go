package bill

import "testing"

// TestNextStatus specifies the shared 单据状态 state machine (GLOSSARY 单据状态):
// 未提交 / 审批中 / 审批通过 / 审批不通过 / 已撤销.
// Legal transitions:
//   - submit  未提交|审批不通过|已撤销 → 审批中
//   - approve 审批中 → 审批通过
//   - reject  审批中 → 审批不通过
//   - revoke  审批中 → 已撤销
func TestNextStatus(t *testing.T) {
	tests := []struct {
		name   string
		from   string
		action Action
		want   string
		ok     bool
	}{
		// submit
		{"submit from unsubmitted", StatusUnsubmitted, ActionSubmit, StatusApproving, true},
		{"resubmit from rejected", StatusRejected, ActionSubmit, StatusApproving, true},
		{"resubmit from revoked", StatusRevoked, ActionSubmit, StatusApproving, true},
		{"submit from approving is illegal", StatusApproving, ActionSubmit, "", false},
		{"submit from approved is illegal", StatusApproved, ActionSubmit, "", false},
		// approve
		{"approve from approving", StatusApproving, ActionApprove, StatusApproved, true},
		{"approve from unsubmitted is illegal", StatusUnsubmitted, ActionApprove, "", false},
		{"approve from approved is illegal", StatusApproved, ActionApprove, "", false},
		{"approve from rejected is illegal", StatusRejected, ActionApprove, "", false},
		{"approve from revoked is illegal", StatusRevoked, ActionApprove, "", false},
		// reject
		{"reject from approving", StatusApproving, ActionReject, StatusRejected, true},
		{"reject from unsubmitted is illegal", StatusUnsubmitted, ActionReject, "", false},
		{"reject from rejected is illegal", StatusRejected, ActionReject, "", false},
		{"reject from approved is illegal", StatusApproved, ActionReject, "", false},
		{"reject from revoked is illegal", StatusRevoked, ActionReject, "", false},
		// revoke
		{"revoke from approving", StatusApproving, ActionRevoke, StatusRevoked, true},
		{"revoke from unsubmitted is illegal", StatusUnsubmitted, ActionRevoke, "", false},
		{"revoke from approved is illegal", StatusApproved, ActionRevoke, "", false},
		{"revoke from rejected is illegal", StatusRejected, ActionRevoke, "", false},
		{"revoke from revoked is illegal", StatusRevoked, ActionRevoke, "", false},
		// unknown inputs
		{"unknown action is illegal", StatusUnsubmitted, Action("delete"), "", false},
		{"unknown status is illegal", "deleted", ActionSubmit, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NextStatus(tt.from, tt.action)
			if ok != tt.ok || got != tt.want {
				t.Errorf("NextStatus(%q, %q) = (%q, %v), want (%q, %v)",
					tt.from, tt.action, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// TestCanTransit checks the boolean form agrees with NextStatus.
func TestCanTransit(t *testing.T) {
	if !CanTransit(StatusUnsubmitted, ActionSubmit) {
		t.Error("CanTransit(unsubmitted, submit) = false, want true")
	}
	if CanTransit(StatusApproved, ActionSubmit) {
		t.Error("CanTransit(approved, submit) = true, want false")
	}
}
