package attendanceconfig

import (
	"strings"
	"context"
	"net/http"
	"time"

	"github.com/jzero-io/jzero/core/stores/condition"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_attendance_config"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_org"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendanceconfig"
)

type List struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewList(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *List {
	return &List{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx, r: r,
	}
}

// List 组织默认规则分页列表（按组织名称筛选）.
func (l *List) List(req *types.ListRequest) (resp *types.ListResponse, err error) {
	rows, total, err := l.svcCtx.Model.HrmAttendanceConfig.PageByCondition(l.ctx, nil,
		condition.NewChain().
			Page(req.Current, req.Size).
			OrderByDesc(hrm_attendance_config.Id).
			Build()...)
	if err != nil {
		return nil, err
	}

	// 组织名称筛选: 命中组织的规则保留
	orgNames := make(map[string]string)
	if req.OrgName != "" {
		orgs, err := l.svcCtx.Model.HrmOrg.FindByCondition(l.ctx, nil)
		if err != nil {
			return nil, err
		}
		for _, org := range orgs {
			orgNames[org.Uuid] = org.OrgName
		}
	}

	records := make([]types.HrmAttendanceConfig, 0, len(rows))
	for _, row := range rows {
		orgName := ""
		if name, ok := orgNames[row.OrgUuid]; ok {
			orgName = name
			if req.OrgName != "" && !containsStr([]string{orgName}, req.OrgName) &&
				!substringMatch(orgName, req.OrgName) {
				continue
			}
		} else if req.OrgName != "" {
			continue
		} else {
			org, err := l.svcCtx.Model.HrmOrg.FindOneByUuid(l.ctx, nil, row.OrgUuid)
			if err != nil {
				if err == hrm_org.ErrNotFound {
					org = nil
				} else {
					return nil, err
				}
			}
			if org != nil {
				orgName = org.OrgName
			}
		}
		records = append(records, types.HrmAttendanceConfig{
			Uuid:       row.Uuid,
			OrgUuid:    row.OrgUuid,
			OrgName:    orgName,
			StartTime:  row.StartTime,
			EndTime:    row.EndTime,
			WorkDays:   csvToInt64s(row.WorkDays),
			Remark:     row.Remark,
			CreateTime: row.CreateTime.Format(time.DateTime),
			UpdateTime: row.UpdateTime.Format(time.DateTime),
		})
	}

	// OrgName 筛选时 Total 以过滤后为准
	totalFiltered := total
	if req.OrgName != "" {
		totalFiltered = int64(len(records))
	}

	return &types.ListResponse{PageResponse: types.PageResponse{
		Current: req.Current,
		Size:    req.Size,
		Total:   totalFiltered,
	}, Records: records}, nil
}

func substringMatch(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func containsStr(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

func csvToInt64s(v string) []int64 {
	parts := splitCsv(v)
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		if n, ok := parseInt64(p); ok {
			out = append(out, n)
		}
	}
	return out
}

func splitCsv(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseInt64(s string) (int64, bool) {
	var n int64
	if s == "" {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int64(c-'0')
	}
	return n, true
}
