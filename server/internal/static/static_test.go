package static

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("index.html", "<html>index</html>")
	mustWrite("assets/app.js", "console.log(1)")
	return dir
}

func serve(t *testing.T, dir, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	h := Handler(dir)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestHandlerNilCases(t *testing.T) {
	if Handler("") != nil {
		t.Fatal("empty dir must return nil")
	}
	if Handler(t.TempDir()) != nil {
		t.Fatal("dir without index.html must return nil")
	}
}

func TestServeIndexAndAssets(t *testing.T) {
	dir := newTestDir(t)
	if rec := serve(t, dir, http.MethodGet, "/"); rec.Body.String() != "<html>index</html>" {
		t.Fatalf("root should serve index, got %q", rec.Body.String())
	}
	if rec := serve(t, dir, http.MethodGet, "/assets/app.js"); rec.Code != http.StatusOK {
		t.Fatalf("asset should serve 200, got %d", rec.Code)
	}
}

func TestSpaFallback(t *testing.T) {
	dir := newTestDir(t)
	if rec := serve(t, dir, http.MethodGet, "/some/deep/route"); rec.Code != http.StatusOK || rec.Body.String() != "<html>index</html>" {
		t.Fatalf("deep link should fall back to index 200, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiPrefixStays404(t *testing.T) {
	dir := newTestDir(t)
	if rec := serve(t, dir, http.MethodGet, "/api/v1/unknown"); rec.Code != http.StatusNotFound {
		t.Fatalf("/api path must stay 404, got %d", rec.Code)
	}
}

func TestNonGetRejected(t *testing.T) {
	dir := newTestDir(t)
	if rec := serve(t, dir, http.MethodPost, "/"); rec.Code != http.StatusNotFound {
		t.Fatalf("POST must 404, got %d", rec.Code)
	}
}
