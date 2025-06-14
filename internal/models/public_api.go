package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PublicAPI represents a publicly available API with its Postman collection
type PublicAPI struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Category    string             `json:"category" bson:"category"`
	BaseURL     string             `json:"base_url" bson:"base_url"`
	AuthType    string             `json:"auth_type" bson:"auth_type"`
	Tags        []string           `json:"tags" bson:"tags"`
	PostmanURL  string             `json:"postman_url" bson:"postman_url"`
	PostmanID   string             `json:"postman_id" bson:"postman_id"`
	IsActive    bool               `json:"is_active" bson:"is_active"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// PostmanPublicCollection represents a collection from Postman's public API
type PostmanPublicCollection struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Fork        struct {
		Label     string `json:"label"`
		CreatedAt string `json:"createdAt"`
	} `json:"fork"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// PostmanPublicAPIResponse represents the response from Postman's public API
type PostmanPublicAPIResponse struct {
	Collections []PostmanPublicCollection `json:"collections"`
	Meta        struct {
		NextCursor string `json:"nextCursor"`
		HasMore    bool   `json:"hasMore"`
		Total      int    `json:"total"`
	} `json:"meta"`
}

// PostmanCollectionDetail represents detailed collection information
type PostmanCollectionDetail struct {
	Collection struct {
		Info struct {
			PostmanID   string `json:"_postman_id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Schema      string `json:"schema"`
		} `json:"info"`
		Item []PostmanItem `json:"item"`
		Auth struct {
			Type   string `json:"type"`
			Bearer []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
				Type  string `json:"type"`
			} `json:"bearer"`
		} `json:"auth"`
		Event    []PostmanEvent `json:"event"`
		Variable []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"variable"`
	} `json:"collection"`
}

// PostmanItem represents a request item in a Postman collection
type PostmanItem struct {
	Name     string         `json:"name"`
	Request  PostmanRequest `json:"request"`
	Response []interface{}  `json:"response"`
	Item     []PostmanItem  `json:"item,omitempty"` // for nested folders
}

// PostmanRequest represents a request in a Postman collection
type PostmanRequest struct {
	Method string `json:"method"`
	Header []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"header"`
	Body struct {
		Mode    string `json:"mode"`
		Raw     string `json:"raw"`
		Options struct {
			Raw struct {
				Language string `json:"language"`
			} `json:"raw"`
		} `json:"options"`
	} `json:"body"`
	URL struct {
		Raw   string   `json:"raw"`
		Host  []string `json:"host"`
		Path  []string `json:"path"`
		Query []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"query"`
	} `json:"url"`
	Description string `json:"description"`
}

// PostmanEvent represents an event in a Postman collection
type PostmanEvent struct {
	Listen string `json:"listen"`
	Script struct {
		Type string   `json:"type"`
		Exec []string `json:"exec"`
	} `json:"script"`
}

// PopularAPIs contains a curated list of popular APIs for the demo
var PopularAPIs = []PublicAPI{
	{
		Name:        "JSONPlaceholder",
		Description: "Fake Online REST API for Testing and Prototyping",
		Category:    "Testing",
		BaseURL:     "https://jsonplaceholder.typicode.com",
		AuthType:    "none",
		Tags:        []string{"testing", "json", "rest", "fake-data"},
		PostmanURL:  "https://www.postman.com/postman/workspace/published-postman-templates/collection/631643-f695cab7-6878-eb55-7943-ad88e1ccfd65",
		PostmanID:   "631643-f695cab7-6878-eb55-7943-ad88e1ccfd65",
		IsActive:    true,
	},
	{
		Name:        "OpenWeatherMap API",
		Description: "Weather data for any location including over 200,000 cities",
		Category:    "Weather",
		BaseURL:     "https://api.openweathermap.org/data/2.5",
		AuthType:    "apikey",
		Tags:        []string{"weather", "climate", "forecast"},
		PostmanURL:  "https://www.postman.com/postman/workspace/postman-public-workspace/collection/12959542-dcf7d8b4-434c-4f5e-b3f4-6d90b82c9a4c",
		PostmanID:   "12959542-dcf7d8b4-434c-4f5e-b3f4-6d90b82c9a4c",
		IsActive:    true,
	},
	{
		Name:        "GitHub API",
		Description: "Access GitHub's REST API for repositories, users, and organizations",
		Category:    "Developer Tools",
		BaseURL:     "https://api.github.com",
		AuthType:    "bearer",
		Tags:        []string{"github", "git", "repository", "developer"},
		PostmanURL:  "https://www.postman.com/postman/workspace/github-rest-api/collection/10131015-37136a85-47b1-44ac-a2cc-343450d9a3eb",
		PostmanID:   "10131015-37136a85-47b1-44ac-a2cc-343450d9a3eb",
		IsActive:    true,
	},
	{
		Name:        "REST Countries",
		Description: "Get information about countries via REST API",
		Category:    "Geography",
		BaseURL:     "https://restcountries.com/v3.1",
		AuthType:    "none",
		Tags:        []string{"countries", "geography", "flags", "currency"},
		PostmanURL:  "https://www.postman.com/postman/workspace/rest-countries-api/collection/1559645-8ac79ce4-dfcb-4019-8344-6b5c68bb29aa",
		PostmanID:   "1559645-8ac79ce4-dfcb-4019-8344-6b5c68bb29aa",
		IsActive:    true,
	},
	{
		Name:        "Dog CEO API",
		Description: "Random dog images and breed information",
		Category:    "Fun",
		BaseURL:     "https://dog.ceo/api",
		AuthType:    "none",
		Tags:        []string{"dogs", "images", "animals", "fun"},
		PostmanURL:  "https://www.postman.com/postman/workspace/published-postman-templates/collection/631643-cc045093-49b5-4445-a582-3f69779e93c1",
		PostmanID:   "631643-cc045093-49b5-4445-a582-3f69779e93c1",
		IsActive:    true,
	},
	{
		Name:        "Stripe API",
		Description: "Accept payments online and in mobile apps",
		Category:    "Payments",
		BaseURL:     "https://api.stripe.com/v1",
		AuthType:    "bearer",
		Tags:        []string{"payments", "stripe", "billing", "e-commerce"},
		PostmanURL:  "https://www.postman.com/stripe-dev/workspace/stripe-developers/collection/665823-be3e5b87-7ffc-4308-a0f9-643e4fe81ec0",
		PostmanID:   "665823-be3e5b87-7ffc-4308-a0f9-643e4fe81ec0",
		IsActive:    true,
	},
	{
		Name:        "SpaceX API",
		Description: "SpaceX launches, rockets, capsules, and company information",
		Category:    "Space",
		BaseURL:     "https://api.spacexdata.com/v4",
		AuthType:    "none",
		Tags:        []string{"spacex", "rockets", "launches", "space"},
		PostmanURL:  "https://www.postman.com/spacex-api/workspace/spacex/collection/10927015-bccfe645-091b-4c86-826c-2ad8b8e1c0e1",
		PostmanID:   "10927015-bccfe645-091b-4c86-826c-2ad8b8e1c0e1",
		IsActive:    true,
	},
	{
		Name:        "Cat Facts API",
		Description: "Daily cat facts for your applications",
		Category:    "Fun",
		BaseURL:     "https://catfact.ninja",
		AuthType:    "none",
		Tags:        []string{"cats", "facts", "animals", "fun"},
		PostmanURL:  "https://www.postman.com/postman/workspace/published-postman-templates/collection/631643-3cbb2e4e-4bd7-4e4e-a85b-24e5fc7e5726",
		PostmanID:   "631643-3cbb2e4e-4bd7-4e4e-a85b-24e5fc7e5726",
		IsActive:    true,
	},
}
