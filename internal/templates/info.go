package templates

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// TemplateInfo holds all metadata and structural details for a template
type TemplateInfo struct {
	Config          TemplateConfig
	Ports           []string
	SupportedAddons []string
	FileTree        string
	TotalFiles      int
}

// GetTemplatePorts returns the known default network ports for a template
func GetTemplatePorts(templateID string) []string {
	switch templateID {
	case "go-fiber", "go-gin", "go-echo", "rust-actix":
		return []string{"API: 8080"}
	case "fastify-api", "hono-api", "node-express", "nestjs-api", "rust-axum", "nextjs-tailwind", "bun-elysia", "go-htmx":
		return []string{"App: 3000"}
	case "python-fastapi":
		return []string{"API: 8000"}
	case "ai-fastapi-starter":
		return []string{"API: 8000", "ChromaDB: 8001"}
	case "react-vite-ts", "vue-vite-ts", "svelte-vite-ts":
		return []string{"Dev Server: 5173"}
	case "astro-tailwind":
		return []string{"Dev Server: 4321"}
	case "fullstack-go-react":
		return []string{"API: 8080", "Frontend: 5173"}
	case "fullstack-ts-monorepo":
		return []string{"API: 8080", "Frontend: 3000"}
	default:
		return []string{"None (CLI Tool)"}
	}
}

// GetSupportedAddons returns which addons can be injected into this template
func GetSupportedAddons(templateID string) []string {
	switch templateID {
	case "go-fiber", "go-gin", "go-echo", "fastify-api", "hono-api", "node-express", "nestjs-api", "python-fastapi", "ai-fastapi-starter", "fullstack-go-react", "fullstack-ts-monorepo", "bun-elysia", "go-htmx":
		return []string{"🐘 PostgreSQL", "📦 SQLite", "🔴 Redis", "🔐 JWT Auth"}
	case "go-cli", "go-tui":
		return []string{"📦 SQLite", "🔴 Redis", "🐳 Docker"}
	case "rust-actix", "rust-axum":
		return []string{"Docker Compose Services"}
	default:
		return []string{"Standalone Template (Frontend/CLI)"}
	}
}

type treeNode struct {
	name     string
	isDir    bool
	children []*treeNode
}

// GenerateTemplateTree generates an ASCII directory tree of the embedded template files
func GenerateTemplateTree(templateID string) (string, int, error) {
	root := &treeNode{name: templateID, isDir: true}
	fileCount := 0

	err := fs.WalkDir(FS, templateID, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root folder itself
		if p == templateID {
			return nil
		}

		// Skip metadata config
		if d.Name() == "template.json" {
			return nil
		}

		// Rel path inside template
		relPath := strings.TrimPrefix(p, templateID+"/")
		parts := strings.Split(relPath, "/")

		// Insert into tree
		current := root
		for i, part := range parts {
			isLast := (i == len(parts)-1)
			var found *treeNode
			for _, child := range current.children {
				if child.name == part {
					found = child
					break
				}
			}

			if found == nil {
				displayName := part
				// Strip .tmpl for visual aesthetic
				if isLast && !d.IsDir() {
					displayName = strings.TrimSuffix(displayName, ".tmpl")
					fileCount++
				}
				newNode := &treeNode{
					name:  displayName,
					isDir: !isLast || d.IsDir(),
				}
				current.children = append(current.children, newNode)
				found = newNode
			}
			current = found
		}

		return nil
	})

	if err != nil {
		return "", 0, err
	}

	// Sort children (directories first, then alphabetical)
	sortTree(root)

	var sb strings.Builder
	renderTree(root, "", true, &sb, true)
	return sb.String(), fileCount, nil
}

func sortTree(node *treeNode) {
	sort.Slice(node.children, func(i, j int) bool {
		if node.children[i].isDir != node.children[j].isDir {
			return node.children[i].isDir // Dirs first
		}
		return node.children[i].name < node.children[j].name
	})
	for _, child := range node.children {
		if child.isDir {
			sortTree(child)
		}
	}
}

func renderTree(node *treeNode, prefix string, isLast bool, sb *strings.Builder, isRoot bool) {
	if isRoot {
		sb.WriteString(node.name + "/\n")
	} else {
		marker := "├── "
		if isLast {
			marker = "└── "
		}
		suffix := ""
		if node.isDir {
			suffix = "/"
		}
		sb.WriteString(prefix + marker + node.name + suffix + "\n")
	}

	childPrefix := prefix
	if !isRoot {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.children {
		isLastChild := (i == len(node.children)-1)
		renderTree(child, childPrefix, isLastChild, sb, false)
	}
}

// GetTemplateInfo inspects and returns detailed information about a template
func GetTemplateInfo(templateID string) (*TemplateInfo, error) {
	tmpl, err := FindTemplateByID(templateID)
	if err != nil {
		return nil, err
	}

	tree, count, err := GenerateTemplateTree(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tree for template %s: %w", templateID, err)
	}

	ports := GetTemplatePorts(templateID)
	addons := GetSupportedAddons(templateID)

	return &TemplateInfo{
		Config:          *tmpl,
		Ports:           ports,
		SupportedAddons: addons,
		FileTree:        tree,
		TotalFiles:      count,
	}, nil
}
