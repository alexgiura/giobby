package openapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

// Spec is a minimal OpenAPI 3.0 document.
type Spec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers"`
	Tags       []Tag               `json:"tags"`
	Paths      map[string]PathItem `json:"paths"`
	Components *Components         `json:"components,omitempty"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type PathItem map[string]*Operation

type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      *Schema     `json:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

type RequestBody struct {
	Required    bool                 `json:"required,omitempty"`
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema  *Schema     `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Description          string             `json:"description,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Example              interface{}        `json:"example,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty"`
}

type Components struct {
	Schemas         map[string]*Schema        `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Description  string `json:"description,omitempty"`
}

// tagOrder matches handler / router domain grouping.
var tagOrder = []string{
	"Health",
	"Auth",
	"Country",
	"Reference",
	"Company",
	"Contacts",
	"Customers",
	"Vendors",
	"Products",
	"Warehouse",
	"Sales",
	"Purchases",
	"Accounting",
	"CRM",
	"Calendar",
	"PersonalActivity",
	"Users",
	"Settings",
	"Messages",
	"LoggedUser",
	"Ecommerce",
	"Social",
	"Tasks",
	"Emails",
	"Notifications",
	"Plugins",
	"Tilby",
	"OpenAPI",
	"Other",
}

var segmentToTag = map[string]string{
	"healthz":             "Health",
	"health":              "Health",
	"auth":                "Auth",
	"countries":           "Country",
	"cities":              "Reference",
	"currencies":          "Reference",
	"um":                  "Reference",
	"officesType":         "Reference",
	"contactsroles":       "Reference",
	"paymentterms":        "Reference",
	"companies":           "Company",
	"companysettings":     "Company",
	"contacts":            "Contacts",
	"customers":           "Customers",
	"vendors":             "Vendors",
	"products":            "Products",
	"productsgroups":      "Products",
	"attributes":          "Products",
	"pricelists":          "Products",
	"storages":            "Warehouse",
	"storageLocations":    "Warehouse",
	"stocks":              "Warehouse",
	"lots":                "Warehouse",
	"machineDataTracking": "Warehouse",
	"sales":               "Sales",
	"purchases":           "Purchases",
	"accounting":          "Accounting",
	"crmaccounts":         "CRM",
	"calendars":           "Calendar",
	"personalactivities":  "PersonalActivity",
	"users":               "Users",
	"settings":            "Settings",
	"messages":            "Messages",
	"messagegroups":       "Messages",
	"loggeduser":          "LoggedUser",
	"ecommerce":           "Ecommerce",
	"social":              "Social",
	"tasks":               "Tasks",
	"emails":              "Emails",
	"notifications":       "Notifications",
	"plugins":             "Plugins",
	"tilby":               "Tilby",
	"openapi.json":        "OpenAPI",
}

var pathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

// Generate walks the mux router and builds an OpenAPI 3 document
// from implemented handlers (English labels). Optional swaggerPath
// enriches request body schemas / required fields only (no Italian text).
func Generate(router *mux.Router, swaggerPath string) (*Spec, error) {
	enrich := loadSwaggerSchemas(swaggerPath)

	spec := &Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title: "Giobby API Explorer",
			Description: "Auto-generated from this project's handlers.\n\n" +
				"1. Call **Auth → Create auth (register)** or **login**\n" +
				"2. Click **Authorize** and paste `access_token`\n" +
				"3. Open an endpoint, fill **required** body fields (marked), then **Send**\n\n" +
				"Base URL follows the host you open (localhost or server IP) via the explorer proxy.",
			Version: "1.0.0",
		},
		// Relative server = same origin as the explorer UI (works on any host/IP).
		Servers: []Server{
			{URL: "/", Description: "Same origin as this API Explorer (recommended)"},
		},
		Paths: map[string]PathItem{},
		Components: &Components{
			Schemas: map[string]*Schema{},
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
					Description:  "Paste access_token from /api/auth/login (without the word Bearer)",
				},
				"apiKeyAuth": {
					Type:        "apiKey",
					Name:        "X-API-Key",
					In:          "header",
					Description: "Optional API key header",
				},
				"cookieAuth": {
					Type:        "apiKey",
					Name:        "session",
					In:          "cookie",
					Description: "Optional session cookie",
				},
			},
		},
	}

	if enrich != nil {
		for name, raw := range enrich.Definitions {
			spec.Components.Schemas[name] = convertSwaggerSchema(raw, enrich.Definitions)
		}
	}

	usedTags := map[string]bool{}

	err := router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil || pathTemplate == "" {
			return nil
		}
		methods, err := route.GetMethods()
		if err != nil || len(methods) == 0 {
			return nil
		}

		for _, method := range methods {
			method = strings.ToUpper(method)
			if method == http.MethodOptions {
				continue
			}

			tag := tagForPath(pathTemplate)
			usedTags[tag] = true
			authRequired := requiresAuth(pathTemplate, method)

			op := &Operation{
				Tags:        []string{tag},
				Summary:     englishSummary(method, pathTemplate),
				Description: englishDescription(method, pathTemplate, authRequired),
				OperationID: operationID(method, pathTemplate),
				Parameters:  pathParameters(pathTemplate),
				Responses: map[string]Response{
					"200": {Description: "Success", Content: map[string]MediaType{
						"application/json": {Schema: &Schema{Type: "object", AdditionalProperties: true}},
					}},
					"400": {Description: "Bad request"},
					"401": {Description: "Unauthorized"},
					"404": {Description: "Not found"},
					"500": {Description: "Server error"},
				},
			}

			if authRequired {
				op.Security = []map[string][]string{{"bearerAuth": {}}}
			} else {
				op.Security = []map[string][]string{}
			}

			if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
				op.RequestBody = defaultRequestBody(pathTemplate, method)
			}

			applySchemaEnrichment(op, method, pathTemplate, enrich)

			if _, ok := spec.Paths[pathTemplate]; !ok {
				spec.Paths[pathTemplate] = PathItem{}
			}
			spec.Paths[pathTemplate][strings.ToLower(method)] = op
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk routes: %w", err)
	}

	for _, name := range tagOrder {
		if usedTags[name] {
			spec.Tags = append(spec.Tags, Tag{Name: name, Description: tagDescription(name)})
		}
	}
	for name := range usedTags {
		found := false
		for _, t := range spec.Tags {
			if t.Name == name {
				found = true
				break
			}
		}
		if !found {
			spec.Tags = append(spec.Tags, Tag{Name: name})
		}
	}

	return spec, nil
}

func tagForPath(path string) string {
	p := strings.TrimPrefix(path, "/api")
	p = strings.Trim(p, "/")
	if p == "" {
		return "Other"
	}
	seg := strings.Split(p, "/")[0]
	if t, ok := segmentToTag[seg]; ok {
		return t
	}
	return "Other"
}

func tagDescription(name string) string {
	switch name {
	case "Health":
		return "Liveness probes"
	case "Auth":
		return "JWT register, login, refresh, logout, and profile"
	case "Country":
		return "Countries"
	case "Reference":
		return "Cities, currencies, UM, payment terms"
	case "Company":
		return "Company profile and company settings"
	case "Warehouse":
		return "Storages, locations, stocks, lots, machine tracking"
	case "OpenAPI":
		return "Live OpenAPI document for this explorer"
	default:
		return name + " endpoints implemented in this API"
	}
}

func englishSummary(method, path string) string {
	clean := strings.TrimPrefix(path, "/api")
	if clean == "" {
		clean = path
	}
	parts := strings.Split(strings.Trim(clean, "/"), "/")
	resource := "resource"
	if len(parts) > 0 && parts[0] != "" {
		resource = humanizeSegment(parts[0])
	}
	hasID := pathParamRe.MatchString(path)
	tail := ""
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if !strings.HasPrefix(last, "{") {
			tail = " (" + humanizeSegment(last) + ")"
		} else if len(parts) > 2 {
			prev := parts[len(parts)-2]
			if !strings.HasPrefix(prev, "{") {
				tail = " (" + humanizeSegment(prev) + ")"
			}
		}
	}

	switch method {
	case http.MethodGet:
		if hasID {
			return "Get " + singularize(resource) + tail
		}
		return "List " + resource + tail
	case http.MethodPost:
		return "Create " + singularize(resource) + tail
	case http.MethodPut:
		return "Update " + singularize(resource) + tail
	case http.MethodPatch:
		return "Patch " + singularize(resource) + tail
	case http.MethodDelete:
		return "Delete " + singularize(resource) + tail
	default:
		return method + " " + path
	}
}

func englishDescription(method, path string, authRequired bool) string {
	auth := "**Authentication:** not required (public)."
	if authRequired {
		auth = "**Authentication required:** Bearer JWT."
	}
	return fmt.Sprintf("%s\n\nRoute: `%s %s`\n\n%s",
		englishSummary(method, path), method, path, auth)
}

func humanizeSegment(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	// camelCase split
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return s
	}
	return strings.ToLower(out)
}

func singularize(s string) string {
	if strings.HasSuffix(s, "ies") && len(s) > 3 {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "sses") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") && len(s) > 1 {
		return s[:len(s)-1]
	}
	return s
}

func operationID(method, path string) string {
	id := strings.ToLower(method) + "_" + path
	id = strings.ReplaceAll(id, "/", "_")
	id = strings.ReplaceAll(id, "{", "")
	id = strings.ReplaceAll(id, "}", "")
	id = strings.Trim(id, "_")
	return id
}

func pathParameters(path string) []Parameter {
	matches := pathParamRe.FindAllStringSubmatch(path, -1)
	var out []Parameter
	for _, m := range matches {
		name := m[1]
		out = append(out, Parameter{
			Name:        name,
			In:          "path",
			Required:    true,
			Description: "Path parameter `" + name + "`",
			Schema:      &Schema{Type: "string", Example: samplePathValue(name)},
			Example:     samplePathValue(name),
		})
	}
	return out
}

func samplePathValue(name string) string {
	switch strings.ToLower(name) {
	case "idum":
		return "PZ"
	case "code", "idsourcecurr":
		return "EUR"
	case "idcustomer":
		return "C00001"
	case "idvendor":
		return "V00001"
	case "idvat":
		return "TVA19"
	case "idlot":
		return "L00001"
	default:
		return "1"
	}
}

func requiresAuth(path, method string) bool {
	if path == "/healthz" || path == "/health" {
		return false
	}
	if path == "/api/openapi.json" {
		return false
	}
	if strings.HasPrefix(path, "/api/auth/") {
		switch path {
		case "/api/auth/register", "/api/auth/login", "/api/auth/refresh":
			return false
		}
	}
	return true
}

// Handler returns an http.HandlerFunc that serves the OpenAPI JSON.
func Handler(spec *Spec) http.HandlerFunc {
	body, err := json.Marshal(spec)
	if err != nil {
		body = []byte(`{"openapi":"3.0.3","info":{"title":"error","version":"0"},"paths":{}}`)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}
