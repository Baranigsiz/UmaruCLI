package generator

import (
	"os"
	"path/filepath"
	"strings"
)

// AddonStatus represents the installation status of an individual addon
type AddonStatus struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Installed     bool     `json:"installed"`
	DetectedFiles []string `json:"detected_files,omitempty"`
}

// ProjectAddonAudit holds the full audit report for a project
type ProjectAddonAudit struct {
	ProjectName string        `json:"project_name"`
	Framework   string        `json:"framework"`
	Type        ProjectType   `json:"type"`
	TargetDir   string        `json:"target_dir"`
	Addons      []AddonStatus `json:"addons"`
}

// AuditProjectAddons inspects the given target directory and reports which addons are installed
func AuditProjectAddons(targetDir string) (*ProjectAddonAudit, error) {
	proj, err := DetectProject(targetDir)
	if err != nil {
		return nil, err
	}

	audit := &ProjectAddonAudit{
		ProjectName: proj.ProjectName,
		Framework:   proj.Framework,
		Type:        proj.Type,
		TargetDir:   proj.TargetDir,
	}

	dummyConfig := proj.ToProjectConfig(AddonConfig{})
	baseDir := GetAddonBaseDir(dummyConfig)

	cleanRel := func(f string) string {
		rel, err := filepath.Rel(targetDir, f)
		if err != nil {
			rel = f
		}
		return filepath.ToSlash(rel)
	}

	// 1. Docker
	dockerFiles := getDockerFiles(baseDir)
	var foundDocker []string
	hasDockerCore := false
	for _, f := range dockerFiles {
		if fileExists(f) {
			foundDocker = append(foundDocker, cleanRel(f))
			base := filepath.Base(f)
			if base == "Dockerfile" || strings.HasPrefix(base, "docker-compose") {
				hasDockerCore = true
			}
		}
	}
	audit.Addons = append(audit.Addons, AddonStatus{
		ID:            "docker",
		Name:          "Docker & Compose",
		Installed:     hasDockerCore,
		DetectedFiles: foundDocker,
	})

	// 2. CI/CD
	ciFiles := getCIFiles(dummyConfig.TargetDir)
	var foundCI []string
	for _, f := range ciFiles {
		if fileExists(f) {
			foundCI = append(foundCI, cleanRel(f))
		}
	}
	audit.Addons = append(audit.Addons, AddonStatus{
		ID:            "ci",
		Name:          "GitHub Actions CI/CD",
		Installed:     len(foundCI) > 0,
		DetectedFiles: foundCI,
	})

	isPostgresContent := func(filePath string) bool {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return false
		}
		s := strings.ToLower(string(data))
		return strings.Contains(s, "pg") || strings.Contains(s, "postgres") || strings.Contains(s, "asyncpg") || strings.Contains(s, "psycopg")
	}

	isSQLiteContent := func(filePath string) bool {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return false
		}
		s := strings.ToLower(string(data))
		return strings.Contains(s, "sqlite") || strings.Contains(s, "better-sqlite3") || strings.Contains(s, "aiosqlite")
	}

	// 3. PostgreSQL
	pgConfig := dummyConfig
	pgConfig.Addons.Database = "postgres"
	pgFiles := getDatabaseFiles(pgConfig, baseDir)
	var foundPG []string
	for _, f := range pgFiles {
		if fileExists(f) {
			if proj.Type == ProjectTypeNode || proj.Type == ProjectTypePython {
				if isPostgresContent(f) {
					foundPG = append(foundPG, cleanRel(f))
				}
			} else {
				foundPG = append(foundPG, cleanRel(f))
			}
		}
	}
	audit.Addons = append(audit.Addons, AddonStatus{
		ID:            "postgres",
		Name:          "PostgreSQL Database",
		Installed:     len(foundPG) > 0,
		DetectedFiles: foundPG,
	})

	// 4. SQLite
	sqliteConfig := dummyConfig
	sqliteConfig.Addons.Database = "sqlite"
	sqliteFiles := getDatabaseFiles(sqliteConfig, baseDir)
	var foundSQLite []string
	for _, f := range sqliteFiles {
		if fileExists(f) {
			if proj.Type == ProjectTypeNode || proj.Type == ProjectTypePython {
				if isSQLiteContent(f) {
					foundSQLite = append(foundSQLite, cleanRel(f))
				}
			} else {
				foundSQLite = append(foundSQLite, cleanRel(f))
			}
		}
	}
	audit.Addons = append(audit.Addons, AddonStatus{
		ID:            "sqlite",
		Name:          "SQLite Database",
		Installed:     len(foundSQLite) > 0,
		DetectedFiles: foundSQLite,
	})

	// 5. JWT Auth
	authConfig := dummyConfig
	authConfig.Addons.Auth = "jwt"
	authFiles := getAuthFiles(authConfig, baseDir)
	var foundAuth []string
	for _, f := range authFiles {
		if fileExists(f) {
			foundAuth = append(foundAuth, cleanRel(f))
		}
	}
	audit.Addons = append(audit.Addons, AddonStatus{
		ID:            "jwt",
		Name:          "JWT Authentication",
		Installed:     len(foundAuth) > 0,
		DetectedFiles: foundAuth,
	})

	// 6. Redis
	redisConfig := dummyConfig
	redisConfig.Addons.Redis = true
	redisFiles := getRedisFiles(redisConfig, baseDir)
	var foundRedis []string
	for _, f := range redisFiles {
		if fileExists(f) {
			foundRedis = append(foundRedis, cleanRel(f))
		}
	}
	audit.Addons = append(audit.Addons, AddonStatus{
		ID:            "redis",
		Name:          "Redis Caching Client",
		Installed:     len(foundRedis) > 0,
		DetectedFiles: foundRedis,
	})

	return audit, nil
}
