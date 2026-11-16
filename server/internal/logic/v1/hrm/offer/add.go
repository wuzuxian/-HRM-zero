package offer

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	candidate "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/candidate"
	position "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/position"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_offer_bill"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_candidate"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/offer"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增录用审批单（HR208）：对 offer（已发Offer）阶段候选人发起录用。
// 事务内守卫：候选人必须 status='offer'（10653）、职位必须 status='open'（10654）；
// 候选人姓名/岗位名称/部门（来源需求单快照）落快照。初始状态 未提交，
// 提交/审批/撤销走共享单据底座 v1/hrm/bill（类型码 208）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateOffer(req.CandidateUuid, req.PositionUuid, req.PlanEntryDate, req.RankLevel); err != nil {
		return nil, err
	}

	user := hrmuser.Current(l.ctx)

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		cand, err := ensureCandidateOfferable(l.ctx, l.svcCtx, session, req.CandidateUuid)
		if err != nil {
			return err
		}
		pos, err := ensurePositionOpen(l.ctx, l.svcCtx, session, req.PositionUuid)
		if err != nil {
			return err
		}
		orgUuid, orgName := positionOrgSnapshot(l.ctx, l.svcCtx, session, pos)

		no, err := bill.GenerateBillNo(l.ctx, session, BillTypeCode, timeNow())
		if err != nil {
			return err
		}
		billNo = no
		return l.svcCtx.Model.HrmOfferBill.InsertV2(l.ctx, session, &hrm_offer_bill.HrmOfferBill{
			Uuid:            uuid.New().String(),
			BillNo:          no,
			Status:          bill.StatusUnsubmitted,
			CandidateUuid:   req.CandidateUuid,
			CandidateName:   cand.Name,
			PositionUuid:    req.PositionUuid,
			PostName:        pos.PostName,
			OrgUuid:         orgUuid,
			OrgName:         orgName,
			PlanEntryDate:   req.PlanEntryDate,
			ProbationMonths: req.ProbationMonths,
			RankLevel:       defaultRank(req.RankLevel),
			ReporterName:    req.ReporterName,
			EntryBillUuid:   "",
			Applied:         0,
			CreateBy:        user.Uuid,
			CreatorName:     user.Username,
		})
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("offer: 录用审批单已创建 billNo=%s candidate=%s position=%s planEntry=%s",
		billNo, req.CandidateUuid, req.PositionUuid, req.PlanEntryDate)
	return &types.AddResponse{}, nil
}

// defaultRank 空 职级 默认 初级（junior）.
func defaultRank(rankLevel string) string {
	if rankLevel == "" {
		return RankLevelJunior
	}
	return rankLevel
}

// ensureCandidateOfferable loads the candidate and rejects any status other than
// offer（已发Offer）— 录用审批只对已发Offer的候选人发起（HRM-26 契约）.
func ensureCandidateOfferable(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, candidateUuid string) (*hrm_recruitment_candidate.HrmRecruitmentCandidate, error) {
	m := candidateModelOf(svcCtx, session)
	row, err := m.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_recruitment_candidate.Uuid, candidateUuid).Build()...)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.OfferCandidateNotOfferableCode, "候选人不存在")
	}
	if row.Status != candidate.CandidateStatusOffer {
		return nil, status.ErrorMessage(errcodes.OfferCandidateNotOfferableCode,
			"候选人不在 已发Offer 阶段，不可发起录用审批")
	}
	return row, nil
}

// ensurePositionOpen loads the position and rejects any status other than
// open（招聘中）— 已满/已关闭职位不可发起录用（HRM-25 契约）.
func ensurePositionOpen(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, positionUuid string) (*hrm_recruitment_position.HrmRecruitmentPosition, error) {
	m := positionModelOf(svcCtx, session)
	row, err := m.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_recruitment_position.Uuid, positionUuid).Build()...)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.OfferPositionNotOpenableCode, "招聘职位不存在")
	}
	if row.Status != position.PositionStatusOpen {
		return nil, status.ErrorMessage(errcodes.OfferPositionNotOpenableCode,
			"招聘职位不在 招聘中 状态，不可发起录用审批")
	}
	return row, nil
}

// candidateModelOf / positionModelOf 事务内构造 model（session==nil 用聚合 Model，
// 与 leavebalance.modelOf 同款）.
func candidateModelOf(svcCtx *svc.ServiceContext, session sqlx.Session) hrm_recruitment_candidate.HrmRecruitmentCandidateModel {
	if session == nil {
		return svcCtx.Model.HrmRecruitmentCandidate
	}
	return hrm_recruitment_candidate.NewHrmRecruitmentCandidateModel(sqlx.NewSqlConnFromSession(session))
}

func positionModelOf(svcCtx *svc.ServiceContext, session sqlx.Session) hrm_recruitment_position.HrmRecruitmentPositionModel {
	if session == nil {
		return svcCtx.Model.HrmRecruitmentPosition
	}
	return hrm_recruitment_position.NewHrmRecruitmentPositionModel(sqlx.NewSqlConnFromSession(session))
}

// positionOrgSnapshot 部门快照：职位本身不挂部门，取来源需求单的 org 快照
// （需求单 org_uuid/org_name）；需求单缺失时留空（演示站宽松口径）.
func positionOrgSnapshot(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, pos *hrm_recruitment_position.HrmRecruitmentPosition) (string, string) {
	var m hrm_recruitment_requirement.HrmRecruitmentRequirementModel
	if session == nil {
		m = svcCtx.Model.HrmRecruitmentRequirement
	} else {
		m = hrm_recruitment_requirement.NewHrmRecruitmentRequirementModel(sqlx.NewSqlConnFromSession(session))
	}
	req, err := m.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_recruitment_requirement.Uuid, pos.RequirementUuid).Build()...)
	if err != nil {
		return "", ""
	}
	return req.OrgUuid, req.OrgName
}
