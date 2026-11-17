package scheduleboard

// HRM-15 排班表 logic 测试（svctest，sqlite 后端）。
// 覆盖: 矩阵读写 / 周期生成相位断言（含历史锚点）/ 复制周 / 导入-导出 round-trip /
// 发布标记与发布保护 / 权限口径内考勤组（组外成员拒绝）。

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/jzero-io/jzero/core/stores/condition"

	coreSvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/config"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	attendancegrouplogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/attendancegroup"
	employeelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/employee"
	orglogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/org"
	schedulecyclelogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/schedulecycle"
	shiftlogic "github.com/jzero-io/jzero-admin/server/internal/logic/v1/hrm/shift"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_employee"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_schedule_entry"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	grouptypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/attendancegroup"
	employeetypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/employee"
	orgtypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/org"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/scheduleboard"
	cycletypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/schedulecycle"
	shifttypes "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/shift"
	"github.com/jzero-io/jzero/core/status"
	"github.com/xuri/excelize/v2"
)

func newSvcCtx(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{
		ServiceContext: &coreSvc.ServiceContext{SqlxConn: env.SqlConn},
		ConfigCenter:   config.NewStaticConfigCenter(config.Config{Hrm: config.HrmConf{InitialUserPassword: "test-pwd-123"}}),
		Model:          env.Model,
	}
}

// seeded 测试环境: 公司A→技术部（张三/李四）→产品部（王五）; 启用考勤组挂技术部; 三个班次。
type seeded struct {
	orgA, orgTech, orgProd   string
	empZhang, empLi, empWang string // 张三/李四在技术部（组内），王五在产品部（组外）
	group                    string // 启用考勤组（适用范围=技术部）
	shiftDay, shiftNight     string // 白班/夜班 uuid
}

func seedAll(t *testing.T, svcCtx *svc.ServiceContext) seeded {
	t.Helper()
	ctx := context.Background()

	addOrg := func(parentUuid, name string) string {
		req := &orgtypes.AddRequest{OrgName: name, OrgType: "2", OrderNum: 1, Status: "1"}
		if parentUuid != "" {
			req.ParentUuid = parentUuid
		}
		if _, err := orglogic.NewAdd(ctx, svcCtx, nil).Add(req); err != nil {
			t.Fatalf("add org %s: %v", name, err)
		}
		rows, err := svcCtx.Model.HrmOrg.FindByCondition(ctx, nil)
		if err != nil {
			t.Fatalf("find org %s: %v", name, err)
		}
		for _, row := range rows {
			if row.OrgName == name {
				return row.Uuid
			}
		}
		t.Fatalf("org %s not found", name)
		return ""
	}
	s := seeded{}
	s.orgA = addOrg("", "公司A")
	s.orgTech = addOrg(s.orgA, "技术部")
	s.orgProd = addOrg(s.orgA, "产品部")

	addEmp := func(name, no, orgUuid string) string {
		if _, err := employeelogic.NewAdd(ctx, svcCtx, nil).Add(&employeetypes.AddRequest{
			Name: name, EmployeeNo: no, OrgUuid: orgUuid, PersonStatus: "1",
		}); err != nil {
			t.Fatalf("add employee %s: %v", name, err)
		}
		row, err := svcCtx.Model.HrmEmployee.FindOneByCondition(ctx, nil,
			condition.NewChain().Equal(hrm_employee.EmployeeNo, no).Build()...)
		if err != nil {
			t.Fatalf("find employee %s: %v", no, err)
		}
		return row.Uuid
	}
	s.empZhang = addEmp("张三", "E0001", s.orgTech)
	s.empLi = addEmp("李四", "E0002", s.orgTech)
	s.empWang = addEmp("王五", "E0003", s.orgProd)

	addShift := func(name, start, end string) string {
		if _, err := shiftlogic.NewAdd(ctx, svcCtx, nil).Add(&shifttypes.AddRequest{
			ShiftName: name, StartTime: start, EndTime: end,
		}); err != nil {
			t.Fatalf("add shift %s: %v", name, err)
		}
		records, err := shiftlogic.NewList(ctx, svcCtx, nil).List(&shifttypes.ListRequest{PageRequest: shifttypes.PageRequest{Current: 1, Size: 100}})
		if err != nil {
			t.Fatalf("list shifts: %v", err)
		}
		for _, rec := range records.Records {
			if rec.ShiftName == name {
				return rec.Uuid
			}
		}
		t.Fatalf("shift %s not found", name)
		return ""
	}
	s.shiftDay = addShift("白班", "09:00", "18:00")
	s.shiftNight = addShift("夜班", "22:00", "06:00")

	// 启用考勤组: 适用范围=技术部（部门含下级口径）
	if _, err := attendancegrouplogic.NewAdd(ctx, svcCtx, nil).Add(&grouptypes.AddRequest{
		GroupName: "技术考勤组", GroupType: "2", OrgUuids: []string{s.orgTech},
		Managers: []string{"王经理"}, Status: "1",
	}); err != nil {
		t.Fatalf("add group: %v", err)
	}
	groups, err := svcCtx.Model.HrmAttendanceGroup.FindByCondition(ctx, nil,
		condition.NewChain().Build()...)
	if err != nil {
		t.Fatalf("find group: %v", err)
	}
	for _, g := range groups {
		if g.GroupName == "技术考勤组" {
			s.group = g.Uuid
		}
	}
	if s.group == "" {
		t.Fatal("group not found")
	}
	return s
}

func findCell(t *testing.T, svcCtx *svc.ServiceContext, employeeUuid, date string) *hrm_schedule_entry.HrmScheduleEntry {
	t.Helper()
	row, err := svcCtx.Model.HrmScheduleEntry.FindOneByCondition(context.Background(), nil,
		condition.NewChain().
			Equal(hrm_schedule_entry.ScheduleDate, date).
			Equal(hrm_schedule_entry.EmployeeUuid, employeeUuid).
			Build()...)
	if err != nil && err != hrm_schedule_entry.ErrNotFound {
		t.Fatalf("find cell: %v", err)
	}
	return row
}

func assertErrCode(t *testing.T, err error, wantCode int, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error %d, got nil", what, wantCode)
	}
	if sc := status.FromError(err); sc == nil || sc.Code() != status.Code(wantCode) {
		t.Fatalf("%s: error = %v, want code %d", what, err, wantCode)
	}
}

// TestMatrixReadWrite 矩阵读取: 按考勤组展开成员（技术部两人，产品部王五不出现）+ 单格写读回。
func TestMatrixReadWrite(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedAll(t, svcCtx)
	ctx := context.Background()

	// 空矩阵: 2 行（张三/李四），2026-11 有 30 天，全空班次
	resp, err := NewMatrix(ctx, svcCtx, nil).Matrix(&types.MatrixRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	if len(resp.Rows) != 2 {
		t.Fatalf("rows = %d, want 2 (组外王五不出现)", len(resp.Rows))
	}
	if resp.GroupName != "技术考勤组" || resp.GroupType != "2" {
		t.Fatalf("group meta = %s/%s, want 技术考勤组/2", resp.GroupName, resp.GroupType)
	}
	if len(resp.MonthDays) != 30 || resp.MonthDays[0] != "2026-11-01" || resp.MonthDays[29] != "2026-11-30" {
		t.Fatalf("monthDays wrong: %d days, first=%s last=%s", len(resp.MonthDays), resp.MonthDays[0], resp.MonthDays[29])
	}
	for _, row := range resp.Rows {
		if len(row.Cells) != 30 {
			t.Fatalf("cells = %d, want 30", len(row.Cells))
		}
		for _, c := range row.Cells {
			if c.ShiftUuid != "" || c.Published != "0" {
				t.Fatalf("empty matrix cell not empty: %+v", c)
			}
		}
	}

	// 单格写: 张三 11-02 白班
	if _, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-02", ShiftUuid: s.shiftDay,
	}); err != nil {
		t.Fatalf("save cell: %v", err)
	}
	// 单格写: 李四 11-03 空（休）→ 仍落行（一人一天一行，空=休）
	if _, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empLi, Date: "2026-11-03",
	}); err != nil {
		t.Fatalf("save rest cell: %v", err)
	}

	resp, err = NewMatrix(ctx, svcCtx, nil).Matrix(&types.MatrixRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("matrix 2: %v", err)
	}
	for _, row := range resp.Rows {
		for i, c := range row.Cells {
			if row.EmployeeUuid == s.empZhang && c.Date == "2026-11-02" && c.ShiftUuid != s.shiftDay {
				t.Fatalf("zhang 11-02 shift = %s, want %s", c.ShiftUuid, s.shiftDay)
			}
			if row.EmployeeUuid == s.empLi && c.Date == "2026-11-03" {
				if c.ShiftUuid != "" {
					t.Fatalf("li 11-03 should be rest/empty, got %s", c.ShiftUuid)
				}
				if c.Published != "0" {
					t.Fatalf("li 11-03 published = %s", c.Published)
				}
			}
			_ = i
		}
	}
	cell := findCell(t, svcCtx, s.empLi, "2026-11-03")
	if cell == nil {
		t.Fatal("rest cell should still persist as a row (空=休落行)")
	}

	// 覆盖写: 张三 11-02 改夜班（同键覆盖）
	if _, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-02", ShiftUuid: s.shiftNight,
	}); err != nil {
		t.Fatalf("overwrite cell: %v", err)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-02")
	if cell.ShiftUuid != s.shiftNight {
		t.Fatalf("overwrite failed: %s", cell.ShiftUuid)
	}

	// 组外员工拒绝（王五在产品部，不在考勤组适用范围）
	_, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empWang, Date: "2026-11-02", ShiftUuid: s.shiftDay,
	})
	assertErrCode(t, err, errcodes.ScheduleBoardEmployeeNotFoundCode, "组外员工")

	// 非法班次 uuid
	_, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-04", ShiftUuid: "not-exist",
	})
	assertErrCode(t, err, errcodes.ScheduleBoardShiftInvalidCode, "非法班次")

	// 非法日期
	_, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "20261104",
	})
	assertErrCode(t, err, errcodes.ScheduleBoardInvalidCode, "非法日期")

	// 停用考勤组拒绝
	if err = svcCtx.Model.HrmAttendanceGroup.UpdateFieldsByCondition(ctx, nil,
		map[string]any{"status": "0"},
		condition.NewChain().Equal("uuid", s.group).Build()...); err != nil {
		t.Fatalf("disable group: %v", err)
	}
	_, err = NewMatrix(ctx, svcCtx, nil).Matrix(&types.MatrixRequest{GroupUuid: s.group, Month: "2026-11"})
	assertErrCode(t, err, errcodes.ScheduleBoardGroupNotFoundCode, "停用考勤组")
}

// TestGeneratePhase 周期生成相位断言: 锚点日=序列第 0 项；历史锚点跨月保持绝对相位。
func TestGeneratePhase(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedAll(t, svcCtx)
	ctx := context.Background()

	// 周期 [白班, 夜班]（长度 2）
	if _, err := schedulecyclelogic.NewAdd(ctx, svcCtx, nil).Add(&cycletypes.AddRequest{
		CycleName: "两班倒", ShiftUuids: []string{s.shiftDay, s.shiftNight}, Status: "1",
	}); err != nil {
		t.Fatalf("add cycle: %v", err)
	}
	cycles, err := svcCtx.Model.HrmScheduleCycle.FindByCondition(ctx, nil,
		condition.NewChain().Build()...)
	if err != nil {
		t.Fatalf("find cycle: %v", err)
	}
	cycleUuid := cycles[0].Uuid

	// 锚点缺省（=2026-11-01）: 11-01 白班, 11-02 夜班, 11-03 白班 …
	if _, err = NewGenerate(ctx, svcCtx, nil).Generate(&types.GenerateRequest{
		GroupUuid: s.group, Month: "2026-11", CycleUuid: cycleUuid,
	}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	want := func(day int) string {
		if day%2 == 1 {
			return s.shiftDay // 1,3,5.. → 白班（序列第 0 项锚定 11-01）
		}
		return s.shiftNight
	}
	for day := 1; day <= 30; day++ {
		date := fmt.Sprintf("2026-11-%02d", day)
		cell := findCell(t, svcCtx, s.empZhang, date)
		if cell == nil || cell.ShiftUuid != want(day) {
			t.Fatalf("phase anchor default: 11-%02d cell=%v want %s", day, cell, want(day))
		}
	}

	// 清空重跑: 历史锚点 2026-10-31（=序列第 0 项）→ 11-01 = 序列第 1 项（夜班）
	if err = svcCtx.Model.HrmScheduleEntry.DeleteByCondition(ctx, nil,
		condition.NewChain().Build()...); err != nil {
		t.Fatalf("clear entries: %v", err)
	}
	if _, err = NewGenerate(ctx, svcCtx, nil).Generate(&types.GenerateRequest{
		GroupUuid: s.group, Month: "2026-11", CycleUuid: cycleUuid, AnchorDate: "2026-10-31", Overwrite: "1",
	}); err != nil {
		t.Fatalf("generate anchored: %v", err)
	}
	cell := findCell(t, svcCtx, s.empZhang, "2026-11-01")
	if cell == nil || cell.ShiftUuid != s.shiftNight {
		t.Fatalf("historical anchor phase: 11-01 = %v, want 夜班 %s", cell, s.shiftNight)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-02")
	if cell == nil || cell.ShiftUuid != s.shiftDay {
		t.Fatalf("historical anchor phase: 11-02 = %v, want 白班 %s", cell, s.shiftDay)
	}

	// 同锚点再次生成（保持绝对相位连续）: 已有排班 + overwrite 空 → 不覆盖，written=0
	if _, err = NewGenerate(ctx, svcCtx, nil).Generate(&types.GenerateRequest{
		GroupUuid: s.group, Month: "2026-11", CycleUuid: cycleUuid, AnchorDate: "2026-10-31",
	}); err != nil {
		t.Fatalf("generate idempotent: %v", err)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-01")
	if cell.ShiftUuid != s.shiftNight {
		t.Fatalf("idempotent generate overwrote: %s", cell.ShiftUuid)
	}

	// 周期不存在
	_, err = NewGenerate(ctx, svcCtx, nil).Generate(&types.GenerateRequest{
		GroupUuid: s.group, Month: "2026-11", CycleUuid: "missing",
	})
	assertErrCode(t, err, errcodes.ScheduleBoardCycleNotFoundCode, "周期不存在")

	// 非法月份
	_, err = NewGenerate(ctx, svcCtx, nil).Generate(&types.GenerateRequest{
		GroupUuid: s.group, Month: "2026-13", CycleUuid: cycleUuid,
	})
	assertErrCode(t, err, errcodes.ScheduleBoardInvalidCode, "非法月份")
}

// TestGeneratePublishedProtection 已发布日不受周期生成覆盖（发布保护）。
func TestGeneratePublishedProtection(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedAll(t, svcCtx)
	ctx := context.Background()

	// 先手工保存 11-01 白班并发布
	if _, err := NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-01", ShiftUuid: s.shiftDay,
	}); err != nil {
		t.Fatalf("save cell: %v", err)
	}
	pub, err := NewPublish(ctx, svcCtx, nil).Publish(&types.PublishRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if pub.PublishedCount != 1 {
		t.Fatalf("published = %d, want 1", pub.PublishedCount)
	}

	// 生成（overwrite=1）→ 已发布的 11-01 不被改写
	if _, err := schedulecyclelogic.NewAdd(ctx, svcCtx, nil).Add(&cycletypes.AddRequest{
		CycleName: "单班", ShiftUuids: []string{s.shiftNight}, Status: "1",
	}); err != nil {
		t.Fatalf("add cycle: %v", err)
	}
	cycleRows, err := svcCtx.Model.HrmScheduleCycle.FindByCondition(ctx, nil,
		condition.NewChain().Build()...)
	if err != nil {
		t.Fatalf("find cycle: %v", err)
	}
	if _, err = NewGenerate(ctx, svcCtx, nil).Generate(&types.GenerateRequest{
		GroupUuid: s.group, Month: "2026-11", CycleUuid: cycleRows[0].Uuid, Overwrite: "1",
	}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	cell := findCell(t, svcCtx, s.empZhang, "2026-11-01")
	if cell.ShiftUuid != s.shiftDay {
		t.Fatalf("published day was overwritten: %s, want 白班", cell.ShiftUuid)
	}
	// 其他日（11-02）被生成
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-02")
	if cell == nil || cell.ShiftUuid != s.shiftNight {
		t.Fatalf("generate should fill 11-02 with 夜班, got %v", cell)
	}
}

// TestCopyWeek 复制周: 源周→目标周按星期对齐；非周一拒绝；overwrite 口径。
func TestCopyWeek(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedAll(t, svcCtx)
	ctx := context.Background()

	// 源周 2026-11-02（周一）..11-08: 张三 白班/夜班 交替，李四只排 11-03
	if _, err := NewBatchSave(ctx, svcCtx, nil).BatchSave(&types.BatchSaveRequest{
		GroupUuid: s.group,
		Cells: []types.CellItem{
			{EmployeeUuid: s.empZhang, Date: "2026-11-02", ShiftUuid: s.shiftDay},
			{EmployeeUuid: s.empZhang, Date: "2026-11-03", ShiftUuid: s.shiftNight},
			{EmployeeUuid: s.empZhang, Date: "2026-11-04", ShiftUuid: s.shiftDay},
			{EmployeeUuid: s.empLi, Date: "2026-11-03", ShiftUuid: s.shiftDay},
		},
	}); err != nil {
		t.Fatalf("batch save: %v", err)
	}

	// 非周一拒绝
	_, err := NewCopyWeek(ctx, svcCtx, nil).CopyWeek(&types.CopyWeekRequest{
		GroupUuid: s.group, SourceWeekStart: "2026-11-03", TargetWeekStart: "2026-11-09",
	})
	assertErrCode(t, err, errcodes.ScheduleBoardInvalidCode, "源周非周一")
	_, err = NewCopyWeek(ctx, svcCtx, nil).CopyWeek(&types.CopyWeekRequest{
		GroupUuid: s.group, SourceWeekStart: "2026-11-02", TargetWeekStart: "2026-11-10",
	})
	assertErrCode(t, err, errcodes.ScheduleBoardInvalidCode, "目标周非周一")
	_, err = NewCopyWeek(ctx, svcCtx, nil).CopyWeek(&types.CopyWeekRequest{
		GroupUuid: s.group, SourceWeekStart: "2026-11-02", TargetWeekStart: "2026-11-02",
	})
	assertErrCode(t, err, errcodes.ScheduleBoardInvalidCode, "同周")

	// 复制到 2026-11-09 周
	resp, err := NewCopyWeek(ctx, svcCtx, nil).CopyWeek(&types.CopyWeekRequest{
		GroupUuid: s.group, SourceWeekStart: "2026-11-02", TargetWeekStart: "2026-11-09",
	})
	if err != nil {
		t.Fatalf("copy week: %v", err)
	}
	if resp.WrittenCount != 4 {
		t.Fatalf("written = %d, want 4", resp.WrittenCount)
	}
	checks := []struct{ emp, date, shift string }{
		{s.empZhang, "2026-11-09", s.shiftDay},   // 周一↔周一
		{s.empZhang, "2026-11-10", s.shiftNight}, // 周二↔周二
		{s.empZhang, "2026-11-11", s.shiftDay},
		{s.empLi, "2026-11-10", s.shiftDay},
	}
	for _, ck := range checks {
		cell := findCell(t, svcCtx, ck.emp, ck.date)
		if cell == nil || cell.ShiftUuid != ck.shift {
			t.Fatalf("copied %s %s = %v, want %s", ck.emp, ck.date, cell, ck.shift)
		}
	}
	// 源周休班日（11-04 李四无班）不复制
	if cell := findCell(t, svcCtx, s.empLi, "2026-11-11"); cell != nil {
		t.Fatalf("empty source day should not copy, got %+v", cell)
	}

	// 目标周已有排班 + overwrite 空 → 仅填充空格: 张三 11-09 已是白班，源 11-16 周改夜班再复制不覆盖
	if _, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-16", ShiftUuid: s.shiftNight,
	}); err != nil {
		t.Fatalf("save source 11-16: %v", err)
	}
	if _, err = NewCopyWeek(ctx, svcCtx, nil).CopyWeek(&types.CopyWeekRequest{
		GroupUuid: s.group, SourceWeekStart: "2026-11-16", TargetWeekStart: "2026-11-09",
	}); err != nil {
		t.Fatalf("copy into occupied week: %v", err)
	}
	cell := findCell(t, svcCtx, s.empZhang, "2026-11-09")
	if cell.ShiftUuid != s.shiftDay {
		t.Fatalf("occupied cell overwritten without overwrite flag: %s", cell.ShiftUuid)
	}
	// overwrite=1 → 覆盖
	if _, err = NewCopyWeek(ctx, svcCtx, nil).CopyWeek(&types.CopyWeekRequest{
		GroupUuid: s.group, SourceWeekStart: "2026-11-16", TargetWeekStart: "2026-11-09", Overwrite: "1",
	}); err != nil {
		t.Fatalf("copy overwrite: %v", err)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-09")
	if cell.ShiftUuid != s.shiftNight {
		t.Fatalf("overwrite copy failed: %s, want 夜班", cell.ShiftUuid)
	}
}

// TestImportExportRoundTrip 导入→导出→再导入 round-trip: 值=班次名称编码，休=空。
func TestImportExportRoundTrip(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedAll(t, svcCtx)
	ctx := context.Background()

	// 构造 xlsx: 表头 employee_no,1..30; 张三 1=白班 2=夜班 3=休; 李四 1=夜班 其余空
	buildXlsx := func() []byte {
		header := []string{"employee_no"}
		for i := 1; i <= 30; i++ {
			header = append(header, fmt.Sprintf("%d", i))
		}
		rows := [][]string{
			header,
			append([]string{"E0001"}, dayCells(map[int]string{1: "白班", 2: "夜班", 3: "休"})...),
			append([]string{"E0002"}, dayCells(map[int]string{1: "夜班"})...),
		}
		f := newXlsx(rows)
		buf, err := f.WriteToBuffer()
		if err != nil {
			t.Fatalf("xlsx buffer: %v", err)
		}
		return buf.Bytes()
	}

	resp, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		GroupUuid: s.group, Month: "2026-11",
		FileName: "board.xlsx", FileContent: base64.StdEncoding.EncodeToString(buildXlsx()),
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if resp.SuccessCount != 60 { // 2 人 × 30 天（休/空也落行清空）
		t.Fatalf("successCount = %d, want 60", resp.SuccessCount)
	}
	cell := findCell(t, svcCtx, s.empZhang, "2026-11-01")
	if cell == nil || cell.ShiftUuid != s.shiftDay {
		t.Fatalf("import zhang 11-01 = %v, want 白班", cell)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-02")
	if cell == nil || cell.ShiftUuid != s.shiftNight {
		t.Fatalf("import zhang 11-02 = %v, want 夜班", cell)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-03")
	if cell == nil || cell.ShiftUuid != "" {
		t.Fatalf("import zhang 11-03 = %v, want 休(空)", cell)
	}
	cell = findCell(t, svcCtx, s.empLi, "2026-11-01")
	if cell == nil || cell.ShiftUuid != s.shiftNight {
		t.Fatalf("import li 11-01 = %v, want 夜班", cell)
	}

	// 导出 → 值应为班次名称
	exportResp, err := NewExport(ctx, svcCtx, nil).Export(&types.ExportRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(exportResp.Content)
	if err != nil {
		t.Fatalf("export base64: %v", err)
	}
	grid, err := readXlsxRows(raw)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if len(grid) != 3 {
		t.Fatalf("export rows = %d, want 3 (header+2)", len(grid))
	}
	if grid[0][0] != "employee_no" {
		t.Fatalf("export header[0] = %s", grid[0][0])
	}
	// 注: excelize GetRows 会裁掉行尾空单元格 —— grid[1] = [E0001 张三 白班 夜班 ...截至最后非空]
	if grid[1][0] != "E0001" || grid[1][2] != "白班" || grid[1][3] != "夜班" {
		t.Fatalf("export zhang row = %v", grid[1])
	}
	if grid[2][2] != "夜班" {
		t.Fatalf("export li row = %v", grid[2])
	}

	// round-trip: 再导入导出内容 → 幂等
	resp2, err := NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		GroupUuid: s.group, Month: "2026-11",
		FileName: exportResp.FileName, FileContent: exportResp.Content,
	})
	if err != nil {
		t.Fatalf("re-import: %v", err)
	}
	if resp2.SuccessCount != 60 {
		t.Fatalf("re-import count = %d, want 60", resp2.SuccessCount)
	}
	cell = findCell(t, svcCtx, s.empZhang, "2026-11-01")
	if cell.ShiftUuid != s.shiftDay {
		t.Fatalf("round-trip broken: %s", cell.ShiftUuid)
	}

	// 未知工号 → 整批拒绝（组外王五的工号 E0003）
	badXlsx := newXlsx([][]string{
		{"employee_no", "1"},
		{"E0003", "白班"},
	})
	buf, _ := badXlsx.WriteToBuffer()
	_, err = NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		GroupUuid: s.group, Month: "2026-11",
		FileName: "bad.xlsx", FileContent: base64.StdEncoding.EncodeToString(buf.Bytes()),
	})
	assertErrCode(t, err, errcodes.ScheduleBoardEmployeeNotFoundCode, "导入组外工号")

	// 未知班次编码 → 整批拒绝
	badXlsx2 := newXlsx([][]string{
		{"employee_no", "1"},
		{"E0001", "神秘班"},
	})
	buf2, _ := badXlsx2.WriteToBuffer()
	_, err = NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		GroupUuid: s.group, Month: "2026-11",
		FileName: "bad2.xlsx", FileContent: base64.StdEncoding.EncodeToString(buf2.Bytes()),
	})
	assertErrCode(t, err, errcodes.ScheduleBoardImportInvalidCode, "导入未知班次编码")

	// 表头第 1 列错 → 拒绝
	badXlsx3 := newXlsx([][]string{{"no", "1"}, {"E0001", "白班"}})
	buf3, _ := badXlsx3.WriteToBuffer()
	_, err = NewImport(ctx, svcCtx, nil).Import(&types.ImportRequest{
		GroupUuid: s.group, Month: "2026-11",
		FileName: "bad3.xlsx", FileContent: base64.StdEncoding.EncodeToString(buf3.Bytes()),
	})
	assertErrCode(t, err, errcodes.ScheduleBoardImportInvalidCode, "导入表头错")
}

// TestPublishFlow 发布/撤销发布 + 已发布行禁止单格覆盖。
func TestPublishFlow(t *testing.T) {
	svcCtx := newSvcCtx(t)
	s := seedAll(t, svcCtx)
	ctx := context.Background()

	if _, err := NewBatchSave(ctx, svcCtx, nil).BatchSave(&types.BatchSaveRequest{
		GroupUuid: s.group,
		Cells: []types.CellItem{
			{EmployeeUuid: s.empZhang, Date: "2026-11-01", ShiftUuid: s.shiftDay},
			{EmployeeUuid: s.empLi, Date: "2026-11-01", ShiftUuid: s.shiftNight},
		},
	}); err != nil {
		t.Fatalf("batch save: %v", err)
	}

	pub, err := NewPublish(ctx, svcCtx, nil).Publish(&types.PublishRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if pub.PublishedCount != 2 {
		t.Fatalf("published = %d, want 2", pub.PublishedCount)
	}
	if c := findCell(t, svcCtx, s.empZhang, "2026-11-01"); c.Published != "1" {
		t.Fatalf("zhang published = %s", c.Published)
	}

	// 再次发布 → 幂等 0
	pub, err = NewPublish(ctx, svcCtx, nil).Publish(&types.PublishRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("publish again: %v", err)
	}
	if pub.PublishedCount != 0 {
		t.Fatalf("republished = %d, want 0", pub.PublishedCount)
	}

	// 已发布行单格修改 → 10536
	_, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-01", ShiftUuid: s.shiftNight,
	})
	assertErrCode(t, err, errcodes.ScheduleBoardPublishedLockedCode, "已发布行修改")

	// 撤销发布 → 回 0，可再改
	unpub, err := NewPublish(ctx, svcCtx, nil).Publish(&types.PublishRequest{GroupUuid: s.group, Month: "2026-11", Unpublish: "1"})
	if err != nil {
		t.Fatalf("unpublish: %v", err)
	}
	if unpub.PublishedCount != 2 {
		t.Fatalf("unpublished = %d, want 2", unpub.PublishedCount)
	}
	if c := findCell(t, svcCtx, s.empZhang, "2026-11-01"); c.Published != "0" {
		t.Fatalf("zhang after unpublish = %s", c.Published)
	}
	if _, err = NewSaveCell(ctx, svcCtx, nil).SaveCell(&types.CellRequest{
		GroupUuid: s.group, EmployeeUuid: s.empZhang, Date: "2026-11-01", ShiftUuid: s.shiftNight,
	}); err != nil {
		t.Fatalf("edit after unpublish: %v", err)
	}

	// 矩阵 publishedCount 统计
	if _, err = NewPublish(ctx, svcCtx, nil).Publish(&types.PublishRequest{GroupUuid: s.group, Month: "2026-11"}); err != nil {
		t.Fatalf("publish 2: %v", err)
	}
	resp, err := NewMatrix(ctx, svcCtx, nil).Matrix(&types.MatrixRequest{GroupUuid: s.group, Month: "2026-11"})
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	for _, row := range resp.Rows {
		if row.PublishedCount != 1 {
			t.Fatalf("publishedCount = %d, want 1 (row %s)", row.PublishedCount, row.EmployeeName)
		}
	}
}

// ---- xlsx helpers (测试构造/解析) ----

func dayCells(m map[int]string) []string {
	out := make([]string, 30)
	for i := range out {
		out[i] = m[i+1]
	}
	return out
}

func newXlsx(rows [][]string) *excelize.File {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			panic(err)
		}
	}
	return f
}

func readXlsxRows(raw []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return f.GetRows(f.GetSheetName(0))
}
