package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionSource defines the source of the collection data
type CollectionSource string

const (
	CollectionSourcePostman       CollectionSource = "postman"
	CollectionSourceOpenAPI       CollectionSource = "openapi"
	CollectionSourcePostmanPublic CollectionSource = "postman_public" // New source for publicly imported collections
	CollectionSourceKonfig        CollectionSource = "konfig"
)

// Collection represents a Postman collection or API specification
type Collection struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name" validate:"required"`
	Description    string             `json:"description" bson:"description"`
	UserID         string             `json:"user_id" bson:"user_id"`
	PostmanData    interface{}        `json:"postman_data,omitempty" bson:"postman_data,omitempty"`         // Keep if direct Postman JSON is stored
	RawPostmanJSON string             `json:"raw_postman_json,omitempty" bson:"raw_postman_json,omitempty"` // For collections imported from Postman API
	OpenAPISpec    string             `json:"openapi_spec,omitempty" bson:"openapi_spec,omitempty"`         // For OpenAPI specs
	Source         CollectionSource   `json:"source" bson:"source"`                                         // To distinguish between Postman, OpenAPI, etc.
	SourceDetail   string             `json:"source_detail,omitempty" bson:"source_detail,omitempty"`       // E.g., Postman Collection UID, Konfig ID
	Endpoints      []Endpoint         `json:"endpoints" bson:"endpoints"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	ID          string            `json:"id" bson:"id"`
	Name        string            `json:"name" bson:"name"`
	Method      string            `json:"method" bson:"method"`
	URL         string            `json:"url" bson:"url"`
	Headers     map[string]string `json:"headers" bson:"headers"`
	Parameters  map[string]string `json:"parameters" bson:"parameters"`
	Body        interface{}       `json:"body" bson:"body"`
	Description string            `json:"description" bson:"description"`
}

// CreateCollectionRequest represents the request to create a collection
type CreateCollectionRequest struct {
	Name           string           `json:"name" validate:"required"`
	Description    string           `json:"description"`
	PostmanData    interface{}      `json:"postman_data,omitempty"`
	RawPostmanJSON string           `json:"raw_postman_json,omitempty"` // Added for importing
	OpenAPISpec    string           `json:"openapi_spec,omitempty"`     // Added for importing
	Source         CollectionSource `json:"source" validate:"required"`
	SourceDetail   string           `json:"source_detail"`
}

// UpdateCollectionRequest represents the request to update a collection
type UpdateCollectionRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PostmanData interface{} `json:"postman_data"`
}
