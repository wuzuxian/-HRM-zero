package offer

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_offer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/offer"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Edit 修改录用审批单：仅 未提交/审批不通过/已撤销 状态可修改（演示站规则）。
// 单据编号/状态/快照字段由系统维护，编辑只改 计划入职日期/试用期/职级/汇报人
// 与候选人/职位引用（重新过守卫）.
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if err = validateOffer(req.CandidateUuid, req.PositionUuid, req.PlanEntryDate, req.RankLevel); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmOfferBill.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_offer_bill.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.OfferNotFoundCode, "录用审批单不存在")
		}
		return nil, err
	}
	if !canEdit(row.Status) {
		return nil, status.ErrorMessage(errcodes.OfferNotEditableCode,
			"当前单据状态不允许修改（仅 未提交/审批不通过/已撤销 可改）")
	}

	cand, err := ensureCandidateOfferable(l.ctx, l.svcCtx, nil, req.CandidateUuid)
	if err != nil {
		return nil, err
	}
	pos, err := ensurePositionOpen(l.ctx, l.svcCtx, nil, req.PositionUuid)
	if err != nil {
		return nil, err
	}
	orgUuid, orgName := positionOrgSnapshot(l.ctx, l.svcCtx, nil, pos)

	fields := map[string]any{
		"candidate_uuid":   req.CandidateUuid,
		"candidate_name":   cand.Name,
		"position_uuid":    req.PositionUuid,
		"post_name":        pos.PostName,
		"org_uuid":         orgUuid,
		"org_name":         orgName,
		"plan_entry_date":  req.PlanEntryDate,
		"probation_months": req.ProbationMonths,
		"rank_level":       defaultRank(req.RankLevel),
		"reporter_name":    req.ReporterName,
	}
	if err = l.svcCtx.Model.HrmOfferBill.UpdateFieldsByCondition(l.ctx, nil, fields,
		condition.NewChain().Equal(hrm_offer_bill.Uuid, req.Uuid).Build()...); err != nil {
		return nil, err
	}

	l.Infof("offer: 录用审批单已修改 billNo=%s", row.BillNo)
	return &types.EditResponse{}, nil
}
