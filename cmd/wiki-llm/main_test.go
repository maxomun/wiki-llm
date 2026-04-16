package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunGenerateAPI_MissingRequiredFlags(t *testing.T) {
	t.Parallel()

	exitCode := run([]string{"generate", "api"})
	if exitCode == 0 {
		t.Fatalf("se esperaba exit code != 0 cuando faltan flags obligatorios")
	}
}

func TestRunGenerateAPI_Success(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.0.3
info:
  title: CLI Test API
  version: 1.0.0
paths:
  /health:
    get:
      operationId: Health
      responses:
        "200":
          description: OK
`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "openapi.yaml")
	outputPath := filepath.Join(tmpDir, "out")

	if err := os.WriteFile(sourcePath, []byte(spec), 0o644); err != nil {
		t.Fatalf("escribir spec temporal: %v", err)
	}

	exitCode := run([]string{
		"generate",
		"api",
		"--source", sourcePath,
		"--output", outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("se esperaba exit code 0 en ejecucion valida, actual=%d", exitCode)
	}

	indexPath := filepath.Join(outputPath, "index.md")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("se esperaba archivo generado %s: %v", indexPath, err)
	}
}
