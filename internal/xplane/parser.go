package xplane

import (
	"encoding/json"
	"fmt"
	"io"
)

func Parse(r io.ReadCloser) (*Resource, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from stdin: %w", err)
	}

	var data *Resource
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("Failed to decode JSON: %w", err)
	}

	return data, nil
}
