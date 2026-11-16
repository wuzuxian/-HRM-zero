package signin

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

type GetSigninStats struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetSigninStats(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetSigninStats {
	return &GetSigninStats{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetSigninStats 统计口径（demo-inventory 6：应到/已签/未到）：
//   - 应到 Expected = participant 名册行数（HRM-28 契约：名册=应到名单）；
//   - 已签 Signed   = sign_in_time 非空的名册行数（扫码自签 + 管理员补签都算已签）；
//   - 未到 Absent   = Expected - Signed。
func (l *GetSigninStats) GetSigninStats(req *types.GetSigninStatsRequest) (resp *types.GetSigninStatsResponse, err error) {
	if _, err = l.svcCtx.Model.HrmTrainingSession.FindOneByCondition(l.ctx, nil,
		signinSessionConditionByUuid(req.SessionUuid)...); err != nil {
		if err == hrm_training_session.ErrNotFound {
			return nil, status.ErrorMessage(errcodes.SigninSessionNotFoundCode, "培训班次不存在")
		}
		return nil, err
	}

	rows, err := l.svcCtx.Model.HrmTrainingSessionParticipant.FindByCondition(l.ctx, nil,
		condition.NewChain().Equal(hrm_training_session_participant.SessionUuid, req.SessionUuid).Build()...)
	if err != nil {
		return nil, err
	}

	expected := int64(len(rows))
	var signed int64
	for _, p := range rows {
		if p.SignInTime.Valid {
			signed++
		}
	}

	return &types.GetSigninStatsResponse{
		Expected: expected,
		Signed:   signed,
		Absent:   expected - signed,
	}, nil
}
