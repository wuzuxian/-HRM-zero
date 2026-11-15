package position

import (
	"context"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_position"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
)

// IncreaseHiredCount 已录用人数事务化递增（HRM-27 录用审批 seam 契约）。
//
// 契约（供 HRM-27 录用审批单审批通过后调用）：
//   - IncreaseHiredCount(ctx, svcCtx, positionUuid) — svcCtx 须为生效的
//     *svc.ServiceContext（SqlxConn + Model）。本函数内部自起事务
//     （SqlxConn.TransactCtx），行级锁定后 校验→递增→状态联动，调用方
//     若已在更大事务内，可传 session 版本 IncreaseHiredCountTx。
//   - 递增前校验：职位必须存在（10580）、状态不可为 closed（10584）、
//     hired_count+1 不得超过 headcount（10585 超招拒绝）。
//   - 达到 headcount 时自动置 status='full'（已满）。
//   - 幂等性：无幂等键——同一录用单请勿重复调用；去重由 HRM-27 侧
//     （offer 单审批通过即终态）保证。
func IncreaseHiredCount(ctx context.Context, svcCtx *svc.ServiceContext, positionUuid string) error {
	if positionUuid == "" {
		return status.ErrorMessage(errcodes.PositionInvalidCode, "招聘职位 uuid 必填")
	}
	return svcCtx.SqlxConn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return IncreaseHiredCountTx(ctx, svcCtx, session, positionUuid)
	})
}

// IncreaseHiredCountTx is the session-aware variant of IncreaseHiredCount for
// callers that already run inside a transaction (e.g. HRM-27 approving the
// offer bill in one big tx). Same contract, no nested transaction.
func IncreaseHiredCountTx(ctx context.Context, svcCtx *svc.ServiceContext, session sqlx.Session, positionUuid string) error {
	row, err := svcCtx.Model.HrmRecruitmentPosition.FindOneByCondition(ctx, session,
		condition.NewChain().Equal(hrm_recruitment_position.Uuid, positionUuid).Build()...)
	if err != nil {
		return status.ErrorMessage(errcodes.PositionNotFoundCode, "招聘职位不存在")
	}
	if row.Status == PositionStatusClosed {
		return status.ErrorMessage(errcodes.PositionStatusTransitionInvalidCode,
			"职位已关闭，不可递增已录用人数")
	}
	if row.HiredCount+1 > row.Headcount {
		return status.ErrorMessage(errcodes.PositionHiredCountExceededCode,
			"已录用人数将超出招聘人数，不可继续录用")
	}
	row.HiredCount++
	if row.HiredCount >= row.Headcount {
		row.Status = PositionStatusFull
	}
	return svcCtx.Model.HrmRecruitmentPosition.Update(ctx, session, row)
}
