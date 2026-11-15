package training_session

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
)

type SetSessionParticipants struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSetSessionParticipants(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SetSessionParticipants {
	return &SetSessionParticipants{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// SetSessionParticipants 选人（维护班次学员）：全量覆盖 hrm_training_session_participant
// 绑定行（删全插，幂等）。HRM-29 签到、HRM-31 我的班次、HRM-32 证书发放都以这批行为准.
func (l *SetSessionParticipants) SetSessionParticipants(req *types.SetSessionParticipantsRequest) (resp *types.SetSessionParticipantsResponse, err error) {
	// 班次须存在
	if _, err = l.svcCtx.Model.HrmTrainingSession.FindOneByUuid(l.ctx, nil, req.SessionUuid); err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TrainingSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}

	// dedup while preserving order
	seen := make(map[string]struct{}, len(req.UserUuids))
	userUuids := make([]string, 0, len(req.UserUuids))
	for _, u := range req.UserUuids {
		if u == "" {
			continue
		}
		if _, dup := seen[u]; dup {
			continue
		}
		seen[u] = struct{}{}
		userUuids = append(userUuids, u)
	}

	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// full replace: delete existing bindings, insert the new selection
		if err := l.svcCtx.Model.HrmTrainingSessionParticipant.DeleteByCondition(ctx, session,
			condition.NewChain().Equal(hrm_training_session_participant.SessionUuid, req.SessionUuid).Build()...); err != nil {
			return err
		}
		for _, u := range userUuids {
			if err := l.svcCtx.Model.HrmTrainingSessionParticipant.InsertV2(ctx, session,
				&hrm_training_session_participant.HrmTrainingSessionParticipant{
					Uuid:        uuid.New().String(),
					SessionUuid: req.SessionUuid,
					UserUuid:    u,
				}); err != nil {
				return err
			}
		}
		return nil
	}); txErr != nil {
		return nil, txErr
	}

	return &types.SetSessionParticipantsResponse{}, nil
}
