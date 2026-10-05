package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ResourceConfig defines the parameters for generating a fullstack/backend resource
type ResourceConfig struct {
	Name      string // Raw input (e.g. "Product", "user_profile")
	TargetDir string // Project root directory
	Force     bool   // Overwrite existing files
	DryRun    bool   // Simulate without writing
}

// ResourceFile describes a single generated file
type ResourceFile struct {
	Path    string `json:"path"`
	RelPath string `json:"rel_path"`
	Content string `json:"content"`
	Action  string `json:"action"` // "created", "skipped (already exists)", "overwritten"
}

// ResourceResult contains the report of resource generation
type ResourceResult struct {
	ResourceName string         `json:"resource_name"`
	Language     string         `json:"language"`
	Framework    string         `json:"framework"`
	Files        []ResourceFile `json:"files"`
}

// ToPascalCase converts strings like "user_profile", "user-profile", "user" into "UserProfile"
func ToPascalCase(s string) string {
	s = Transliterate(s)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})
	if len(parts) == 0 {
		return "Resource"
	}
	var sb strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		sb.WriteRune(unicode.ToUpper(runes[0]))
		for i := 1; i < len(runes); i++ {
			sb.WriteRune(unicode.ToLower(runes[i]))
		}
	}
	res := sb.String()
	if res == "" {
		return "Resource"
	}
	return res
}

// ToCamelCase converts string into camelCase (e.g. "UserProfile" -> "userProfile")
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return "resource"
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// Pluralize adds simple English pluralization (e.g. "user" -> "users", "category" -> "categories")
func Pluralize(s string) string {
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		lastTwo := lower[len(lower)-2:]
		if !strings.ContainsAny(string(lastTwo[0]), "aeiou") {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}

// GenerateResource inspects project framework and generates model, handler, service, and repository boilerplate
func GenerateResource(cfg ResourceConfig) (*ResourceResult, error) {
	targetDir := cfg.TargetDir
	if targetDir == "" {
		targetDir = "."
	}
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory: %w", err)
	}

	proj, err := DetectProject(absDir)
	if err != nil {
		return nil, fmt.Errorf("could not detect project structure: %w", err)
	}

	rawName := strings.TrimSpace(cfg.Name)
	if rawName == "" {
		return nil, fmt.Errorf("resource name cannot be empty")
	}

	pascalName := ToPascalCase(rawName)
	slugName := Slugify(rawName)
	camelName := ToCamelCase(rawName)
	pluralSlug := Pluralize(slugName)
	pluralPascal := Pluralize(pascalName)

	baseDir := GetAddonBaseDir(proj.ToProjectConfig(AddonConfig{}))

	var files []ResourceFile

	switch proj.Type {
	case ProjectTypeGo:
		files = generateGoResource(baseDir, pascalName, slugName, camelName, pluralSlug, pluralPascal, proj.Framework)
	case ProjectTypeNode:
		files = generateNodeResource(baseDir, pascalName, slugName, camelName, pluralSlug, pluralPascal, proj.Framework)
	case ProjectTypePython:
		files = generatePythonResource(baseDir, pascalName, slugName, camelName, pluralSlug, pluralPascal)
	case ProjectTypeRust:
		files = generateRustResource(baseDir, pascalName, slugName, camelName, pluralSlug, pluralPascal)
	default:
		return nil, fmt.Errorf("resource generation is not supported for project type: %s", proj.Type)
	}

	result := &ResourceResult{
		ResourceName: pascalName,
		Language:     string(proj.Type),
		Framework:    proj.Framework,
		Files:        make([]ResourceFile, 0, len(files)),
	}

	for _, f := range files {
		action := "created"
		if fileExists(f.Path) {
			if !cfg.Force {
				action = "skipped (already exists)"
				f.Action = action
				result.Files = append(result.Files, f)
				continue
			}
			action = "overwritten"
		}

		if !cfg.DryRun {
			if err := os.MkdirAll(filepath.Dir(f.Path), 0755); err != nil {
				return nil, fmt.Errorf("failed creating directory for %s: %w", f.RelPath, err)
			}
			if err := os.WriteFile(f.Path, []byte(f.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed writing %s: %w", f.RelPath, err)
			}
		}
		f.Action = action
		result.Files = append(result.Files, f)
	}

	return result, nil
}

func generateGoResource(baseDir, pascal, slug, camel, pluralSlug, pluralPascal, framework string) []ResourceFile {
	replacer := strings.NewReplacer(
		"{{Pascal}}", pascal,
		"{{Slug}}", slug,
		"{{Camel}}", camel,
		"{{PluralSlug}}", pluralSlug,
		"{{PluralPascal}}", pluralPascal,
	)

	var files []ResourceFile

	// 1. Model (internal/models/<slug>.go)
	modelPath := filepath.Join(baseDir, "internal", "models", fmt.Sprintf("%s.go", slug))
	modelTmpl := `package models

import (
	"time"
)

// {{Pascal}} represents the core entity
type {{Pascal}} struct {
	ID        string    ` + "`json:\"id\" db:\"id\"`" + `
	Name      string    ` + "`json:\"name\" db:\"name\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" db:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" db:\"updated_at\"`" + `
}

// Create{{Pascal}}Request defines the payload for creating a {{Pascal}}
type Create{{Pascal}}Request struct {
	Name string ` + "`json:\"name\" validate:\"required,min=2\"`" + `
}

// Update{{Pascal}}Request defines the payload for updating a {{Pascal}}
type Update{{Pascal}}Request struct {
	Name string ` + "`json:\"name,omitempty\"`" + `
}
`
	files = append(files, ResourceFile{
		Path:    modelPath,
		RelPath: filepath.ToSlash(filepath.Join("internal", "models", fmt.Sprintf("%s.go", slug))),
		Content: replacer.Replace(modelTmpl),
	})

	// 2. Repository (internal/repository/<slug>_repository.go)
	repoPath := filepath.Join(baseDir, "internal", "repository", fmt.Sprintf("%s_repository.go", slug))
	repoTmpl := `package repository

import (
	"context"
	"errors"
	"sync"
	"time"
	"umaru/internal/models"
)

var (
	Err{{Pascal}}NotFound = errors.New("{{Slug}} not found")
)

// {{Pascal}}Repository defines data access methods for {{Pascal}}
type {{Pascal}}Repository interface {
	FindAll(ctx context.Context) ([]models.{{Pascal}}, error)
	FindByID(ctx context.Context, id string) (*models.{{Pascal}}, error)
	Create(ctx context.Context, item *models.{{Pascal}}) error
	Delete(ctx context.Context, id string) error
}

type memory{{Pascal}}Repository struct {
	mu    sync.RWMutex
	items map[string]models.{{Pascal}}
}

// New{{Pascal}}Repository creates a new in-memory {{Pascal}} repository
func New{{Pascal}}Repository() {{Pascal}}Repository {
	return &memory{{Pascal}}Repository{
		items: make(map[string]models.{{Pascal}}),
	}
}

func (r *memory{{Pascal}}Repository) FindAll(ctx context.Context) ([]models.{{Pascal}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]models.{{Pascal}}, 0, len(r.items))
	for _, v := range r.items {
		result = append(result, v)
	}
	return result, nil
}

func (r *memory{{Pascal}}Repository) FindByID(ctx context.Context, id string) (*models.{{Pascal}}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[id]
	if !exists {
		return nil, Err{{Pascal}}NotFound
	}
	return &item, nil
}

func (r *memory{{Pascal}}Repository) Create(ctx context.Context, item *models.{{Pascal}}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	r.items[item.ID] = *item
	return nil
}

func (r *memory{{Pascal}}Repository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return Err{{Pascal}}NotFound
	}
	delete(r.items, id)
	return nil
}
`
	files = append(files, ResourceFile{
		Path:    repoPath,
		RelPath: filepath.ToSlash(filepath.Join("internal", "repository", fmt.Sprintf("%s_repository.go", slug))),
		Content: replacer.Replace(repoTmpl),
	})

	// 3. Service (internal/service/<slug>_service.go)
	servicePath := filepath.Join(baseDir, "internal", "service", fmt.Sprintf("%s_service.go", slug))
	serviceTmpl := `package service

import (
	"context"
	"fmt"
	"time"
	"umaru/internal/models"
	"umaru/internal/repository"
)

// {{Pascal}}Service defines business logic operations
type {{Pascal}}Service interface {
	GetAll(ctx context.Context) ([]models.{{Pascal}}, error)
	GetByID(ctx context.Context, id string) (*models.{{Pascal}}, error)
	Create(ctx context.Context, req models.Create{{Pascal}}Request) (*models.{{Pascal}}, error)
	Delete(ctx context.Context, id string) error
}

type {{Camel}}ServiceImpl struct {
	repo repository.{{Pascal}}Repository
}

// New{{Pascal}}Service constructs a new {{Pascal}}Service
func New{{Pascal}}Service(repo repository.{{Pascal}}Repository) {{Pascal}}Service {
	return &{{Camel}}ServiceImpl{repo: repo}
}

func (s *{{Camel}}ServiceImpl) GetAll(ctx context.Context) ([]models.{{Pascal}}, error) {
	return s.repo.FindAll(ctx)
}

func (s *{{Camel}}ServiceImpl) GetByID(ctx context.Context, id string) (*models.{{Pascal}}, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *{{Camel}}ServiceImpl) Create(ctx context.Context, req models.Create{{Pascal}}Request) (*models.{{Pascal}}, error) {
	item := &models.{{Pascal}}{
		ID:   fmt.Sprintf("{{Slug}}_%d", time.Now().UnixNano()),
		Name: req.Name,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *{{Camel}}ServiceImpl) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
`
	files = append(files, ResourceFile{
		Path:    servicePath,
		RelPath: filepath.ToSlash(filepath.Join("internal", "service", fmt.Sprintf("%s_service.go", slug))),
		Content: replacer.Replace(serviceTmpl),
	})

	// 4. Handler (internal/handlers/<slug>.go)
	handlerPath := filepath.Join(baseDir, "internal", "handlers", fmt.Sprintf("%s.go", slug))
	var handlerTmpl string

	if strings.Contains(framework, "gin") {
		handlerTmpl = `package handlers

import (
	"net/http"
	"umaru/internal/models"
	"umaru/internal/service"

	"github.com/gin-gonic/gin"
)

type {{Pascal}}Handler struct {
	service service.{{Pascal}}Service
}

func New{{Pascal}}Handler(s service.{{Pascal}}Service) *{{Pascal}}Handler {
	return &{{Pascal}}Handler{service: s}
}

func (h *{{Pascal}}Handler) RegisterRoutes(r *gin.RouterGroup) {
	routes := r.Group("/{{PluralSlug}}")
	{
		routes.GET("", h.List)
		routes.GET("/:id", h.Get)
		routes.POST("", h.Create)
		routes.DELETE("/:id", h.Delete)
	}
}

func (h *{{Pascal}}Handler) List(c *gin.Context) {
	items, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *{{Pascal}}Handler) Get(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "{{Slug}} not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *{{Pascal}}Handler) Create(c *gin.Context) {
	var req models.Create{{Pascal}}Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *{{Pascal}}Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
`
	} else if strings.Contains(framework, "echo") {
		handlerTmpl = `package handlers

import (
	"net/http"
	"umaru/internal/models"
	"umaru/internal/service"

	"github.com/labstack/echo/v4"
)

type {{Pascal}}Handler struct {
	service service.{{Pascal}}Service
}

func New{{Pascal}}Handler(s service.{{Pascal}}Service) *{{Pascal}}Handler {
	return &{{Pascal}}Handler{service: s}
}

func (h *{{Pascal}}Handler) RegisterRoutes(g *echo.Group) {
	routes := g.Group("/{{PluralSlug}}")
	routes.GET("", h.List)
	routes.GET("/:id", h.Get)
	routes.POST("", h.Create)
	routes.DELETE("/:id", h.Delete)
}

func (h *{{Pascal}}Handler) List(c echo.Context) error {
	items, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *{{Pascal}}Handler) Get(c echo.Context) error {
	id := c.Param("id")
	item, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "{{Slug}} not found"})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *{{Pascal}}Handler) Create(c echo.Context) error {
	var req models.Create{{Pascal}}Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	created, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *{{Pascal}}Handler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}
`
	} else {
		// Default Fiber
		handlerTmpl = `package handlers

import (
	"umaru/internal/models"
	"umaru/internal/service"

	"github.com/gofiber/fiber/v2"
)

type {{Pascal}}Handler struct {
	service service.{{Pascal}}Service
}

func New{{Pascal}}Handler(s service.{{Pascal}}Service) *{{Pascal}}Handler {
	return &{{Pascal}}Handler{service: s}
}

func (h *{{Pascal}}Handler) RegisterRoutes(router fiber.Router) {
	routes := router.Group("/{{PluralSlug}}")
	routes.Get("/", h.List)
	routes.Get("/:id", h.Get)
	routes.Post("/", h.Create)
	routes.Delete("/:id", h.Delete)
}

func (h *{{Pascal}}Handler) List(c *fiber.Ctx) error {
	items, err := h.service.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *{{Pascal}}Handler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "{{Slug}} not found"})
	}
	return c.JSON(item)
}

func (h *{{Pascal}}Handler) Create(c *fiber.Ctx) error {
	var req models.Create{{Pascal}}Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	created, err := h.service.Create(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *{{Pascal}}Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
`
	}

	files = append(files, ResourceFile{
		Path:    handlerPath,
		RelPath: filepath.ToSlash(filepath.Join("internal", "handlers", fmt.Sprintf("%s.go", slug))),
		Content: replacer.Replace(handlerTmpl),
	})

	return files
}

func generateNodeResource(baseDir, pascal, slug, camel, pluralSlug, pluralPascal, framework string) []ResourceFile {
	var files []ResourceFile

	// 1. Model / Schema (src/models/<slug>.model.ts)
	modelPath := filepath.Join(baseDir, "src", "models", fmt.Sprintf("%s.model.ts", slug))
	modelContent := fmt.Sprintf(`export interface %s {
  id: string;
  name: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Create%sDTO {
  name: string;
}

export interface Update%sDTO {
  name?: string;
}
`, pascal, pascal, pascal)

	files = append(files, ResourceFile{
		Path:    modelPath,
		RelPath: filepath.ToSlash(filepath.Join("src", "models", fmt.Sprintf("%s.model.ts", slug))),
		Content: modelContent,
	})

	// 2. Service (src/services/<slug>.service.ts)
	servicePath := filepath.Join(baseDir, "src", "services", fmt.Sprintf("%s.service.ts", slug))
	serviceContent := fmt.Sprintf(`import { %s, Create%sDTO, Update%sDTO } from '../models/%s.model';

export class %sService {
  private items: Map<string, %s> = new Map();

  async findAll(): Promise<%s[]> {
    return Array.from(this.items.values());
  }

  async findById(id: string): Promise<%s | null> {
    return this.items.get(id) || null;
  }

  async create(dto: Create%sDTO): Promise<%s> {
    const item: %s = {
      id: '%s_' + Date.now(),
      name: dto.name,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    this.items.set(item.id, item);
    return item;
  }

  async delete(id: string): Promise<boolean> {
    return this.items.delete(id);
  }
}
`, pascal, pascal, pascal, slug, pascal, pascal, pascal, pascal, pascal, pascal, pascal, slug)

	files = append(files, ResourceFile{
		Path:    servicePath,
		RelPath: filepath.ToSlash(filepath.Join("src", "services", fmt.Sprintf("%s.service.ts", slug))),
		Content: serviceContent,
	})

	// 3. Controller (src/controllers/<slug>.controller.ts)
	controllerPath := filepath.Join(baseDir, "src", "controllers", fmt.Sprintf("%s.controller.ts", slug))
	controllerContent := fmt.Sprintf(`import { Request, Response } from 'express';
import { %sService } from '../services/%s.service';

const service = new %sService();

export class %sController {
  static async list(req: Request, res: Response) {
    const items = await service.findAll();
    return res.json(items);
  }

  static async get(req: Request, res: Response) {
    const item = await service.findById(req.params.id);
    if (!item) {
      return res.status(404).json({ error: '%s not found' });
    }
    return res.json(item);
  }

  static async create(req: Request, res: Response) {
    const created = await service.create(req.body);
    return res.status(201).json(created);
  }

  static async delete(req: Request, res: Response) {
    const deleted = await service.delete(req.params.id);
    if (!deleted) {
      return res.status(404).json({ error: '%s not found' });
    }
    return res.status(204).send();
  }
}
`, pascal, slug, pascal, pascal, pascal, pascal)

	files = append(files, ResourceFile{
		Path:    controllerPath,
		RelPath: filepath.ToSlash(filepath.Join("src", "controllers", fmt.Sprintf("%s.controller.ts", slug))),
		Content: controllerContent,
	})

	// 4. Routes (src/routes/<slug>.routes.ts)
	routesPath := filepath.Join(baseDir, "src", "routes", fmt.Sprintf("%s.routes.ts", slug))
	routesContent := fmt.Sprintf(`import { Router } from 'express';
import { %sController } from '../controllers/%s.controller';

const router = Router();

router.get('/', %sController.list);
router.get('/:id', %sController.get);
router.post('/', %sController.create);
router.delete('/:id', %sController.delete);

export default router;
`, pascal, slug, pascal, pascal, pascal, pascal)

	files = append(files, ResourceFile{
		Path:    routesPath,
		RelPath: filepath.ToSlash(filepath.Join("src", "routes", fmt.Sprintf("%s.routes.ts", slug))),
		Content: routesContent,
	})

	return files
}

func generatePythonResource(baseDir, pascal, slug, camel, pluralSlug, pluralPascal string) []ResourceFile {
	var files []ResourceFile

	// 1. Schemas (app/schemas/<slug>.py)
	schemaPath := filepath.Join(baseDir, "app", "schemas", fmt.Sprintf("%s.py", slug))
	schemaContent := fmt.Sprintf(`from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field

class %sBase(BaseModel):
    name: str = Field(..., min_length=2, description="Name of the %s")

class %sCreate(%sBase):
    pass

class %sUpdate(BaseModel):
    name: Optional[str] = None

class %sResponse(%sBase):
    id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
`, pascal, slug, pascal, pascal, pascal, pascal, pascal)

	files = append(files, ResourceFile{
		Path:    schemaPath,
		RelPath: filepath.ToSlash(filepath.Join("app", "schemas", fmt.Sprintf("%s.py", slug))),
		Content: schemaContent,
	})

	// 2. Router (app/api/endpoints/<slug>.py or app/routers/<slug>.py)
	routerRel := filepath.Join("app", "api", "endpoints", fmt.Sprintf("%s.py", slug))
	if fileExists(filepath.Join(baseDir, "app", "routers")) || !fileExists(filepath.Join(baseDir, "app", "api")) {
		routerRel = filepath.Join("app", "routers", fmt.Sprintf("%s.py", slug))
	}
	routerPath := filepath.Join(baseDir, routerRel)

	routerContent := fmt.Sprintf(`import time
from datetime import datetime
from typing import List
from fastapi import APIRouter, HTTPException, status
from app.schemas.%s import %sCreate, %sResponse, %sUpdate

router = APIRouter(prefix="/%s", tags=["%s"])

# In-memory mock storage
_items = {}

@router.get("", response_model=List[%sResponse])
async def list_%s():
    """Retrieve all %s items."""
    return list(_items.values())

@router.get("/{item_id}", response_model=%sResponse)
async def get_%s(item_id: str):
    """Retrieve a single %s by ID."""
    if item_id not in _items:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="%s not found")
    return _items[item_id]

@router.post("", response_model=%sResponse, status_code=status.HTTP_201_CREATED)
async def create_%s(payload: %sCreate):
    """Create a new %s."""
    item_id = f"%s_{int(time.time()*1000)}"
    now = datetime.utcnow()
    item = {
        "id": item_id,
        "name": payload.name,
        "created_at": now,
        "updated_at": now,
    }
    _items[item_id] = item
    return item

@router.delete("/{item_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_%s(item_id: str):
    """Delete a %s by ID."""
    if item_id not in _items:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="%s not found")
    del _items[item_id]
`, slug, pascal, pascal, pascal, pluralSlug, pluralPascal, pascal, pluralSlug, pluralSlug, pascal, slug, slug, pascal, pascal, slug, pascal, slug, slug, slug, slug, pascal)

	files = append(files, ResourceFile{
		Path:    routerPath,
		RelPath: filepath.ToSlash(routerRel),
		Content: routerContent,
	})

	return files
}

func generateRustResource(baseDir, pascal, slug, camel, pluralSlug, pluralPascal string) []ResourceFile {
	var files []ResourceFile

	// Model (src/models/<slug>.rs)
	modelPath := filepath.Join(baseDir, "src", "models", fmt.Sprintf("%s.rs", slug))
	modelContent := fmt.Sprintf(`use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct %s {
    pub id: String,
    pub name: String,
}

#[derive(Debug, Deserialize)]
pub struct Create%sPayload {
    pub name: String,
}
`, pascal, pascal)

	files = append(files, ResourceFile{
		Path:    modelPath,
		RelPath: filepath.ToSlash(filepath.Join("src", "models", fmt.Sprintf("%s.rs", slug))),
		Content: modelContent,
	})

	return files
}
