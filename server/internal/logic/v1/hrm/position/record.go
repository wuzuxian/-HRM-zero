package position

import (
	"time"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
)

// positionRecord converts a model row into the API record; requirementBillNo is
// the 来源需求单单据编号 snapshot for list display.
func positionRecord(row *hrm_recruitment_position.HrmRecruitmentPosition, requirementBillNo string) types.HrmPosition {
	return types.HrmPosition{
		Uuid:              row.Uuid,
		PostName:          row.PostName,
		RequirementUuid:   row.RequirementUuid,
		RequirementBillNo: requirementBillNo,
		PostCategory:      row.PostCategory,
		OwnerUuid:         row.OwnerUuid,
		OwnerName:         row.OwnerName,
		Headcount:         row.Headcount,
		HiredCount:        row.HiredCount,
		PublishDate:       row.PublishDate,
		Status:            row.Status,
		Remark:            row.Remark,
		CreateTime:        row.CreateTime.Format(time.DateTime),
		UpdateTime:        row.UpdateTime.Format(time.DateTime),
	}
}
