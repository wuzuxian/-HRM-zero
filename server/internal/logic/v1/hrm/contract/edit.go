package contract

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/contract"
)

type Edit struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewEdit(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Edit {
	return &Edit{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Uuid == "" {
		return nil, status.ErrorMessage(errcodes.ContractNotFoundCode, "uuid 不能为空")
	}
	if !validDate(req.StartDate) || !validDate(req.EndDate) {
		return nil, status.ErrorMessage(errcodes.ContractInvalidCode, "生效日/到期日格式须为 YYYY-MM-DD")
	}
	if req.EndDate < req.StartDate {
		return nil, status.ErrorMessage(errcodes.ContractInvalidCode, "到期日不能早于生效日")
	}
	if req.Status != "" && !validContractStatus(req.Status) {
		return nil, status.ErrorMessage(errcodes.ContractInvalidCode, "合同状态非法")
	}

	row, err := findContractByUuid(l.ctx, l.svcCtx, req.Uuid)
	if err != nil {
		return nil, err
	}

	statusValue := req.Status
	if statusValue == "" {
		statusValue = row.Status
	}

	row.ContractNo = req.ContractNo
	row.StartDate = req.StartDate
	row.EndDate = req.EndDate
	row.Status = statusValue
	row.Remark = req.Remark

	if err = l.svcCtx.Model.HrmEmployeeContract.Update(l.ctx, nil, row); err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
