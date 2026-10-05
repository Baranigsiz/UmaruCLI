package generator

import (
	"fmt"
	"path/filepath"
)

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
