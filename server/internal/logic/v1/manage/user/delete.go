package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/manage/user"
)

type Delete struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewDelete(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Delete {
	return &Delete{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

func (l *Delete) Delete(req *types.DeleteRequest) (resp *types.DeleteResponse, err error) {
	if len(req.Uuids) == 0 {
		return nil, nil
	}

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// HRM-52 删除一致性 (ADR 0007): 账号是员工档案的投影, 档案单向持有账号指针,
		// 被引用的账号不可删除 (提示先在员工档案解除关联)。校验与删除在同一事务内,
		// 任一 uuid 被引用即整批拒绝 (全或无, 不部分成功)。一致性问题由应用层保证,
		// 不加外键/唯一索引。
		uuids := make([]string, 0, len(req.Uuids))
		for _, u := range req.Uuids {
			if u != "" {
				uuids = append(uuids, u)
			}
		}
		if len(uuids) > 0 {
			referenced, err := l.svcCtx.Model.HrmEmployee.FindByCondition(ctx, session, condition.Condition{
				Field:    hrm_employee.UserUuid,
				Operator: condition.In,
				Value:    uuids,
			})
			if err != nil {
				return err
			}
			if len(referenced) > 0 {
				first := referenced[0]
				msg := fmt.Sprintf("账号已被员工档案引用（工号 %s %s），请先在员工档案中解除关联后再删除",
					first.EmployeeNo, first.Name)
				if len(referenced) > 1 {
					msg = fmt.Sprintf("%d 个账号已被员工档案引用（如工号 %s %s），请先在员工档案中解除关联后再删除",
						len(referenced), first.EmployeeNo, first.Name)
				}
				return status.ErrorMessage(errcodes.EmployeeUserReferencedCode, msg)
			}
		}

		if err = l.svcCtx.Model.ManageUser.DeleteByCondition(l.ctx, session, condition.Condition{
			Field:    manage_user.Uuid,
			Operator: condition.In,
			Value:    req.Uuids,
		}); err != nil {
			return err
		}
		if err = l.svcCtx.Model.ManageUserRole.DeleteByCondition(l.ctx, session, condition.NewChain().In("user_uuid", req.Uuids).Build()...); err != nil {
			return err
		}
		return nil
	})

	return
}
