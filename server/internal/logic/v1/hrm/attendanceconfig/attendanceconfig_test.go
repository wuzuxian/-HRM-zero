package attendanceconfig

import (
	"context"
	"testing"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	orglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendanceconfig"
)

func newCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func mustList(t *testing.T, svcCtx *svc.ServiceContext, req types.ListRequest) []types.HrmAttendanceConfig {
	t.Helper()
	if req.Current == 0 {
		req.Current = 1
	}
	if req.Size == 0 {
		req.Size = 100
	}
	resp, err := NewList(context.Background(), svcCtx, nil).List(&req)
	if err != nil {
		t.Fatalf("list configs: %v", err)
	}
	return resp.Records
}

// TestConfigCrudUpsertAndValidation walks 组织默认规则维护:
// 新组织首次保存=新增; 同组织再次保存=覆盖; 校验拒绝非法时间/考勤日/未知组织.
func TestConfigCrudUpsertAndValidation(t *testing.T) {
	svcCtx := newCtx(t)
	ctx := context.Background()

	if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(&orgtypes.AddRequest{
		OrgName: "公司A", OrgType: "1", OrderNum: 1, Status: "1",
	}); err != nil {
		t.Fatalf("add org: %v", err)
	}
	rows, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil)
	if err != nil {
		t.Fatalf("find org: %v", err)
	}
	orgUuid := rows[0].Uuid

	// 快速开始: 首次保存新增
	if _, err = NewSave(ctx, svcCtx, nil).Save(&types.SaveRequest{
		OrgUuid: orgUuid, StartTime: "09:00", EndTime: "18:00", WorkDays: []int64{1, 2, 3, 4, 5},
	}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if records := mustList(t, svcCtx, types.ListRequest{}); len(records) != 1 ||
		records[0].OrgName != "公司A" || records[0].StartTime != "09:00" || len(records[0].WorkDays) != 5 {
		t.Fatalf("unexpected records: %+v", records)
	}

	// 再次保存同组织 = 编辑（不产生第二条）
	if _, err = NewSave(ctx, svcCtx, nil).Save(&types.SaveRequest{
		OrgUuid: orgUuid, StartTime: "09:30", EndTime: "18:30", WorkDays: []int64{1, 2, 3, 4},
	}); err != nil {
		t.Fatalf("resave config: %v", err)
	}
	records := mustList(t, svcCtx, types.ListRequest{})
	if len(records) != 1 || records[0].StartTime != "09:30" || len(records[0].WorkDays) != 4 {
		t.Fatalf("upsert should overwrite, got: %+v", records)
	}

	// 校验: 非法组合
	for _, bad := range []types.SaveRequest{
		{OrgUuid: orgUuid, StartTime: "9:00", EndTime: "18:00"},              // 时间格式
		{OrgUuid: orgUuid, StartTime: "09:00", EndTime: "09:00"},             // 相同
		{OrgUuid: orgUuid, StartTime: "09:00", EndTime: "18:00", WorkDays: []int64{0, 8}}, // 考勤日越界
		{OrgUuid: orgUuid, StartTime: "09:00", EndTime: ""},                  // 只填一半
	} {
		if _, err = NewSave(ctx, svcCtx, nil).Save(&bad); err == nil {
			t.Fatalf("save %+v should fail", bad)
		}
	}
	// 未知组织
	if _, err = NewSave(ctx, svcCtx, nil).Save(&types.SaveRequest{
		OrgUuid: "no-such-org", StartTime: "09:00", EndTime: "18:00",
	}); err == nil {
		t.Fatal("save for missing org should fail")
	}

	// 删除
	if _, err = NewDelete(ctx, svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{records[0].Uuid}}); err != nil {
		t.Fatalf("delete config: %v", err)
	}
	if records := mustList(t, svcCtx, types.ListRequest{}); len(records) != 0 {
		t.Fatalf("expect empty after delete, got %+v", records)
	}
}
