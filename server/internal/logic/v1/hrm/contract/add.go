package contract

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee_contract"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/contract"
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

// Add 新增劳动合同（spec 用户故事 17，员工档案上下文维护入口）：
// 员工档案 uuid 必填（带出工号/姓名/部门快照），生效日/到期日必填且到期日不早于生效日。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if req.EmployeeUuid == "" {
		return nil, status.ErrorMessage(errcodes.ContractInvalidCode, "员工档案 uuid 不能为空")
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

	emp, err := resolveEmployee(l.ctx, l.svcCtx, req.EmployeeUuid)
	if err != nil {
		return nil, err
	}

	statusValue := req.Status
	if statusValue == "" {
		statusValue = ContractStatusActive
	}

	_, err = l.svcCtx.Model.HrmEmployeeContract.Insert(l.ctx, nil, &hrm_employee_contract.HrmEmployeeContract{
		Uuid:         uuidNew(),
		EmployeeUuid: req.EmployeeUuid,
		EmployeeNo:   emp.EmployeeNo,
		EmployeeName: emp.Name,
		OrgUuid:      emp.OrgUuid,
		ContractNo:   req.ContractNo,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		RenewCount:   0,
		Status:       statusValue,
		Remark:       req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.AddResponse{}, nil
}
