package signin

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

type SigninForParticipant struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewSigninForParticipant(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *SigninForParticipant {
	return &SigninForParticipant{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// SigninForParticipant 管理员补签（demo-inventory 6：未到名单 + 确认补签，记录方式/操作人/原因）：
// 方式=2 补签，操作人=当前登录管理员，原因=弹窗录入（可选）。幂等：已签行重复补签直接返回成功
// （明细保持首次记录，不覆盖）——补签本身是登记性质操作。
// 班次已结束仍允许补签（课后核对名单是补签的主要场景），已结束拒绝的是扫码自签与动态码刷新。
func (l *SigninForParticipant) SigninForParticipant(req *types.SigninForParticipantRequest) (resp *types.SigninForParticipantResponse, err error) {
	participant, err := l.svcCtx.Model.HrmTrainingSessionParticipant.FindOneByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_training_session_participant.Uuid, req.Uuid).Build()...)
	if err != nil {
		if err == hrm_training_session_participant.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SigninParticipantNotFoundCode, "签到明细行不存在")
		}
		return nil, err
	}

	if _, err = l.svcCtx.Model.HrmTrainingSession.FindOneByCondition(l.ctx, nil,
		signinSessionConditionByUuid(participant.SessionUuid)...); err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SigninSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}

	// 幂等：已签（自签或补签）不覆盖首次记录
	if participant.SignInTime.Valid {
		return &types.SigninForParticipantResponse{}, nil
	}

	user := hrmuser.Current(l.ctx)

	if err = l.svcCtx.Model.HrmTrainingSessionParticipant.UpdateFieldsByCondition(l.ctx, nil,
		map[string]any{
			"sign_in_time":   time.Now(),
			"sign_in_method": SignInMethodFor,
			"sign_in_by":     user.Uuid,
			"sign_in_reason": req.Reason,
		},
		condition.NewChain().Equal(hrm_training_session_participant.Uuid, participant.Uuid).Build()...); err != nil {
		return nil, err
	}

	l.Infof("signin: 管理员补签 participant=%s operator=%s reason=%s", participant.Uuid, user.Username, req.Reason)
	return &types.SigninForParticipantResponse{}, nil
}
