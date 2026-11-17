package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jzero-io/jzero-admin/core-engine/helper/auth"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user_role"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/auth"
)

func CreateToken(secret string, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// userStatusDisabled manage_user.status 真值 '2'=禁用 (ADR 0007: '1' 启用 / '2' 禁用)。
// HRM-53 账号状态真实开关: 登录是状态的消费方, 禁用账号即使密码正确也拒绝。
const userStatusDisabled = "2"

type PwdLogin struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewPwdLogin(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *PwdLogin {
	return &PwdLogin{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

func (l *PwdLogin) PwdLogin(req *types.PwdLoginRequest) (resp *types.LoginResponse, err error) {
	config, err := l.svcCtx.ConfigCenter.GetConfig()
	if err != nil {
		return nil, err
	}

	user, err := l.svcCtx.Model.ManageUser.FindOneByUsername(l.ctx, nil, req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if req.Password != user.Password {
		return nil, errors.New("用户名或密码错误")
	}
	// HRM-53 (ADR 0007 生命周期一致性): 密码校验通过后再看状态, 禁用账号明确拒绝,
	// 文案与"用户名或密码错误"区分开 — 已存在的禁用账号从此真的登不进去 (预期行为变更)。
	if user.Status == userStatusDisabled {
		return nil, errors.New("账号已禁用, 无法登录")
	}
	userRoles, err := l.svcCtx.Model.ManageUserRole.FindByCondition(l.ctx, nil, condition.NewChain().
		Equal(manage_user_role.UserUuid, user.Uuid).
		Build()...)
	if err != nil {
		return nil, err
	}
	var roleIds []string
	for _, userRole := range userRoles {
		roleIds = append(roleIds, userRole.RoleUuid)
	}

	marshal, err := json.Marshal(auth.Auth{
		Uuid:      user.Uuid,
		Username:  user.Username,
		RoleUuids: roleIds,
	})
	if err != nil {
		return nil, err
	}

	var claims map[string]any
	err = json.Unmarshal(marshal, &claims)
	if err != nil {
		return nil, err
	}

	// token 过期时间
	expirationTime := time.Now().Add(time.Duration(config.Jwt.AccessExpire) * time.Second).Unix()
	claims["exp"] = expirationTime

	token, err := CreateToken(l.svcCtx.MustGetConfig().Jwt.AccessSecret, claims)
	if err != nil {
		return nil, err
	}

	claims["exp"] = time.Now().Add(time.Duration(config.Jwt.RefreshExpire) * time.Second).Unix()
	refreshToken, err := CreateToken(l.svcCtx.MustGetConfig().Jwt.AccessSecret, claims)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}
