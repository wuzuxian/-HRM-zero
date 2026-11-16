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
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_signin_token"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

type DoSignin struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDoSignin(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *DoSignin {
	return &DoSignin{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// FindActiveSigninToken resolves the current token row by token string (nil when missing/expired).
func FindActiveSigninToken(ctx context.Context, model hrm_training_signin_token.HrmTrainingSigninTokenModel, token string, now time.Time) (*hrm_training_signin_token.HrmTrainingSigninToken, error) {
	row, err := model.FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(hrm_training_signin_token.Token, token).Build()...)
	if err != nil {
		if err == hrm_training_signin_token.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if !row.ExpireTime.After(now) {
		return nil, nil
	}
	return row, nil
}

// DoSignin 学员扫码签到（jwt+token 双验）：token 活跃 + 当前登录人在该班次名册中 →
// 写 sign_in_time（方式=1 扫码自签，操作人=本人）。幂等：已签直接返回原签到时间（不覆盖）。
// 班次已结束拒绝签到。
func (l *DoSignin) DoSignin(req *types.DoSigninRequest) (resp *types.DoSigninResponse, err error) {
	if req.Token == "" {
		return nil, status.ErrorMessage(errcodes.SigninTokenInvalidCode, "签到码无效，请重新扫码")
	}

	tokenRow, err := FindActiveSigninToken(l.ctx, l.svcCtx.Model.HrmTrainingSigninToken, req.Token, time.Now())
	if err != nil {
		return nil, err
	}
	if tokenRow == nil {
		return nil, status.ErrorMessage(errcodes.SigninTokenInvalidCode, "签到码无效或已过期，请重新扫码")
	}

	session, err := l.svcCtx.Model.HrmTrainingSession.FindOneByCondition(l.ctx, nil,
		signinSessionConditionByUuid(tokenRow.SessionUuid)...)
	if err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SigninSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}
	if session.Status == SessionStatusFinished {
		return nil, status.ErrorMessage(errcodes.SigninSessionFinishedCode, "班次已结束，不再接受签到")
	}

	user := hrmuser.Current(l.ctx)
	if user.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.SigninNotParticipantCode, "请先登录后再扫码签到")
	}

	participant, err := l.svcCtx.Model.HrmTrainingSessionParticipant.FindOneByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_training_session_participant.SessionUuid, tokenRow.SessionUuid).
			Equal(hrm_training_session_participant.UserUuid, user.Uuid).
			Build()...)
	if err != nil {
		if err == hrm_training_session_participant.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SigninNotParticipantCode, "您不在该班次学员名册中")
		}
		return nil, err
	}

	// 幂等：已签返回原签到时间（不覆盖，保留首次真实签到）
	if participant.SignInTime.Valid {
		return &types.DoSigninResponse{
			SessionUuid: tokenRow.SessionUuid,
			SessionName: session.SessionName,
			SignInTime:  formatNullTime(participant.SignInTime),
		}, nil
	}

	now := time.Now()
	if err = l.svcCtx.Model.HrmTrainingSessionParticipant.UpdateFieldsByCondition(l.ctx, nil,
		map[string]any{
			"sign_in_time":   now,
			"sign_in_method": SignInMethodSelf,
			"sign_in_by":     user.Uuid,
			"sign_in_reason": "",
		},
		condition.NewChain().Equal(hrm_training_session_participant.Uuid, participant.Uuid).Build()...); err != nil {
		return nil, err
	}

	l.Infof("signin: 学员扫码签到 sessionNo=%s user=%s", session.SessionNo, user.Username)
	return &types.DoSigninResponse{
		SessionUuid: tokenRow.SessionUuid,
		SessionName: session.SessionName,
		SignInTime:  now.Format(time.DateTime),
	}, nil
}
