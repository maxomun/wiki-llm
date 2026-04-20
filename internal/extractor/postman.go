package extractor

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/max/wiki-llm/internal/normalizer"
)

type postmanCollection struct {
	Info postmanInfo   `json:"info"`
	Item []postmanItem `json:"item"`
}

type postmanInfo struct {
	Name        string `json:"name"`
	Description any    `json:"description"`
	Version     any    `json:"version"`
}

type postmanItem struct {
	Name        string        `json:"name"`
	Description any           `json:"description"`
	Item        []postmanItem `json:"item"`
	Request     *postmanReq   `json:"request"`
	Response    []postmanResp `json:"response"`
}

type postmanReq struct {
	Method      string          `json:"method"`
	Description any             `json:"description"`
	Header      []postmanHeader `json:"header"`
	Body        *postmanBody    `json:"body"`
	URL         postmanURL      `json:"url"`
}

type postmanHeader struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description any    `json:"description"`
}

type postmanBody struct {
	Mode    string         `json:"mode"`
	Raw     string         `json:"raw"`
	Options map[string]any `json:"options"`
}

type postmanURL struct {
	Raw      string            `json:"raw"`
	Path     any               `json:"path"`
	Query    []postmanQuery    `json:"query"`
	Variable []postmanVariable `json:"variable"`
}

type postmanQuery struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Disabled    bool   `json:"disabled"`
}

type postmanVariable struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type postmanResp struct {
	Code   int             `json:"code"`
	Status string          `json:"status"`
	Header []postmanHeader `json:"header"`
	Body   string          `json:"body"`
}

// ExtractPostmanCollection extrae un APIDocument desde una coleccion Postman v2.1.
func ExtractPostmanCollection(sourcePath string) (normalizer.APIDocument, error) {
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		return normalizer.APIDocument{}, fmt.Errorf("leer postman collection: %w", err)
	}

	var collection postmanCollection
	if err := json.Unmarshal(raw, &collection); err != nil {
		return normalizer.APIDocument{}, fmt.Errorf("parsear postman collection json: %w", err)
	}
	if strings.TrimSpace(collection.Info.Name) == "" {
		return normalizer.APIDocument{}, fmt.Errorf("collection invalida: falta info.name")
	}

	doc := normalizer.APIDocument{
		Title:          strings.TrimSpace(collection.Info.Name),
		Description:    strings.TrimSpace(normalizeAny(collection.Info.Description)),
		Version:        normalizeUsefulVersion(normalizeAny(collection.Info.Version)),
		SourcePath:     sourcePath,
		Endpoints:      make([]normalizer.Endpoint, 0),
		ContractSource: normalizer.SourcePostman,
	}

	unique := make(map[string]normalizer.Endpoint)
	order := make([]string, 0)
	collectPostmanEndpoints(collection.Item, nil, unique, &order)

	for _, key := range order {
		doc.Endpoints = append(doc.Endpoints, unique[key])
	}

	return doc, nil
}

func collectPostmanEndpoints(items []postmanItem, folderTags []string, unique map[string]normalizer.Endpoint, order *[]string) {
	for _, item := range items {
		if len(item.Item) > 0 {
			nextTags := append([]string{}, folderTags...)
			if strings.TrimSpace(item.Name) != "" {
				nextTags = append(nextTags, strings.TrimSpace(item.Name))
			}
			collectPostmanEndpoints(item.Item, nextTags, unique, order)
			continue
		}

		if item.Request == nil {
			continue
		}

		method := strings.ToUpper(strings.TrimSpace(item.Request.Method))
		if method == "" {
			method = "GET"
		}

		path := normalizePostmanPath(item.Request.URL)
		opID := fallbackOperationID(strings.ToLower(method), path)
		if strings.TrimSpace(item.Name) != "" {
			opID = strings.ToUpper(method) + "_" + slugToken(item.Name)
		}

		endpoint := normalizer.Endpoint{
			Path:        path,
			Method:      method,
			OperationID: opID,
			Summary:     strings.TrimSpace(item.Name),
			Description: firstNonEmpty(
				strings.TrimSpace(normalizeAny(item.Request.Description)),
				strings.TrimSpace(normalizeAny(item.Description)),
			),
			Tags:        normalizeStringList(folderTags),
			Parameters:  postmanParameters(item.Request),
			RequestBody: postmanRequestBody(item.Request),
			Responses:   postmanResponses(item.Response),
			Sources:     []normalizer.SourceType{normalizer.SourcePostman},
		}

		key := method + "|" + path
		if existing, found := unique[key]; found {
			merged := mergeEndpoint(existing, endpoint)
			unique[key] = merged
			continue
		}
		unique[key] = endpoint
		*order = append(*order, key)
	}
}

func postmanParameters(req *postmanReq) []normalizer.Parameter {
	out := make([]normalizer.Parameter, 0)
	for _, v := range req.URL.Variable {
		name := strings.TrimSpace(v.Key)
		if name == "" {
			continue
		}
		out = append(out, normalizer.Parameter{
			Name:        name,
			In:          "path",
			Required:    true,
			Type:        "string",
			Description: strings.TrimSpace(v.Description),
		})
	}
	for _, q := range req.URL.Query {
		if q.Disabled {
			continue
		}
		name := strings.TrimSpace(q.Key)
		if name == "" {
			continue
		}
		out = append(out, normalizer.Parameter{
			Name:        name,
			In:          "query",
			Required:    false,
			Type:        "string",
			Description: strings.TrimSpace(q.Description),
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

func postmanRequestBody(req *postmanReq) *normalizer.RequestBody {
	if req.Body == nil || strings.TrimSpace(req.Body.Mode) == "" {
		return nil
	}

	contentType := "application/json"
	return &normalizer.RequestBody{
		Required:     true,
		Description:  "Body definido en Postman",
		ContentTypes: []string{contentType},
		SchemaRef:    strings.TrimSpace(req.Body.Mode),
		Example:      strings.TrimSpace(req.Body.Raw),
	}
}

func postmanResponses(responses []postmanResp) []normalizer.Response {
	if len(responses) == 0 {
		return nil
	}

	out := make([]normalizer.Response, 0, len(responses))
	for _, resp := range responses {
		contentTypes := make([]string, 0)
		for _, h := range resp.Header {
			if strings.EqualFold(strings.TrimSpace(h.Key), "content-type") {
				val := strings.TrimSpace(h.Value)
				if val != "" {
					contentTypes = append(contentTypes, val)
				}
			}
		}
		out = append(out, normalizer.Response{
			StatusCode:   fmt.Sprintf("%d", resp.Code),
			Description:  strings.TrimSpace(resp.Status),
			ContentTypes: normalizeStringList(contentTypes),
			Example:      strings.TrimSpace(resp.Body),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StatusCode < out[j].StatusCode
	})
	return out
}

func normalizePostmanPath(url postmanURL) string {
	segments := make([]string, 0)
	switch p := url.Path.(type) {
	case []any:
		for _, segment := range p {
			s := strings.TrimSpace(fmt.Sprint(segment))
			if s == "" {
				continue
			}
			segments = append(segments, normalizePathToken(s))
		}
	case []string:
		for _, segment := range p {
			s := strings.TrimSpace(segment)
			if s == "" {
				continue
			}
			segments = append(segments, normalizePathToken(s))
		}
	}

	if len(segments) == 0 && strings.TrimSpace(url.Raw) != "" {
		raw := url.Raw
		if idx := strings.Index(raw, "://"); idx >= 0 {
			raw = raw[idx+3:]
			if slash := strings.Index(raw, "/"); slash >= 0 {
				raw = raw[slash+1:]
			} else {
				raw = ""
			}
		}
		for _, seg := range strings.Split(raw, "/") {
			s := strings.TrimSpace(seg)
			if s == "" {
				continue
			}
			segments = append(segments, normalizePathToken(s))
		}
	}

	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

func normalizePathToken(token string) string {
	if strings.HasPrefix(token, ":") {
		return "{" + strings.TrimPrefix(token, ":") + "}"
	}
	return token
}

func mergeEndpoint(base, incoming normalizer.Endpoint) normalizer.Endpoint {
	if base.Summary == "" {
		base.Summary = incoming.Summary
	}
	if base.Description == "" {
		base.Description = incoming.Description
	}
	base.Tags = normalizeStringList(append(base.Tags, incoming.Tags...))
	base.Sources = mergeSourceTypes(base.Sources, incoming.Sources)

	if len(base.Parameters) == 0 && len(incoming.Parameters) > 0 {
		base.Parameters = incoming.Parameters
	}
	if base.RequestBody == nil && incoming.RequestBody != nil {
		base.RequestBody = incoming.RequestBody
	}
	if len(base.Responses) == 0 && len(incoming.Responses) > 0 {
		base.Responses = incoming.Responses
	}
	return base
}

func mergeSourceTypes(base, incoming []normalizer.SourceType) []normalizer.SourceType {
	set := make(map[normalizer.SourceType]struct{}, len(base)+len(incoming))
	out := make([]normalizer.SourceType, 0, len(base)+len(incoming))
	for _, src := range append(base, incoming...) {
		if src == "" {
			continue
		}
		if _, ok := set[src]; ok {
			continue
		}
		set[src] = struct{}{}
		out = append(out, src)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func slugToken(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return "endpoint"
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "endpoint"
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func normalizeUsefulVersion(v string) string {
	version := strings.TrimSpace(v)
	if version == "" {
		return ""
	}
	if strings.EqualFold(version, "postman") {
		return ""
	}
	return version
}
