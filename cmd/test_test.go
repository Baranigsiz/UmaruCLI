package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestCmd_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)

	output, err := executeCommand("test", "--dry-run", tempDir)
	if err != nil {
		t.Fatalf("executeCommand(test --dry-run) failed: %v, output: %s", err, output)
	}

	if !strings.Contains(output, "Umaru Test Runner") {
		t.Errorf("expected header 'Umaru Test Runner' in output, got: %s", output)
	}
	if !strings.Contains(output, "go test") {
		t.Errorf("expected 'go test' in command, got: %s", output)
	}
	if !strings.Contains(output, "Dry-run mode") {
		t.Errorf("expected dry-run note in output, got: %s", output)
	}
}

func TestTestCmd_JSON(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testapp\ngo 1.24"), 0644)

	output, err := executeCommand("test", "--json", "--verbose", "--coverage", "-f", "TestUnit", tempDir)
	if err != nil {
		t.Fatalf("executeCommand(test --json) failed: %v, output: %s", err, output)
	}

	var res map[string]interface{}
	if err := json.Unmarshal([]byte(output), &res); err != nil {
		t.Fatalf("failed parsing JSON output: %v, raw: %s", err, output)
	}

	if res["language"] != "go" {
		t.Errorf("expected language go, got %v", res["language"])
	}
	cmdLine, ok := res["command_line"].(string)
	if !ok || !strings.Contains(cmdLine, "go test") {
		t.Errorf("expected go test in command_line, got %v", res["command_line"])
	}
	if !strings.Contains(cmdLine, "-cover") {
		t.Errorf("expected -cover in command_line, got %v", cmdLine)
	}
	if !strings.Contains(cmdLine, "TestUnit") {
		t.Errorf("expected TestUnit in command_line, got %v", cmdLine)
	}
}

func TestTestCmd_NodeProject(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name":"myapp","scripts":{"test":"vitest run"}}`), 0644)

	output, err := executeCommand("test", "--json", tempDir)
	if err != nil {
		t.Fatalf("executeCommand(test node) failed: %v", err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal([]byte(output), &res); err != nil {
		t.Fatalf("failed parsing JSON: %v", err)
	}

	if res["language"] != "node" {
		t.Errorf("expected language node, got %v", res["language"])
	}
	if res["command_line"] != "npm test" {
		t.Errorf("expected 'npm test', got %v", res["command_line"])
	}
}
