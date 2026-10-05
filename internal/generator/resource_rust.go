package generator

import (
	"fmt"
	"path/filepath"
)

func generateRustResource(baseDir string, ctx resourceContext) []ResourceFile {
	var files []ResourceFile

	// Model (src/models/<slug>.rs)
	modelPath := filepath.Join(baseDir, "src", "models", fmt.Sprintf("%s.rs", ctx.slug))
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
`, ctx.pascal, ctx.pascal)

	files = append(files, ResourceFile{
		Path:    modelPath,
		RelPath: filepath.ToSlash(filepath.Join("src", "models", fmt.Sprintf("%s.rs", ctx.slug))),
		Content: modelContent,
	})

	return files
}
