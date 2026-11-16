package position

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
)

// statusNotFoundRequirement maps a missing 来源需求单 to the 1058x range.
func statusNotFoundRequirement() error {
	return status.ErrorMessage(errcodes.PositionInvalidCode,
		"来源需求单不存在（须引用已审批通过的招聘需求单）")
}

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

// Add 新增招聘职位（spec 用户故事 44）：初始状态 招聘中（open），已录用人数 0。
// 来源需求单接缝（HRM-24 契约）：事务内加载需求单，仅 status='approved' 可引用，
// 非 approved（未提交/审批中/审批不通过/已撤销）拒绝 10582。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validatePosition(req.PostName, req.PostCategory, req.PublishDate, req.Headcount, req.RequirementUuid); err != nil {
		return nil, err
	}

	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := ensureRequirementApproved(ctx, l.svcCtx, session, req.RequirementUuid); err != nil {
			return err
		}
		return l.svcCtx.Model.HrmRecruitmentPosition.InsertV2(ctx, session, &hrm_recruitment_position.HrmRecruitmentPosition{
			Uuid:            uuid.New().String(),
			PostName:        req.PostName,
			RequirementUuid: req.RequirementUuid,
			PostCategory:    req.PostCategory,
			OwnerUuid:       req.OwnerUuid,
			OwnerName:       req.OwnerName,
			Headcount:       req.Headcount,
			HiredCount:      0,
			PublishDate:     req.PublishDate,
			Status:          PositionStatusOpen,
			Remark:          req.Remark,
		})
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("position: 招聘职位已创建 post=%s(%s) requirement=%s headcount=%d owner=%s",
		req.PostName, req.PostCategory, req.RequirementUuid, req.Headcount, req.OwnerName)
	return &types.AddResponse{}, nil
}

// ensureRequirementApproved loads the 来源需求单 and rejects non-approved references
// (HRM-24 seam contract: only status='approved' 需求单 may be published as a position).
func ensureRequirementApproved(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, requirementUuid string) error {
	row, err := svcCtx.Model.HrmRecruitmentRequirement.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_recruitment_requirement.Uuid, requirementUuid).Build()...)
	if err != nil {
		return statusNotFoundRequirement()
	}
	if row.Status != "approved" {
		return status.ErrorMessage(errcodes.PositionRequirementNotApprovedCode,
			"来源需求单未审批通过，不可发布职位（仅审批通过的需求单可引用）")
	}
	return nil
}
