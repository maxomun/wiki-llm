package discoverer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/max/wiki-llm/internal/extractor"
)

// SourceInput representa una fuente resuelta para el pipeline.
type SourceInput struct {
	Path       string
	SourceType string
}

// ResolveGenerateAPISources resuelve fuentes explicitas y descubrimiento por --code.
func ResolveGenerateAPISources(codeRoot string, explicitSources []string, sourceType string) ([]SourceInput, []string, error) {
	logs := make([]string, 0)
	resolved := make([]SourceInput, 0, len(explicitSources)+1)

	explicitHasOpenAPI := false
	for _, source := range explicitSources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		effectiveType := strings.ToLower(strings.TrimSpace(sourceType))
		if effectiveType == "" {
			effectiveType = extractor.SourceTypeAuto
		}
		if effectiveType == extractor.SourceTypeAuto {
			detected, err := extractor.DetectSourceType(source)
			if err != nil {
				return nil, nil, fmt.Errorf("detectar tipo de source %s: %w", source, err)
			}
			effectiveType = detected
		}
		if effectiveType == extractor.SourceTypeOpenAPI {
			explicitHasOpenAPI = true
			logs = append(logs, fmt.Sprintf("usando OpenAPI desde --source: %s", source))
		}
		if effectiveType == extractor.SourceTypePostman {
			logs = append(logs, fmt.Sprintf("usando Postman como complemento: %s", source))
		}
		resolved = append(resolved, SourceInput{Path: source, SourceType: effectiveType})
	}

	if explicitHasOpenAPI {
		return resolved, logs, nil
	}

	if strings.TrimSpace(codeRoot) == "" {
		if len(resolved) == 0 {
			return nil, nil, fmt.Errorf("se requiere al menos --source o --code")
		}
		return resolved, logs, nil
	}

	logs = append(logs, fmt.Sprintf("no se entrego OpenAPI explicito, buscando swagger.json en proyecto: %s", codeRoot))
	swaggerPath, err := FindSwaggerJSON(codeRoot)
	if err != nil {
		return nil, nil, err
	}
	logs = append(logs, fmt.Sprintf("swagger.json encontrado en: %s", swaggerPath))

	// Swagger siempre se agrega como contrato base al inicio.
	out := make([]SourceInput, 0, len(resolved)+1)
	out = append(out, SourceInput{Path: swaggerPath, SourceType: extractor.SourceTypeOpenAPI})
	out = append(out, resolved...)
	return out, logs, nil
}

// FindSwaggerJSON busca recursivamente un swagger.json priorizando:
// 1) docs/swagger.json
// 2) swagger.json en raiz
// 3) cualquier otro encontrado
func FindSwaggerJSON(codeRoot string) (string, error) {
	root := strings.TrimSpace(codeRoot)
	if root == "" {
		return "", fmt.Errorf("ruta --code vacia")
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("no se puede acceder a --code: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("--code debe ser un directorio: %s", root)
	}

	candidates := make([]string, 0)
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "swagger.json") {
			candidates = append(candidates, path)
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("error buscando swagger.json: %w", err)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no se encontro swagger.json dentro de --code: %s", root)
	}

	sort.Slice(candidates, func(i, j int) bool {
		ri := swaggerPriority(root, candidates[i])
		rj := swaggerPriority(root, candidates[j])
		if ri == rj {
			return candidates[i] < candidates[j]
		}
		return ri < rj
	})

	return candidates[0], nil
}

func swaggerPriority(root, absPath string) int {
	rel, err := filepath.Rel(root, absPath)
	if err != nil {
		return 99
	}
	rel = filepath.ToSlash(strings.ToLower(rel))
	switch rel {
	case "docs/swagger.json":
		return 0
	case "swagger.json":
		return 1
	default:
		if strings.HasSuffix(rel, "/docs/swagger.json") {
			return 2
		}
		return 3
	}
}
