// Package svctest provides DB-backed test environments for logic-layer tests.
//
// Two backends are supported:
//
//   - MySQL (primary seam): starts a disposable Docker MySQL container, applies
//     the real mysql migrations from desc/sql_migration/mysql, and returns a
//     ServiceContext-ready model.Model. Requires a local docker binary and a
//     mysql:8.0 image (any mirror tag works; see mysqlImage).
//   - SQLite (fallback): applies desc/sql_migration/sqlite migrations against a
//     temp-file sqlite database. Works offline, no docker needed.
//
// Selection: HRM_TEST_DB=sqlite|mysql (default "sqlite"). CI and quick local
// runs use sqlite; set HRM_TEST_DB=mysql to exercise the docker harness.
package svctest

import (
	stdsql "database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	// modernc.org/sqlite registers the "sqlite" database driver used by the
	// golang-migrate sqlite database driver.
	_ "modernc.org/sqlite"

	"github.com/jzero-io/jzero/core/stores/migrate"
	"github.com/jzero-io/jzero/core/stores/modelx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/jzero-io/jzero-admin/server/internal/model"
)

const (
	mysqlRootPassword = "root"
	mysqlDatabase     = "jzero_admin_test"
	// mysqlImage is the image started by the docker harness. Override with
	// HRM_TEST_MYSQL_IMAGE when the docker hub image is not locally available
	// (e.g. a registry mirror tag).
	mysqlImageEnv = "HRM_TEST_MYSQL_IMAGE"
)

var mysqlImage = "mysql:8.0"

// Env is a database-backed test environment. It carries the raw connection
// plus the aggregate model so tests can construct logic structs with a
// minimal *svc.ServiceContext.
type Env struct {
	SqlConn sqlx.SqlConn
	Model   model.Model
	Dsn     string

	cleanup func()
}

// Close releases the environment (container, temp files). Called automatically
// via t.Cleanup when created through New/NewMySQL/NewSqlite.
func (e *Env) Close() {
	if e.cleanup != nil {
		e.cleanup()
	}
}

// New returns the environment selected by HRM_TEST_DB (default sqlite).
func New(t *testing.T) *Env {
	t.Helper()
	switch strings.ToLower(os.Getenv("HRM_TEST_DB")) {
	case "mysql":
		return NewMySQL(t)
	default:
		return NewSqlite(t)
	}
}

// migrationDir walks up from the test working directory until it finds the
// server module root (identified by the desc/sql_migration directory).
func migrationDir(t *testing.T, driver string) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		candidate := filepath.Join(dir, "desc", "sql_migration", driver)
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate desc/sql_migration/%s above %s", driver, os.Getenv("PWD"))
		}
		dir = parent
	}
}

func applyMigrations(t *testing.T, driver, dsn string) {
	t.Helper()
	conf := sqlx.SqlConf{DataSource: dsn, DriverName: driver}
	source := "file://" + migrationDir(t, driver)
	m, err := migrate.NewMigrate(conf, migrate.WithSource(source))
	if err != nil {
		t.Fatalf("new migrate: %v", err)
	}
	if err = m.Up(); err != nil {
		t.Fatalf("apply %s migrations: %v", driver, err)
	}
	if err1, err2 := m.Close(); err1 != nil || err2 != nil {
		t.Logf("migrate close: %v %v", err1, err2)
	}
}

// NewSqlite applies the sqlite migrations to a fresh temp-file database.
func NewSqlite(t *testing.T) *Env {
	t.Helper()

	dbFile := filepath.Join(t.TempDir(), "hrm_test.db")
	dsn := "file:" + dbFile + "?_pragma=busy_timeout(5000)&_txlock=immediate"

	applyMigrations(t, "sqlite", dsn)

	conn := modelx.MustNewConn(sqlx.SqlConf{DataSource: dsn, DriverName: "sqlite"})
	env := &Env{SqlConn: conn, Model: model.NewModel(conn), Dsn: dsn}
	t.Cleanup(env.Close)
	return env
}

// NewMySQL starts a disposable docker MySQL container, applies the real mysql
// migrations and returns the environment. Skips the test when docker is not
// usable so sqlite-only machines can still run `go test ./...`.
func NewMySQL(t *testing.T) *Env {
	t.Helper()

	if img := os.Getenv(mysqlImageEnv); img != "" {
		mysqlImage = img
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not found in PATH; skip mysql logic test (use HRM_TEST_DB=sqlite)")
	}

	name := fmt.Sprintf("jzero-admin-hrm-test-%d", time.Now().UnixNano())
	cid := runDocker(t,
		"run", "-d", "--rm", "--name", name,
		"-e", "MYSQL_ROOT_PASSWORD="+mysqlRootPassword,
		"-e", "MYSQL_DATABASE="+mysqlDatabase,
		"-p", "127.0.0.1::3306",
		mysqlImage,
	)
	t.Cleanup(func() {
		_ = exec.Command("docker", "rm", "-f", name).Run()
	})

	port := dockerPort(t, cid, "3306")

	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlRootPassword, port, mysqlDatabase)

	waitForMySQL(t, dsn)

	applyMigrations(t, "mysql", dsn)

	// plain mysql conn; sqlbuilder default flavor is MySQL already
	conn := sqlx.NewMysql(dsn)
	env := &Env{SqlConn: conn, Model: model.NewModel(conn), Dsn: dsn}
	env.cleanup = func() { _ = exec.Command("docker", "rm", "-f", name).Run() }
	t.Cleanup(env.Close)
	return env
}

func runDocker(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		t.Skipf("docker %v failed: %v\n%s (is a mysql image available locally? set %s)",
			args, err, out, mysqlImageEnv)
	}
	return strings.TrimSpace(string(out))
}

func dockerPort(t *testing.T, cid, containerPort string) string {
	t.Helper()
	out, err := exec.Command("docker", "port", cid, containerPort).CombinedOutput()
	if err != nil {
		t.Fatalf("docker port: %v\n%s", err, out)
	}
	// "127.0.0.1:55432" or "0.0.0.0:55432" (possibly multiple lines)
	for _, line := range strings.Fields(string(out)) {
		if host, p, err := net.SplitHostPort(line); err == nil {
			_ = host
			return p
		}
	}
	t.Fatalf("cannot parse docker port output: %s", out)
	return ""
}

func waitForMySQL(t *testing.T, dsn string) {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for {
		// poll with a plain database/sql handle: go-zero pools connections by
		// DSN, so closing a pooled handle here would break the later conn
		db, err := stdsql.Open("mysql", dsn)
		if err == nil {
			pingErr := db.Ping()
			_ = db.Close()
			if pingErr == nil {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("mysql container not ready within deadline")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
