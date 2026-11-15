package training_session

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
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

// Delete 删除/批量删除培训班次（级联清理学员绑定行）.
func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return &types.DeleteResponse{}, nil
	}

	if err = l.svcCtx.Model.HrmTrainingSession.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_training_session.Uuid,
		Operator: condition.In,
		Value:    req.Uuids,
	}); err != nil {
		return nil, err
	}

	// cascade: 学员绑定行 (hrm_training_session_participant) 不留孤儿
	if err = l.svcCtx.Model.HrmTrainingSessionParticipant.DeleteByCondition(l.ctx, nil, condition.Condition{
		Field:    hrm_training_session_participant.SessionUuid,
		Operator: condition.In,
		Value:    req.Uuids,
	}); err != nil {
		return nil, err
	}

	return &types.DeleteResponse{}, nil
}
