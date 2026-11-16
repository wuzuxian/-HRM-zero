package signin

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/signin"
)

type GetSigninDetailList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetSigninDetailList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetSigninDetailList {
	return &GetSigninDetailList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// signinDetailEnrich carries the roster row's user/employee snapshot for display.
type signinDetailEnrich struct {
	username   string
	nickName   string
	employeeNo string
	orgName    string
}

// loadDetailEnrich batch-enriches 名册 rows: 学员账号/姓名 (manage_user), 工号/部门
// (hrm_employee via user_uuid, one IN query each — no N+1).
func (l *GetSigninDetailList) loadDetailEnrich(rows []*hrm_training_session_participant.HrmTrainingSessionParticipant) map[string]signinDetailEnrich {
	out := make(map[string]signinDetailEnrich, len(rows))

	userUuids := make([]string, 0, len(rows))
	for _, p := range rows {
		userUuids = append(userUuids, p.UserUuid)
	}

	users := make(map[string]*manage_user.ManageUser, len(userUuids))
	if len(userUuids) > 0 {
		urows, err := l.svcCtx.Model.ManageUser.FindByCondition(l.ctx, nil,
			condition.NewChain().In(manage_user.Uuid, userUuids).Build()...)
		if err == nil {
			for _, u := range urows {
				users[u.Uuid] = u
			}
		}
	}

	emps := make(map[string]*hrm_employee.HrmEmployee, len(userUuids))
	if len(userUuids) > 0 {
		erows, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_employee.UserUuid, userUuids).Build()...)
		if err == nil {
			for _, e := range erows {
				emps[e.UserUuid] = e
			}
		}
	}

	orgUuids := make([]string, 0, len(emps))
	for _, e := range emps {
		if e.OrgUuid != "" {
			orgUuids = append(orgUuids, e.OrgUuid)
		}
	}
	orgs := make(map[string]*hrm_org.HrmOrg, len(orgUuids))
	if len(orgUuids) > 0 {
		orows, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil,
			condition.NewChain().In(hrm_org.Uuid, orgUuids).Build()...)
		if err == nil {
			for _, o := range orows {
				orgs[o.Uuid] = o
			}
		}
	}

	for _, p := range rows {
		e := signinDetailEnrich{}
		if u := users[p.UserUuid]; u != nil {
			e.username = u.Username
			e.nickName = u.Nickname
		}
		if emp := emps[p.UserUuid]; emp != nil {
			e.employeeNo = emp.EmployeeNo
			if o := orgs[emp.OrgUuid]; o != nil {
				e.orgName = o.OrgName
			}
		}
		out[p.Uuid] = e
	}
	return out
}

// GetSigninDetailList 签到明细表（demo-inventory 6：姓名/工号/部门/签到时间/方式/操作人/原因）：
// 名册=应到名单，一行一位学员；未签行 签到时间/方式 为空。
func (l *GetSigninDetailList) GetSigninDetailList(req *types.GetSigninDetailListRequest) (resp *types.GetSigninDetailListResponse, err error) {
	rows, err := l.svcCtx.Model.HrmTrainingSessionParticipant.FindByCondition(l.ctx, nil,
		condition.NewChain().
			Equal(hrm_training_session_participant.SessionUuid, req.SessionUuid).
			OrderByAsc(hrm_training_session_participant.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	enrich := l.loadDetailEnrich(rows)

	// 操作人姓名（自签=本人= nickName；补签=管理员 user_uuid → nickname）
	operatorUuids := make([]string, 0, len(rows))
	for _, p := range rows {
		if p.SignInBy != "" {
			operatorUuids = append(operatorUuids, p.SignInBy)
		}
	}
	operators := make(map[string]string, len(operatorUuids))
	if len(operatorUuids) > 0 {
		orows, oerr := l.svcCtx.Model.ManageUser.FindByCondition(l.ctx, nil,
			condition.NewChain().In(manage_user.Uuid, operatorUuids).Build()...)
		if oerr == nil {
			for _, u := range orows {
				operators[u.Uuid] = u.Nickname
			}
		}
	}

	records := make([]types.SigninDetail, 0, len(rows))
	for _, p := range rows {
		e := enrich[p.Uuid]
		rec := types.SigninDetail{
			Uuid:         p.Uuid,
			SessionUuid:  p.SessionUuid,
			UserUuid:     p.UserUuid,
			Username:     e.username,
			NickName:     e.nickName,
			EmployeeNo:   e.employeeNo,
			OrgName:      e.orgName,
			SignInTime:   formatNullTime(p.SignInTime),
			SignInMethod: p.SignInMethod,
			SignInBy:     p.SignInBy,
			SignInByName: operators[p.SignInBy],
			SignInReason: p.SignInReason,
		}
		records = append(records, rec)
	}

	return &types.GetSigninDetailListResponse{Records: records}, nil
}
