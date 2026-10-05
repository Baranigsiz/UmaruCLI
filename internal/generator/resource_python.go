package generator

import (
	"fmt"
	"path/filepath"
)

func generatePythonResource(baseDir string, ctx resourceContext) []ResourceFile {
	var files []ResourceFile

	// 1. Schemas (app/schemas/<slug>.py)
	schemaPath := filepath.Join(baseDir, "app", "schemas", fmt.Sprintf("%s.py", ctx.slug))
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
`, ctx.pascal, ctx.slug, ctx.pascal, ctx.pascal, ctx.pascal, ctx.pascal, ctx.pascal)

	files = append(files, ResourceFile{
		Path:    schemaPath,
		RelPath: filepath.ToSlash(filepath.Join("app", "schemas", fmt.Sprintf("%s.py", ctx.slug))),
		Content: schemaContent,
	})

	// 2. Router (app/api/endpoints/<slug>.py or app/routers/<slug>.py)
	routerRel := filepath.Join("app", "api", "endpoints", fmt.Sprintf("%s.py", ctx.slug))
	if fileExists(filepath.Join(baseDir, "app", "routers")) || !fileExists(filepath.Join(baseDir, "app", "api")) {
		routerRel = filepath.Join("app", "routers", fmt.Sprintf("%s.py", ctx.slug))
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
`, ctx.slug, ctx.pascal, ctx.pascal, ctx.pascal, ctx.pluralSlug, ctx.pluralPascal, ctx.pascal, ctx.pluralSlug, ctx.pluralSlug, ctx.pascal, ctx.slug, ctx.slug, ctx.pascal, ctx.pascal, ctx.slug, ctx.pascal, ctx.slug, ctx.slug, ctx.slug, ctx.slug, ctx.pascal)

	files = append(files, ResourceFile{
		Path:    routerPath,
		RelPath: filepath.ToSlash(routerRel),
		Content: routerContent,
	})

	return files
}
