package renderer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/max/wiki-llm/internal/extractor"
)

func TestRenderAPI_SnapshotOpenAPIExample(t *testing.T) {
	t.Parallel()

	sourcePath := filepath.Join("..", "..", "docs", "openapi.yaml")
	doc, err := extractor.ExtractOpenAPI(sourcePath)
	if err != nil {
		t.Fatalf("extraer openapi de ejemplo: %v", err)
	}

	files := RenderAPI(doc)

	assertGolden(t, files["index.md"], filepath.Join("testdata", "snapshots", "index.md.golden"))
	assertGolden(t, files["apis/api-cif.md"], filepath.Join("testdata", "snapshots", "api-cif.md.golden"))
}

func assertGolden(t *testing.T, got, goldenPath string) {
	t.Helper()

	expectedRaw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("leer golden %s: %v", goldenPath, err)
	}

	expected := string(expectedRaw)
	if got != expected {
		t.Fatalf("snapshot no coincide para %s", goldenPath)
	}
}
