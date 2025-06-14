package models

// EnhancedEndpoint represents a more detailed endpoint structure for SDK generation
type EnhancedEndpoint struct {
	Name        string                     `json:"name"`
	Summary     string                     `json:"summary"`
	Description string                     `json:"description"`
	OperationID string                     `json:"operation_id"`
	Method      string                     `json:"method"`
	Path        string                     `json:"path"`
	PathParams  map[string]ParameterSchema `json:"path_params"`
	QueryParams map[string]ParameterSchema `json:"query_params"`
	Headers     map[string]ParameterSchema `json:"headers"`
	RequestBody *RequestBodySchema         `json:"request_body,omitempty"`
	Responses   []ResponseSchema           `json:"responses"`
	Tags        []string                   `json:"tags"`
	Security    []SecurityRequirement      `json:"security,omitempty"`
	Examples    map[string]ExampleValue    `json:"examples,omitempty"`
}

// ParameterSchema represents parameter information
type ParameterSchema struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Example     interface{} `json:"example,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Format      string      `json:"format,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	MinLength   *int        `json:"min_length,omitempty"`
	MaxLength   *int        `json:"max_length,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
}

// RequestBodySchema represents request body structure
type RequestBodySchema struct {
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// ResponseSchema represents response structure
type ResponseSchema struct {
	StatusCode  int                    `json:"status_code"`
	Description string                 `json:"description"`
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Headers     map[string]interface{} `json:"headers,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// SecurityRequirement represents security requirements
type SecurityRequirement struct {
	Type   string   `json:"type"`   // bearer, apikey, basic, oauth2
	Name   string   `json:"name"`   // security scheme name
	Scopes []string `json:"scopes"` // for oauth2
}

// ExampleValue represents example values for endpoints
type ExampleValue struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
}

// SchemaDefinition represents reusable schema components
type SchemaDefinition struct {
	Type                 string                     `json:"type"`
	Properties           map[string]ParameterSchema `json:"properties,omitempty"`
	Required             []string                   `json:"required,omitempty"`
	Items                *ParameterSchema           `json:"items,omitempty"` // for arrays
	AdditionalProperties *ParameterSchema           `json:"additional_properties,omitempty"`
	Discriminator        *DiscriminatorSchema       `json:"discriminator,omitempty"`
	Example              interface{}                `json:"example,omitempty"`
	Description          string                     `json:"description,omitempty"`
}

// DiscriminatorSchema for polymorphic schemas
type DiscriminatorSchema struct {
	PropertyName string            `json:"property_name"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

// APISpecification represents the complete API specification
type APISpecification struct {
	Info      APIInfo                       `json:"info"`
	Servers   []ServerInfo                  `json:"servers"`
	Endpoints []EnhancedEndpoint            `json:"endpoints"`
	Schemas   map[string]SchemaDefinition   `json:"schemas"`
	Security  map[string]SecuritySchemeInfo `json:"security"`
	Tags      []TagInfo                     `json:"tags"`
}

// APIInfo represents API metadata
type APIInfo struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	Contact     ContactInfo `json:"contact,omitempty"`
	License     LicenseInfo `json:"license,omitempty"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// LicenseInfo represents license information
type LicenseInfo struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// ServerInfo represents server information
type ServerInfo struct {
	URL         string                 `json:"url"`
	Description string                 `json:"description,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

// SecuritySchemeInfo represents security scheme definitions
type SecuritySchemeInfo struct {
	Type             string      `json:"type"`
	Scheme           string      `json:"scheme,omitempty"`             // for http
	BearerFormat     string      `json:"bearer_format,omitempty"`      // for bearer
	Name             string      `json:"name,omitempty"`               // for apiKey
	In               string      `json:"in,omitempty"`                 // for apiKey: header, query, cookie
	Flows            OAuth2Flows `json:"flows,omitempty"`              // for oauth2
	OpenIDConnectURL string      `json:"openid_connect_url,omitempty"` // for openIdConnect
}

// OAuth2Flows represents OAuth2 flow configurations
type OAuth2Flows struct {
	Implicit          *OAuth2Flow `json:"implicit,omitempty"`
	Password          *OAuth2Flow `json:"password,omitempty"`
	ClientCredentials *OAuth2Flow `json:"client_credentials,omitempty"`
	AuthorizationCode *OAuth2Flow `json:"authorization_code,omitempty"`
}

// OAuth2Flow represents a single OAuth2 flow
type OAuth2Flow struct {
	AuthorizationURL string            `json:"authorization_url,omitempty"`
	TokenURL         string            `json:"token_url,omitempty"`
	RefreshURL       string            `json:"refresh_url,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// TagInfo represents tag information for grouping endpoints
type TagInfo struct {
	Name         string                `json:"name"`
	Description  string                `json:"description,omitempty"`
	ExternalDocs ExternalDocumentation `json:"external_docs,omitempty"`
}

// ExternalDocumentation represents external documentation
type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}
