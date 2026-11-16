package salary_payment

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	hrmuser "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_payment"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_payment"
)

type Register struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewRegister(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Register {
	return &Register{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// 发薪方式（demo 8.4：银行代发/现金/其他）。
const (
	PayModeBank  = "bank"
	PayModeCash  = "cash"
	PayModeOther = "other"
)

var payModes = map[string]struct{}{
	PayModeBank:  {},
	PayModeCash:  {},
	PayModeOther: {},
}

// Register 发薪登记（HR210，demo 8.4 /hrm/salary/payment-list）：
//   - 守卫：批次须 已确认未发薪（confirm_status='1' 且 pay_status='0'）。
//     未确认 → 10714（审批门禁：未确认不可发薪）；已发薪 → 重复登记拒绝。
//   - 发薪金额留空（<=0）默认取批次实发合计（可改）。
//   - 事务内：登记行插入 + 批次 pay_status='1'（一批次一条登记 UNIQUE(batch_uuid)）。
//     HRM-39 工资条发布与薪资统计台账透传 pay_status / pay_amount（消费契约）。
func (l *Register) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	batch, err := salary_batch.FindBatchByUuid(l.ctx, l.svcCtx, req.BatchUuid)
	if err != nil {
		return nil, err
	}
	// 审批门禁：confirm_status 须 '1'（HRM-37 未确认批次落 ''/'0'，同义未推进）；
	// pay_status=''/'0' 同义未发薪，'1' 已发薪 → 重复登记拒绝。
	if batch.ConfirmStatus != "1" {
		return nil, statusBatchNotConfirmed("批次尚未通过薪资确认审批，不可发薪登记")
	}
	if batch.PayStatus != "" && batch.PayStatus != "0" {
		return nil, statusBatchNotConfirmed("批次已发薪，不可重复发薪登记")
	}
	payMode := req.PayMode
	if payMode == "" {
		payMode = PayModeBank
	}
	if _, ok := payModes[payMode]; !ok {
		return nil, statusInvalid(fmt.Sprintf("发薪方式非法（bank/cash/other），当前 %s", payMode))
	}
	payAmount := req.PayAmount
	if payAmount <= 0 {
		payAmount = batch.NetTotal
	}

	op := hrmuser.Current(l.ctx)

	var paymentUuid string
	txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		paymentUuid = uuid.New().String()
		if err := model.NewHrmSalaryPaymentModel(sqlx.NewSqlConnFromSession(session)).InsertV2(ctx, session,
			&model.HrmSalaryPayment{
				Uuid:         paymentUuid,
				BatchUuid:    batch.Uuid,
				BatchNo:      batch.BatchNo,
				BatchName:    batch.BatchName,
				SubjectUuid:  batch.SubjectUuid,
				SubjectName:  batch.SubjectName,
				PayAmount:    payAmount,
				PayMode:      payMode,
				ReceiptNo:    req.ReceiptNo,
				ArriveTime:   req.ArriveTime,
				RegisterBy:   op.Uuid,
				RegisterName: op.Username,
				RegisterTime: time.Now().Format(time.DateTime),
			}); err != nil {
			return err
		}
		return hrm_salary_batch.NewHrmSalaryBatchModel(sqlx.NewSqlConnFromSession(session)).
			UpdateFieldsByCondition(ctx, session, map[string]any{
				"pay_status": "1",
			}, condition.NewChain().Equal(hrm_salary_batch.Uuid, batch.Uuid).Build()...)
	})
	if txErr != nil {
		return nil, txErr
	}
	l.Logger.Infof("salary_payment: 发薪登记 batch=%s amount=%.2f mode=%s", batch.Uuid, payAmount, payMode)
	return &types.RegisterResponse{Uuid: paymentUuid}, nil
}
