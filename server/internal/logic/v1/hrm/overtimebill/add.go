package overtimebill

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/bill"
	"github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/hrmuser"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_overtime_bill"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/overtimebill"
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

// timeNow 可测试钩子（单据编号日期部分用当天）.
var timeNow = time.Now

// Add 新增加班申请单（HR217）：事务内生成单据编号 + 按起止时间自动计算申请时长
// （按命中加班规则的最小计算单位向下取整，未命中规则按默认 30 分钟），初始状态 未提交。
// 提交/审批/撤销走共享单据底座 v1/hrm/bill（类型码 217）。
func (l *Add) Add(req *types.AddRequest) (resp *types.AddResponse, err error) {
	user := hrmuser.Current(l.ctx)
	applyHours, err := validateOvertime(req.DayType, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	// 员工快照补齐：传了员工档案 uuid 时以档案带出部门快照.
	orgUuid, orgName := req.OrgUuid, req.OrgName
	if req.EmployeeUuid != "" {
		emp, err := l.svcCtx.Model.HrmEmployee.FindOneByUuid(l.ctx, nil, req.EmployeeUuid)
		if err == nil && emp.OrgUuid != "" {
			orgUuid = emp.OrgUuid
			if orgName == "" {
				orgName = orgNameOfUuid(l.ctx, l.svcCtx, orgUuid)
			}
		}
	}

	var billNo string
	if txErr := l.svcCtx.SqlxConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		no, err := bill.GenerateBillNo(ctx, session, BillTypeCode, timeNow())
		if err != nil {
			return err
		}
		billNo = no
		_, insertErr := modelOf(l.svcCtx, session).Insert(ctx, session, &model.HrmOvertimeBill{
			Uuid:           uuid.New().String(),
			BillNo:         no,
			Status:         StatusUnsubmitted,
			EmployeeUuid:   req.EmployeeUuid,
			EmployeeNo:     req.EmployeeNo,
			EmployeeName:   req.EmployeeName,
			OrgUuid:        orgUuid,
			OrgName:        orgName,
			OvertimeDate:   req.OvertimeDate,
			DayType:        req.DayType,
			StartTime:      req.StartTime,
			EndTime:        req.EndTime,
			ApplyHours:     applyHours,
			Reason:         req.Reason,
			Source:         SourceEmployee,
			CompMultiplier: 1,
			Settled:        0,
			CreateBy:       user.Uuid,
			CreatorName:    user.Username,
		})
		return insertErr
	}); txErr != nil {
		return nil, txErr
	}

	l.Infof("overtimebill: 加班申请单已创建 billNo=%s employee=%s(%s) dayType=%s applyHours=%.1f",
		billNo, req.EmployeeName, req.EmployeeNo, req.DayType, applyHours)
	return &types.AddResponse{}, nil
}

// orgNameOfUuid 取组织名称（找不到返回空串，快照允许为空）.
func orgNameOfUuid(ctx context.Context, svcCtx *svc.ServiceContext, orgUuid string) string {
	if orgUuid == "" {
		return ""
	}
	row, err := hrm_org.NewHrmOrgModel(svcCtx.SqlxConn).FindOneByUuid(ctx, nil, orgUuid)
	if err != nil {
		return ""
	}
	return row.OrgName
}
