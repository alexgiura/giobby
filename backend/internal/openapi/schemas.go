package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type swaggerFile struct {
	Paths       map[string]map[string]json.RawMessage `json:"paths"`
	Definitions map[string]json.RawMessage            `json:"definitions"`
}

func loadSwaggerSchemas(path string) *swaggerFile {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var s swaggerFile
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}

func defaultRequestBody(path, method string) *RequestBody {
	// Built-in English schemas for local auth endpoints (not in Giobby swagger).
	switch path {
	case "/api/auth/register":
		return &RequestBody{
			Required:    true,
			Description: "Required fields: username, password",
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Required: []string{"username", "password"},
						Properties: map[string]*Schema{
							"username": {Type: "string", Example: "demo"},
							"password": {Type: "string", Format: "password", Example: "password123"},
							"email":    {Type: "string", Format: "email", Example: "demo@example.com"},
						},
					},
					Example: map[string]interface{}{
						"username": "demo",
						"password": "password123",
						"email":    "demo@example.com",
					},
				},
			},
		}
	case "/api/auth/login":
		return &RequestBody{
			Required:    true,
			Description: "Required fields: username, password",
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Required: []string{"username", "password"},
						Properties: map[string]*Schema{
							"username": {Type: "string", Example: "demo"},
							"password": {Type: "string", Format: "password", Example: "password123"},
						},
					},
					Example: map[string]interface{}{"username": "demo", "password": "password123"},
				},
			},
		}
	case "/api/auth/refresh":
		return &RequestBody{
			Required:    true,
			Description: "Required fields: refresh_token",
			Content: map[string]MediaType{
				"application/json": {
					Schema: &Schema{
						Type: "object",
						Required: []string{"refresh_token"},
						Properties: map[string]*Schema{
							"grant_type":    {Type: "string", Example: "refresh_token"},
							"refresh_token": {Type: "string", Example: "paste-refresh-token"},
						},
					},
					Example: map[string]interface{}{
						"grant_type":    "refresh_token",
						"refresh_token": "paste-refresh-token",
					},
				},
			},
		}
	}

	return &RequestBody{
		Required:    method != http.MethodPatch,
		Description: "JSON request body — fill required fields marked in the schema",
		Content: map[string]MediaType{
			"application/json": {
				Schema:  &Schema{Type: "object", AdditionalProperties: true},
				Example: map[string]interface{}{},
			},
		},
	}
}

// applySchemaEnrichment adds body/query schemas from swagger without Italian text.
func applySchemaEnrichment(op *Operation, method, muxPath string, enrich *swaggerFile) {
	if enrich == nil {
		return
	}
	swPath := strings.TrimPrefix(muxPath, "/api")
	if swPath == "" {
		swPath = "/"
	}
	methods, ok := enrich.Paths[swPath]
	if !ok {
		norm := func(p string) string {
			return regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(p, "{}")
		}
		target := norm(swPath)
		for p, m := range enrich.Paths {
			if norm(p) == target {
				methods = m
				ok = true
				break
			}
		}
	}
	if !ok {
		return
	}
	raw, ok := methods[strings.ToLower(method)]
	if !ok {
		return
	}
	var swOp map[string]interface{}
	if err := json.Unmarshal(raw, &swOp); err != nil {
		return
	}

	params, _ := swOp["parameters"].([]interface{})
	var extra []Parameter
	for _, p := range params {
		pm, _ := p.(map[string]interface{})
		if pm == nil {
			continue
		}
		in, _ := pm["in"].(string)
		name, _ := pm["name"].(string)
		req, _ := pm["required"].(bool)
		desc, _ := pm["description"].(string)
		// Keep English-only short labels for query/header; strip long Italian by using name.
		if in == "query" || in == "header" {
			schema := &Schema{Type: stringVal(pm["type"], "string")}
			if f, ok := pm["format"].(string); ok {
				schema.Format = f
			}
			label := name
			if desc != "" && isMostlyASCII(desc) && !looksItalian(desc) {
				label = desc
			}
			extra = append(extra, Parameter{
				Name: name, In: in, Required: req, Description: label, Schema: schema,
			})
			continue
		}
		if in == "body" {
			schema := bodySchemaFromParam(pm, enrich.Definitions)
			applyKnownRequired(schema, muxPath, method, enrich.Definitions)
			example := exampleFromSchema(schema, enrich.Definitions)
			reqFields := schemaRequiredNames(schema, enrich.Definitions)
			propFields := schemaPropertyNames(schema, enrich.Definitions)
			desc := "JSON request body"
			if len(reqFields) > 0 {
				desc = "Required fields: " + strings.Join(reqFields, ", ")
			} else if len(propFields) > 0 {
				desc = "Body fields: " + strings.Join(propFields, ", ")
			}
			op.RequestBody = &RequestBody{
				Required:    true,
				Description: desc,
				Content: map[string]MediaType{
					"application/json": {Schema: schema, Example: example},
				},
			}
		}
	}
	if len(extra) > 0 {
		op.Parameters = append(op.Parameters, extra...)
	}
}

func bodySchemaFromParam(pm map[string]interface{}, defs map[string]json.RawMessage) *Schema {
	sch, _ := pm["schema"].(map[string]interface{})
	if sch == nil {
		return &Schema{Type: "object", AdditionalProperties: true}
	}
	if ref, ok := sch["$ref"].(string); ok {
		name := strings.TrimPrefix(ref, "#/definitions/")
		return &Schema{Ref: "#/components/schemas/" + name}
	}
	raw, _ := json.Marshal(sch)
	return convertSwaggerSchema(raw, defs)
}

func convertSwaggerSchema(raw json.RawMessage, defs map[string]json.RawMessage) *Schema {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return &Schema{Type: "object"}
	}
	if ref, ok := m["$ref"].(string); ok {
		name := strings.TrimPrefix(ref, "#/definitions/")
		return &Schema{Ref: "#/components/schemas/" + name}
	}
	s := &Schema{
		Type:   stringVal(m["type"], ""),
		Format: stringVal(m["format"], ""),
	}
	s.Example = coerceExample(m["example"], s.Type, s.Format)
	// Skip Italian descriptions from legacy swagger.
	if req, ok := m["required"].([]interface{}); ok {
		for _, r := range req {
			if rs, ok := r.(string); ok {
				s.Required = append(s.Required, rs)
			}
		}
	}
	if en, ok := m["enum"].([]interface{}); ok {
		s.Enum = en
	}
	if props, ok := m["properties"].(map[string]interface{}); ok {
		s.Properties = map[string]*Schema{}
		keys := make([]string, 0, len(props))
		for k := range props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			pr, _ := json.Marshal(props[k])
			s.Properties[k] = convertSwaggerSchema(pr, defs)
		}
	}
	if items, ok := m["items"].(map[string]interface{}); ok {
		ir, _ := json.Marshal(items)
		s.Items = convertSwaggerSchema(ir, defs)
	}
	if s.Type == "" && s.Ref == "" && s.Properties != nil {
		s.Type = "object"
	}
	return s
}

func exampleFromSchema(s *Schema, defs map[string]json.RawMessage) interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	if s.Example != nil {
		return coerceExample(s.Example, s.Type, s.Format)
	}
	if s.Ref != "" {
		name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		if raw, ok := defs[name]; ok {
			return exampleFromSchema(convertSwaggerSchema(raw, defs), defs)
		}
		return map[string]interface{}{}
	}
	switch s.Type {
	case "object":
		out := map[string]interface{}{}
		for _, req := range s.Required {
			if v, ok := s.Properties[req]; ok {
				out[req] = exampleFromSchema(v, defs)
			}
		}
		for k, v := range s.Properties {
			if _, exists := out[k]; !exists {
				out[k] = exampleFromSchema(v, defs)
			}
		}
		return out
	case "array":
		return []interface{}{exampleFromSchema(s.Items, defs)}
	case "integer":
		return 1
	case "number":
		return 0
	case "boolean":
		return true
	case "string":
		if len(s.Enum) > 0 {
			return s.Enum[0]
		}
		if s.Format == "date-time" {
			return "2026-01-01T00:00:00Z"
		}
		if s.Format == "date" {
			return "2026-01-01"
		}
		if s.Format == "email" {
			return "user@example.com"
		}
		if s.Format == "password" {
			return "password123"
		}
		return "string"
	default:
		return map[string]interface{}{}
	}
}

// coerceExample fixes legacy swagger examples that put numbers/bools as quoted strings.
func coerceExample(v interface{}, typ, format string) interface{} {
	if v == nil {
		return nil
	}
	switch typ {
	case "integer":
		switch t := v.(type) {
		case float64:
			return int64(t)
		case int:
			return t
		case int32:
			return t
		case int64:
			return t
		case json.Number:
			n, err := t.Int64()
			if err == nil {
				return n
			}
		case string:
			s := strings.TrimSpace(t)
			if s == "" {
				return 1
			}
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				return n
			}
			return 1
		}
	case "number":
		switch t := v.(type) {
		case float64:
			return t
		case int:
			return float64(t)
		case int64:
			return float64(t)
		case json.Number:
			n, err := t.Float64()
			if err == nil {
				return n
			}
		case string:
			s := strings.TrimSpace(t)
			if s == "" {
				return 0.0
			}
			if n, err := strconv.ParseFloat(s, 64); err == nil {
				return n
			}
			return 0.0
		}
	case "boolean":
		switch t := v.(type) {
		case bool:
			return t
		case string:
			s := strings.TrimSpace(strings.ToLower(t))
			if s == "true" || s == "1" {
				return true
			}
			if s == "false" || s == "0" {
				return false
			}
			return true
		case float64:
			return t != 0
		}
	case "string":
		switch t := v.(type) {
		case string:
			// Replace useless blank placeholders from swagger.
			if strings.TrimSpace(t) == "" {
				if format == "email" {
					return "user@example.com"
				}
				return "string"
			}
			return t
		case float64, int, int64, bool:
			return fmt.Sprint(t)
		}
	}
	return v
}

func schemaRequiredNames(s *Schema, defs map[string]json.RawMessage) []string {
	if s == nil {
		return nil
	}
	if s.Ref != "" {
		name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		if raw, ok := defs[name]; ok {
			return convertSwaggerSchema(raw, defs).Required
		}
	}
	return s.Required
}

func schemaPropertyNames(s *Schema, defs map[string]json.RawMessage) []string {
	if s == nil {
		return nil
	}
	if s.Ref != "" {
		name := strings.TrimPrefix(s.Ref, "#/components/schemas/")
		if raw, ok := defs[name]; ok {
			s = convertSwaggerSchema(raw, defs)
		}
	}
	if s.Properties == nil {
		return nil
	}
	keys := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// knownRequiredFields supplements swagger defs that omit "required".
var knownRequiredFields = map[string][]string{
	"POST /api/messages":                 {"body"},
	"POST /api/users":                    {"username", "password"},
	"POST /api/settings":                 {"key1"},
	"POST /api/loggeduser/changelanguage": {"language"},
	"POST /api/loggeduser/devicetoken":   {"deviceToken"},
	"POST /api/loggeduser/image":         {"imageUrl"},
	"POST /api/contacts":                 {"name"},
	"POST /api/customers":                {"idcustomer"},
	"POST /api/vendors":                  {"idvendor"},
	"POST /api/products/product":         {"idmaterial"},
	"POST /api/plugins/bindcommerce/updatestock": {"idMaterials"},
}

func applyKnownRequired(schema *Schema, muxPath, method string, defs map[string]json.RawMessage) {
	key := method + " " + muxPath
	fields, ok := knownRequiredFields[key]
	if !ok || schema == nil {
		return
	}
	target := schema
	if schema.Ref != "" {
		name := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
		if raw, ok := defs[name]; ok {
			// Mutating converted copy won't update components; set inline schema instead.
			converted := convertSwaggerSchema(raw, defs)
			converted.Required = mergeRequired(converted.Required, fields)
			*schema = *converted
			schema.Ref = ""
			return
		}
	}
	target.Required = mergeRequired(target.Required, fields)
}

func mergeRequired(existing, extra []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range existing {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	for _, r := range extra {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

func stringVal(v interface{}, def string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return def
}

func looksItalian(s string) bool {
	lower := strings.ToLower(s)
	for _, w := range []string{"restituisce", "lista", "delle", "del ", "una ", "tutti", "messaggi", "utente"} {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

func isMostlyASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
