package employee

import (
	"context"
	"net/http"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
)

type LinkableUserList struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewLinkableUserList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *LinkableUserList {
	return &LinkableUserList{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// LinkableUserList 候选账号分页 (HRM-51): 只返回未被任何档案关联的账号
// (uuid+用户名+昵称), 关键字模糊匹配用户名/昵称 (OR 组, huandu sqlbuilder 自带括号)。
// 已被占用的账号不进候选 (验收: 候选列表只出现未关联的账号)。
func (l *LinkableUserList) LinkableUserList(req *types.LinkableUserListRequest) (resp *types.LinkableUserListResponse, err error) {
	// 占用集合 = 档案侧所有非空指针 (含 NULL 语义: user_uuid 为空的行天然不在内)
	linked, err := l.svcCtx.Model.HrmEmployee.FindByCondition(l.ctx, nil,
		condition.NewChain().NotEqual(hrm_employee.UserUuid, "").Build()...)
	if err != nil {
		return nil, err
	}
	occupied := make([]any, 0, len(linked))
	for _, row := range linked {
		if row.UserUuid != "" {
			occupied = append(occupied, row.UserUuid)
		}
	}

	chain := condition.NewChain().OrderByAsc(manage_user.Id)
	if len(occupied) > 0 {
		chain = chain.NotIn(manage_user.Uuid, occupied)
	}
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		chain = chain.Or([]condition.Field{manage_user.Username, manage_user.Nickname},
			[]condition.Operator{condition.Like, condition.Like}, []any{kw, kw})
	}
	users, total, err := l.svcCtx.Model.ManageUser.PageByCondition(l.ctx, nil,
		chain.Page(req.Current, req.Size).Build()...)
	if err != nil {
		return nil, err
	}

	records := make([]types.LinkableUser, 0, len(users))
	for _, user := range users {
		records = append(records, types.LinkableUser{
			Uuid:     user.Uuid,
			Username: user.Username,
			Nickname: user.Nickname,
		})
	}
	return &types.LinkableUserListResponse{
		Records: records,
		PageResponse: types.PageResponse{
			Current: req.Current,
			Size:    req.Size,
			Total:   total,
		},
	}, nil
}
