package entry

import (
	"context"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_work"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
)

// insertChildren writes the three 入职单 snapshot sub-tables (工作经历/教育经历/
// 家庭成员) inside the caller's transaction. Only non-empty rows are persisted.
func insertChildren(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, entryUuid string,
	works []types.HrmEntryBillWork, educations []types.HrmEntryBillEducation, families []types.HrmEntryBillFamily) error {
	if len(works) > 0 {
		rows := make([]*hrm_entry_bill_work.HrmEntryBillWork, 0, len(works))
		for _, item := range works {
			rows = append(rows, &hrm_entry_bill_work.HrmEntryBillWork{
				Uuid:      uuid.New().String(),
				EntryUuid: entryUuid,
				Company:   item.Company,
				Position:  item.Position,
				StartDate: item.StartDate,
				EndDate:   item.EndDate,
				Remark:    item.Remark,
			})
		}
		if err := svcCtx.Model.HrmEntryBillWork.BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}

	if len(educations) > 0 {
		rows := make([]*hrm_entry_bill_education.HrmEntryBillEducation, 0, len(educations))
		for _, item := range educations {
			rows = append(rows, &hrm_entry_bill_education.HrmEntryBillEducation{
				Uuid:      uuid.New().String(),
				EntryUuid: entryUuid,
				School:    item.School,
				Major:     item.Major,
				Education: item.Education,
				StartDate: item.StartDate,
				EndDate:   item.EndDate,
				Remark:    item.Remark,
			})
		}
		if err := svcCtx.Model.HrmEntryBillEducation.BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}

	if len(families) > 0 {
		rows := make([]*hrm_entry_bill_family.HrmEntryBillFamily, 0, len(families))
		for _, item := range families {
			rows = append(rows, &hrm_entry_bill_family.HrmEntryBillFamily{
				Uuid:      uuid.New().String(),
				EntryUuid: entryUuid,
				Name:      item.Name,
				Relation:  item.Relation,
				Phone:     item.Phone,
				WorkUnit:  item.WorkUnit,
				Remark:    item.Remark,
			})
		}
		if err := svcCtx.Model.HrmEntryBillFamily.BulkInsert(ctx, session, rows); err != nil {
			return err
		}
	}
	return nil
}
