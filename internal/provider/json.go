package provider

import (
	"encoding/json"
	"fmt"
)

// DecodeJSON unmarshals JSON into dest (shared path for API fixtures and live responses).
func DecodeJSON(data []byte, dest any) error {
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("provider json: %w", err)
	}
	return nil
}
