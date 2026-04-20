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
	index.WriteString("- BasePath principal: `")
	index.WriteString(nonEmpty(doc.BasePath, "/"))
	index.WriteString("`\n")
	index.WriteString("- Fuente principal del contrato: `")
	index.WriteString(nonEmpty(string(doc.ContractSource), "unknown"))
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
		index.WriteString("| Metodo | BasePath | Path | OperationId | Sources |\n")
		index.WriteString("|---|---|---|---|---|\n")
		for _, endpoint := range sorted {
			index.WriteString("| `")
			index.WriteString(escapeCell(endpoint.Method))
			index.WriteString("` | `")
			index.WriteString(escapeCell(nonEmpty(endpoint.BasePath, "/")))
			index.WriteString("` | `")
			index.WriteString(escapeCell(endpoint.Path))
			index.WriteString("` | `")
			index.WriteString(escapeCell(nonEmpty(endpoint.OperationID, "-")))
			index.WriteString("` | `")
			index.WriteString(escapeCell(nonEmpty(joinSources(endpoint.Sources), "-")))
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
	api.WriteString("BasePath principal: `")
	api.WriteString(nonEmpty(doc.BasePath, "/"))
	api.WriteString("`\n\n")
	api.WriteString("Fuente principal del contrato: `")
	api.WriteString(nonEmpty(string(doc.ContractSource), "unknown"))
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
		api.WriteString("- BasePath: `")
		api.WriteString(nonEmpty(endpoint.BasePath, "/"))
		api.WriteString("`\n")
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
		if len(endpoint.Sources) > 0 {
			api.WriteString("- Sources: `")
			api.WriteString(strings.Join(sourceStrings(endpoint.Sources), "`, `"))
			api.WriteString("`\n")
		}
		if strings.TrimSpace(string(endpoint.Confidence)) != "" {
			api.WriteString("- Confidence: `")
			api.WriteString(string(endpoint.Confidence))
			api.WriteString("`\n")
		}
		if endpoint.Implementation != nil {
			api.WriteString("\n### Implementacion interna\n\n")
			api.WriteString("- Handler: `")
			api.WriteString(nonEmpty(endpoint.Implementation.HandlerName, "-"))
			api.WriteString("`\n")
			api.WriteString("- Archivo: `")
			api.WriteString(nonEmpty(endpoint.Implementation.HandlerFile, "-"))
			api.WriteString("`\n")

			api.WriteString("\n### Dependencias\n\n")
			api.WriteString("- Base de datos: `")
			api.WriteString(boolString(endpoint.Implementation.UsesDatabase))
			api.WriteString("`\n")
			if len(endpoint.Implementation.DatabaseTypes) > 0 {
				api.WriteString("- Tipo(s) BD: `")
				api.WriteString(strings.Join(endpoint.Implementation.DatabaseTypes, "`, `"))
				api.WriteString("`\n")
			}
			if len(endpoint.Implementation.DatabaseTables) > 0 {
				api.WriteString("- Tablas: `")
				api.WriteString(strings.Join(endpoint.Implementation.DatabaseTables, "`, `"))
				api.WriteString("`\n")
			}
			if len(endpoint.Implementation.DatabaseSPs) > 0 {
				api.WriteString("- Stored Procedures: `")
				api.WriteString(strings.Join(endpoint.Implementation.DatabaseSPs, "`, `"))
				api.WriteString("`\n")
			}
			if len(endpoint.Implementation.DatabaseCollections) > 0 {
				api.WriteString("- Collections: `")
				api.WriteString(strings.Join(endpoint.Implementation.DatabaseCollections, "`, `"))
				api.WriteString("`\n")
			}
			if len(endpoint.Implementation.DatabaseQueries) > 0 {
				api.WriteString("- Queries detectadas:\n")
				for _, query := range endpoint.Implementation.DatabaseQueries {
					api.WriteString("  - `")
					api.WriteString(escapeCell(query))
					api.WriteString("`\n")
				}
			}
			api.WriteString("- Mensajeria: `")
			api.WriteString(boolString(endpoint.Implementation.UsesMessaging))
			api.WriteString("`\n")

			if len(endpoint.Implementation.ExternalAPICalls) > 0 {
				api.WriteString("- APIs externas: `")
				api.WriteString(strings.Join(endpoint.Implementation.ExternalAPICalls, "`, `"))
				api.WriteString("`\n")
			} else {
				api.WriteString("- APIs externas: `false`\n")
			}

			api.WriteString("\n### Flujo resumido\n\n")
			for _, step := range summarizeFlow(endpoint) {
				api.WriteString("- ")
				api.WriteString(step)
				api.WriteString("\n")
			}
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
			if endpoint.RequestBody.Example != "" {
				api.WriteString("- Example: `")
				api.WriteString(escapeCell(endpoint.RequestBody.Example))
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
				if resp.Example != "" {
					api.WriteString("|  | Example | `")
					api.WriteString(escapeCell(resp.Example))
					api.WriteString("` |  |\n")
				}
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

func sourceStrings(values []normalizer.SourceType) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, string(v))
	}
	return out
}

func joinSources(values []normalizer.SourceType) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(sourceStrings(values), ", ")
}

func summarizeFlow(endpoint normalizer.Endpoint) []string {
	impl := endpoint.Implementation
	if impl == nil {
		return nil
	}
	steps := make([]string, 0, 10)
	if strings.TrimSpace(impl.HandlerName) != "" {
		steps = append(steps, "Enruta la solicitud al handler: "+impl.HandlerName)
	} else {
		steps = append(steps, "Resuelve el handler y valida parametros de entrada")
	}

	operation := inferOperationType(endpoint.Method, impl.DatabaseQueries)
	entities := inferEntities(impl.DatabaseTables, impl.DatabaseCollections)
	steps = append(steps, buildSemanticOperationStep(operation, entities))

	if len(impl.ServiceCalls) > 0 {
		steps = append(steps, "Orquesta logica de aplicacion via servicios: "+joinLimited(impl.ServiceCalls, 4))
	}
	if len(impl.RepositoryCalls) > 0 {
		steps = append(steps, "Consulta repositorios: "+joinLimited(impl.RepositoryCalls, 4))
	}
	if impl.UsesDatabase {
		details := make([]string, 0, 4)
		if len(impl.DatabaseTypes) > 0 {
			details = append(details, "tipo(s): "+strings.Join(impl.DatabaseTypes, ", "))
		}
		if len(impl.DatabaseTables) > 0 {
			details = append(details, "tablas: "+joinLimited(impl.DatabaseTables, 3))
		}
		if len(impl.DatabaseSPs) > 0 {
			details = append(details, "stored procedures: "+joinLimited(impl.DatabaseSPs, 2))
		}
		if len(impl.DatabaseCollections) > 0 {
			details = append(details, "collections: "+joinLimited(impl.DatabaseCollections, 2))
		}

		dbStep := "Ejecuta operaciones de base de datos"
		if len(details) > 0 {
			dbStep += " (" + strings.Join(details, "; ") + ")"
		}
		steps = append(steps, dbStep)

		if len(impl.DatabaseQueries) > 0 {
			querySummary := summarizeSQLOperations(impl.DatabaseQueries)
			steps = append(steps, "Aplica consultas SQL detectadas ("+querySummary+"): "+joinLimited(impl.DatabaseQueries, 1))
		}
	}
	if len(impl.ExternalAPICalls) > 0 {
		steps = append(steps, "Consume APIs externas: "+joinLimited(impl.ExternalAPICalls, 4))
	}
	if impl.UsesMessaging {
		steps = append(steps, "Publica eventos o mensajes")
	}
	steps = append(steps, "Retorna respuesta al cliente")
	return steps
}

func joinLimited(values []string, max int) string {
	if len(values) == 0 || max <= 0 {
		return ""
	}
	if len(values) <= max {
		return strings.Join(values, ", ")
	}
	return strings.Join(values[:max], ", ") + fmt.Sprintf(" (+%d mas)", len(values)-max)
}

func inferOperationType(method string, queries []string) string {
	sqlOps := extractSQLOperations(queries)
	if len(sqlOps) > 0 {
		return strings.Join(sqlOps, " y ")
	}
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD":
		return "consulta"
	case "POST":
		return "creacion"
	case "PUT", "PATCH":
		return "actualizacion"
	case "DELETE":
		return "eliminacion"
	default:
		return "procesamiento"
	}
}

func extractSQLOperations(queries []string) []string {
	ops := make([]string, 0, 4)
	seen := make(map[string]struct{})
	add := func(op string) {
		if op == "" {
			return
		}
		if _, ok := seen[op]; ok {
			return
		}
		seen[op] = struct{}{}
		ops = append(ops, op)
	}
	for _, q := range queries {
		clean := strings.ToUpper(strings.TrimSpace(q))
		switch {
		case strings.HasPrefix(clean, "SELECT "):
			add("consulta")
		case strings.HasPrefix(clean, "INSERT "):
			add("creacion")
		case strings.HasPrefix(clean, "UPDATE "):
			add("actualizacion")
		case strings.HasPrefix(clean, "DELETE "):
			add("eliminacion")
		case strings.HasPrefix(clean, "EXEC ") || strings.HasPrefix(clean, "EXECUTE "):
			add("ejecucion")
		}
	}
	return ops
}

func summarizeSQLOperations(queries []string) string {
	ops := extractSQLOperations(queries)
	if len(ops) == 0 {
		return fmt.Sprintf("%d query(s)", len(queries))
	}
	return strings.Join(ops, "/") + fmt.Sprintf(", %d query(s)", len(queries))
}

func inferEntities(tables, collections []string) []string {
	out := make([]string, 0, len(tables)+len(collections))
	seen := make(map[string]struct{}, len(tables)+len(collections))
	add := func(v string) {
		item := normalizeEntityName(v)
		if item == "" {
			return
		}
		if _, ok := seen[item]; ok {
			return
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	for _, t := range tables {
		add(t)
	}
	for _, c := range collections {
		add(c)
	}
	sort.Strings(out)
	return out
}

func normalizeEntityName(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return ""
	}
	if idx := strings.LastIndex(v, "."); idx >= 0 && idx+1 < len(v) {
		v = v[idx+1:]
	}
	v = strings.TrimPrefix(v, "tb_")
	v = strings.TrimPrefix(v, "tbl_")
	return strings.TrimSpace(v)
}

func buildSemanticOperationStep(operation string, entities []string) string {
	op := strings.TrimSpace(operation)
	if op == "" {
		op = "procesamiento"
	}
	if len(entities) == 0 {
		return "Ejecuta una operacion de " + op + " sobre las entidades detectadas"
	}
	return "Ejecuta una operacion de " + op + " sobre entidades: " + joinLimited(entities, 3)
}
