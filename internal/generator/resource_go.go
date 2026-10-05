package generator

import (
	"fmt"
	"path/filepath"
	"strings"
)

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
