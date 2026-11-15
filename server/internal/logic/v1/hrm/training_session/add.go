package training_session

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzero-io/jzero/core/status"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_plan"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training_session"
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

// timeLayout is the 起止时间 format (YYYY-MM-DD HH:mm:ss).
const timeLayout = time.DateTime

// sessionStatuses are the 班次状态 codes: 1=未开始 2=进行中 3=已结束 (HRM-31).
const (
	SessionStatusNotStarted = "1"
	SessionStatusOngoing    = "2"
	SessionStatusFinished   = "3"
)

// validateSession is the parameter validation shared by Add/Edit: 名称必填、
// 时间格式与先后顺序.
func validateSession(sessionName, startTime, endTime string) error {
	if sessionName == "" {
		return status.ErrorMessage(errcodes.TrainingSessionInvalidCode, "班次名称不能为空")
	}
	sd, err := time.ParseInLocation(timeLayout, startTime, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.TrainingSessionInvalidCode,
			"开始时间格式非法（须为 YYYY-MM-DD HH:mm:ss）: "+startTime)
	}
	ed, err := time.ParseInLocation(timeLayout, endTime, time.Local)
	if err != nil {
		return status.ErrorMessage(errcodes.TrainingSessionInvalidCode,
			"结束时间格式非法（须为 YYYY-MM-DD HH:mm:ss）: "+endTime)
	}
	if !ed.After(sd) {
		return status.ErrorMessage(errcodes.TrainingSessionInvalidCode,
			"结束时间必须晚于开始时间")
	}
	return nil
}

// Add 新增培训班次（管理端）：事务内生成班次编号 PX-年-流水，初始状态 1=未开始.
// 关联培训计划必须已发布（published=1，发布计划后可建班次）。
// 签到（HRM-29）/证书发放（HRM-32）为后续工单，本处只落 plan_uuid/participant 缝隙。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	if err = validateSession(req.SessionName, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.PlanUuid) == "" {
		return nil, status.ErrorMessage(errcodes.TrainingSessionInvalidCode, "关联培训计划必填")
	}

	var sessionNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 发布计划后可建班次：plan must exist and be published (published=1)
		plan, err := l.svcCtx.Model.HrmTrainingPlan.FindOneByUuid(ctx, session, req.PlanUuid)
		if err != nil {
			if err == hrm_training_plan.ErrNotFound {
				return status.ErrorMessage(errcodes.TrainingSessionPlanNotPublishedCode, "关联的培训计划不存在")
			}
			return err
		}
		if plan.Published != 1 {
			return status.ErrorMessage(errcodes.TrainingSessionPlanNotPublishedCode,
				"关联的培训计划未发布（仅已发布计划可建班次）")
		}

		startTime, perr := time.ParseInLocation(timeLayout, req.StartTime, time.Local)
		if perr != nil {
			return perr
		}
		no, err := GenerateSessionNo(ctx, session, startTime.Year(), startTime)
		if err != nil {
			return err
		}
		sessionNo = no
		return l.svcCtx.Model.HrmTrainingSession.InsertV2(ctx, session, &hrm_training_session.HrmTrainingSession{
			Uuid:        uuid.New().String(),
			SessionNo:   no,
			SessionName: req.SessionName,
			PlanUuid:    req.PlanUuid,
			Lecturer:    req.Lecturer,
			Location:    req.Location,
			Hours:       req.Hours,
			StartTime:   startTime,
			EndTime:     mustParseTime(req.EndTime),
			Status:      SessionStatusNotStarted,
			Remark:      req.Remark,
		})
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("training_session: 培训班次已创建 sessionNo=%s plan=%s", sessionNo, req.PlanUuid)
	return &types.AddResponse{}, nil
}

func mustParseTime(s string) time.Time {
	t, _ := time.ParseInLocation(timeLayout, s, time.Local)
	return t
}
