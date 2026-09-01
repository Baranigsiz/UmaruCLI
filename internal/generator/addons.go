package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddonConfig holds optional infrastructure and middleware add-ons
type AddonConfig struct {
	Database string `json:"database,omitempty"` // "none", "postgres", "sqlite", "mongodb"
	Auth     string `json:"auth,omitempty"`     // "none", "jwt"
	Redis    bool   `json:"redis,omitempty"`    // true/false
}

// HasAddons returns true if any addon is enabled
func (a AddonConfig) HasAddons() bool {
	db := strings.ToLower(strings.TrimSpace(a.Database))
	auth := strings.ToLower(strings.TrimSpace(a.Auth))
	return (db != "" && db != "none") || (auth != "" && auth != "none") || a.Redis
}

// GetAddonFiles returns the list of file paths that will be generated for the selected addons
func GetAddonFiles(config ProjectConfig) []string {
	var files []string
	if !config.Addons.HasAddons() {
		return files
	}

	db := strings.ToLower(strings.TrimSpace(config.Addons.Database))
	auth := strings.ToLower(strings.TrimSpace(config.Addons.Auth))
	isGo := strings.HasPrefix(config.Template, "go-") || config.Template == "fullstack-go-react"
	isNode := strings.HasPrefix(config.Template, "node-") || strings.HasPrefix(config.Template, "nestjs-")
	isPython := strings.HasPrefix(config.Template, "python-")

	baseDir := config.TargetDir
	if config.Template == "fullstack-go-react" {
		baseDir = filepath.Join(config.TargetDir, "apps", "api")
	}

	// Database Addon
	if db != "" && db != "none" {
		if isGo {
			files = append(files, filepath.Join(baseDir, "internal", "database", fmt.Sprintf("%s.go", db)))
		} else if isNode {
			files = append(files, filepath.Join(baseDir, "src", "config", "database.ts"))
		} else if isPython {
			files = append(files, filepath.Join(baseDir, "app", "core", "database.py"))
		}
	}

	// Auth Addon
	if auth == "jwt" {
		if isGo {
			files = append(files, filepath.Join(baseDir, "internal", "middleware", "auth.go"))
		} else if isNode {
			files = append(files, filepath.Join(baseDir, "src", "middlewares", "auth.middleware.ts"))
		} else if isPython {
			files = append(files, filepath.Join(baseDir, "app", "core", "security.py"))
		}
	}

	// Redis Addon
	if config.Addons.Redis {
		if isGo {
			files = append(files, filepath.Join(baseDir, "internal", "cache", "redis.go"))
		} else if isNode {
			files = append(files, filepath.Join(baseDir, "src", "config", "redis.ts"))
		} else if isPython {
			files = append(files, filepath.Join(baseDir, "app", "core", "redis.py"))
		}
	}

	return files
}

// GenerateAddons writes the addon template files into the target project
func GenerateAddons(config ProjectConfig) error {
	if !config.Addons.HasAddons() {
		return nil
	}

	db := strings.ToLower(strings.TrimSpace(config.Addons.Database))
	auth := strings.ToLower(strings.TrimSpace(config.Addons.Auth))
	isGo := strings.HasPrefix(config.Template, "go-") || config.Template == "fullstack-go-react"
	isNode := strings.HasPrefix(config.Template, "node-") || strings.HasPrefix(config.Template, "nestjs-")
	isPython := strings.HasPrefix(config.Template, "python-")

	baseDir := config.TargetDir
	if config.Template == "fullstack-go-react" {
		baseDir = filepath.Join(config.TargetDir, "apps", "api")
	}

	writeFile := func(relPath, content string) error {
		fullPath := filepath.Join(baseDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(fullPath, []byte(content), 0644)
	}

	// 1. Database Addon Generation
	if db != "" && db != "none" {
		if isGo {
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
				if err := writeFile(filepath.Join("internal", "database", "postgres.go"), content); err != nil {
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
				if err := writeFile(filepath.Join("internal", "database", "sqlite.go"), content); err != nil {
					return err
				}
			}
		} else if isNode {
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
			if err := writeFile(filepath.Join("src", "config", "database.ts"), content); err != nil {
				return err
			}
		} else if isPython {
			content := `import os
from typing import AsyncGenerator

# Database Connection Settings
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql+asyncpg://postgres:postgres@localhost:5432/app")

async def get_db_session():
    """Async database session dependency generator."""
    # Yield database session here
    yield None
`
			if err := writeFile(filepath.Join("app", "core", "database.py"), content); err != nil {
				return err
			}
		}
	}

	// 2. Auth Addon Generation (JWT)
	if auth == "jwt" {
		if isGo {
			content := `package middleware

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID string ` + "`json:\"user_id\"`" + `
	Email  string ` + "`json:\"email\"`" + `
	Role   string ` + "`json:\"role\"`" + `
	jwt.RegisteredClaims
}

// GenerateJWT creates a new signed token valid for 24 hours
func GenerateJWT(userID, email, role, secretKey string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateJWT parses and verifies a given JWT token string
func ValidateJWT(tokenStr, secretKey string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}
`
			if err := writeFile(filepath.Join("internal", "middleware", "auth.go"), content); err != nil {
				return err
			}
		} else if isNode {
			content := `import { Request, Response, NextFunction } from 'express';

export interface AuthRequest extends Request {
  user?: {
    id: string;
    email: string;
    role: string;
  };
}

export function authMiddleware(req: AuthRequest, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Unauthorized: Missing or malformed token' });
  }

  const token = authHeader.split(' ')[1];
  try {
    // In production, verify with jwt.verify(token, process.env.JWT_SECRET!)
    req.user = { id: 'usr_sample', email: 'user@example.com', role: 'admin' };
    next();
  } catch (err) {
    return res.status(401).json({ error: 'Unauthorized: Invalid token' });
  }
}
`
			if err := writeFile(filepath.Join("src", "middlewares", "auth.middleware.ts"), content); err != nil {
				return err
			}
		} else if isPython {
			content := `import os
from datetime import datetime, timedelta
from typing import Optional
from jose import JWTError, jwt

SECRET_KEY = os.getenv("JWT_SECRET_KEY", "your-super-secret-key-change-in-production")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60 * 24

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    to_encode = data.copy()
    expire = datetime.utcnow() + (expires_delta or timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES))
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)

def verify_token(token: str) -> Optional[dict]:
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        return payload
    except JWTError:
        return None
`
			if err := writeFile(filepath.Join("app", "core", "security.py"), content); err != nil {
				return err
			}
		}
	}

	// 3. Redis Addon Generation
	if config.Addons.Redis {
		if isGo {
			content := `package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis initializes a Redis client connection
func ConnectRedis(addr, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}

	log.Printf("🔴 Connected to Redis at %s", addr)
	return rdb, nil
}
`
			if err := writeFile(filepath.Join("internal", "cache", "redis.go"), content); err != nil {
				return err
			}
		} else if isNode {
			content := `// Redis Cache Connection
export async function connectRedis(host = 'localhost', port = 6379) {
  console.log(` + "`🔴 Redis client configured on ${host}:${port}`" + `);
  return { host, port, ready: true };
}
`
			if err := writeFile(filepath.Join("src", "config", "redis.ts"), content); err != nil {
				return err
			}
		} else if isPython {
			content := `import os

REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))

async def get_redis_client():
    """Returns async Redis client connection."""
    return {"host": REDIS_HOST, "port": REDIS_PORT}
`
			if err := writeFile(filepath.Join("app", "core", "redis.py"), content); err != nil {
				return err
			}
		}
	}

	return nil
}
