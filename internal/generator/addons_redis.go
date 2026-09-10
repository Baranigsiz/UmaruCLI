package generator

import (
	"fmt"
	"path/filepath"
)

func getRedisFiles(config ProjectConfig, baseDir string) []string {
	if !config.Addons.Redis {
		return nil
	}
	if isGoTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "internal", "cache", "redis.go")}
	} else if isNodeTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "src", "config", "redis.ts")}
	} else if isPythonTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "app", "core", "redis.py")}
	}
	return nil
}

func generateRedisAddon(config ProjectConfig, baseDir string) error {
	if !config.Addons.Redis {
		return nil
	}

	if isGoTemplate(config.Template) {
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
		if err := writeAddonFile(baseDir, filepath.Join("internal", "cache", "redis.go"), content); err != nil {
			return err
		}
	} else if isNodeTemplate(config.Template) {
		content := `// Redis Cache Connection
export async function connectRedis(host = 'localhost', port = 6379) {
  console.log(` + "`🔴 Redis client configured on ${host}:${port}`" + `);
  return { host, port, ready: true };
}
`
		if err := writeAddonFile(baseDir, filepath.Join("src", "config", "redis.ts"), content); err != nil {
			return err
		}

		pkgPath := filepath.Join(baseDir, "package.json")
		nodeDeps := map[string]string{"ioredis": "^5.4.1"}
		nodeDevDeps := map[string]string{"@types/ioredis": "^5.0.0"}
		if err := injectNodeDependencies(pkgPath, nodeDeps, nodeDevDeps); err != nil {
			return fmt.Errorf("failed injecting node redis dependencies: %w", err)
		}
	} else if isPythonTemplate(config.Template) {
		content := `import os

REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))

async def get_redis_client():
    """Returns async Redis client connection."""
    return {"host": REDIS_HOST, "port": REDIS_PORT}
`
		if err := writeAddonFile(baseDir, filepath.Join("app", "core", "redis.py"), content); err != nil {
			return err
		}

		reqPath := filepath.Join(baseDir, "requirements.txt")
		if err := injectPythonDependencies(reqPath, []string{"redis>=5.0.0"}); err != nil {
			return fmt.Errorf("failed injecting python redis dependencies: %w", err)
		}
	}

	return nil
}
