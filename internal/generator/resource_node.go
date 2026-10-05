package generator

import (
	"fmt"
	"path/filepath"
)

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
