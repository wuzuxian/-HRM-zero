package offer

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_offer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/offer"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 录用审批单分页列表，支持按 单据编号（模糊）/单据状态/候选人 筛选
// （demo-inventory §5：单据编号/状态/候选人）。默认按「我的单据」过滤
// （创建人=当前用户，spec 用户故事 4）；管理员级角色（持共享审批码 v1:hrm:bill:approve）可看全部，
// 传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !hrmuser.CanSeeAllBills(l.ctx, l.svcCtx, user)

	chain := condition.NewChain().
		Equal(hrm_offer_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_offer_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_offer_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Equal(hrm_offer_bill.CandidateUuid, req.CandidateUuid, condition.WithSkip(req.CandidateUuid == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_offer_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmOfferBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmOffer, 0, len(rows))
	for _, row := range rows {
		records = append(records, offerRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func offerRecord(row *hrm_offer_bill.HrmOfferBill) types.HrmOffer {
	return types.HrmOffer{
		Uuid:            row.Uuid,
		BillNo:          row.BillNo,
		Status:          row.Status,
		CandidateUuid:   row.CandidateUuid,
		CandidateName:   row.CandidateName,
		PositionUuid:    row.PositionUuid,
		PostName:        row.PostName,
		OrgUuid:         row.OrgUuid,
		OrgName:         row.OrgName,
		PlanEntryDate:   row.PlanEntryDate,
		ProbationMonths: row.ProbationMonths,
		RankLevel:       row.RankLevel,
		ReporterName:    row.ReporterName,
		EntryBillUuid:   row.EntryBillUuid,
		CreatorUuid:     row.CreateBy,
		CreatorName:     row.CreatorName,
		CreateTime:      row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:      row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
