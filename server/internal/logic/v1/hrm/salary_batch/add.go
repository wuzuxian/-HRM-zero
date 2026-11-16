package salary_batch

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_salary_batch"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/salary_batch"
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

// Add 新增薪资批次（主体+年月+批次类型唯一范围；批次编号在插入事务内分配 SA 序列）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.BatchName == "" {
		return nil, statusInvalid("批次名称不能为空")
	}
	if req.BatchType == "" {
		req.BatchType = BatchTypeMonthly
	}
	if req.BatchType != BatchTypeMonthly {
		return nil, statusInvalid("批次类型仅支持 monthly（月度批次）")
	}
	subject, err := findSubjectByUuid(l.ctx, l.svcCtx, req.SubjectUuid)
	if err != nil {
		return nil, err
	}
	batchMonthFirst, _, err := validateBatchMonth(req.BatchYear, req.BatchMonth)
	if err != nil {
		return nil, err
	}

	// 唯一范围守卫（主体+年月+批次类型）
	existing, err := l.svcCtx.Model.HrmSalaryBatch.FindOneByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_salary_batch.SubjectUuid, req.SubjectUuid).
			Equal(hrm_salary_batch.BatchYear, req.BatchYear).
			Equal(hrm_salary_batch.BatchMonth, req.BatchMonth).
			Equal(hrm_salary_batch.BatchType, req.BatchType).
			Build()...)
	if err != nil && err != hrm_salary_batch.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, statusDuplicate("该主体下该年月已存在同类型批次")
	}

	var batchUuid string
	txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		batchNo, err := GenerateBatchNo(ctx, session, batchMonthFirst)
		if err != nil {
			return err
		}
		batchUuid = uuid.New().String()
		return hrm_salary_batch.NewHrmSalaryBatchModel(sqlx.NewSqlConnFromSession(session)).InsertV2(ctx, session,
			&hrm_salary_batch.HrmSalaryBatch{
				Uuid:        batchUuid,
				BatchNo:     batchNo,
				BatchName:   req.BatchName,
				BatchType:   req.BatchType,
				SubjectUuid: subject.Uuid,
				SubjectName: subject.SubjectName,
				BatchYear:   req.BatchYear,
				BatchMonth:  req.BatchMonth,
				TrialStatus: "0",
				Version:     0,
				Remark:      req.Remark,
			})
	})
	if txErr != nil {
		return nil, txErr
	}
	l.Logger.Infof("salary batch added %s %s %d-%02d", batchUuid, subject.SubjectName, req.BatchYear, req.BatchMonth)
	return &types.AddResponse{Uuid: batchUuid}, nil
}
