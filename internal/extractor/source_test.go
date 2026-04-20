package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectSourceType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	postmanPath := filepath.Join(tmpDir, "collection.json")
	openAPIPath := filepath.Join(tmpDir, "openapi.yaml")
	swaggerPath := filepath.Join(tmpDir, "swagger.json")

	if err := os.WriteFile(postmanPath, []byte(`{"info":{"schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"}}`), 0o644); err != nil {
		t.Fatalf("escribir postman temporal: %v", err)
	}
	if err := os.WriteFile(openAPIPath, []byte("openapi: 3.0.3\ninfo: {}\npaths: {}\n"), 0o644); err != nil {
		t.Fatalf("escribir openapi temporal: %v", err)
	}
	if err := os.WriteFile(swaggerPath, []byte(`{"swagger":"2.0","info":{},"paths":{}}`), 0o644); err != nil {
		t.Fatalf("escribir swagger temporal: %v", err)
	}

	postmanType, err := DetectSourceType(postmanPath)
	if err != nil {
		t.Fatalf("detectar postman: %v", err)
	}
	if postmanType != SourceTypePostman {
		t.Fatalf("tipo postman inesperado: %s", postmanType)
	}

	openAPIType, err := DetectSourceType(openAPIPath)
	if err != nil {
		t.Fatalf("detectar openapi: %v", err)
	}
	if openAPIType != SourceTypeOpenAPI {
		t.Fatalf("tipo openapi inesperado: %s", openAPIType)
	}

	swaggerType, err := DetectSourceType(swaggerPath)
	if err != nil {
		t.Fatalf("detectar swagger: %v", err)
	}
	if swaggerType != SourceTypeOpenAPI {
		t.Fatalf("tipo swagger inesperado: %s", swaggerType)
	}
}
