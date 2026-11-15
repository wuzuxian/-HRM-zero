package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/status"
	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
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

// Edit 编辑员工: 主表字段更新 + 三个从表全删全插, 全部在同一事务内 (spec user story 9).
func (l *Edit) Edit(req *types.EditRequest) (resp *types.EditResponse, err error) {
	if req.Name == "" {
		return nil, status.ErrorMessage(errcodes.EmployeeInvalidCode, "姓名不能为空")
	}
	if err = validateEmployeeBasic(nil, &req.Gender, &req.MaritalStatus, &req.PersonStatus); err != nil {
		return nil, err
	}

	err = l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		row, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(ctx, session, req.Uuid)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return status.ErrorMessage(errcodes.EmployeeNotFoundCode, "员工不存在")
			}
			return err
		}

		// 工号变更时查重; 为空表示保持原工号
		if req.EmployeeNo != "" && req.EmployeeNo != row.EmployeeNo {
			_, dupErr := l.svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, session,
				condition.NewChain().Equal(hrm_employee.EmployeeNo, req.EmployeeNo).Build()...)
			if dupErr == nil {
				return status.ErrorMessage(errcodes.EmployeeNoDuplicateCode, "工号已存在: "+req.EmployeeNo)
			}
			if dupErr != sqlx.ErrNotFound {
				return dupErr
			}
			row.EmployeeNo = req.EmployeeNo
		}
		if req.EmployeeNo == "" {
			req.EmployeeNo = row.EmployeeNo
		}

		row.Name = req.Name
		row.Gender = req.Gender
		row.Birthday = req.Birthday
		row.Nation = req.Nation
		row.PoliticalStatus = req.PoliticalStatus
		row.MaritalStatus = req.MaritalStatus
		row.NativePlace = req.NativePlace
		row.BloodType = req.BloodType
		row.IdCard = req.IdCard
		row.Height = req.Height
		row.Weight = req.Weight
		row.HouseholdAddress = req.HouseholdAddress
		row.Address = req.Address
		row.EmergencyContact = req.EmergencyContact
		row.EmergencyPhone = req.EmergencyPhone
		row.Phone = req.Phone
		row.Email = req.Email
		row.Avatar = req.Avatar
		row.BankName = req.BankName
		row.BankAccount = req.BankAccount
		row.CompanyUuid = req.CompanyUuid
		row.OrgUuid = req.OrgUuid
		row.Position = req.Position
		row.Post = req.Post
		if req.PersonStatus != "" {
			row.PersonStatus = req.PersonStatus
		}
		row.HireDate = req.HireDate
		row.FormalDate = req.FormalDate
		row.Remark = req.Remark

		if err = l.svcCtx.Model.HrmEmployee.Update(ctx, session, row); err != nil {
			return err
		}

		// 员工↔账号双向同步 (HRM-04): 已生成账号的员工, 编辑档案同步账号资料
		// (昵称=姓名, 手机号, 邮箱); 未生成账号 (user_uuid '') 则无账号可同步.
		if row.UserUuid != "" {
			user, userErr := l.svcCtx.Model.ManageUser.FindOneByUuid(ctx, session, row.UserUuid)
			if userErr != nil && userErr != sqlx.ErrNotFound {
				return userErr
			}
			if userErr == nil {
				user.Nickname = row.Name
				user.Phone = row.Phone
				user.Email = row.Email
				if err = l.svcCtx.Model.ManageUser.Update(ctx, session, user); err != nil {
					return err
				}
			}
		}

		return saveChildren(ctx, session, l.svcCtx, row.Uuid, req.Educations, req.Works, req.Families)
	})
	if err != nil {
		return nil, err
	}

	return &types.EditResponse{}, nil
}
