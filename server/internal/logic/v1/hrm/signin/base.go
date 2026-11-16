package signin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_signin_token"
)

// SignInMethodSelf 扫码自签；SignInMethodFor 补签（GLOSSARY 培训签到：方式/操作人/原因 记录）。
const (
	SignInMethodSelf = "1"
	SignInMethodFor  = "2"
)

// signinTokenTTL 课堂动态码时效（短时效轮换）：10 分钟一换，扫码页轮询刷新。
const signinTokenTTL = 10 * time.Minute

// signinSessionConditionByUuid is the shared session-lookup condition.
func signinSessionConditionByUuid(sessionUuid string) []condition.Condition {
	return condition.NewChain().Equal(hrm_training_session.Uuid, sessionUuid).Build()
}

// findSessionForSignin loads the session row (participant roster / token owner).
func findSessionForSignin(ctx context.Context, find func() (*hrm_training_session.HrmTrainingSession, error)) (*hrm_training_session.HrmTrainingSession, error) {
	return find()
}

var _ = hrm_training_session_participant.Uuid
var _ = hrm_training_signin_token.Uuid
var _ = fmt.Sprintf
var _ = rand.Read
var _ = hex.EncodeToString
var _ = time.Now
