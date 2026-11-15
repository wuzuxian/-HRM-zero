package training

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_certificate"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session"
	"github.com/jzero-io/jzero-admin/server/internal/model/hrm_training_session_participant"
	"github.com/jzero-io/jzero-admin/server/internal/svc"
	"github.com/jzero-io/jzero-admin/server/internal/svctest"
	"github.com/jzero-io/jzero-admin/server/internal/errcodes"
	types "github.com/jzero-io/jzero-admin/server/internal/types/v1/hrm/training"
	"github.com/jzero-io/jzero/core/status"
)

// userCtx 模拟当前登录用户（JWT 载荷 seam：helper/auth.Info 读取 ctx "uuid"）。
func userCtx(userUuid string) context.Context {
	return context.WithValue(context.Background(), "uuid", userUuid)
}

func newEnv(t *testing.T) *svc.ServiceContext {
	t.Helper()
	env := svctest.New(t)
	return &svc.ServiceContext{Model: env.Model}
}

type fixture struct {
	session1 string // 入职培训，admin 学员
	session2 string // 安全培训，admin 学员
	session3 string // 管理层培训，other 学员
	cert1    string // admin 的已发证书（带扫描件）
	cert2    string // admin 的待发证书（无扫描件）
	cert3    string // other 的证书
}

func seed(t *testing.T, svcCtx *svc.ServiceContext) fixture {
	t.Helper()
	ctx := context.Background()
	f := fixture{}

	const admin = "1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f"
	const other = "11111111-2222-4333-8444-555555555555"
	mustTime := func(s string) time.Time {
		tv, err := time.Parse(time.DateTime, s)
		if err != nil {
			t.Fatalf("parse time %q: %v", s, err)
		}
		return tv
	}

	sessions := []hrm_training_session.HrmTrainingSession{
		{Uuid: "s-1", SessionNo: "PX-TEST-000001", SessionName: "入职培训(测试)", Lecturer: "王老师", Location: "总部一楼培训室", Hours: 14, StartTime: mustTime("2026-11-20 09:00:00"), EndTime: mustTime("2026-11-21 17:00:00"), Status: "1"},
		{Uuid: "s-2", SessionNo: "PX-TEST-000002", SessionName: "安全生产培训", Lecturer: "李老师", Location: "二楼会议室", Hours: 6, StartTime: mustTime("2026-10-10 09:00:00"), EndTime: mustTime("2026-10-10 17:00:00"), Status: "3"},
		{Uuid: "s-3", SessionNo: "PX-TEST-000003", SessionName: "管理层培训", Lecturer: "赵老师", Location: "总部五楼", Hours: 4, StartTime: mustTime("2026-11-25 09:00:00"), EndTime: mustTime("2026-11-25 13:00:00"), Status: "1"},
	}
	for i := range sessions {
		if _, err := svcCtx.Model.HrmTrainingSession.Insert(ctx, nil, &sessions[i]); err != nil {
			t.Fatalf("seed session %d: %v", i, err)
		}
	}
	f.session1, f.session2, f.session3 = sessions[0].Uuid, sessions[1].Uuid, sessions[2].Uuid

	participants := []hrm_training_session_participant.HrmTrainingSessionParticipant{
		{Uuid: "p-1", SessionUuid: f.session1, UserUuid: admin},
		{Uuid: "p-2", SessionUuid: f.session2, UserUuid: admin},
		{Uuid: "p-3", SessionUuid: f.session3, UserUuid: other},
	}
	for i := range participants {
		if _, err := svcCtx.Model.HrmTrainingSessionParticipant.Insert(ctx, nil, &participants[i]); err != nil {
			t.Fatalf("seed participant %d: %v", i, err)
		}
	}

	// 最小 PDF 的 base64（演示扫描件）
	pdfContent := base64.StdEncoding.EncodeToString([]byte("%PDF-1.4 demo scan"))
	certs := []hrm_training_certificate.HrmTrainingCertificate{
		{Uuid: "c-1", CertNo: "JN-TEST-000001", CertName: "结业证书", CertType: "结业证", Source: "1", SessionUuid: f.session2, UserUuid: admin, IssueTime: sql.NullTime{Time: mustTime("2026-10-12 10:00:00"), Valid: true}, Status: "1", FileName: "结业证书.pdf", FileContent: sql.NullString{String: pdfContent, Valid: true}},
		{Uuid: "c-2", CertNo: "AQ-TEST-000001", CertName: "安全培训合格证", CertType: "安全培训证", Source: "1", SessionUuid: f.session2, UserUuid: admin, Status: "0"},
		{Uuid: "c-3", CertNo: "JN-TEST-000002", CertName: "他人结业证", Source: "1", SessionUuid: f.session3, UserUuid: other, Status: "1", FileName: "other.pdf", FileContent: sql.NullString{String: pdfContent, Valid: true}},
	}
	for i := range certs {
		if _, err := svcCtx.Model.HrmTrainingCertificate.Insert(ctx, nil, &certs[i]); err != nil {
			t.Fatalf("seed certificate %d: %v", i, err)
		}
	}
	f.cert1, f.cert2, f.cert3 = certs[0].Uuid, certs[1].Uuid, certs[2].Uuid

	return f
}

// TestMySessionListScopedToParticipant 覆盖 User Story 52 的一半：
// 我的班次只显示当前登录用户作为学员（hrm_training_session_participant）的班次，
// 且可按状态筛选、分页。
func TestMySessionListScopedToParticipant(t *testing.T) {
	svcCtx := newEnv(t)
	f := seed(t, svcCtx)

	// admin 是 s-1/s-2 的学员（迁移演示数据还给 admin 预置了 2 个班次，
	// 所以这里做相对断言：测试班次必须在、别人的班次必须不在）
	resp, err := NewGetMySessionList(userCtx("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f"), svcCtx, nil).GetMySessionList(&types.GetMySessionListRequest{PageRequest: types.PageRequest{Current: 1, Size: 100}})
	if err != nil {
		t.Fatalf("get my sessions: %v", err)
	}
	if resp.Total < 2 || len(resp.Records) < 2 {
		t.Fatalf("expected at least 2 sessions for admin, got total=%d records=%d", resp.Total, len(resp.Records))
	}
	names := map[string]types.HrmMyTrainingSession{}
	for _, r := range resp.Records {
		names[r.SessionName] = r
	}
	s1, ok := names["入职培训(测试)"]
	if !ok {
		t.Fatalf("入职培训(测试) not in admin's sessions: %+v", resp.Records)
	}
	if s1.SessionNo != "PX-TEST-000001" || s1.Location != "总部一楼培训室" || s1.StartTime == "" {
		t.Fatalf("session fields incomplete: %+v", s1)
	}
	if _, ok := names["管理层培训"]; ok {
		t.Fatalf("别人的班次不能出现在我的班次里")
	}

	// other 只看到 s-3（迁移演示数据只绑定了 admin，所以这里可做绝对断言）
	respOther, err := NewGetMySessionList(userCtx("11111111-2222-4333-8444-555555555555"), svcCtx, nil).GetMySessionList(&types.GetMySessionListRequest{PageRequest: types.PageRequest{Current: 1, Size: 100}})
	if err != nil {
		t.Fatalf("get other's sessions: %v", err)
	}
	if respOther.Total != 1 || len(respOther.Records) != 1 || respOther.Records[0].SessionName != "管理层培训" {
		t.Fatalf("other user should only see 管理层培训: %+v", respOther.Records)
	}

	// 状态筛选：所有返回记录状态都是 '1' 且含 s-1
	respStarted, err := NewGetMySessionList(userCtx("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f"), svcCtx, nil).GetMySessionList(&types.GetMySessionListRequest{PageRequest: types.PageRequest{Current: 1, Size: 100}, Status: "1"})
	if err != nil {
		t.Fatalf("get my sessions by status: %v", err)
	}
	startedNames := map[string]bool{}
	for _, r := range respStarted.Records {
		startedNames[r.SessionName] = true
	}
	if !startedNames["入职培训(测试)"] {
		t.Fatalf("status filter lost s-1: %+v", respStarted.Records)
	}
	for _, r := range respStarted.Records {
		if r.Status != "1" {
			t.Fatalf("status filter returned status %q: %+v", r.Status, r)
		}
	}

	// 分页：size=1 时每页一条，total 与全量一致
	respPage, err := NewGetMySessionList(userCtx("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f"), svcCtx, nil).GetMySessionList(&types.GetMySessionListRequest{PageRequest: types.PageRequest{Current: 2, Size: 1}})
	if err != nil {
		t.Fatalf("get my sessions page 2: %v", err)
	}
	if respPage.Total != resp.Total || len(respPage.Records) != 1 {
		t.Fatalf("pagination failed: total=%d len=%d want total=%d", respPage.Total, len(respPage.Records), resp.Total)
	}

	// 未登录上下文 → 空列表（不做越权暴露）
	respAnon, err := NewGetMySessionList(context.Background(), svcCtx, nil).GetMySessionList(&types.GetMySessionListRequest{PageRequest: types.PageRequest{Current: 1, Size: 100}})
	if err != nil {
		t.Fatalf("anonymous list should not error: %v", err)
	}
	if respAnon.Total != 0 || len(respAnon.Records) != 0 {
		t.Fatalf("anonymous must see nothing, got total=%d", respAnon.Total)
	}

	_ = f
}

// TestMyCertificateListAndDownload 覆盖 User Story 52 的另一半：
// 我的培训证书只显示本人证书；扫描件可预览/下载（base64）；
// 非本人证书与无扫描件分别报 10360/10361。
func TestMyCertificateListAndDownload(t *testing.T) {
	svcCtx := newEnv(t)
	f := seed(t, svcCtx)
	admin := userCtx("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f")

	// 列表：admin 只看到自己的证书（含迁移演示预置的 2 张，做相对断言）
	resp, err := NewGetMyCertificateList(admin, svcCtx, nil).GetMyCertificateList(&types.GetMyCertificateListRequest{PageRequest: types.PageRequest{Current: 1, Size: 100}})
	if err != nil {
		t.Fatalf("get my certificates: %v", err)
	}
	if resp.Total < 2 || len(resp.Records) < 2 {
		t.Fatalf("expected at least 2 certificates for admin, got total=%d records=%d", resp.Total, len(resp.Records))
	}
	byUuid := map[string]types.HrmMyTrainingCertificate{}
	for _, r := range resp.Records {
		byUuid[r.Uuid] = r
		if r.Uuid == f.cert3 {
			t.Fatalf("别人的证书不能出现在我的证书里: %+v", r)
		}
	}
	var issued *types.HrmMyTrainingCertificate
	for i := range resp.Records {
		if resp.Records[i].Uuid == f.cert1 {
			issued = &resp.Records[i]
		}
	}
	if issued == nil {
		t.Fatalf("cert1 missing: %+v", resp.Records)
	}
	if issued.CertNo != "JN-TEST-000001" || issued.CertName != "结业证书" || issued.Status != "1" {
		t.Fatalf("cert1 fields wrong: %+v", issued)
	}
	if issued.IssueTime == "" {
		t.Fatalf("发证时间 missing on issued cert: %+v", issued)
	}
	if !issued.HasFile || issued.FileName != "结业证书.pdf" {
		t.Fatalf("hasFile/fileName wrong: %+v", issued)
	}
	// 待发证书：无发证时间、无扫描件
	for _, r := range resp.Records {
		if r.Uuid == f.cert2 && (r.HasFile || r.IssueTime != "" || r.Status != "0") {
			t.Fatalf("pending cert wrong: %+v", r)
		}
	}

	// 状态筛选：所有返回记录状态都是 '1' 且含 c-1
	respIssued, err := NewGetMyCertificateList(admin, svcCtx, nil).GetMyCertificateList(&types.GetMyCertificateListRequest{PageRequest: types.PageRequest{Current: 1, Size: 100}, Status: "1"})
	if err != nil {
		t.Fatalf("get issued certs: %v", err)
	}
	if _, ok := byUuid[f.cert1]; !ok {
		t.Fatalf("status filter lost cert1: %+v", respIssued.Records)
	}
	for _, r := range respIssued.Records {
		if r.Status != "1" {
			t.Fatalf("status filter returned status %q: %+v", r.Status, r)
		}
	}

	// 下载扫描件：内容是 seed 时写入的 base64
	dl, err := NewDownloadMyCertificateFile(admin, svcCtx, nil).DownloadMyCertificateFile(&types.DownloadMyCertificateFileRequest{Uuid: f.cert1})
	if err != nil {
		t.Fatalf("download cert1 file: %v", err)
	}
	if dl.FileName != "结业证书.pdf" {
		t.Fatalf("download fileName wrong: %q", dl.FileName)
	}
	if raw, decErr := base64.StdEncoding.DecodeString(dl.Content); decErr != nil || string(raw) != "%PDF-1.4 demo scan" {
		t.Fatalf("download content wrong: %v %q", decErr, dl.Content)
	}

	// 无扫描件的证书 → 10361
	_, err = NewDownloadMyCertificateFile(admin, svcCtx, nil).DownloadMyCertificateFile(&types.DownloadMyCertificateFileRequest{Uuid: f.cert2})
	if !errors.Is(err, ErrCertificateFileNotFound) {
		t.Fatalf("download cert without file should be ErrCertificateFileNotFound, got %v", err)
	}
	if st := status.FromError(err); st == nil || st.Code() != errcodes.TrainingCertificateFileNotFoundCode {
		t.Fatalf("file-not-found must map to code 10361, got %+v", st)
	}

	// 非本人证书 → 10360
	_, err = NewDownloadMyCertificateFile(admin, svcCtx, nil).DownloadMyCertificateFile(&types.DownloadMyCertificateFileRequest{Uuid: f.cert3})
	if !errors.Is(err, ErrCertificateNotFound) {
		t.Fatalf("download someone else's cert should be ErrCertificateNotFound, got %v", err)
	}
	if st := status.FromError(err); st == nil || st.Code() != errcodes.TrainingCertificateNotFoundCode {
		t.Fatalf("not-found must map to code 10360, got %+v", st)
	}

	// 不存在的证书 → 10360
	_, err = NewDownloadMyCertificateFile(admin, svcCtx, nil).DownloadMyCertificateFile(&types.DownloadMyCertificateFileRequest{Uuid: "no-such-uuid"})
	if !errors.Is(err, ErrCertificateNotFound) {
		t.Fatalf("download unknown cert should be ErrCertificateNotFound, got %v", err)
	}
}
