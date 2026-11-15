package requirement

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_recruitment_requirement"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/requirement"
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

// Add 新增招聘需求单（HR207）：事务内生成单据编号，初始状态 未提交.
// 单据编号/状态列与共享审批底座（HRM-06）对齐，提交/审批走共享端点；
// 审批通过后需求单可被招聘职位（HRM-25）引用发布（spec 用户故事 43）.
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateRequirement(req.PostName, req.PostCategory, req.Urgency,
		req.ExpectDate, req.Headcount, req.OrgUuid); err != nil {
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
		err = l.svcCtx.Model.HrmRecruitmentRequirement.InsertV2(ctx, session, &hrm_recruitment_requirement.HrmRecruitmentRequirement{
			Uuid:            uuid.New().String(),
			BillNo:          no,
			Status:          bill.StatusUnsubmitted,
			PostName:        req.PostName,
			PostCategory:    req.PostCategory,
			Headcount:       req.Headcount,
			Urgency:         req.Urgency,
			ExpectDate:      req.ExpectDate,
			OrgUuid:         req.OrgUuid,
			OrgName:         req.OrgName,
			RequirementDesc: req.RequirementDesc,
			CreateBy:        user.Uuid,
			CreatorName:     user.Username,
		})
		return err
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("requirement: 招聘需求单已创建 billNo=%s post=%s(%s) headcount=%d urgency=%s",
		billNo, req.PostName, req.PostCategory, req.Headcount, req.Urgency)
	return &types.AddResponse{}, nil
}
