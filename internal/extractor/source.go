package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/max/wiki-llm/internal/normalizer"
)

const (
	SourceTypeAuto    = "auto"
	SourceTypeOpenAPI = "openapi"
	SourceTypePostman = "postman"
)

// ExtractSource despacha la extraccion segun el tipo de fuente.
func ExtractSource(sourcePath, sourceType string) (normalizer.APIDocument, error) {
	resolvedType := strings.ToLower(strings.TrimSpace(sourceType))
	if resolvedType == "" {
		resolvedType = SourceTypeAuto
	}

	if resolvedType == SourceTypeAuto {
		detected, err := DetectSourceType(sourcePath)
		if err != nil {
			return normalizer.APIDocument{}, err
		}
		resolvedType = detected
	}

	switch resolvedType {
	case SourceTypeOpenAPI:
		return ExtractOpenAPI(sourcePath)
	case SourceTypePostman:
		return ExtractPostmanCollection(sourcePath)
	default:
		return normalizer.APIDocument{}, fmt.Errorf("source-type no soportado: %s", sourceType)
	}
}

// DetectSourceType intenta inferir el tipo de fuente desde el archivo.
func DetectSourceType(sourcePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(sourcePath))
	switch ext {
	case ".yaml", ".yml":
		return SourceTypeOpenAPI, nil
	case ".json":
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			return "", fmt.Errorf("leer source para deteccion: %w", err)
		}
		content := strings.ToLower(string(raw))
		if strings.Contains(content, "schema.getpostman.com/json/collection") {
			return SourceTypePostman, nil
		}
		if strings.Contains(content, "\"openapi\"") {
			return SourceTypeOpenAPI, nil
		}
		return "", fmt.Errorf("no se pudo detectar tipo de source para archivo json")
	default:
		return "", fmt.Errorf("extension no soportada para deteccion automatica: %s", ext)
	}
}
