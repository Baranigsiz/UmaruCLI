package generator

import (
	"path/filepath"
	"strings"
)

func getCIFiles(baseDir string) []string {
	return []string{
		filepath.Join(baseDir, ".github", "workflows", "ci.yml"),
	}
}

func generateCIAddon(config ProjectConfig, baseDir string) error {
	var content string

	switch {
	case config.Template == "fullstack-go-react":
		content = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test-api:
    name: Test & Build Go API
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: apps/api
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Verify Dependencies
        run: go mod verify

      - name: Run Tests
        run: go test -v -race ./...

  test-web:
    name: Test & Build Frontend
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: apps/web
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20

      - name: Install Dependencies
        run: npm ci || npm install

      - name: Build Application
        run: npm run build --if-present
`
	case isGoTemplate(config.Template):
		content = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Test & Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Verify Dependencies
        run: go mod verify

      - name: Run Tests
        run: go test -v -race ./...

      - name: Build Binary
        run: go build -v ./...
`
	case isRustTemplate(config.Template):
		content = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Check, Test & Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Rust Toolchain
        uses: dtolnay/rust-toolchain@stable

      - name: Cargo Cache
        uses: actions/cache@v4
        with:
          path: |
            ~/.cargo/bin/
            ~/.cargo/registry/index/
            ~/.cargo/registry/cache/
            ~/.cargo/git/db/
            target/
          key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}
          restore-keys: ${{ runner.os }}-cargo-

      - name: Check
        run: cargo check --verbose

      - name: Run Tests
        run: cargo test --verbose

      - name: Build Release
        run: cargo build --release --verbose
`
	case strings.HasPrefix(config.Template, "bun-"):
		content = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Lint, Test & Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Bun
        uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - name: Install Dependencies
        run: bun install --frozen-lockfile || bun install

      - name: Run Tests
        run: bun test --if-present

      - name: Build Application
        run: bun run build --if-present
`
	case isPythonTemplate(config.Template):
		content = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Lint & Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
          cache: 'pip'

      - name: Install Dependencies
        run: |
          python -m pip install --upgrade pip
          if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
          pip install flake8 pytest

      - name: Lint with flake8
        run: |
          flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics
          flake8 . --count --exit-zero --max-complexity=10 --max-line-length=127 --statistics

      - name: Run Tests
        run: pytest
`
	default: // Node.js / TypeScript
		content = `name: CI

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Lint, Test & Build
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'

      - name: Install Dependencies
        run: npm ci || npm install

      - name: Run Linter & Tests
        run: |
          npm run lint --if-present
          npm test --if-present

      - name: Build Application
        run: npm run build --if-present
`
	}

	return writeAddonFile(baseDir, filepath.Join(".github", "workflows", "ci.yml"), content)
}
