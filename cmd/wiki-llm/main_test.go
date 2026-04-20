package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateAPI_MissingRequiredFlags(t *testing.T) {
	t.Parallel()

	exitCode := run([]string{"generate", "api", "--output", "./output"})
	if exitCode == 0 {
		t.Fatalf("se esperaba exit code != 0 cuando faltan --source/--code")
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

func TestRunGenerateAPI_InvalidSourcePath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "out")

	exitCode := run([]string{
		"generate",
		"api",
		"--source", filepath.Join(tmpDir, "missing.yaml"),
		"--output", outputPath,
	})
	if exitCode == 0 {
		t.Fatalf("se esperaba error con source inexistente")
	}
}

func TestRunGenerateAPI_InvalidOutputPath(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.0.3
info:
  title: CLI Test API
  version: 1.0.0
paths: {}`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "openapi.yaml")
	outputFile := filepath.Join(tmpDir, "out-file")

	if err := os.WriteFile(sourcePath, []byte(spec), 0o644); err != nil {
		t.Fatalf("escribir spec temporal: %v", err)
	}
	if err := os.WriteFile(outputFile, []byte("no-dir"), 0o644); err != nil {
		t.Fatalf("escribir archivo output temporal: %v", err)
	}

	exitCode := run([]string{
		"generate",
		"api",
		"--source", sourcePath,
		"--output", outputFile,
	})
	if exitCode == 0 {
		t.Fatalf("se esperaba error con output que apunta a archivo")
	}
}

func TestRunGenerateAPI_PostmanSourceType(t *testing.T) {
	t.Parallel()

	collection := `{
  "info": {
    "name": "CLI Postman",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Get Health",
      "request": {
        "method": "GET",
        "url": {
          "path": ["api", "health"]
        }
      },
      "response": []
    }
  ]
}`

	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "collection.json")
	outputPath := filepath.Join(tmpDir, "out")
	if err := os.WriteFile(sourcePath, []byte(collection), 0o644); err != nil {
		t.Fatalf("escribir collection temporal: %v", err)
	}

	exitCode := run([]string{
		"generate",
		"api",
		"--source", sourcePath,
		"--source-type", "postman",
		"--output", outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("se esperaba exit code 0 en source-type postman, actual=%d", exitCode)
	}
}

func TestRunGenerateAPI_MultiSourceMerge(t *testing.T) {
	t.Parallel()

	openapi := `openapi: 3.0.3
info:
  title: Unified API
  version: 1.0.0
paths:
  /health:
    get:
      operationId: Health
      responses:
        "200":
          description: OK
`
	postman := `{
  "info": {
    "name": "Unified API Collection",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Get Ready",
      "request": {
        "method": "GET",
        "url": {
          "path": ["ready"]
        }
      },
      "response": []
    }
  ]
}`

	tmpDir := t.TempDir()
	openapiPath := filepath.Join(tmpDir, "openapi.yaml")
	postmanPath := filepath.Join(tmpDir, "collection.json")
	outputPath := filepath.Join(tmpDir, "out")
	if err := os.WriteFile(openapiPath, []byte(openapi), 0o644); err != nil {
		t.Fatalf("escribir openapi temporal: %v", err)
	}
	if err := os.WriteFile(postmanPath, []byte(postman), 0o644); err != nil {
		t.Fatalf("escribir postman temporal: %v", err)
	}

	exitCode := run([]string{
		"generate", "api",
		"--source", openapiPath,
		"--source", postmanPath,
		"--output", outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("se esperaba exit code 0 en merge multi-source, actual=%d", exitCode)
	}

	apiPath := filepath.Join(outputPath, "apis", "unified-api.md")
	raw, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatalf("leer markdown fusionado: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "## `GET /health`") || !strings.Contains(content, "## `GET /ready`") {
		t.Fatalf("el markdown fusionado no contiene endpoints de ambas fuentes")
	}
}

func TestRunGenerateAPI_CodeOnlyFindsSwagger(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatalf("crear docs dir: %v", err)
	}
	swagger := `{
  "swagger": "2.0",
  "info": {
    "title": "Code API",
    "version": "1.0"
  },
  "paths": {
    "/health": {
      "get": {
        "responses": {
          "200": {
            "description": "OK"
          }
        }
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(docsDir, "swagger.json"), []byte(swagger), 0o644); err != nil {
		t.Fatalf("crear swagger: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "out")
	exitCode := run([]string{
		"generate", "api",
		"--code", tmpDir,
		"--output", outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("se esperaba exit code 0 en modo solo code, actual=%d", exitCode)
	}
	if _, err := os.Stat(filepath.Join(outputPath, "index.md")); err != nil {
		t.Fatalf("se esperaba index.md generado: %v", err)
	}
}
