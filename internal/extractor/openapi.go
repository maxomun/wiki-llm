package extractor

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/max/wiki-llm/internal/normalizer"
	"gopkg.in/yaml.v3"
)

type openAPISpec struct {
	OpenAPI  string                     `yaml:"openapi"`
	Swagger  string                     `yaml:"swagger"`
	Info     openAPIInfo                `yaml:"info"`
	Paths    map[string]openAPIPathItem `yaml:"paths"`
	Servers  []openAPIServer            `yaml:"servers"`
	BasePath string                     `yaml:"basePath"`
}

type openAPIInfo struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type openAPIServer struct {
	URL string `yaml:"url"`
}

type openAPIOperation struct {
	OperationID string                     `yaml:"operationId"`
	Summary     string                     `yaml:"summary"`
	Description string                     `yaml:"description"`
	Tags        []string                   `yaml:"tags"`
	Deprecated  bool                       `yaml:"deprecated"`
	Parameters  []openAPIParameter         `yaml:"parameters"`
	RequestBody *openAPIRequestBody        `yaml:"requestBody"`
	Responses   map[string]openAPIResponse `yaml:"responses"`
	Security    []map[string][]string      `yaml:"security"`
}

type openAPIPathItem struct {
	Parameters []openAPIParameter `yaml:"parameters"`
	Get        *openAPIOperation  `yaml:"get"`
	Post       *openAPIOperation  `yaml:"post"`
	Put        *openAPIOperation  `yaml:"put"`
	Patch      *openAPIOperation  `yaml:"patch"`
	Delete     *openAPIOperation  `yaml:"delete"`
	Head       *openAPIOperation  `yaml:"head"`
	Options    *openAPIOperation  `yaml:"options"`
	Trace      *openAPIOperation  `yaml:"trace"`
}

type openAPIParameter struct {
	Name        string        `yaml:"name"`
	In          string        `yaml:"in"`
	Required    bool          `yaml:"required"`
	Description string        `yaml:"description"`
	Schema      openAPISchema `yaml:"schema"`
	Example     any           `yaml:"example"`
}

type openAPIRequestBody struct {
	Required    bool                        `yaml:"required"`
	Description string                      `yaml:"description"`
	Content     map[string]openAPIMediaType `yaml:"content"`
}

type openAPIResponse struct {
	Description string                      `yaml:"description"`
	Content     map[string]openAPIMediaType `yaml:"content"`
}

type openAPIMediaType struct {
	Schema openAPISchema `yaml:"schema"`
}

type openAPISchema struct {
	Ref    string `yaml:"$ref"`
	Type   string `yaml:"type"`
	Format string `yaml:"format"`
}

// ExtractOpenAPI lee una especificacion OpenAPI y la transforma al modelo interno.
func ExtractOpenAPI(sourcePath string) (normalizer.APIDocument, error) {
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		return normalizer.APIDocument{}, fmt.Errorf("leer openapi: %w", err)
	}

	var spec openAPISpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return normalizer.APIDocument{}, fmt.Errorf("parsear openapi yaml/json: %w", err)
	}

	if strings.TrimSpace(spec.OpenAPI) == "" && strings.TrimSpace(spec.Swagger) == "" {
		return normalizer.APIDocument{}, fmt.Errorf("documento invalido: falta campo openapi/swagger")
	}

	api := normalizer.APIDocument{
		Title:          strings.TrimSpace(spec.Info.Title),
		Version:        strings.TrimSpace(spec.Info.Version),
		Description:    strings.TrimSpace(spec.Info.Description),
		SourcePath:     sourcePath,
		Endpoints:      make([]normalizer.Endpoint, 0),
		ContractSource: normalizer.SourceOpenAPI,
	}
	basePath := extractBasePath(spec.Servers, spec.BasePath)
	api.BasePath = basePath

	paths := sortedKeys(spec.Paths)
	for _, path := range paths {
		pathItem := spec.Paths[path]
		for _, methodOp := range extractPathOperations(pathItem) {
			method := methodOp.Method
			op := methodOp.Operation
			api.Endpoints = append(api.Endpoints, normalizer.Endpoint{
				BasePath:     basePath,
				Path:         path,
				Method:       strings.ToUpper(method),
				OperationID:  nonEmpty(strings.TrimSpace(op.OperationID), fallbackOperationID(method, path)),
				Summary:      strings.TrimSpace(op.Summary),
				Description:  strings.TrimSpace(op.Description),
				Tags:         normalizeStringList(op.Tags),
				Parameters:   mapParameters(pathItem.Parameters, op.Parameters),
				RequestBody:  mapRequestBody(op.RequestBody),
				Responses:    mapResponses(op.Responses),
				Deprecated:   op.Deprecated,
				SecurityRefs: mapSecurityRefs(op.Security),
				Sources:      []normalizer.SourceType{normalizer.SourceOpenAPI},
			})
		}
	}

	return api, nil
}

func mapParameters(pathParams, opParams []openAPIParameter) []normalizer.Parameter {
	merged := make(map[string]openAPIParameter)
	order := make([]string, 0, len(pathParams)+len(opParams))

	upsert := func(p openAPIParameter) {
		key := parameterKey(p.In, p.Name)
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = p
	}

	for _, p := range pathParams {
		upsert(p)
	}
	for _, p := range opParams {
		upsert(p)
	}

	out := make([]normalizer.Parameter, 0, len(merged))
	for _, key := range order {
		p := merged[key]
		required := p.Required
		if strings.EqualFold(p.In, "path") {
			required = true
		}

		schemaRef := strings.TrimSpace(p.Schema.Ref)
		pType := strings.TrimSpace(p.Schema.Type)
		if pType == "" && schemaRef != "" {
			pType = "object"
		}

		out = append(out, normalizer.Parameter{
			Name:        strings.TrimSpace(p.Name),
			In:          strings.ToLower(strings.TrimSpace(p.In)),
			Required:    required,
			Description: strings.TrimSpace(p.Description),
			SchemaRef:   schemaRef,
			Type:        pType,
			Format:      strings.TrimSpace(p.Schema.Format),
			Example:     normalizeAny(p.Example),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].In == out[j].In {
			return out[i].Name < out[j].Name
		}
		return out[i].In < out[j].In
	})
	return out
}

func mapRequestBody(body *openAPIRequestBody) *normalizer.RequestBody {
	if body == nil {
		return nil
	}

	contentTypes := sortedKeys(body.Content)
	return &normalizer.RequestBody{
		Required:     body.Required,
		Description:  strings.TrimSpace(body.Description),
		ContentTypes: contentTypes,
		SchemaRef:    firstSchemaRef(body.Content),
	}
}

func mapResponses(responses map[string]openAPIResponse) []normalizer.Response {
	statusCodes := sortedKeys(responses)
	out := make([]normalizer.Response, 0, len(statusCodes))
	for _, status := range statusCodes {
		resp := responses[status]
		out = append(out, normalizer.Response{
			StatusCode:   status,
			Description:  strings.TrimSpace(resp.Description),
			ContentTypes: sortedKeys(resp.Content),
			SchemaRef:    firstSchemaRef(resp.Content),
		})
	}
	return out
}

func mapSecurityRefs(security []map[string][]string) []string {
	if len(security) == 0 {
		return nil
	}

	set := make(map[string]struct{})
	for _, secItem := range security {
		for name := range secItem {
			set[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func firstSchemaRef(content map[string]openAPIMediaType) string {
	contentTypes := sortedKeys(content)
	for _, contentType := range contentTypes {
		schema := content[contentType].Schema
		if schema.Ref != "" {
			return strings.TrimSpace(schema.Ref)
		}
		if schema.Type != "" {
			return strings.TrimSpace(schema.Type)
		}
	}
	return ""
}

func normalizeAny(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func sortedKeys[T any](m map[string]T) []string {
	if len(m) == 0 {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

type methodOperation struct {
	Method    string
	Operation openAPIOperation
}

func extractPathOperations(pathItem openAPIPathItem) []methodOperation {
	out := make([]methodOperation, 0, 8)
	appendIf := func(method string, op *openAPIOperation) {
		if op == nil {
			return
		}
		out = append(out, methodOperation{Method: method, Operation: *op})
	}

	appendIf("delete", pathItem.Delete)
	appendIf("get", pathItem.Get)
	appendIf("head", pathItem.Head)
	appendIf("options", pathItem.Options)
	appendIf("patch", pathItem.Patch)
	appendIf("post", pathItem.Post)
	appendIf("put", pathItem.Put)
	appendIf("trace", pathItem.Trace)
	return out
}

func parameterKey(in, name string) string {
	return strings.ToLower(strings.TrimSpace(in)) + "|" + strings.TrimSpace(name)
}

func fallbackOperationID(method, path string) string {
	pathSlug := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_").Replace(path)
	pathSlug = strings.Trim(pathSlug, "_")
	if pathSlug == "" {
		pathSlug = "root"
	}
	return strings.ToUpper(method) + "_" + pathSlug
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	set := make(map[string]struct{})
	for _, v := range values {
		n := strings.TrimSpace(v)
		if n == "" {
			continue
		}
		set[n] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}

	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func extractBasePath(servers []openAPIServer, fallback string) string {
	for _, server := range servers {
		url := strings.TrimSpace(server.URL)
		if url == "" {
			continue
		}
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			if idx := strings.Index(url, "://"); idx >= 0 {
				tmp := url[idx+3:]
				if slash := strings.Index(tmp, "/"); slash >= 0 {
					url = tmp[slash:]
				} else {
					url = "/"
				}
			}
		}
		url = strings.TrimSpace(strings.SplitN(url, "?", 2)[0])
		if url == "" || url == "/" {
			continue
		}
		if !strings.HasPrefix(url, "/") {
			url = "/" + url
		}
		return url
	}
	fallback = strings.TrimSpace(fallback)
	if fallback == "" || fallback == "/" {
		return ""
	}
	if !strings.HasPrefix(fallback, "/") {
		fallback = "/" + fallback
	}
	return fallback
}
