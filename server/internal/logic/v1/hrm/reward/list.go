package reward

import (
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_reward_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/reward"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// List 奖惩单分页列表，支持按单据编号/单据状态/奖惩类型/员工工号筛选.
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 管理员级角色（持共享审批码 v1:hrm:bill:approve）可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !hrmuser.CanSeeAllBills(l.ctx, l.svcCtx, user)

	chain := condition.NewChain().
		Equal(hrm_reward_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_reward_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_reward_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Equal(hrm_reward_bill.RewardType, req.RewardType, condition.WithSkip(req.RewardType == "")).
		Like(hrm_reward_bill.EmployeeNo, "%"+req.EmployeeNo+"%", condition.WithSkip(req.EmployeeNo == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_reward_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmRewardBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmReward, 0, len(rows))
	for _, row := range rows {
		records = append(records, rewardRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func rewardRecord(row *hrm_reward_bill.HrmRewardBill) types.HrmReward {
	return types.HrmReward{
		Uuid:         row.Uuid,
		BillNo:       row.BillNo,
		Status:       row.Status,
		EmployeeUuid: row.EmployeeUuid,
		EmployeeNo:   row.EmployeeNo,
		EmployeeName: row.EmployeeName,
		RewardType:   row.RewardType,
		RewardLevel:  row.RewardLevel,
		DocNo:        row.DocNo,
		OccurDate:    row.OccurDate,
		Amount:       row.Amount,
		Reason:       row.Reason,
		SalaryMonth:  row.SalaryMonth,
		Synced:       row.Synced,
		CreatorUuid:  row.CreateBy,
		CreatorName:  row.CreatorName,
		CreateTime:   row.CreateTime.Format(time.DateTime),
		UpdateTime:   row.UpdateTime.Format(time.DateTime),
	}
}
