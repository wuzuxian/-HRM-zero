package signin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_signin_token"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

type RefreshSigninToken struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRefreshSigninToken(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *RefreshSigninToken {
	return &RefreshSigninToken{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GenerateSigninToken renders a fresh random token (32 hex chars).
func GenerateSigninToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// RefreshSigninToken 课堂动态码刷新（demo-inventory 6：动态二维码 + 刷新二维码按钮）：
// 每班次仅保一条活跃 token，刷新 = 事务内删旧插新（旧码立即失效，短时效轮换防转发截图）。
// SigninUrl 返回前端路由路径（/hrm/training/signin/scan?token=...），由前端拼 origin 渲染二维码。
func (l *RefreshSigninToken) RefreshSigninToken(req *types.RefreshSigninTokenRequest) (resp *types.RefreshSigninTokenResponse, err error) {
	session, err := l.svcCtx.Model.HrmTrainingSession.FindOneByCondition(l.ctx, nil,
		signinSessionConditionByUuid(req.SessionUuid)...)
	if err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SigninSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}
	if session.Status == SessionStatusFinished {
		return nil, status.ErrorMessage(errcodes.SigninSessionFinishedCode, "班次已结束，不允许刷新签到码")
	}

	token, err := GenerateSigninToken()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expire := now.Add(signinTokenTTL)

	userUuid := currentUserUuid(l.ctx)

	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 单活跃 token：先清掉本班次旧码（轮换即失效）
		if err := l.svcCtx.Model.HrmTrainingSigninToken.DeleteByCondition(ctx, tx,
			condition.NewChain().Equal(hrm_training_signin_token.SessionUuid, req.SessionUuid).Build()...); err != nil {
			return err
		}
		return l.svcCtx.Model.HrmTrainingSigninToken.InsertV2(ctx, tx, &hrm_training_signin_token.HrmTrainingSigninToken{
			Uuid:        uuid.New().String(),
			SessionUuid: req.SessionUuid,
			Token:       token,
			ExpireTime:  expire,
			CreateBy:    userUuid,
		})
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("signin: 刷新课堂动态码 sessionNo=%s token=%s expire=%s", session.SessionNo, maskToken(token), expire.Format(time.DateTime))
	return &types.RefreshSigninTokenResponse{
		Token:      token,
		ExpireTime: expire.Format(time.DateTime),
		SigninUrl:  fmt.Sprintf("/hrm/training/signin/scan?token=%s", token),
	}, nil
}

// maskToken hides the token body in logs.
func maskToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:8] + "..."
}
