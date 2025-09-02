package util

import (
	"encoding/json"
	"log"
)

func StructToMapJSON(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	log.Printf("StructToMapJSON: %+v", m)
	return m, nil
}
