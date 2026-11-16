package perf_result

import (
	"context"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/google/uuid"

	hrm_employee "github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	hrm_org "github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	model "github.com/jzero-io/jzero-admin/server/internal/model/hrm_perf_result_approval"
)

// 系数映射（HRM-43 契约口径，hrm-43-notes：等级 → 绩效系数；
// 等级制非 A/B/C/D 值 → 1.0 默认）。
const (
	CoefficientA = 1.2
	CoefficientB = 1.0
	CoefficientC = 0.8
	CoefficientD = 0.6
)

// GradeCoefficient maps 最终等级 to 绩效系数:
//
//	A→1.2 / B→1.0 / C→0.8 / D→0.6；非 A/B/C/D（等级制原值）→ 1.0 默认。
//	百分比制等级由 hrm_perf_assessment.PercentGrade 产出；HRM-49 联薪照抄本映射。
func GradeCoefficient(grade string) float64 {
	switch grade {
	case "A":
		return CoefficientA
	case "B":
		return CoefficientB
	case "C":
		return CoefficientC
	case "D":
		return CoefficientD
	default:
		return 1.0
	}
}

// FindApprovalByUuid loads one 审批单; ErrResultApprovalNotFound when missing.
func FindApprovalByUuid(ctx context.Context, conn sqlx.SqlConn, uuid string) (*model.HrmPerfResultApproval, error) {
	row, err := model.NewHrmPerfResultApprovalModel(conn).FindOneByUuid(ctx, nil, uuid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrResultApprovalNotFound
		}
		return nil, err
	}
	return row, nil
}

// FindApprovalByPeriod loads the 审批单 of a 批次 (一批次一张审批单);
// ErrResultApprovalNotFound when the 批次 has none yet.
func FindApprovalByPeriod(ctx context.Context, conn sqlx.SqlConn, periodUuid string) (*model.HrmPerfResultApproval, error) {
	row, err := model.NewHrmPerfResultApprovalModel(conn).FindOneByCondition(ctx, nil,
		condition.NewChain().Equal(model.PeriodUuid, periodUuid).Build()...)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrResultApprovalNotFound
		}
		return nil, err
	}
	return row, nil
}

// round2 rounds to 2 decimals (金额/得分 decimal 精度口径).
func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}

// snapshotNamesByUuid resolves uuids to display names in one IN query; missing
// uuids resolve to "".
func employeeRowsByUuids(ctx context.Context, conn sqlx.SqlConn, uuids []string) (map[string]*hrm_employee.HrmEmployee, error) {
	out := make(map[string]*hrm_employee.HrmEmployee, len(uuids))
	if len(uuids) == 0 {
		return out, nil
	}
	rows, err := hrm_employee.NewHrmEmployeeModel(conn).FindByCondition(ctx, nil,
		condition.NewChain().In(hrm_employee.Uuid, uuids).Build()...)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.Uuid] = r
	}
	return out, nil
}

// orgNameByUuid resolves org uuid → org_name (single IN query; missing → "").
func orgNameByUuid(ctx context.Context, conn sqlx.SqlConn, orgUuid string) (string, error) {
	if orgUuid == "" {
		return "", nil
	}
	row, err := hrm_org.NewHrmOrgModel(conn).FindOneByUuid(ctx, nil, orgUuid)
	if err != nil {
		if err == hrm_org.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return row.OrgName, nil
}

// timeFormat is the varchar(20) time column format (YYYY-MM-DD HH:MM:SS).
const timeFormat = "2006-01-02 15:04:05"

// nowStr formats a timestamp for the varchar(20) time columns.
func nowStr() string {
	return time.Now().Format(timeFormat)
}

// insertApproval persists one 审批单 row (事务内调用; bill_no 由调用方先分配).
func insertApproval(ctx context.Context, session sqlx.Session, row *model.HrmPerfResultApproval) error {
	row.Uuid = uuid.New().String()
	return model.NewHrmPerfResultApprovalModel(sqlx.NewSqlConnFromSession(session)).InsertV2(ctx, session, row)
}
