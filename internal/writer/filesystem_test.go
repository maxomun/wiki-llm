package writer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFiles_WritesNestedArtifacts(t *testing.T) {
	t.Parallel()

	outputDir := filepath.Join(t.TempDir(), "output")
	files := map[string]string{
		"index.md":      "# Index\n",
		"apis/api-a.md": "# API A\n",
	}

	if err := WriteFiles(outputDir, files); err != nil {
		t.Fatalf("WriteFiles devolvio error: %v", err)
	}

	indexPath := filepath.Join(outputDir, "index.md")
	apiPath := filepath.Join(outputDir, "apis", "api-a.md")

	indexRaw, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("leer index.md: %v", err)
	}
	apiRaw, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatalf("leer apis/api-a.md: %v", err)
	}

	if string(indexRaw) != "# Index\n" {
		t.Fatalf("contenido index.md inesperado: %q", string(indexRaw))
	}
	if string(apiRaw) != "# API A\n" {
		t.Fatalf("contenido api-a.md inesperado: %q", string(apiRaw))
	}
}

func TestValidateOutputDir_FailsWhenPathIsFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "no-dir.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("crear archivo temporal: %v", err)
	}

	if err := ValidateOutputDir(filePath); err == nil {
		t.Fatalf("se esperaba error cuando output es archivo")
	}
}
