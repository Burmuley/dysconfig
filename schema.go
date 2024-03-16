package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type FieldParameters struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Default  any    `json:"default"`
}

type Fields map[string]FieldParameters

type Schema []struct {
	StructName string `json:"struct_name"`
	JsonTags   bool   `json:"json_tags"`
	Fields     Fields `json:"fields"`
}

func parseSchema(file string) (Schema, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	schema := make(Schema, 0)
	jsonParser := json.NewDecoder(f)
	if err := jsonParser.Decode(&schema); err != nil {
		return nil, fmt.Errorf("error parsing json: %w", err)
	}

	return schema, nil
}
