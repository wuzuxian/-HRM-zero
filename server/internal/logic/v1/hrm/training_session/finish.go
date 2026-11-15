package training_session

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
)

type Finish struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewFinish(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Finish {
	return &Finish{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Finish 班次结束（行操作）：状态置 3=已结束；已结束 is terminal.
// 签到明细（HRM-29）与证书发放（HRM-32）在班次结束后流转。
func (l *Finish) Finish(req *types.FinishRequest) (resp *types.FinishResponse, err error) {
	row, err := l.svcCtx.Model.HrmTrainingSession.FindOneByUuid(l.ctx, nil, req.Uuid)
	if err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.TrainingSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}
	if row.Status == SessionStatusFinished {
		return nil, status.ErrorMessage(errcodes.TrainingSessionStatusInvalidCode,
			"班次已结束，不允许重复结束")
	}

	if err = l.svcCtx.Model.HrmTrainingSession.UpdateFieldsByCondition(l.ctx, nil,
		map[string]any{"status": SessionStatusFinished},
		sessionConditionByUuid(req.Uuid)...); err != nil {
		return nil, err
	}

	l.Infof("training_session: 班次已结束 sessionNo=%s", row.SessionNo)
	return &types.FinishResponse{}, nil
}
