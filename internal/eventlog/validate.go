package eventlog

import (
	"encoding/json"
	"fmt"
)

// validateSliceCreated checks that the data of a slice.created event has
// name, title, description, base, and base_commit. A field set to null
// counts as missing. It returns the first missing field.
func validateSliceCreated(rawData json.RawMessage) error {
	required := []string{"name", "title", "description", "base", "base_commit"}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(rawData, &data); err != nil {
		return err
	}

	for _, field := range required {
		value, exists := data[field]

		if !exists || string(value) == "null" {
			return fmt.Errorf("slice.created data is missing %s", field)
		}
	}
	return nil
}
