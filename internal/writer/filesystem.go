package writer

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFiles escribe artefactos en disco bajo outputDir.
// Las claves del mapa son rutas relativas dentro de outputDir.
func WriteFiles(outputDir string, files map[string]string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("crear directorio de salida: %w", err)
	}

	for relPath, content := range files {
		fullPath := filepath.Join(outputDir, relPath)
		parent := filepath.Dir(fullPath)
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("crear directorio %s: %w", parent, err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("escribir archivo %s: %w", fullPath, err)
		}
	}

	return nil
}
