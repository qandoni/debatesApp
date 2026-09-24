// Package integration содержит интеграционные тесты, покрывающие места,
// которые юнит-тесты на моках защитить не могут: реальные SQL-ограничения,
// транзакции, оптимистичные блокировки и гонки между конкурентными запросами.
//
// Тесты требуют запущенный Postgres (docker compose: make env-up && make env-port-forward).
// Параметры подключения берутся из переменных окружения TEST_POSTGRES_*,
// при их отсутствии — из POSTGRES_* (как в .env), иначе из дефолтов локальной среды.
//
// Запуск: make test-integration
package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	core_pgx_pool "github.com/qandoni/debatesApp/internal/core/repository/postgres/pool/pgx"
)

var (
	itPool      *core_pgx_pool.Pool
	itTxManager *core_pgx_pool.TransactionManager
	itTimeout   = 5 * time.Second
	itDBName    string
)

type postgresEnv struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// loadPostgresEnv читает конфигурацию тестовой БД.
// Приоритет: TEST_POSTGRES_* -> POSTGRES_* -> дефолты локальной среды разработки.
func loadPostgresEnv() postgresEnv {
	get := func(testKey, appKey, fallback string) string {
		if v := os.Getenv(testKey); v != "" {
			return v
		}
		if v := os.Getenv(appKey); v != "" {
			return v
		}
		return fallback
	}

	return postgresEnv{
		Host:     get("TEST_POSTGRES_HOST", "POSTGRES_HOST", "localhost"),
		Port:     get("TEST_POSTGRES_PORT", "POSTGRES_PORT", "5432"),
		User:     get("TEST_POSTGRES_USER", "POSTGRES_USER", "localhost"),
		Password: get("TEST_POSTGRES_PASSWORD", "POSTGRES_PASSWORD", "ywerdsjfnv23"),
		Database: get("TEST_POSTGRES_DB", "POSTGRES_DB", "testDB"),
	}
}

func (e postgresEnv) dsn(database string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		e.User, e.Password, e.Host, e.Port, database,
	)
}

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

func runMain(m *testing.M) int {
	env := loadPostgresEnv()
	ctx := context.Background()

	// Уникальная тестовая БД на прогон: полная изоляция от боевой/дев-БД
	// и от параллельных прогонов.
	itDBName = fmt.Sprintf("debates_it_%d", time.Now().UnixNano()%1_000_000_000)

	admin, err := connectRaw(ctx, env.dsn(env.Database))
	if err != nil {
		fmt.Printf("[integration] SKIP: postgres недоступен (%s:%s/%s): %v\n", env.Host, env.Port, env.Database, err)
		return 0
	}

	if _, err := admin.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %q`, itDBName)); err != nil {
		fmt.Printf("[integration] SKIP: не удалось создать тестовую БД %s: %v\n", itDBName, err)
		_ = admin.Close(ctx)
		return 0
	}

	code := 0
	func() {
		defer func() {
			// Гасим оставшиеся соединения и удаляем тестовую БД.
			dropCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_, _ = admin.Exec(dropCtx,
				`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`,
				itDBName)
			_, _ = admin.Exec(dropCtx, fmt.Sprintf(`DROP DATABASE IF EXISTS %q`, itDBName))
			_ = admin.Close(dropCtx)
		}()

		cfg := core_pgx_pool.Config{
			Host:     env.Host,
			Port:     env.Port,
			User:     env.User,
			Password: env.Password,
			Database: itDBName,
			Timeout:  itTimeout,
		}

		itPool, err = core_pgx_pool.NewPool(ctx, cfg)
		if err != nil {
			fmt.Printf("[integration] FAIL: подключение к тестовой БД %s: %v\n", itDBName, err)
			code = 1
			return
		}
		defer itPool.Close()

		if err := applyMigrations(ctx, itPool); err != nil {
			fmt.Printf("[integration] FAIL: применение миграций: %v\n", err)
			code = 1
			return
		}

		itTxManager = core_pgx_pool.NewTransactionManager(itPool)
		wireDependencies()

		code = m.Run()
	}()

	return code
}

// applyMigrations применяет *.up.sql из каталога migrations в порядке нумерации.
func applyMigrations(ctx context.Context, pool *core_pgx_pool.Pool) error {
	dir := os.Getenv("TEST_MIGRATIONS_DIR")
	if dir == "" {
		dir = "../../migrations"
	}

	entries, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("миграции не найдены в %s (cwd=%s)", dir, mustCWD())
	}
	sort.Strings(entries)

	for _, path := range entries {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply %s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

func mustCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "?"
	}
	return cwd
}

// connectRaw — прямое pgx-соединение для административных операций
// (CREATE/DROP DATABASE), минуя пул приложения.
func connectRaw(ctx context.Context, dsn string) (*pgxConn, error) {
	return newPgxConn(ctx, dsn)
}
