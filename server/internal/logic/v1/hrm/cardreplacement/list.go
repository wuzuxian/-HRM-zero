package cardreplacement

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_card_replacement_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/cardreplacement"
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

// List 补卡单分页列表，支持按 单据编号/单据状态/员工工号/补卡类型 筛选.
// 默认按「我的单据」过滤（创建人=当前用户，spec 用户故事 4）；
// 超级管理员可看全部，传 mine=1 时管理员也只看自己的单据.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	user := hrmuser.Current(l.ctx)
	onlyMine := req.Mine == 1 || !user.IsSuperAdmin()

	chain := condition.NewChain().
		Equal(hrm_card_replacement_bill.CreateBy, user.Uuid, condition.WithSkip(!onlyMine)).
		Like(hrm_card_replacement_bill.BillNo, "%"+req.BillNo+"%", condition.WithSkip(req.BillNo == "")).
		Equal(hrm_card_replacement_bill.Status, req.Status, condition.WithSkip(req.Status == "")).
		Equal(hrm_card_replacement_bill.EmployeeNo, req.EmployeeNo, condition.WithSkip(req.EmployeeNo == "")).
		Equal(hrm_card_replacement_bill.CardType, req.CardType, condition.WithSkip(req.CardType == "")).
		Page(req.Current, req.Size).
		OrderByDesc(hrm_card_replacement_bill.Id)

	rows, total, err := l.svcCtx.Model.HrmCardReplacementBill.PageByCondition(l.ctx, nil, chain.Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.HrmCardReplacement, 0, len(rows))
	for _, row := range rows {
		records = append(records, cardReplacementRecord(row))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   total,
	}, Records: records}, nil
}

func cardReplacementRecord(row *hrm_card_replacement_bill.HrmCardReplacementBill) types.HrmCardReplacement {
	return types.HrmCardReplacement{
		Uuid:             row.Uuid,
		BillNo:           row.BillNo,
		Status:           row.Status,
		EmployeeUuid:     row.EmployeeUuid,
		EmployeeNo:       row.EmployeeNo,
		EmployeeName:     row.EmployeeName,
		OrgUuid:          row.OrgUuid,
		OrgName:          row.OrgName,
		PunchDate:        row.PunchDate,
		CardType:         row.CardType,
		PunchTime:        row.PunchTime,
		Reason:           row.Reason,
		OrigClockInTime:  row.OrigClockInTime,
		OrigClockOutTime: row.OrigClockOutTime,
		CreatorUuid:      row.CreateBy,
		CreatorName:      row.CreatorName,
		CreateTime:       row.CreateTime.Format("2006-01-02 15:04:05"),
		UpdateTime:       row.UpdateTime.Format("2006-01-02 15:04:05"),
	}
}
