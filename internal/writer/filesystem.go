package writer

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateOutputDir verifica que el directorio de salida se pueda crear y escribir.
func ValidateOutputDir(outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("directorio de salida vacio")
	}

	info, err := os.Stat(outputDir)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("la ruta de salida existe y no es un directorio: %s", outputDir)
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("no se puede inspeccionar directorio de salida: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("crear directorio de salida: %w", err)
	}

	probePath := filepath.Join(outputDir, ".write-check.tmp")
	if err := os.WriteFile(probePath, []byte("ok"), 0o600); err != nil {
		return fmt.Errorf("directorio de salida no es escribible: %w", err)
	}
	if err := os.Remove(probePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("no se pudo limpiar archivo temporal de validacion: %w", err)
	}

	return nil
}

// WriteFiles escribe artefactos en disco bajo outputDir.
// Las claves del mapa son rutas relativas dentro de outputDir.
func WriteFiles(outputDir string, files map[string]string) error {
	if err := ValidateOutputDir(outputDir); err != nil {
		return err
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
