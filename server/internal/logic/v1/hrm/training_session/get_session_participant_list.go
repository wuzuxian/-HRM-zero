package training_session

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
)

type GetSessionParticipantList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetSessionParticipantList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetSessionParticipantList {
	return &GetSessionParticipantList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetSessionParticipantList 班次学员列表（选人弹窗回显）：绑定行 + 学员账号/姓名.
func (l *GetSessionParticipantList) GetSessionParticipantList(req *types.GetSessionParticipantListRequest) (resp *types.GetSessionParticipantListResponse, err error) {
	participants, err := l.svcCtx.Model.HrmTrainingSessionParticipant.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_training_session_participant.SessionUuid, req.SessionUuid).
			OrderByAsc(hrm_training_session_participant.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	// enrich 学员信息 (账号/姓名) in one batched IN query
	userUuids := make([]string, 0, len(participants))
	for _, p := range participants {
		userUuids = append(userUuids, p.UserUuid)
	}
	users := make(map[string]*manage_user.ManageUser, len(userUuids))
	if len(userUuids) > 0 {
		rows, uerr := l.svcCtx.Model.ManageUser.FindByCondition(l.ctx, nil,
			condition.NewChain().In(manage_user.Uuid, userUuids).Build()...)
		if uerr != nil {
			return nil, uerr
		}
		for _, u := range rows {
			users[u.Uuid] = u
		}
	}

	records := make([]types.SessionParticipant, 0, len(participants))
	for _, p := range participants {
		rec := types.SessionParticipant{
			Uuid:     p.Uuid,
			UserUuid: p.UserUuid,
		}
		if u := users[p.UserUuid]; u != nil {
			rec.Username = u.Username
			rec.NickName = u.Nickname
		}
		records = append(records, rec)
	}

	return &types.GetSessionParticipantListResponse{Records: records}, nil
}
