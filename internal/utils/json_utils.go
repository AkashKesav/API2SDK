package utils

import (
	"encoding/json"
	"fmt"
)

// MapToJSONString converts a map[string]interface{} to a JSON string.
func MapToJSONString(data map[string]interface{}) (string, error) {
	if data == nil {
		// Return empty JSON object string if map is nil, as some JS libraries might expect an object
		return "{}", nil
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal map to JSON: %w", err)
	}
	return string(jsonBytes), nil
}
