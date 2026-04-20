package sourcecode

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/max/wiki-llm/internal/normalizer"
)

var routePattern = regexp.MustCompile(`(?i)\.(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s*\(\s*([^,\n]+)\s*,\s*([^\)\n]+)\)`)
var funcPattern = regexp.MustCompile(`func\s*(\([^\)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
var routePlaceholderPattern = regexp.MustCompile(`\{[^/]+\}`)
var quotedStringPattern = regexp.MustCompile("\"([^\"\\\\]*(\\\\.[^\"\\\\]*)*)\"|`([^`]*)`")
var whitespacePattern = regexp.MustCompile(`\s+`)

const maxCallChainDepth = 6

// EnrichDocument enriquece endpoints con metadata de implementacion.
func EnrichDocument(doc normalizer.APIDocument, codeRoot string) (normalizer.APIDocument, error) {
	root := strings.TrimSpace(codeRoot)
	if root == "" {
		return doc, nil
	}
	info, err := os.Stat(root)
	if err != nil {
		return doc, fmt.Errorf("acceder a --code: %w", err)
	}
	if !info.IsDir() {
		return doc, fmt.Errorf("--code debe ser directorio: %s", root)
	}

	goFiles, err := collectGoFiles(root)
	if err != nil {
		return doc, err
	}
	routes, err := scanRoutes(goFiles)
	if err != nil {
		return doc, err
	}
	handlers, err := scanHandlers(goFiles)
	if err != nil {
		return doc, err
	}

	out := doc
	out.Endpoints = make([]normalizer.Endpoint, 0, len(doc.Endpoints))
	for _, endpoint := range doc.Endpoints {
		ref, ok := matchRouteRef(routes, endpoint)
		impl := (*normalizer.ImplementationInfo)(nil)

		if ok {
			combined := aggregateHandlerAnalysis(ref.HandlerName, handlers, map[string]bool{}, 0, maxCallChainDepth)
			item := normalizer.ImplementationInfo{
				HandlerName: ref.HandlerName,
				HandlerFile: ref.HandlerFile,
			}
			applyHandlerAnalysis(&item, combined)
			impl = &item
		} else if analyses, found := handlers[endpoint.OperationID]; found && len(analyses) > 0 {
			combined := aggregateHandlerAnalysis(endpoint.OperationID, handlers, map[string]bool{}, 0, maxCallChainDepth)
			item := normalizer.ImplementationInfo{
				HandlerName: endpoint.OperationID,
				HandlerFile: analyses[0].HandlerFile,
			}
			applyHandlerAnalysis(&item, combined)
			impl = &item
		}

		endpoint.Implementation = impl
		if impl != nil {
			endpoint.Sources = appendSourceIfMissing(endpoint.Sources, normalizer.SourceCode)
		}
		out.Endpoints = append(out.Endpoints, endpoint)
	}
	return out, nil
}

type routeRef struct {
	HandlerName string
	HandlerFile string
}

type handlerAnalysis struct {
	HandlerFile         string
	InvokedMethods      []string
	ServiceCalls        []string
	RepositoryCalls     []string
	ExternalAPICalls    []string
	UsesDatabase        bool
	DatabaseTypes       []string
	DatabaseTables      []string
	DatabaseSPs         []string
	DatabaseCollections []string
	DatabaseQueries     []string
	UsesMessaging       bool
}

func collectGoFiles(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".go") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("recorrer codigo fuente: %w", err)
	}
	sort.Strings(paths)
	return paths, nil
}

func scanRoutes(paths []string) (map[string]routeRef, error) {
	routes := make(map[string]routeRef)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("leer archivo go %s: %w", path, err)
		}
		matches := routePattern.FindAllStringSubmatch(string(raw), -1)
		for _, m := range matches {
			method := strings.ToUpper(strings.TrimSpace(m[1]))
			routePath := normalizeRoutePath(extractRoutePath(m[2]))
			handlerName := parseHandlerName(m[3])
			if handlerName == "" || routePath == "" {
				continue
			}
			key := normalizeRouteKey(method, routePath)
			if _, exists := routes[key]; !exists {
				routes[key] = routeRef{HandlerName: handlerName, HandlerFile: path}
			}
		}
	}
	return routes, nil
}

func scanHandlers(paths []string) (map[string][]handlerAnalysis, error) {
	out := make(map[string][]handlerAnalysis)
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("leer archivo go %s: %w", path, err)
		}
		content := string(raw)
		locs := funcPattern.FindAllStringSubmatchIndex(content, -1)
		for _, loc := range locs {
			name := content[loc[4]:loc[5]]
			start := loc[1]
			body, ok := extractFunctionBody(content, start)
			if !ok {
				continue
			}
			analysis := analyzeHandlerBody(body)
			analysis.HandlerFile = path
			out[name] = append(out[name], analysis)
		}
	}
	return out, nil
}

func aggregateHandlerAnalysis(entry string, handlers map[string][]handlerAnalysis, visited map[string]bool, depth, maxDepth int) handlerAnalysis {
	name := strings.TrimSpace(entry)
	if name == "" || depth > maxDepth || visited[name] {
		return handlerAnalysis{}
	}
	analyses, ok := handlers[name]
	if !ok || len(analyses) == 0 {
		return handlerAnalysis{}
	}
	visited[name] = true

	combined := handlerAnalysis{}
	for _, analysis := range analyses {
		combined = mergeAnalyses(combined, analysis)
	}
	for _, method := range combined.InvokedMethods {
		child := aggregateHandlerAnalysis(method, handlers, visited, depth+1, maxDepth)
		combined = mergeAnalyses(combined, child)
	}
	return combined
}

func mergeAnalyses(base, incoming handlerAnalysis) handlerAnalysis {
	base.InvokedMethods = mergeStringSlices(base.InvokedMethods, incoming.InvokedMethods)
	base.ServiceCalls = mergeStringSlices(base.ServiceCalls, incoming.ServiceCalls)
	base.RepositoryCalls = mergeStringSlices(base.RepositoryCalls, incoming.RepositoryCalls)
	base.ExternalAPICalls = mergeStringSlices(base.ExternalAPICalls, incoming.ExternalAPICalls)
	base.DatabaseTypes = mergeStringSlices(base.DatabaseTypes, incoming.DatabaseTypes)
	base.DatabaseTables = mergeStringSlices(base.DatabaseTables, incoming.DatabaseTables)
	base.DatabaseSPs = mergeStringSlices(base.DatabaseSPs, incoming.DatabaseSPs)
	base.DatabaseCollections = mergeStringSlices(base.DatabaseCollections, incoming.DatabaseCollections)
	base.DatabaseQueries = mergeStringSlices(base.DatabaseQueries, incoming.DatabaseQueries)
	base.UsesDatabase = base.UsesDatabase || incoming.UsesDatabase
	base.UsesMessaging = base.UsesMessaging || incoming.UsesMessaging
	if strings.TrimSpace(base.HandlerFile) == "" {
		base.HandlerFile = incoming.HandlerFile
	}
	return base
}

func mergeStringSlices(base, incoming []string) []string {
	set := make(map[string]struct{}, len(base)+len(incoming))
	out := make([]string, 0, len(base)+len(incoming))
	add := func(v string) {
		item := strings.TrimSpace(v)
		if item == "" {
			return
		}
		if _, ok := set[item]; ok {
			return
		}
		set[item] = struct{}{}
		out = append(out, item)
	}
	for _, v := range base {
		add(v)
	}
	for _, v := range incoming {
		add(v)
	}
	sort.Strings(out)
	return out
}

func applyHandlerAnalysis(impl *normalizer.ImplementationInfo, analysis handlerAnalysis) {
	impl.ServiceCalls = analysis.ServiceCalls
	impl.RepositoryCalls = analysis.RepositoryCalls
	impl.ExternalAPICalls = analysis.ExternalAPICalls
	impl.UsesDatabase = analysis.UsesDatabase
	impl.DatabaseTypes = analysis.DatabaseTypes
	impl.DatabaseTables = analysis.DatabaseTables
	impl.DatabaseSPs = analysis.DatabaseSPs
	impl.DatabaseCollections = analysis.DatabaseCollections
	impl.DatabaseQueries = analysis.DatabaseQueries
	impl.UsesMessaging = analysis.UsesMessaging
}

func analyzeHandlerBody(body string) handlerAnalysis {
	analysis := handlerAnalysis{
		InvokedMethods:   findInvokedMethods(body),
		ServiceCalls:     findCalls(body, `(?i)\b\w*service\w*\.\w+\s*\(`),
		RepositoryCalls:  findCalls(body, `(?i)\b\w*(repo|repository)\w*\.\w+\s*\(`),
		ExternalAPICalls: findCalls(body, `(?i)\b(http\.\w+|resty\.\w+|\w+\.Do)\s*\(`),
	}
	lower := strings.ToLower(body)
	analysis.UsesDatabase = strings.Contains(lower, "db.") ||
		strings.Contains(lower, ".query(") ||
		strings.Contains(lower, ".queryrow(") ||
		strings.Contains(lower, ".querycontext(") ||
		strings.Contains(lower, ".queryrowcontext(") ||
		strings.Contains(lower, ".exec(") ||
		strings.Contains(lower, "gorm") ||
		strings.Contains(lower, "mongo")
	analysis.UsesMessaging = strings.Contains(lower, ".publish(") ||
		strings.Contains(lower, "sendmessage") ||
		strings.Contains(lower, ".produce(") ||
		strings.Contains(lower, "kafka")

	analysis.DatabaseTypes = detectDatabaseTypes(lower)
	analysis.DatabaseQueries = findDatabaseQueries(body)
	analysis.DatabaseTables = findTableNames(body, analysis.DatabaseQueries)
	analysis.DatabaseSPs = findStoredProcedures(body)
	analysis.DatabaseCollections = findCollections(body)

	if len(analysis.DatabaseTypes) > 0 || len(analysis.DatabaseTables) > 0 || len(analysis.DatabaseSPs) > 0 || len(analysis.DatabaseCollections) > 0 || len(analysis.DatabaseQueries) > 0 {
		analysis.UsesDatabase = true
	}
	return analysis
}

func detectDatabaseTypes(lowerBody string) []string {
	types := make([]string, 0, 3)
	add := func(v string) {
		for _, t := range types {
			if t == v {
				return
			}
		}
		types = append(types, v)
	}

	if strings.Contains(lowerBody, "mongo") || strings.Contains(lowerBody, ".collection(") {
		add("mongodb")
	}
	if strings.Contains(lowerBody, "gorm") {
		add("gorm")
	}
	if strings.Contains(lowerBody, "db.") ||
		strings.Contains(lowerBody, ".query(") ||
		strings.Contains(lowerBody, ".queryrow(") ||
		strings.Contains(lowerBody, ".querycontext(") ||
		strings.Contains(lowerBody, ".queryrowcontext(") ||
		strings.Contains(lowerBody, ".exec(") ||
		strings.Contains(lowerBody, ".raw(") {
		add("sql")
	}
	sort.Strings(types)
	return types
}

func findTableNames(body string, queries []string) []string {
	set := make(map[string]struct{})
	out := make([]string, 0)
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		if _, ok := set[v]; ok {
			return
		}
		set[v] = struct{}{}
		out = append(out, v)
	}

	// 1) Tablas inferidas desde SQL detectado.
	sqlTableRe := regexp.MustCompile(`(?i)\b(from|join|into|update|table)\s+([a-zA-Z_][a-zA-Z0-9_\.]*)`)
	for _, q := range queries {
		matches := sqlTableRe.FindAllStringSubmatch(q, -1)
		for _, m := range matches {
			if len(m) > 2 {
				add(strings.ToLower(m[2]))
			}
		}
	}

	// 2) Tablas inferidas desde ORM/table helpers.
	tableCallRe := regexp.MustCompile(`(?i)\.table\(\s*"([^"]+)"\s*\)`)
	for _, m := range tableCallRe.FindAllStringSubmatch(body, -1) {
		if len(m) > 1 {
			add(strings.ToLower(strings.TrimSpace(m[1])))
		}
	}

	sort.Strings(out)
	return out
}

func findStoredProcedures(body string) []string {
	re := regexp.MustCompile(`(?i)\bexec(?:ute)?\s+([a-zA-Z_][a-zA-Z0-9_\.]*)`)
	matches := re.FindAllStringSubmatch(body, -1)
	return dedupeCapture(matches, 1)
}

func findCollections(body string) []string {
	re := regexp.MustCompile(`(?i)\.collection\(\s*"([^"]+)"\s*\)`)
	matches := re.FindAllStringSubmatch(body, -1)
	return dedupeCapture(matches, 1)
}

func findDatabaseQueries(body string) []string {
	re := regexp.MustCompile(`(?is)\.(Query|QueryRow|QueryContext|QueryRowContext|Exec|Raw)\s*\(\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	addQuery := func(raw string) {
		query := normalizeQuerySnippet(raw)
		if query == "" {
			return
		}
		if _, ok := seen[query]; ok {
			return
		}
		seen[query] = struct{}{}
		out = append(out, query)
	}

	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		addQuery(m[2])
	}

	// Captura SQL en variables/literales multilinea (ej: qry := `UPDATE ...`).
	for _, q := range extractSQLLiterals(body) {
		addQuery(q)
	}

	sort.Strings(out)
	return out
}

func extractSQLLiterals(body string) []string {
	out := make([]string, 0)

	backtick := regexp.MustCompile("(?s)`([^`]*)`")
	for _, m := range backtick.FindAllStringSubmatch(body, -1) {
		if len(m) < 2 {
			continue
		}
		if looksLikeSQL(m[1]) {
			out = append(out, m[1])
		}
	}

	quoted := regexp.MustCompile(`"([^"\\]*(\\.[^"\\]*)*)"`)
	for _, m := range quoted.FindAllStringSubmatch(body, -1) {
		if len(m) < 2 {
			continue
		}
		if looksLikeSQL(m[1]) {
			out = append(out, m[1])
		}
	}

	return out
}

func looksLikeSQL(v string) bool {
	l := strings.ToLower(strings.TrimSpace(v))
	l = whitespacePattern.ReplaceAllString(l, " ")
	return strings.Contains(l, "select ") ||
		strings.Contains(l, "insert into ") ||
		strings.Contains(l, "update ") ||
		strings.Contains(l, "delete from ") ||
		strings.Contains(l, " from ")
}

func extractRoutePath(routeExpr string) string {
	expr := strings.TrimSpace(routeExpr)
	if expr == "" {
		return ""
	}
	segments := quotedStringPattern.FindAllStringSubmatch(expr, -1)
	if len(segments) == 0 {
		return expr
	}
	parts := make([]string, 0, len(segments))
	for _, m := range segments {
		part := strings.TrimSpace(firstNonEmptyString(m[1], m[3]))
		if part == "" {
			continue
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return expr
	}

	pathParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.Contains(part, "/") || strings.HasPrefix(part, ":") {
			pathParts = append(pathParts, part)
		}
	}
	if len(pathParts) == 0 {
		return parts[len(parts)-1]
	}
	joined := strings.Join(pathParts, "")
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func normalizeQuerySnippet(raw string) string {
	query := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	if query == "" {
		return ""
	}
	if len(query) > 180 {
		query = query[:177] + "..."
	}
	return query
}

func dedupeCapture(matches [][]string, group int) []string {
	set := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) <= group {
			continue
		}
		v := strings.TrimSpace(m[group])
		if v == "" {
			continue
		}
		if _, ok := set[v]; ok {
			continue
		}
		set[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func findCalls(body string, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(body, -1)
	set := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		call := strings.TrimSpace(strings.TrimSuffix(m, "("))
		if _, ok := set[call]; ok {
			continue
		}
		set[call] = struct{}{}
		out = append(out, call)
	}
	sort.Strings(out)
	return out
}

func findInvokedMethods(body string) []string {
	dottedCalls := regexp.MustCompile(`\.\s*([A-Za-z_][A-Za-z0-9_]*)\s*\(`).FindAllStringSubmatch(body, -1)
	directCalls := regexp.MustCompile(`(?:^|[^.\w])([A-Za-z_][A-Za-z0-9_]*)\s*\(`).FindAllStringSubmatch(body, -1)

	keywords := map[string]struct{}{
		"if": {}, "for": {}, "switch": {}, "select": {}, "return": {}, "go": {}, "defer": {}, "func": {},
		"make": {}, "new": {}, "len": {}, "cap": {}, "append": {}, "copy": {}, "delete": {}, "panic": {},
		"recover": {}, "close": {}, "print": {}, "println": {}, "string": {}, "int": {}, "int32": {},
		"int64": {}, "float32": {}, "float64": {}, "bool": {}, "byte": {}, "rune": {}, "error": {},
	}

	set := make(map[string]struct{}, len(dottedCalls)+len(directCalls))
	out := make([]string, 0, len(dottedCalls)+len(directCalls))
	add := func(matches [][]string) {
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			name := strings.TrimSpace(m[1])
			if name == "" {
				continue
			}
			if _, isKeyword := keywords[strings.ToLower(name)]; isKeyword {
				continue
			}
			if _, ok := set[name]; ok {
				continue
			}
			set[name] = struct{}{}
			out = append(out, name)
		}
	}

	add(dottedCalls)
	add(directCalls)
	sort.Strings(out)
	return out
}

func extractFunctionBody(content string, from int) (string, bool) {
	start := strings.Index(content[from:], "{")
	if start < 0 {
		return "", false
	}
	start += from
	depth := 0
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[start+1 : i], true
			}
		}
	}
	return "", false
}

func parseHandlerName(handlerExpr string) string {
	expr := strings.TrimSpace(handlerExpr)
	expr = strings.TrimSuffix(expr, ",")
	if idx := strings.LastIndex(expr, "."); idx >= 0 {
		expr = expr[idx+1:]
	}
	expr = strings.TrimSpace(expr)
	expr = strings.TrimSuffix(expr, ")")
	expr = strings.TrimSpace(expr)
	re := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	if !re.MatchString(expr) {
		return ""
	}
	return expr
}

func normalizeRouteKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + "|" + routePathSignature(normalizeRoutePath(path))
}

func routePathSignature(path string) string {
	normalized := normalizeRoutePath(path)
	return routePlaceholderPattern.ReplaceAllString(normalized, "{}")
}

func matchRouteRef(routes map[string]routeRef, endpoint normalizer.Endpoint) (routeRef, bool) {
	method := strings.ToUpper(strings.TrimSpace(endpoint.Method))
	candidates := []string{
		normalizeRoutePath(endpoint.Path),
		normalizeRoutePath(joinPaths(endpoint.BasePath, endpoint.Path)),
	}
	for _, path := range candidates {
		key := method + "|" + routePathSignature(path)
		if ref, ok := routes[key]; ok {
			return ref, true
		}
	}
	return routeRef{}, false
}

func normalizeRoutePath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	p = strings.Join(parts, "/")
	p = strings.ReplaceAll(p, "//", "/")
	return p
}

func joinPaths(basePath, endpointPath string) string {
	base := strings.TrimSpace(basePath)
	endpoint := strings.TrimSpace(endpointPath)
	if base == "" || base == "/" {
		return endpoint
	}
	if endpoint == "" || endpoint == "/" {
		return base
	}
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(endpoint, "/")
}

func appendSourceIfMissing(values []normalizer.SourceType, source normalizer.SourceType) []normalizer.SourceType {
	for _, current := range values {
		if current == source {
			return values
		}
	}
	return append(values, source)
}
