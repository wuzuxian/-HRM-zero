package reward

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/reward"
)

type Add struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAdd(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Add {
	return &Add{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// Add 新增奖惩单（HR215）：事务内生成单据编号，初始状态 未提交.
// 单据编号/状态列与共享审批底座（HRM-06）对齐，提交/审批走共享端点；
// 审批通过后本单成为薪资项目 奖惩奖励/奖惩扣款 的数据源（spec 用户故事 53）.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateReward(req.EmployeeNo, req.EmployeeName, req.RewardType, req.RewardLevel,
		req.OccurDate, req.SalaryMonth, req.Amount); err != nil {
		return nil, err
	}

	user := hrmuser.Current(l.ctx)

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, time.Now())
		if err != nil {
			return err
		}
		billNo = no
		err = l.svcCtx.Model.HrmRewardBill.InsertV2(ctx, session, &hrm_reward_bill.HrmRewardBill{
			Uuid:         uuid.New().String(),
			BillNo:       no,
			Status:       bill.StatusUnsubmitted,
			EmployeeUuid: req.EmployeeUuid,
			EmployeeNo:   req.EmployeeNo,
			EmployeeName: req.EmployeeName,
			RewardType:   req.RewardType,
			RewardLevel:  req.RewardLevel,
			DocNo:        req.DocNo,
			OccurDate:    req.OccurDate,
			Amount:       req.Amount,
			Reason:       req.Reason,
			SalaryMonth:  normalizeSalaryMonth(req.SalaryMonth, req.OccurDate),
			Synced:       0,
			CreateBy:     user.Uuid,
			CreatorName:  user.Username,
		})
		return err
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("reward: 奖惩单已创建 billNo=%s employee=%s(%s) type=%s amount=%.2f",
		billNo, req.EmployeeName, req.EmployeeNo, req.RewardType, req.Amount)
	return &types.AddResponse{}, nil
}
