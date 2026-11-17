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

type CreateAndLinkUser struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewCreateAndLinkUser(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *CreateAndLinkUser {
	return &CreateAndLinkUser{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// CreateAndLinkUser 新建账号并关联 (HRM-51): 员工侧一次建号+关联, 建号动作全部走
// 唯一接缝 provisionAccountWith (ADR 0007 — 不自带一份实现), 仅覆写用户名/密码:
// 用户名空取工号, 密码空取配置 (兜底 123456); 工号/用户名撞已有账号报 10461。
// 已关联其他账号的档案拒绝二次建号 (10722); 离职补录不开号 (10723, 明确报错
// 而非静默空响应 — 与批量「生成用户」的跳过计数语义不同, 这是单次用户动作)。
func (l *CreateAndLinkUser) CreateAndLinkUser(req *types.CreateAndLinkUserRequest) (resp *types.CreateAndLinkUserResponse, err error) {
	var userUuid string
	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		emp, err := findEmployeeOrNotFound(ctx, session, l.svcCtx, req.EmployeeUuid)
		if err != nil {
			return err
		}
		if emp.UserUuid != "" {
			return status.ErrorMessage(errcodes.EmployeeUserAlreadyLinkedCode,
				"员工 "+emp.EmployeeNo+"("+emp.Name+") 已关联系统账号, 请先解除关联")
		}

		userUuid, err = provisionAccountWith(ctx, l.svcCtx, session, req.EmployeeUuid,
			accountOverrides{username: req.Username, password: req.Password})
		if err != nil {
			return err
		}
		if userUuid == "" {
			// 接缝对离职 (person_status='5') 返回空串: 显式动作必须有明确结果
			return status.ErrorMessage(errcodes.EmployeeUserResignedCode,
				"离职补录员工不开通系统账号: "+emp.EmployeeNo+"("+emp.Name+")")
		}

		created, err := l.svcCtx.Model.ManageUser.FindOneByUuid(ctx, session, userUuid)
		if err != nil {
			return err
		}
		resp = &types.CreateAndLinkUserResponse{
			UserUuid:        userUuid,
			Username:        created.Username,
			InitialPassword: created.Password,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
