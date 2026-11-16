package perf_assessment

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/perf_assessment"
)

type GetMyWorkbench struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewGetMyWorkbench(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *GetMyWorkbench {
	return &GetMyWorkbench{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

// GetMyWorkbench 考核任务工作台（当前登录人视角只读聚合：
// 全周期待办汇总卡 + 最近待办列表 + 申诉中单列；无状态变更）。
func (l *GetMyWorkbench) GetMyWorkbench(req *types.GetMyAssessmentWorkbenchRequest) (resp *types.GetMyAssessmentWorkbenchResponse, err error) {
	actor, err := resolveActor(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	cards, todos, appeals, err := MyAssessmentWorkbench(l.ctx, l.svcCtx, actor)
	if err != nil {
		return nil, err
	}

	resp = &types.GetMyAssessmentWorkbenchResponse{
		Cards:   cards,
		Todos:   make([]types.AssessmentWorkbenchTodo, 0, len(todos)),
		Appeals: make([]types.AssessmentWorkbenchTodo, 0, len(appeals)),
	}
	for _, t := range todos {
		resp.Todos = append(resp.Todos, todoToTypes(t))
	}
	for _, t := range appeals {
		resp.Appeals = append(resp.Appeals, todoToTypes(t))
	}
	return resp, nil
}

// todoToTypes maps the domain todo row to the API record.
func todoToTypes(t workbenchTodo) types.AssessmentWorkbenchTodo {
	return types.AssessmentWorkbenchTodo{
		Uuid:         t.uuid,
		AssessmentNo: t.assessmentNo,
		PeriodUuid:   t.periodUuid,
		PeriodNo:     t.periodNo,
		PeriodName:   t.periodName,
		EmployeeUuid: t.employeeUuid,
		EmployeeName: t.employeeName,
		MyRole:       t.myRole,
		CurrentNode:  t.currentNode,
		TaskStatus:   t.taskStatus,
		AppealStatus: t.appealStatus,
		UpdateTime:   t.updateTime,
	}
}
