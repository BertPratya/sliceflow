package eventlog

import (
	"encoding/json"
	"fmt"
)

// decodeSliceCreated checks that the data of a slice.created event has
// name, title, description, base, and base_commit, and returns it as a
// SliceCreated. A field set to null counts as missing. It returns the
// first missing field, or an error when a field is not a string.
func decodeSliceCreated(rawData json.RawMessage) (any, error) {
	required := []string{"name", "title", "description", "base", "base_commit"}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}

	for _, field := range required {
		value, exists := data[field]

		if !exists || string(value) == "null" {
			return nil, fmt.Errorf("slice.created data is missing %s", field)
		}
	}

	var created SliceCreated
	if err := json.Unmarshal(rawData, &created); err != nil {
		return nil, fmt.Errorf("slice.created data: %w", err)
	}
	return created, nil
}

// decodeSliceEdited checks that the data of a slice.edited event holds
// title, description, or both, each a string, and no other key, and
// returns it as a SliceEdited. A field set to null is not a string.
func decodeSliceEdited(rawData json.RawMessage) (any, error) {
	var data map[string]json.RawMessage
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("slice.edited data has no title or description")
	}

	var edited SliceEdited
	for key, value := range data {
		var field **string
		switch key {
		case "title":
			field = &edited.Title
		case "description":
			field = &edited.Description
		default:
			return nil, fmt.Errorf("slice.edited data has the key %s, which it may not hold", key)
		}
		if err := json.Unmarshal(value, field); err != nil || *field == nil {
			return nil, fmt.Errorf("slice.edited data has a %s that is not a string", key)
		}
	}
	return edited, nil
}
