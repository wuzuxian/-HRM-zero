package training_session

import (
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
)

// sessionConditionByUuid is the shared uuid-condition helper for Edit/Finish.
func sessionConditionByUuid(sessionUuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_training_session.Uuid, sessionUuid).Build()
}
