package normalizer

import (
	"regexp"
	"sort"
	"strings"
)

var (
	uuidPattern        = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	numericPattern     = regexp.MustCompile(`^\d+$`)
	likelyTokenPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{12,}$`)
	versionPattern     = regexp.MustCompile(`^(v\d+|\d+\.\d+)$`)
)

// NormalizeDocument aplica normalizacion avanzada a endpoints.
func NormalizeDocument(doc APIDocument) APIDocument {
	out := doc
	out.Endpoints = make([]Endpoint, 0, len(doc.Endpoints))
	for _, endpoint := range doc.Endpoints {
		out.Endpoints = append(out.Endpoints, normalizeEndpoint(endpoint))
	}
	return out
}

func normalizeEndpoint(endpoint Endpoint) Endpoint {
	endpoint.Method = strings.ToUpper(strings.TrimSpace(endpoint.Method))
	normalizedPath := normalizeDynamicPath(endpoint.Path)
	basePath, endpointPath := splitBaseAndEndpointPath(normalizedPath, endpoint.BasePath)
	endpoint.BasePath = basePath
	endpoint.Path = endpointPath

	endpoint.PathParams = detectPathParams(endpoint.Path)
	endpoint.Parameters = normalizeAndDeduplicateParameters(endpoint.Parameters, endpoint.PathParams)
	endpoint.PathParams = syncPathParamsWithParameters(endpoint.PathParams, endpoint.Parameters)
	endpoint.Sources = dedupeSourceTypes(endpoint.Sources)

	return endpoint
}

func normalizeDynamicPath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return "/"
	}
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		if idx := strings.Index(p, "://"); idx >= 0 {
			tmp := p[idx+3:]
			if slash := strings.Index(tmp, "/"); slash >= 0 {
				p = tmp[slash:]
			} else {
				p = "/"
			}
		}
	}
	p = strings.SplitN(p, "?", 2)[0]
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return "/"
	}

	rawSegments := strings.Split(p, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		s := strings.TrimSpace(segment)
		if s == "" {
			continue
		}
		segments = append(segments, normalizeSegment(s))
	}
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

func normalizeSegment(segment string) string {
	if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
		name := strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
		return "{" + normalizeParamName(name) + "}"
	}
	if strings.HasPrefix(segment, ":") {
		return "{" + normalizeParamName(strings.TrimPrefix(segment, ":")) + "}"
	}
	if strings.Contains(segment, "{{") && strings.Contains(segment, "}}") {
		return "{param}"
	}
	if uuidPattern.MatchString(segment) || numericPattern.MatchString(segment) {
		return "{id}"
	}
	if likelyTokenPattern.MatchString(segment) && containsNumberAndLetter(segment) {
		return "{param}"
	}
	return segment
}

func splitBaseAndEndpointPath(path string, currentBase string) (string, string) {
	normalizedBase := normalizeBasePath(currentBase)
	p := normalizeDynamicPath(path)
	segments := splitPathSegments(p)
	if len(segments) == 0 {
		return normalizedBase, "/"
	}

	if normalizedBase != "" {
		baseSegments := splitPathSegments(normalizedBase)
		if len(baseSegments) > 0 && hasPrefixSegments(segments, baseSegments) {
			rest := segments[len(baseSegments):]
			return normalizedBase, joinPath(rest)
		}
	}

	versionIdx := -1
	for idx, segment := range segments {
		if versionPattern.MatchString(strings.ToLower(segment)) {
			versionIdx = idx
			break
		}
	}
	if versionIdx >= 0 && versionIdx < len(segments)-1 {
		base := joinPath(segments[:versionIdx+1])
		endpoint := joinPath(segments[versionIdx+1:])
		return base, endpoint
	}

	return normalizedBase, joinPath(segments)
}

func detectPathParams(path string) []Parameter {
	segments := splitPathSegments(path)
	out := make([]Parameter, 0)
	for _, seg := range segments {
		if !strings.HasPrefix(seg, "{") || !strings.HasSuffix(seg, "}") {
			continue
		}
		name := normalizeParamName(strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}"))
		if name == "" {
			name = "id"
		}
		out = append(out, Parameter{
			Name:     name,
			In:       "path",
			Required: true,
			Type:     "string",
		})
	}
	out = dedupePathParameters(out)
	return out
}

func normalizeAndDeduplicateParameters(params, pathParams []Parameter) []Parameter {
	out := make([]Parameter, 0, len(params)+len(pathParams))
	index := make(map[string]int, len(params)+len(pathParams))
	placeholderOrder := make([]string, 0, len(pathParams))
	pathParamSet := make(map[string]struct{}, len(pathParams))
	for _, pp := range pathParams {
		placeholderOrder = append(placeholderOrder, pp.Name)
		pathParamSet[pp.Name] = struct{}{}
	}
	aliasToCanonical := make(map[string]string)
	assignedCanonical := make(map[string]struct{})

	for _, p := range params {
		normalized := normalizeParameter(p)
		if normalized.In == "path" {
			normalized.Name = canonicalizePathParamName(
				normalized.Name,
				placeholderOrder,
				pathParamSet,
				aliasToCanonical,
				assignedCanonical,
			)
		}
		key := normalized.In + "|" + normalized.Name
		if pos, ok := index[key]; ok {
			out[pos] = mergeSingleParameter(out[pos], normalized)
			continue
		}
		index[key] = len(out)
		out = append(out, normalized)
	}

	// Garantiza que todo placeholder del path exista en lista de parametros.
	for _, pp := range pathParams {
		key := "path|" + pp.Name
		if pos, ok := index[key]; ok {
			out[pos] = mergeSingleParameter(out[pos], pp)
			out[pos].In = "path"
			out[pos].Required = true
			continue
		}
		index[key] = len(out)
		out = append(out, pp)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].In == out[j].In {
			return out[i].Name < out[j].Name
		}
		return out[i].In < out[j].In
	})
	return out
}

func syncPathParamsWithParameters(pathParams, params []Parameter) []Parameter {
	out := make([]Parameter, 0, len(pathParams))
	for _, pp := range pathParams {
		for _, p := range params {
			if strings.EqualFold(p.In, "path") && p.Name == pp.Name {
				out = append(out, p)
				break
			}
		}
	}
	return dedupePathParameters(out)
}

func dedupePathParameters(params []Parameter) []Parameter {
	set := make(map[string]struct{}, len(params))
	out := make([]Parameter, 0, len(params))
	for _, p := range params {
		key := strings.TrimSpace(p.Name)
		if _, ok := set[key]; ok {
			continue
		}
		set[key] = struct{}{}
		out = append(out, p)
	}
	return out
}

func normalizeParamName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "id"
	}
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	name = replacer.Replace(name)
	return name
}

func normalizeParameter(p Parameter) Parameter {
	p.In = strings.ToLower(strings.TrimSpace(p.In))
	p.Name = normalizeParamName(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.SchemaRef = strings.TrimSpace(p.SchemaRef)
	p.Type = strings.TrimSpace(p.Type)
	p.Format = strings.TrimSpace(p.Format)
	p.Example = strings.TrimSpace(p.Example)

	if p.In == "path" {
		p.Required = true
		if strings.TrimSpace(p.Type) == "" {
			p.Type = "string"
		}
	}
	return p
}

func canonicalizePathParamName(
	name string,
	placeholderOrder []string,
	pathParamSet map[string]struct{},
	aliasToCanonical map[string]string,
	assignedCanonical map[string]struct{},
) string {
	if _, ok := pathParamSet[name]; ok {
		assignedCanonical[name] = struct{}{}
		return name
	}
	if mapped, ok := aliasToCanonical[name]; ok {
		assignedCanonical[mapped] = struct{}{}
		return mapped
	}

	// Estrategia estructural:
	// - si hay un unico placeholder en path, cualquier alias de path se mapea a ese.
	// - si hay multiples placeholders, se asignan en orden de aparicion a placeholders disponibles.
	if len(placeholderOrder) == 1 {
		canonical := placeholderOrder[0]
		aliasToCanonical[name] = canonical
		assignedCanonical[canonical] = struct{}{}
		return canonical
	}
	if len(placeholderOrder) > 1 {
		for _, candidate := range placeholderOrder {
			if _, used := assignedCanonical[candidate]; used {
				continue
			}
			aliasToCanonical[name] = candidate
			assignedCanonical[candidate] = struct{}{}
			return candidate
		}
	}

	// Si no hay placeholders estructurales (caso borde), conservar nombre normalizado.
	return name
}

func mergeSingleParameter(base, incoming Parameter) Parameter {
	base.Required = base.Required || incoming.Required
	if base.Description == "" {
		base.Description = incoming.Description
	}
	if base.SchemaRef == "" {
		base.SchemaRef = incoming.SchemaRef
	}
	if base.Type == "" {
		base.Type = incoming.Type
	}
	if base.Format == "" {
		base.Format = incoming.Format
	}
	if base.Example == "" {
		base.Example = incoming.Example
	}
	return base
}

func normalizeBasePath(path string) string {
	p := normalizeDynamicPath(path)
	if p == "/" {
		return ""
	}
	return p
}

func hasPrefixSegments(path, prefix []string) bool {
	if len(prefix) > len(path) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if path[i] != prefix[i] {
			return false
		}
	}
	return true
}

func splitPathSegments(path string) []string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return "/"
	}
	return "/" + strings.Join(parts, "/")
}

func containsNumberAndLetter(v string) bool {
	hasNum := false
	hasAlpha := false
	for _, r := range v {
		if r >= '0' && r <= '9' {
			hasNum = true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasAlpha = true
		}
		if hasNum && hasAlpha {
			return true
		}
	}
	return false
}

func dedupeSourceTypes(values []SourceType) []SourceType {
	if len(values) == 0 {
		return nil
	}
	set := make(map[SourceType]struct{}, len(values))
	out := make([]SourceType, 0, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := set[v]; ok {
			continue
		}
		set[v] = struct{}{}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}
