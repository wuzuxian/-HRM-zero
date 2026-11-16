package position

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/position"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Delete 批量删除招聘职位：普通实体，删除不限状态。
// 每个 uuid 必须存在，任一不存在整批拒绝（10580）。
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return nil, status.ErrorMessage(errcodes.PositionInvalidCode, "请选择要删除的招聘职位")
	}
	for _, uuid := range req.Uuids {
		if _, err = l.svcCtx.Model.HrmRecruitmentPosition.FindOneByCondition(l.ctx, nil,
			condition.NewChain().Equal(hrm_recruitment_position.Uuid, uuid).Build()...); err != nil {
			return nil, status.ErrorMessage(errcodes.PositionNotFoundCode, "招聘职位不存在")
		}
	}
	for _, uuid := range req.Uuids {
		if err = l.svcCtx.Model.HrmRecruitmentPosition.DeleteByCondition(l.ctx, nil,
			condition.NewChain().Equal(hrm_recruitment_position.Uuid, uuid).Build()...); err != nil {
			return nil, err
		}
	}
	l.Infof("position: 招聘职位已删除 count=%d", len(req.Uuids))
	return &types.DeleteResponse{}, nil
}
