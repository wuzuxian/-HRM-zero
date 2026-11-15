package certificate

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	coresvc "github.com/jzero-io/jzero-admin/core-engine/svc"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_cert_template"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/model/manage_user"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/certificate"
)

// Seams under test (per HRM-30 acceptance criteria):
//  1. 批量发证与证书号生成: 证书号=前缀-年-流水（FormatCertNo + GenerateCertNo per 前缀+年, 复用 hrm_bill_sequence）
//  2. 批量发证幂等: 重复调用同一班次+模板跳过已有有效证书的学员（IssuedCount=0）
//  3. 撤销路径: 已发→已撤销；待发/已撤销拒绝
//  4. 导入外部证书: source='2', 证书号唯一, 可带扫描件
//  5. 模板停用后不可发证（CertificateTemplateDisabledCode）

func testSvcCtx(t *testing.T, env *svctest.Env) *svc.ServiceContext {
	t.Helper()
	return &svc.ServiceContext{
		ServiceContext: &coresvc.ServiceContext{SqlxConn: env.SqlConn},
		Model:          env.Model,
	}
}

func seedTemplate(t *testing.T, env *svctest.Env, prefix string, status string) *hrm_training_cert_template.HrmTrainingCertTemplate {
	t.Helper()
	tpl := &hrm_training_cert_template.HrmTrainingCertTemplate{
		Uuid:         uuid.New().String(),
		TemplateName: prefix + "证书模板",
		CertType:     "结业证",
		Source:       "1",
		NoPrefix:     prefix,
		Status:       status,
	}
	if _, err := hrm_training_cert_template.NewHrmTrainingCertTemplateModel(env.SqlConn).Insert(context.Background(), nil, tpl); err != nil {
		t.Fatalf("seed template: %v", err)
	}
	return tpl
}

func seedSessionWithParticipants(t *testing.T, env *svctest.Env, userUuids ...string) string {
	t.Helper()
	ctx := context.Background()
	sessionUuid := uuid.New().String()
	start, _ := time.Parse(time.DateTime, "2026-11-20 09:00:00")
	end, _ := time.Parse(time.DateTime, "2026-11-21 17:00:00")
	if _, err := env.Model.HrmTrainingSession.Insert(ctx, nil, &hrm_training_session.HrmTrainingSession{
		Uuid: sessionUuid, SessionNo: "PX-TEST-" + sessionUuid[:8], SessionName: "测试班次",
		StartTime: start, EndTime: end, Status: "3",
	}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	for _, u := range userUuids {
		if _, err := env.Model.HrmTrainingSessionParticipant.Insert(ctx, nil, &hrm_training_session_participant.HrmTrainingSessionParticipant{
			Uuid: uuid.New().String(), SessionUuid: sessionUuid, UserUuid: u,
		}); err != nil {
			t.Fatalf("seed participant: %v", err)
		}
	}
	return sessionUuid
}

func seedUser(t *testing.T, env *svctest.Env, account string) string {
	t.Helper()
	u := uuid.New().String()
	env.Model.ManageUser.InsertV2(context.Background(), nil, &manage_user.ManageUser{
		Uuid: u, Username: account, Nickname: account + "姓名", Password: "x", Gender: "1", Status: "1",
	})
	return u
}

func TestFormatCertNo(t *testing.T) {
	no, err := FormatCertNo("JN", 2026, 1)
	if err != nil {
		t.Fatalf("FormatCertNo: %v", err)
	}
	if no != "JN-2026-000001" {
		t.Fatalf("FormatCertNo = %q, want JN-2026-000001 (迁移 15 演示格式)", no)
	}
	if _, err = FormatCertNo("JN", 2026, 0); err == nil {
		t.Fatal("seq 0 should be rejected")
	}
	if _, err = FormatCertNo("JN", 2026, 1000000); err == nil {
		t.Fatal("seq 1000000 (6 位流水上限) should be rejected")
	}
	if _, err = FormatCertNo("", 2026, 1); err == nil {
		t.Fatal("empty prefix should be rejected")
	}
}

func TestGenerateCertNoPerPrefixAndYear(t *testing.T) {
	env := svctest.New(t)
	ctx := context.Background()

	n1, err := GenerateCertNo(ctx, env.SqlConn, "JN", 2026)
	if err != nil {
		t.Fatalf("GenerateCertNo #1: %v", err)
	}
	n2, err := GenerateCertNo(ctx, env.SqlConn, "JN", 2026)
	if err != nil {
		t.Fatalf("GenerateCertNo #2: %v", err)
	}
	if n1 != "JN-2026-000001" || n2 != "JN-2026-000002" {
		t.Fatalf("per-year allocation not monotonic: %s then %s", n1, n2)
	}

	// 不同前缀独立流水（同一 hrm_bill_sequence 表, type code = 前缀）
	aq1, err := GenerateCertNo(ctx, env.SqlConn, "AQ", 2026)
	if err != nil || aq1 != "AQ-2026-000001" {
		t.Fatalf("independent prefix sequence: %q err=%v", aq1, err)
	}
	// 不同年重置流水
	jn27, err := GenerateCertNo(ctx, env.SqlConn, "JN", 2027)
	if err != nil || jn27 != "JN-2027-000001" {
		t.Fatalf("per-year reset: %q err=%v", jn27, err)
	}
}

func TestBatchIssueAndIdempotent(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	tpl := seedTemplate(t, env, "TS", "1")
	u1 := seedUser(t, env, "user-a")
	u2 := seedUser(t, env, "user-b")
	sessionUuid := seedSessionWithParticipants(t, env, u1, u2)

	// 第一次：全部学员发证
	resp, err := NewBatchIssue(context.Background(), svcCtx, nil).BatchIssue(&types.BatchIssueRequest{
		SessionUuid: sessionUuid, TemplateUuid: tpl.Uuid,
	})
	if err != nil {
		t.Fatalf("BatchIssue: %v", err)
	}
	if resp.IssuedCount != 2 || resp.SkippedCount != 0 {
		t.Fatalf("first issue = issued %d skipped %d, want 2/0", resp.IssuedCount, resp.SkippedCount)
	}

	// 证书号=前缀-年-流水，source='1'，template_uuid 回填，状态已发
	list, err := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, SessionUuid: sessionUuid,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Records) != 2 {
		t.Fatalf("records = %d, want 2", len(list.Records))
	}
	seqs := map[string]bool{}
	for _, rec := range list.Records {
		if !strings.HasPrefix(rec.CertNo, "TS-2026-") || len(rec.CertNo) != len("TS-2026-000001") {
			t.Fatalf("certNo %q does not follow 前缀-年-流水 format", rec.CertNo)
		}
		if seqs[rec.CertNo] {
			t.Fatalf("duplicate certNo %q", rec.CertNo)
		}
		seqs[rec.CertNo] = true
		if rec.Status != "1" || rec.Source != "1" || rec.TemplateUuid != tpl.Uuid || rec.SessionUuid != sessionUuid {
			t.Fatalf("record mismatch: %+v", rec)
		}
		if rec.IssueTime == "" || rec.TemplateName != tpl.TemplateName || rec.SessionName != "测试班次" {
			t.Fatalf("enrichment mismatch: %+v", rec)
		}
	}

	// 第二次（幂等）：同一班次+模板 → 全部跳过
	resp2, err := NewBatchIssue(context.Background(), svcCtx, nil).BatchIssue(&types.BatchIssueRequest{
		SessionUuid: sessionUuid, TemplateUuid: tpl.Uuid,
	})
	if err != nil {
		t.Fatalf("BatchIssue #2: %v", err)
	}
	if resp2.IssuedCount != 0 || resp2.SkippedCount != 2 {
		t.Fatalf("second issue = issued %d skipped %d, want 0/2", resp2.IssuedCount, resp2.SkippedCount)
	}
}

func TestBatchIssueRejectedOnDisabledTemplate(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	tpl := seedTemplate(t, env, "TD", "2") // 停用
	u1 := seedUser(t, env, "user-a")
	sessionUuid := seedSessionWithParticipants(t, env, u1)

	if _, err := NewBatchIssue(context.Background(), svcCtx, nil).BatchIssue(&types.BatchIssueRequest{
		SessionUuid: sessionUuid, TemplateUuid: tpl.Uuid,
	}); err == nil {
		t.Fatal("batch issue on disabled template should be rejected (模板停用后不可发证)")
	}
}

func TestIssueAndRevokeFlow(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	tpl := seedTemplate(t, env, "TA", "1")
	u1 := seedUser(t, env, "user-a")
	sessionUuid := seedSessionWithParticipants(t, env, u1)

	// 批量发证后撤销（撤销路径 acceptance criterion）
	if _, err := NewBatchIssue(context.Background(), svcCtx, nil).BatchIssue(&types.BatchIssueRequest{
		SessionUuid: sessionUuid, TemplateUuid: tpl.Uuid,
	}); err != nil {
		t.Fatalf("BatchIssue: %v", err)
	}
	list, _ := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, SessionUuid: sessionUuid,
	})
	if len(list.Records) != 1 || list.Records[0].Status != "1" {
		t.Fatalf("expected 1 issued record, got %+v", list.Records)
	}
	certUuid := list.Records[0].Uuid

	// 已发 → 已撤销
	if _, err := NewRevoke(context.Background(), svcCtx, nil).Revoke(&types.RevokeRequest{Uuids: []string{certUuid}}); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if row, err := env.Model.HrmTrainingCertificate.FindOneByUuid(context.Background(), nil, certUuid); err != nil || row.Status != "2" {
		t.Fatalf("status after revoke = %+v err=%v, want 2 (已撤销)", row, err)
	}

	// 已撤销为终态：再撤销拒绝
	if _, err := NewRevoke(context.Background(), svcCtx, nil).Revoke(&types.RevokeRequest{Uuids: []string{certUuid}}); err == nil {
		t.Fatal("revoke on 已撤销 should fail")
	}

	// 撤销后该学员可重新发证（幂等排除已撤销）
	resp, err := NewBatchIssue(context.Background(), svcCtx, nil).BatchIssue(&types.BatchIssueRequest{
		SessionUuid: sessionUuid, TemplateUuid: tpl.Uuid,
	})
	if err != nil {
		t.Fatalf("BatchIssue after revoke: %v", err)
	}
	if resp.IssuedCount != 1 || resp.SkippedCount != 0 {
		t.Fatalf("re-issue after revoke = issued %d skipped %d, want 1/0", resp.IssuedCount, resp.SkippedCount)
	}

	// 待发 → 已发（Issue 行操作）：外部造一张待发证
	pendingUuid := uuid.New().String()
	if _, err := env.Model.HrmTrainingCertificate.Insert(context.Background(), nil, &hrm_training_certificate.HrmTrainingCertificate{
		Uuid: pendingUuid, CertNo: "TA-2026-900001", CertName: "安全培训合格证", CertType: "安全培训证",
		Source: CertSourceSession, SessionUuid: sessionUuid, UserUuid: u1, Status: CertStatusPending,
	}); err != nil {
		t.Fatalf("seed pending cert: %v", err)
	}
	if _, err := NewIssue(context.Background(), svcCtx, nil).Issue(&types.IssueRequest{Uuids: []string{pendingUuid}}); err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if row, _ := env.Model.HrmTrainingCertificate.FindOneByUuid(context.Background(), nil, pendingUuid); row.Status != "1" || !row.IssueTime.Valid {
		t.Fatalf("status after issue = %+v, want 1 with issue_time", row)
	}
	// 已发再发放拒绝
	if _, err := NewIssue(context.Background(), svcCtx, nil).Issue(&types.IssueRequest{Uuids: []string{pendingUuid}}); err == nil {
		t.Fatal("issue on 已发 should fail")
	}
}

func TestImportExternalCertificate(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	u1 := seedUser(t, env, "user-a")

	// 正常导入（带扫描件）: source='2', 导入即发
	if _, err := NewImportExternal(context.Background(), svcCtx, nil).ImportExternal(&types.ImportExternalRequest{
		UserUuid: u1, CertName: "外部结业证书", CertType: "结业证", CertNo: "EXT-2026-0001",
		FileName: "外部结业证书.pdf", FileContent: "JVBERi0xLjQ=",
	}); err != nil {
		t.Fatalf("ImportExternal: %v", err)
	}
	list, err := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, Source: "2",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(list.Records))
	}
	rec := list.Records[0]
	if rec.Source != "2" || rec.Status != "1" || rec.SessionUuid != "" || rec.TemplateUuid != "" || !rec.HasFile {
		t.Fatalf("imported record mismatch: %+v", rec)
	}
	if rec.Username != "user-a" || rec.NickName != "user-a姓名" {
		t.Fatalf("user enrichment mismatch: %+v", rec)
	}

	// 证书号唯一：同号再导入被拒
	if _, err := NewImportExternal(context.Background(), svcCtx, nil).ImportExternal(&types.ImportExternalRequest{
		UserUuid: u1, CertName: "重复证书", CertType: "结业证", CertNo: "EXT-2026-0001",
	}); err == nil {
		t.Fatal("duplicate external certNo should be rejected")
	}

	// 参数校验
	for _, req := range []types.ImportExternalRequest{
		{CertName: "X", CertNo: "N"},  // 缺学员
		{UserUuid: u1, CertNo: "N"},   // 缺名称
		{UserUuid: u1, CertName: "X"}, // 缺证书号
	} {
		if _, err := NewImportExternal(context.Background(), svcCtx, nil).ImportExternal(&req); err == nil {
			t.Fatalf("expected validation error for %+v", req)
		}
	}
}

func TestUploadAndDownloadFile(t *testing.T) {
	env := svctest.New(t)
	svcCtx := testSvcCtx(t, env)
	tpl := seedTemplate(t, env, "TS", "1")
	u1 := seedUser(t, env, "user-a")
	sessionUuid := seedSessionWithParticipants(t, env, u1)

	if _, err := NewBatchIssue(context.Background(), svcCtx, nil).BatchIssue(&types.BatchIssueRequest{
		SessionUuid: sessionUuid, TemplateUuid: tpl.Uuid,
	}); err != nil {
		t.Fatalf("BatchIssue: %v", err)
	}
	list, _ := NewList(context.Background(), svcCtx, nil).List(&types.ListRequest{
		PageRequest: types.PageRequest{Current: 1, Size: 10}, SessionUuid: sessionUuid,
	})
	certUuid := list.Records[0].Uuid

	// 上传扫描件
	if _, err := NewUploadFile(context.Background(), svcCtx, nil).UploadFile(&types.UploadFileRequest{
		Uuid: certUuid, FileName: "结业证书.pdf", FileContent: "JVBERi0xLjQ=",
	}); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	// 下载/预览
	dl, err := NewDownloadFile(context.Background(), svcCtx, nil).DownloadFile(&types.DownloadFileRequest{Uuid: certUuid})
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	if dl.FileName != "结业证书.pdf" || dl.Content != "JVBERi0xLjQ=" {
		t.Fatalf("download mismatch: %+v", dl)
	}

	// 无扫描件的证书 → 下载报错
	pendingUuid := uuid.New().String()
	if _, err := env.Model.HrmTrainingCertificate.Insert(context.Background(), nil, &hrm_training_certificate.HrmTrainingCertificate{
		Uuid: pendingUuid, CertNo: "TS-2026-900002", CertName: "结业证书", Source: CertSourceSession,
		SessionUuid: sessionUuid, UserUuid: u1, Status: CertStatusPending,
	}); err != nil {
		t.Fatalf("seed pending: %v", err)
	}
	if _, err := NewDownloadFile(context.Background(), svcCtx, nil).DownloadFile(&types.DownloadFileRequest{Uuid: pendingUuid}); err == nil {
		t.Fatal("download without file should fail")
	}
}
