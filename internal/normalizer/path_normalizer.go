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
	endpoint.Parameters = mergePathParams(endpoint.Parameters, endpoint.PathParams)
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

func mergePathParams(params, pathParams []Parameter) []Parameter {
	out := make([]Parameter, len(params))
	copy(out, params)

	index := make(map[string]int, len(out))
	for i, p := range out {
		key := strings.ToLower(strings.TrimSpace(p.In)) + "|" + strings.TrimSpace(p.Name)
		index[key] = i
	}

	for _, pp := range pathParams {
		key := "path|" + pp.Name
		if pos, ok := index[key]; ok {
			out[pos].In = "path"
			out[pos].Required = true
			if strings.TrimSpace(out[pos].Type) == "" {
				out[pos].Type = pp.Type
			}
			continue
		}
		out = append(out, pp)
		index[key] = len(out) - 1
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].In == out[j].In {
			return out[i].Name < out[j].Name
		}
		return out[i].In < out[j].In
	})
	return out
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
	if name == "id_cliente" || name == "idcliente" {
		return "id"
	}
	return name
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
