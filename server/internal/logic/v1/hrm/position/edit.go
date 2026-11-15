package position

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero/core/status"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
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

// Edit 编辑招聘职位：普通实体（非单据底座），无审批锁，任意状态可改。
// 约束：headcount 不得小于当前已录用人数；来源需求单变更同样仅 approved 可引用。
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.PositionInvalidCode, "uuid 必填")
	}
	if err = validatePosition(req.PostName, req.PostCategory, req.PublishDate, req.Headcount, req.RequirementUuid); err != nil {
		return nil, err
	}

	row, err := l.svcCtx.Model.HrmRecruitmentPosition.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_recruitment_position.Uuid, req.Uuid).Build()...)
	if err != nil {
		return nil, status.ErrorMessage(errcodes.PositionNotFoundCode, "招聘职位不存在")
	}

	if req.Headcount < row.HiredCount {
		return nil, status.ErrorMessage(errcodes.PositionInvalidCode,
			"招聘人数不得小于当前已录用人数")
	}

	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if req.RequirementUuid != row.RequirementUuid {
			if err := ensureRequirementApproved(ctx, l.svcCtx, session, req.RequirementUuid); err != nil {
				return err
			}
		}
		row.PostName = req.PostName
		row.RequirementUuid = req.RequirementUuid
		row.PostCategory = req.PostCategory
		row.OwnerUuid = req.OwnerUuid
		row.OwnerName = req.OwnerName
		row.Headcount = req.Headcount
		row.PublishDate = req.PublishDate
		row.Remark = req.Remark
		return l.svcCtx.Model.HrmRecruitmentPosition.Update(ctx, session, row)
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("position: 招聘职位已编辑 uuid=%s post=%s headcount=%d", req.Uuid, req.PostName, req.Headcount)
	return &types.EditResponse{}, nil
}
