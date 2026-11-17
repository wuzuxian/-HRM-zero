package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
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

// Add 新增员工 (spec user story 9/10): 工号为空时按 最大工号+1 (8 位补零) 自动生成,
// 三个从表随主表在同一事务内写入。
// HRM-50 建档即开号 (ADR 0007): 档案落库后在同一事务内经唯一接缝 ProvisionAccount
// 开号并挂默认角色, 开号失败整体回滚 (档案不落库); 离职补录 (person_status='5')
// 接缝内跳过开号。响应回带登录用户名与初始密码供前端明示。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateEmployeeBasic(&req.Name, &req.Gender, &req.MaritalStatus, &req.PersonStatus); err != nil {
		return nil, err
	}
	if req.PersonStatus == "" {
		req.PersonStatus = PersonStatusRegular
	}

	var employeeNo, empUuid, userUuid string
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		employeeNo, err = generateEmployeeNo(ctx, session, l.svcCtx, req.EmployeeNo)
		if err != nil {
			return err
		}
		req.EmployeeNo = employeeNo
		empUuid, err = insertEmployee(ctx, session, l.svcCtx, req)
		if err != nil {
			return err
		}
		// 建档即开号: 同一事务内接唯一接缝; 离职跳过返回空串
		userUuid, err = provisionAccount(ctx, l.svcCtx, session, empUuid)
		return err
	})
	if err != nil {
		return nil, err
	}

	if userUuid != "" {
		user, userErr := l.svcCtx.Model.ManageUser.FindOneByUuid(l.ctx, nil, userUuid)
		if userErr != nil {
			return nil, userErr
		}
		initialPassword := l.svcCtx.MustGetConfig().Hrm.InitialUserPassword
		if initialPassword == "" {
			initialPassword = defaultInitialPassword
		}
		return &types.AddResponse{
			UserCreated:     true,
			Username:        user.Username,
			InitialPassword: initialPassword,
		}, nil
	}
	return &types.AddResponse{UserCreated: false}, nil
}

// validateEmployeeBasic 校验公共枚举字段; name/pass 由调用方按 add/edit 语义处理。
func validateEmployeeBasic(name, gender, maritalStatus, personStatus *string) error {
	if name != nil && *name == "" {
		return status.ErrorMessage(errcodes.EmployeeInvalidCode, "姓名不能为空")
	}
	if gender != nil && *gender != "" && !validGender(*gender) {
		return status.ErrorMessage(errcodes.EmployeeInvalidCode, "性别非法")
	}
	if maritalStatus != nil && *maritalStatus != "" && !validMaritalStatus(*maritalStatus) {
		return status.ErrorMessage(errcodes.EmployeeInvalidCode, "婚姻状况非法")
	}
	if personStatus != nil && *personStatus != "" && !validPersonStatus(*personStatus) {
		return status.ErrorMessage(errcodes.EmployeeInvalidCode, "人员状态非法")
	}
	return nil
}
