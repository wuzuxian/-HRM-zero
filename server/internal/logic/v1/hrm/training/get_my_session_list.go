package training

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training"
)

type GetMySessionList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetMySessionList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetMySessionList {
	return &GetMySessionList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetMySessionList 我的班次：当前登录用户作为学员（hrm_training_session_participant）
// 的班次列表，支持按状态筛选、分页。
func (l *GetMySessionList) GetMySessionList(req *types.GetMySessionListRequest) (resp *types.GetMySessionListResponse, err error) {
	userUuid := currentUserUuid(l.ctx)

	// 学员绑定为空 → 空列表，不产生越权暴露
	participants, pErr := l.svcCtx.Model.HrmTrainingSessionParticipant.FindByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_training_session_participant.UserUuid, userUuid).Build()...)
	if pErr != nil {
		return nil, pErr
	}
	sessionUuids := make([]string, 0, len(participants))
	for _, p := range participants {
		sessionUuids = append(sessionUuids, p.SessionUuid)
	}
	if len(sessionUuids) == 0 {
		return &types.GetMySessionListResponse{PageResponse: types.PageResponse{
			Current: req.Current, Size: req.Size, Total: 0,
		}, Records: []types.HrmMyTrainingSession{}}, nil
	}

	rows, total, err := l.svcCtx.Model.HrmTrainingSession.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			In(hrm_training_session.Uuid, sessionUuids).
			Equal(hrm_training_session.Status, req.Status, condition.WithSkip(req.Status == "")).
			OrderByDesc(hrm_training_session.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmMyTrainingSession, 0, len(rows))
	for _, row := range rows {
		records = append(records, types.HrmMyTrainingSession{
			Uuid:        row.Uuid,
			SessionNo:   row.SessionNo,
			SessionName: row.SessionName,
			Lecturer:    row.Lecturer,
			Location:    row.Location,
			Hours:       row.Hours,
			StartTime:   row.StartTime.Format(time.DateTime),
			EndTime:     row.EndTime.Format(time.DateTime),
			Status:      row.Status,
			Remark:      row.Remark,
			CreateTime:  row.CreateTime.Format(time.DateTime),
			UpdateTime:  row.UpdateTime.Format(time.DateTime),
		})
	}

	return &types.GetMySessionListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}
