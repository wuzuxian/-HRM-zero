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
			// 回读失败降级 (code review): 建档+开号事务已提交, 把已成功的建档报成错误
			// 只会误导调用方重试出重复档案。凭据缺失只影响前端明示, 记日志后按已开号
			// 响应: username 给接缝默认规则下的确定值 (本路径无覆写, 用户名=工号),
			// 初始密码无法安全回取则留空。
			l.Errorf("hrm: 员工 %s 建档开号成功但账号回读失败, 响应凭据降级 (不含初始密码): %v", employeeNo, userErr)
			return &types.AddResponse{
				UserCreated: true,
				Username:    employeeNo,
			}, nil
		}
		// 初始密码取接缝已解析的值 (manage_user.password 明文存的是初始密码, 与
		// create_and_link_user 同一来源), 不在响应侧重算配置兜底。
		return &types.AddResponse{
			UserCreated:     true,
			Username:        user.Username,
			InitialPassword: user.Password,
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
