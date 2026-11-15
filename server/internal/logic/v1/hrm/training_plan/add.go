package training_plan

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_plan"
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

// Add 新增培训计划单（HR216）：事务内生成单据编号，初始状态 未提交（发布状态 草稿）。
// 单据编号/状态列与共享审批底座（HRM-06）对齐，提交/审批走共享端点；
// 审批通过由 OnApproved 置 published=1（发布）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validatePlan(req.PlanName, req.PlanYear, req.PlanType, req.StartDate, req.EndDate); err != nil {
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
		err = l.svcCtx.Model.HrmTrainingPlan.InsertV2(ctx, session, &hrm_training_plan.HrmTrainingPlan{
			Uuid:        uuid.New().String(),
			BillNo:      no,
			Status:      bill.StatusUnsubmitted,
			PlanName:    req.PlanName,
			Theme:       req.Theme,
			PlanYear:    req.PlanYear,
			PlanType:    req.PlanType,
			StartDate:   req.StartDate,
			EndDate:     req.EndDate,
			Published:   0,
			CreateBy:    user.Uuid,
			CreatorName: user.Username,
		})
		return err
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("training_plan: 培训计划单已创建 billNo=%s plan=%s", billNo, req.PlanName)
	return &types.AddResponse{}, nil
}
