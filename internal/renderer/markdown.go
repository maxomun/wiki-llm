package renderer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/max/wiki-llm/internal/normalizer"
)

// RenderAPI convierte un APIDocument en artefactos Markdown.
// Devuelve un mapa ruta-relativa -> contenido.
func RenderAPI(doc normalizer.APIDocument) map[string]string {
	files := make(map[string]string)
	apiFileName := fmt.Sprintf("apis/%s.md", slugify(doc.Title))
	sorted := sortedEndpoints(doc.Endpoints)

	var index strings.Builder
	index.WriteString("# ")
	index.WriteString(nonEmpty(doc.Title, "API"))
	index.WriteString("\n\n")
	index.WriteString("- Version: `")
	index.WriteString(nonEmpty(doc.Version, "n/a"))
	index.WriteString("`\n")
	index.WriteString("- Endpoints: `")
	index.WriteString(fmt.Sprintf("%d", len(doc.Endpoints)))
	index.WriteString("`\n")
	index.WriteString("- Fuente: `")
	index.WriteString(doc.SourcePath)
	index.WriteString("`\n\n")
	index.WriteString("## Documentacion\n\n")
	index.WriteString("- [Detalle de la API](")
	index.WriteString(apiFileName)
	index.WriteString(")\n")
	if len(sorted) > 0 {
		index.WriteString("\n## Endpoints\n\n")
		index.WriteString("| Metodo | Path | OperationId |\n")
		index.WriteString("|---|---|---|\n")
		for _, endpoint := range sorted {
			index.WriteString("| `")
			index.WriteString(escapeCell(endpoint.Method))
			index.WriteString("` | `")
			index.WriteString(escapeCell(endpoint.Path))
			index.WriteString("` | `")
			index.WriteString(escapeCell(nonEmpty(endpoint.OperationID, "-")))
			index.WriteString("` |\n")
		}
	}
	files["index.md"] = index.String()

	var api strings.Builder
	api.WriteString("# ")
	api.WriteString(nonEmpty(doc.Title, "API"))
	api.WriteString("\n\n")
	if doc.Description != "" {
		api.WriteString(doc.Description)
		api.WriteString("\n\n")
	}
	api.WriteString("Version: `")
	api.WriteString(nonEmpty(doc.Version, "n/a"))
	api.WriteString("`\n\n")

	api.WriteString("## Tabla de contenidos\n\n")
	for _, endpoint := range sorted {
		api.WriteString("- [")
		api.WriteString(endpoint.Method)
		api.WriteString(" ")
		api.WriteString(endpoint.Path)
		api.WriteString("](#")
		api.WriteString(anchorForEndpoint(endpoint))
		api.WriteString(")\n")
	}
	api.WriteString("\n")

	for _, endpoint := range sorted {
		api.WriteString("## `")
		api.WriteString(endpoint.Method)
		api.WriteString(" ")
		api.WriteString(endpoint.Path)
		api.WriteString("`\n\n")
		if endpoint.OperationID != "" {
			api.WriteString("- OperationId: `")
			api.WriteString(endpoint.OperationID)
			api.WriteString("`\n")
		}
		if endpoint.Summary != "" {
			api.WriteString("- Summary: ")
			api.WriteString(endpoint.Summary)
			api.WriteString("\n")
		}
		if endpoint.Description != "" {
			api.WriteString("- Description: ")
			api.WriteString(endpoint.Description)
			api.WriteString("\n")
		}
		if len(endpoint.Tags) > 0 {
			api.WriteString("- Tags: `")
			api.WriteString(strings.Join(endpoint.Tags, "`, `"))
			api.WriteString("`\n")
		}
		if endpoint.Deprecated {
			api.WriteString("- Deprecated: `true`\n")
		}
		if len(endpoint.SecurityRefs) > 0 {
			api.WriteString("- Security: `")
			api.WriteString(strings.Join(endpoint.SecurityRefs, "`, `"))
			api.WriteString("`\n")
		}

		if len(endpoint.Parameters) > 0 {
			api.WriteString("\n### Parametros\n\n")
			api.WriteString("| Nombre | In | Requerido | Tipo | Formato | Schema | Descripcion |\n")
			api.WriteString("|---|---|---|---|---|---|---|\n")
			for _, p := range endpoint.Parameters {
				api.WriteString("| `")
				api.WriteString(escapeCell(nonEmpty(p.Name, "-")))
				api.WriteString("` | `")
				api.WriteString(escapeCell(nonEmpty(p.In, "-")))
				api.WriteString("` | `")
				api.WriteString(boolString(p.Required))
				api.WriteString("` | `")
				api.WriteString(escapeCell(nonEmpty(p.Type, "-")))
				api.WriteString("` | `")
				api.WriteString(escapeCell(nonEmpty(p.Format, "-")))
				api.WriteString("` | `")
				api.WriteString(escapeCell(nonEmpty(p.SchemaRef, "-")))
				api.WriteString("` | ")
				api.WriteString(escapeCell(nonEmpty(p.Description, "-")))
				api.WriteString(" |\n")
			}
		}

		if endpoint.RequestBody != nil {
			api.WriteString("\n### Request Body\n\n")
			api.WriteString("- Required: `")
			api.WriteString(boolString(endpoint.RequestBody.Required))
			api.WriteString("`\n")
			if endpoint.RequestBody.Description != "" {
				api.WriteString("- Description: ")
				api.WriteString(endpoint.RequestBody.Description)
				api.WriteString("\n")
			}
			if len(endpoint.RequestBody.ContentTypes) > 0 {
				api.WriteString("- Content Types: `")
				api.WriteString(strings.Join(endpoint.RequestBody.ContentTypes, "`, `"))
				api.WriteString("`\n")
			}
			if endpoint.RequestBody.SchemaRef != "" {
				api.WriteString("- Schema: `")
				api.WriteString(endpoint.RequestBody.SchemaRef)
				api.WriteString("`\n")
			}
		}

		if len(endpoint.Responses) > 0 {
			api.WriteString("\n### Responses\n\n")
			api.WriteString("| Status | Descripcion | Content Types | Schema |\n")
			api.WriteString("|---|---|---|---|\n")
			for _, resp := range endpoint.Responses {
				api.WriteString("| `")
				api.WriteString(escapeCell(nonEmpty(resp.StatusCode, "-")))
				api.WriteString("` | ")
				api.WriteString(escapeCell(nonEmpty(resp.Description, "sin descripcion")))
				api.WriteString(" | ")
				api.WriteString(escapeCell(nonEmpty(strings.Join(resp.ContentTypes, ", "), "-")))
				api.WriteString(" | `")
				api.WriteString(escapeCell(nonEmpty(resp.SchemaRef, "-")))
				api.WriteString("` |\n")
			}
		}
		api.WriteString("\n")
	}
	files[apiFileName] = api.String()

	return files
}

func sortedEndpoints(endpoints []normalizer.Endpoint) []normalizer.Endpoint {
	out := make([]normalizer.Endpoint, len(endpoints))
	copy(out, endpoints)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Method < out[j].Method
		}
		return out[i].Path < out[j].Path
	})
	return out
}

func slugify(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return "api"
	}
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-")
	s = replacer.Replace(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "api"
	}
	return out
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func escapeCell(v string) string {
	replacer := strings.NewReplacer("|", "\\|", "\n", " ", "\r", " ")
	return strings.TrimSpace(replacer.Replace(v))
}

func anchorForEndpoint(endpoint normalizer.Endpoint) string {
	raw := strings.ToLower(endpoint.Method + "-" + endpoint.Path)
	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
