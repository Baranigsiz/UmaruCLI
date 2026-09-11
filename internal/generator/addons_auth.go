package generator

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getAuthFiles(config ProjectConfig, baseDir string) []string {
	auth := strings.ToLower(strings.TrimSpace(config.Addons.Auth))
	if auth != "jwt" {
		return nil
	}
	if isGoTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "internal", "middleware", "auth.go")}
	} else if isNodeTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "src", "middlewares", "auth.middleware.ts")}
	} else if isPythonTemplate(config.Template) {
		return []string{filepath.Join(baseDir, "app", "core", "security.py")}
	}
	return nil
}

func generateAuthAddon(config ProjectConfig, baseDir string) error {
	auth := strings.ToLower(strings.TrimSpace(config.Addons.Auth))
	if auth != "jwt" {
		return nil
	}

	if isGoTemplate(config.Template) {
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
		if err := writeAddonFile(baseDir, filepath.Join("internal", "middleware", "auth.go"), content); err != nil {
			return err
		}
	} else if isNodeTemplate(config.Template) {
		var content string
		switch {
		case strings.HasPrefix(config.Template, "hono-") || config.Template == "fullstack-ts-monorepo":
			content = `import { Context, Next } from 'hono';

export interface AuthUser {
  id: string;
  email: string;
  role: string;
}

export async function authMiddleware(c: Context, next: Next) {
  const authHeader = c.req.header('Authorization');
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return c.json({ error: 'Unauthorized: Missing or malformed token' }, 401);
  }

  const token = authHeader.split(' ')[1];
  try {
    // In production, verify token with your JWT secret
    c.set('user', { id: 'usr_sample', email: 'user@example.com', role: 'admin' });
    await next();
  } catch (err) {
    return c.json({ error: 'Unauthorized: Invalid token' }, 401);
  }
}
`
		case strings.HasPrefix(config.Template, "fastify-"):
			content = `import { FastifyRequest, FastifyReply } from 'fastify';

export interface AuthUser {
  id: string;
  email: string;
  role: string;
}

declare module 'fastify' {
  interface FastifyRequest {
    user?: AuthUser;
  }
}

export async function authMiddleware(request: FastifyRequest, reply: FastifyReply) {
  const authHeader = request.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return reply.status(401).send({ error: 'Unauthorized: Missing or malformed token' });
  }

  const token = authHeader.split(' ')[1];
  try {
    // In production, verify token with jwt.verify(token, process.env.JWT_SECRET || 'secret')
    request.user = { id: 'usr_sample', email: 'user@example.com', role: 'admin' };
  } catch (err) {
    return reply.status(401).send({ error: 'Unauthorized: Invalid token' });
  }
}
`
		case strings.HasPrefix(config.Template, "nestjs-"):
			content = `import { Injectable, CanActivate, ExecutionContext, UnauthorizedException } from '@nestjs/common';

@Injectable()
export class AuthGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest();
    const authHeader = request.headers.authorization;

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new UnauthorizedException('Missing or malformed token');
    }

    const token = authHeader.split(' ')[1];
    try {
      request.user = { id: 'usr_sample', email: 'user@example.com', role: 'admin' };
      return true;
    } catch (err) {
      throw new UnauthorizedException('Invalid token');
    }
  }
}
`
		default:
			content = `import { Request, Response, NextFunction } from 'express';

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
		}
		if err := writeAddonFile(baseDir, filepath.Join("src", "middlewares", "auth.middleware.ts"), content); err != nil {
			return err
		}

		pkgPath := filepath.Join(baseDir, "package.json")
		nodeDeps := map[string]string{"jsonwebtoken": "^9.0.2"}
		nodeDevDeps := map[string]string{"@types/jsonwebtoken": "^9.0.6"}
		if err := injectNodeDependencies(pkgPath, nodeDeps, nodeDevDeps); err != nil {
			return fmt.Errorf("failed injecting node auth dependencies: %w", err)
		}
	} else if isPythonTemplate(config.Template) {
		content := `import os
from datetime import datetime, timedelta, timezone
from typing import Optional
from jose import JWTError, jwt

SECRET_KEY = os.getenv("JWT_SECRET_KEY", "your-super-secret-key-change-in-production")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60 * 24

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    to_encode = data.copy()
    expire = datetime.now(timezone.utc) + (expires_delta or timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES))
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)

def verify_token(token: str) -> Optional[dict]:
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        return payload
    except JWTError:
        return None
`
		if err := writeAddonFile(baseDir, filepath.Join("app", "core", "security.py"), content); err != nil {
			return err
		}

		reqPath := filepath.Join(baseDir, "requirements.txt")
		if err := injectPythonDependencies(reqPath, []string{"python-jose[cryptography]>=3.3.0", "passlib[bcrypt]>=1.7.4"}); err != nil {
			return fmt.Errorf("failed injecting python auth dependencies: %w", err)
		}
	}

	return nil
}
