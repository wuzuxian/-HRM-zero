package entry

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_education"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_family"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_entry_bill_work"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/entry"
)

// findChildren loads the three snapshot sub-tables for a 入职单 (detail view).
func findChildren(ctx context.Context, svcCtx *svc.ServiceContext, entryUuid string) ([]types.HrmEntryBillWork, []types.HrmEntryBillEducation, []types.HrmEntryBillFamily, error) {
	works, err := svcCtx.Model.HrmEntryBillWork.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_entry_bill_work.EntryUuid, entryUuid).
			OrderByAsc(hrm_entry_bill_work.Id).Build()...)
	if err != nil {
		return nil, nil, nil, err
	}
	educations, err := svcCtx.Model.HrmEntryBillEducation.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_entry_bill_education.EntryUuid, entryUuid).
			OrderByAsc(hrm_entry_bill_education.Id).Build()...)
	if err != nil {
		return nil, nil, nil, err
	}
	families, err := svcCtx.Model.HrmEntryBillFamily.FindByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_entry_bill_family.EntryUuid, entryUuid).
			OrderByAsc(hrm_entry_bill_family.Id).Build()...)
	if err != nil {
		return nil, nil, nil, err
	}

	workTypes := make([]types.HrmEntryBillWork, 0, len(works))
	for _, row := range works {
		workTypes = append(workTypes, types.HrmEntryBillWork{
			Uuid:      row.Uuid,
			Company:   row.Company,
			Position:  row.Position,
			StartDate: row.StartDate,
			EndDate:   row.EndDate,
			Remark:    row.Remark,
		})
	}
	eduTypes := make([]types.HrmEntryBillEducation, 0, len(educations))
	for _, row := range educations {
		eduTypes = append(eduTypes, types.HrmEntryBillEducation{
			Uuid:      row.Uuid,
			School:    row.School,
			Major:     row.Major,
			Education: row.Education,
			StartDate: row.StartDate,
			EndDate:   row.EndDate,
			Remark:    row.Remark,
		})
	}
	familyTypes := make([]types.HrmEntryBillFamily, 0, len(families))
	for _, row := range families {
		familyTypes = append(familyTypes, types.HrmEntryBillFamily{
			Uuid:     row.Uuid,
			Name:     row.Name,
			Relation: row.Relation,
			Phone:    row.Phone,
			WorkUnit: row.WorkUnit,
			Remark:   row.Remark,
		})
	}
	return workTypes, eduTypes, familyTypes, nil
}

// deleteChildren removes the three snapshot sub-table rows of a 入职单 (used on
// single-document delete inside the caller's transaction).
func deleteChildren(ctx context.Context, session sqlx.Session, svcCtx *svc.ServiceContext, entryUuid string) error {
	if err := svcCtx.Model.HrmEntryBillWork.DeleteByCondition(ctx, session,
		condition.NewChain().Equal(hrm_entry_bill_work.EntryUuid, entryUuid).Build()...); err != nil {
		return err
	}
	if err := svcCtx.Model.HrmEntryBillEducation.DeleteByCondition(ctx, session,
		condition.NewChain().Equal(hrm_entry_bill_education.EntryUuid, entryUuid).Build()...); err != nil {
		return err
	}
	return svcCtx.Model.HrmEntryBillFamily.DeleteByCondition(ctx, session,
		condition.NewChain().Equal(hrm_entry_bill_family.EntryUuid, entryUuid).Build()...)
}
