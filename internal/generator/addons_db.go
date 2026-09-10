package generator

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getDatabaseFiles(config ProjectConfig, baseDir string) []string {
	db := strings.ToLower(strings.TrimSpace(config.Addons.Database))
	if db == "" || db == "none" {
		return nil
	}
	if isGoTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "internal", "database", fmt.Sprintf("%s.go", db))}
	} else if isNodeTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "src", "config", "database.ts")}
	} else if isPythonTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "app", "core", "database.py")}
	}
	return nil
}

func generateDatabaseAddon(config ProjectConfig, baseDir string) error {
	db := strings.ToLower(strings.TrimSpace(config.Addons.Database))
	if db == "" || db == "none" {
		return nil
	}

	if isGoTemplate(config.Template) {
		switch db {
		case "postgres":
			content := fmt.Sprintf(`package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectPostgres establishes a thread-safe connection pool to PostgreSQL
func ConnectPostgres(cfg Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%%s port=%%d user=%%s password=%%s dbname=%%s sslmode=%%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %%w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %%w", err)
	}

	log.Println("🐘 Connected to PostgreSQL successfully")
	return db, nil
}
`)
			if err := writeAddonFile(baseDir, filepath.Join("internal", "database", "postgres.go"), content); err != nil {
				return err
			}
		case "sqlite":
			content := `package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectSQLite opens a SQLite database file with WAL mode enabled
func ConnectSQLite(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	log.Printf("📦 Connected to SQLite database at %s", dbPath)
	return db, nil
}
`
			if err := writeAddonFile(baseDir, filepath.Join("internal", "database", "sqlite.go"), content); err != nil {
				return err
			}
		}
	} else if isNodeTemplate(config.Template) {
		content := `// Database Connection Configuration
export interface DBConfig {
  host?: string;
  port?: number;
  database?: string;
  user?: string;
  password?: string;
}

export async function connectDatabase(config?: DBConfig) {
  console.log('🔌 Database module initialized');
  return { connected: true };
}
`
		if err := writeAddonFile(baseDir, filepath.Join("src", "config", "database.ts"), content); err != nil {
			return err
		}

		pkgPath := filepath.Join(baseDir, "package.json")
		nodeDeps := make(map[string]string)
		nodeDevDeps := make(map[string]string)

		if db == "postgres" {
			nodeDeps["pg"] = "^8.12.0"
			nodeDevDeps["@types/pg"] = "^8.11.6"
		} else if db == "sqlite" {
			nodeDeps["sqlite3"] = "^5.1.7"
			nodeDevDeps["@types/sqlite3"] = "^3.1.11"
		}

		if err := injectNodeDependencies(pkgPath, nodeDeps, nodeDevDeps); err != nil {
			return fmt.Errorf("failed injecting node database dependencies: %w", err)
		}
	} else if isPythonTemplate(config.Template) {
		content := `import os
from typing import AsyncGenerator

# Database Connection Settings
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql+asyncpg://postgres:postgres@localhost:5432/app")

async def get_db_session():
    """Async database session dependency generator."""
    # Yield database session here
    yield None
`
		if err := writeAddonFile(baseDir, filepath.Join("app", "core", "database.py"), content); err != nil {
			return err
		}

		if db == "postgres" {
			reqPath := filepath.Join(baseDir, "requirements.txt")
			if err := injectPythonDependencies(reqPath, []string{"asyncpg>=0.29.0"}); err != nil {
				return fmt.Errorf("failed injecting python database dependencies: %w", err)
			}
		}
	}

	if db == "postgres" && (fileExists(filepath.Join(baseDir, "docker-compose.yml")) || fileExists(filepath.Join(config.TargetDir, "docker-compose.yml"))) {
		_ = appendDockerComposeServices(baseDir, config)
	}

	return nil
}
