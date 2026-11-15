package candidate

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/jzero-io/jzero/core/status"
	"github.com/xuri/excelize/v2"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/candidate"
)

// Seams under test (ticket HRM-26 acceptance criteria), all at the logic-method
// seam against the real migrations (svctest):
//  1. Add: 初始状态 新线索(new)，来源渠道缺省 other。
//  2. Validation: 姓名必填/来源渠道/最高学历枚举 guards (1044x)。
//  3. UpdateStatus: 状态流转合法性 — 不可跳级/不可回退，rejected 入口，hired 终态。
//  4. Import: Excel round-trip (label→code parse, whole-tx rollback on bad row)。
//  5. Delete: 批量删除（普通实体，不限状态）。
func testCtx() context.Context {
	return context.Background()
}

func newSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func addCandidate(t *testing.T, svcCtx *svc.ServiceContext, mutate func(*types.AddRequest)) string {
	t.Helper()
	req := &types.AddRequest{
		Name:            "张三",
		Phone:           "13800000001",
		Email:           "zhangsan@example.com",
		SourceChannel:   SourceChannelRecommend,
		Education:       EducationBachelor,
		CurrentCompany:  "甲公司",
		CurrentPosition: "Go 工程师",
		ExpectSalary:    "20k-30k",
	}
	if mutate != nil {
		mutate(req)
	}
	if _, err := NewAdd(testCtx(), svcCtx, nil).Add(req); err != nil {
		t.Fatalf("Add: %v", err)
	}
	resp, err := NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List after Add: %v", err)
	}
	if len(resp.Records) == 0 {
		t.Fatal("List after Add: no records")
	}
	// 列表按 id 倒序 — 首条即刚插入的候选人
	return resp.Records[0].Uuid
}

func TestAddDefaultsAndRoundtrip(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	resp, err := NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if resp.Total != 0 || len(resp.Records) != 0 {
		t.Fatalf("empty list: total=%d records=%d", resp.Total, len(resp.Records))
	}

	addCandidate(t, svcCtx, func(req *types.AddRequest) {
		req.SourceChannel = "" // 缺省 → other
	})

	recs, err := NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	rec := recs.Records[0]
	if rec.Status != CandidateStatusNew {
		t.Fatalf("new candidate status = %q, want new", rec.Status)
	}
	if rec.SourceChannel != SourceChannelOther {
		t.Fatalf("blank source channel default = %q, want other", rec.SourceChannel)
	}
	if rec.Name != "张三" || rec.Phone != "13800000001" || rec.ExpectSalary != "20k-30k" {
		t.Fatalf("record round-trip mismatch: %+v", rec)
	}
}

func TestAddValidation(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	cases := []struct {
		name   string
		mutate func(*types.AddRequest)
	}{
		{"blank name", func(r *types.AddRequest) { r.Name = "" }},
		{"bad source channel", func(r *types.AddRequest) { r.SourceChannel = "weibo" }},
		{"bad education", func(r *types.AddRequest) { r.Education = "小学" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &types.AddRequest{Name: "张三", SourceChannel: SourceChannelRecommend, Education: EducationBachelor}
			tc.mutate(req)
			_, err := NewAdd(testCtx(), svcCtx, nil).Add(req)
			if err == nil {
				t.Fatal("Add: expected error, got nil")
			}
			code := status.FromError(err).Code()
			_ = code
			if code < 10440 || code > 10443 {
				t.Fatalf("Add: errcode = %v, want 1044x", code)
			}
		})
	}

	// nothing persisted by the rejected adds
	recs, err := NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(recs.Records) != 0 {
		t.Fatalf("rejected adds must not persist, got %d records", len(recs.Records))
	}
}

func TestUpdateStatusTransitionRules(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)
	logic := NewUpdateStatus(testCtx(), svcCtx, nil)

	// full forward chain 新线索→已投递→面试中→已发Offer→录用通过
	uuid := addCandidate(t, svcCtx, nil)
	for _, target := range []string{CandidateStatusDelivered, CandidateStatusInterviewing, CandidateStatusOffer, CandidateStatusHired} {
		if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid, Status: target}); err != nil {
			t.Fatalf("UpdateStatus → %s: %v", target, err)
		}
	}

	// hired is terminal
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid, Status: CandidateStatusRejected}); err == nil {
		t.Fatal("UpdateStatus from hired: expected error, got nil")
	} else if code := status.FromError(err).Code(); code != 10442 {
		t.Fatalf("terminal transition errcode = %d, want 10442", code)
	}

	// 不可跳级: new → interviewing rejected
	uuid2 := addCandidate(t, svcCtx, func(r *types.AddRequest) { r.Name = "李四" })
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid2, Status: CandidateStatusInterviewing}); err == nil {
		t.Fatal("skip-level transition new→interviewing: expected error, got nil")
	} else if code := status.FromError(err).Code(); code != 10442 {
		t.Fatalf("skip-level errcode = %d, want 10442", code)
	}

	// 不可回退: delivered → new rejected
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid2, Status: CandidateStatusDelivered}); err != nil {
		t.Fatalf("new→delivered: %v", err)
	}
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid2, Status: CandidateStatusNew}); err == nil {
		t.Fatal("backward transition delivered→new: expected error, got nil")
	}

	// rejected 可从非终态进入; rejected 自身不可再流转
	uuid3 := addCandidate(t, svcCtx, func(r *types.AddRequest) { r.Name = "王五" })
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid3, Status: CandidateStatusRejected}); err != nil {
		t.Fatalf("rejected from new: %v", err)
	}
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid3, Status: CandidateStatusDelivered}); err == nil {
		t.Fatal("transition from rejected: expected error, got nil")
	}

	// unknown target status
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: uuid2, Status: "archived"}); err == nil {
		t.Fatal("unknown status: expected error, got nil")
	}

	// not found
	if _, err := logic.UpdateStatus(&types.UpdateStatusRequest{Uuid: "no-such-uuid", Status: CandidateStatusDelivered}); err == nil {
		t.Fatal("missing uuid: expected error, got nil")
	} else if code := status.FromError(err).Code(); code != 10440 {
		t.Fatalf("not-found errcode = %d, want 10440", code)
	}
}

func TestImportRoundTripAndRollback(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	buildXlsx := func(t *testing.T, rows [][]string) string {
		t.Helper()
		f := excelize.NewFile()
		sheet := f.GetSheetName(0)
		for i, row := range rows {
			cell, _ := excelize.CoordinatesToCellName(1, i+1)
			if err := f.SetSheetRow(sheet, cell, &row); err != nil {
				t.Fatalf("SetSheetRow: %v", err)
			}
		}
		buf, err := f.WriteToBuffer()
		if err != nil {
			t.Fatalf("WriteToBuffer: %v", err)
		}
		return base64.StdEncoding.EncodeToString(buf.Bytes())
	}

	goodRows := [][]string{
		importHeader,
		{"张三", "13800000001", "zhangsan@example.com", "内部推荐", "本科", "甲公司", "Go 工程师", "20k-30k"},
		{"李四", "13800000002", "lisi@example.com", "招聘网站", "硕士", "乙公司", "前端工程师", "25k"},
		{"王五", "", "", "", "", "", "", ""},
	}
	resp, err := NewImport(testCtx(), svcCtx, nil).Import(&types.ImportRequest{FileName: "candidates.xlsx", Content: buildXlsx(t, goodRows)})
	if err != nil {
		t.Fatalf("Import good file: %v", err)
	}
	if resp.SuccessCount != 3 {
		t.Fatalf("Import SuccessCount = %d, want 3", resp.SuccessCount)
	}

	recs, err := NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(recs.Records) != 3 {
		t.Fatalf("after import: %d records, want 3", len(recs.Records))
	}
	byName := make(map[string]types.HrmCandidate, len(recs.Records))
	for _, rec := range recs.Records {
		byName[rec.Name] = rec
	}
	zs := byName["张三"]
	if zs.SourceChannel != SourceChannelRecommend || zs.Education != EducationBachelor || zs.Status != CandidateStatusNew {
		t.Fatalf("imported 张三 parsed wrong: %+v", zs)
	}
	ww := byName["王五"]
	if ww.SourceChannel != SourceChannelOther || ww.Status != CandidateStatusNew {
		t.Fatalf("imported blank defaults wrong: %+v", ww)
	}

	// bad row (非法来源渠道) → whole import rolled back
	badRows := [][]string{
		importHeader,
		{"赵六", "13800000003", "", "微博", "本科", "", "", ""},
	}
	_, err = NewImport(testCtx(), svcCtx, nil).Import(&types.ImportRequest{FileName: "bad.xlsx", Content: buildXlsx(t, badRows)})
	if err == nil {
		t.Fatal("Import bad file: expected error, got nil")
	} else if code := status.FromError(err).Code(); code != 10443 {
		t.Fatalf("bad import errcode = %d, want 10443", code)
	}
	recs, err = NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(recs.Records) != 3 {
		t.Fatalf("rolled-back import must not persist, got %d records", len(recs.Records))
	}

	// non-xlsx content
	_, err = NewImport(testCtx(), svcCtx, nil).Import(&types.ImportRequest{FileName: "x.xlsx", Content: base64.StdEncoding.EncodeToString([]byte("not excel"))})
	if err == nil {
		t.Fatal("Import non-xlsx: expected error, got nil")
	}
}

func TestBatchDeleteAndListFilters(t *testing.T) {
	env := svctest.New(t)
	svcCtx := newSvcCtx(t, env)

	uuidA := addCandidate(t, svcCtx, func(r *types.AddRequest) { r.Name = "张三"; r.Phone = "13800000001" })
	uuidB := addCandidate(t, svcCtx, func(r *types.AddRequest) { r.Name = "李四"; r.Phone = "13800000002" })
	uuidC := addCandidate(t, svcCtx, func(r *types.AddRequest) { r.Name = "王五"; r.Phone = "13900000003" })

	// filter by name (like)
	resp, err := NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Name: "张"})
	if err != nil {
		t.Fatalf("List by name: %v", err)
	}
	if resp.Total != 1 || resp.Records[0].Name != "张三" {
		t.Fatalf("name filter mismatch: total=%d", resp.Total)
	}

	// filter by phone (like)
	resp, err = NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, Phone: "138"})
	if err != nil {
		t.Fatalf("List by phone: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("phone filter mismatch: total=%d, want 2", resp.Total)
	}

	// filter by source channel
	resp, err = NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}, SourceChannel: SourceChannelRecommend})
	if err != nil {
		t.Fatalf("List by channel: %v", err)
	}
	if resp.Total != 3 {
		t.Fatalf("channel filter mismatch: total=%d, want 3 (default recommend)", resp.Total)
	}

	// batch delete A + B
	if _, err = NewDelete(testCtx(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{uuidA, uuidB}}); err != nil {
		t.Fatalf("BatchDelete: %v", err)
	}
	resp, err = NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if resp.Total != 1 || resp.Records[0].Uuid != uuidC {
		t.Fatalf("after batch delete: total=%d, want 1 (王五)", resp.Total)
	}

	// delete survives status transitions (普通实体不限状态)
	if _, err = NewUpdateStatus(testCtx(), svcCtx, nil).UpdateStatus(&types.UpdateStatusRequest{Uuid: uuidC, Status: CandidateStatusDelivered}); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if _, err = NewDelete(testCtx(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{uuidC}}); err != nil {
		t.Fatalf("Delete after transition: %v", err)
	}
	resp, err = NewList(testCtx(), svcCtx, nil).List(&types.ListRequest{PageRequest: types.PageRequest{Current: 1, Size: 10}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if resp.Total != 0 {
		t.Fatalf("after final delete: total=%d, want 0", resp.Total)
	}

	// empty uuids is a no-op, not an error
	if _, err = NewDelete(testCtx(), svcCtx, nil).Delete(&types.DeleteRequest{Uuids: []string{}}); err != nil {
		t.Fatalf("Delete empty: %v", err)
	}
}
