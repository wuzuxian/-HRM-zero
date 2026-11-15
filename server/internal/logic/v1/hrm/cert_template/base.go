package cert_template

import (
	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
)

// conditionUuid is the shared uuid-condition helper.
func conditionUuid(uuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_training_cert_template.Uuid, uuid).Build()
}
