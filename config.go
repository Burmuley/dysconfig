package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	_ "embed"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schema.json
var jsonSchemaData []byte

var optionals_default bool = true

// VariableType represents a flexible type that can unmarshal various JSON data types.
type VariableType string

// UnmarshalJSON implements the json.Unmarshaler interface for VariableType.
// It attempts to unmarshal the input bytes into different data types (null, number, string, bool, map, array)
// and converts the result to a string representation.
func (vt *VariableType) UnmarshalJSON(b []byte) error {
	{ // try null
		var null any
		if err := json.Unmarshal(b, &null); err == nil {
			if null == nil {
				*vt = VariableType("nil")
				return nil
			}
		}
	}

	{ // try json.Number
		var n json.Number
		if err := json.Unmarshal(b, &n); err == nil {
			*vt = VariableType(n.String())
			return nil
		}
	}

	{ // try string
		var s string
		if err := json.Unmarshal(b, &s); err == nil {
			*vt = VariableType(fmt.Sprintf("%q", s))
			return nil
		}
	}

	{ // try bool
		var bs bool
		if err := json.Unmarshal(b, &bs); err == nil {
			*vt = VariableType(fmt.Sprintf("%t", bs))
			return nil
		}
	}

	{ // try map
		if b[0] == '{' {
			var m map[string]VariableType
			if err := json.Unmarshal(b, &m); err == nil {
				mb := bytes.NewBuffer([]byte{})
				mb.WriteString("{\n")
				for k, v := range m {
					mb.WriteString(fmt.Sprintf("%q: %v,\n", k, v))
				}
				mb.WriteString("}")
				*vt = VariableType(mb.String())
				return nil
			}
		}
	}

	{ // try array
		if b[0] == '[' {
			var a []VariableType
			if err := json.Unmarshal(b, &a); err == nil {
				arrb := bytes.NewBuffer([]byte{})
				arrb.WriteString("{\n")
				for _, v := range a {
					arrb.WriteString(fmt.Sprintf("%v,\n", v))
				}
				arrb.WriteString("}")
				*vt = VariableType(arrb.String())
				return nil
			}
		}
	}

	return errors.New("unknown type of default value")
}

type FieldParameters struct {
	Type            string       `json:"type"`
	Required        bool         `json:"required"`
	Default         VariableType `json:"default"`
	Tags            []string     `json:"tags"`
	OverrideJsonTag string       `json:"override_json_tag"`
}

type Fields map[string]FieldParameters

type Config []struct {
	StructName string `json:"struct_name"`
	JsonTags   bool   `json:"json_tags"`
	Optionals  *bool  `json:"optionals"`
	Fields     Fields `json:"fields"`
}

// validateSchema validates the given config_schema against the JSON schema defined in jsonSchemaData.
// It returns an error if the schema compilation, unmarshalling, or validation fails.
func validateSchema(config_schema []byte) error {
	jsonSchema, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonSchemaData))
	if err != nil {
		return fmt.Errorf("error loading JSON schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", jsonSchema); err != nil {
		return fmt.Errorf("error adding JSON schema resource: %w", err)
	}

	compiledSchema, err := compiler.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("error compiling JSON schema: %w", err)
	}

	var v any
	if err := json.Unmarshal(config_schema, &v); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := compiledSchema.Validate(v); err != nil {
		return fmt.Errorf("error validating config: %w", err)
	}

	return nil
}

// parseConfig reads a configuration file, unmarshal it into a Schema struct, sets default values for "optionals",
// and validates the schema. It returns the parsed Config and any error encountered during the process.
func parseConfig(file string) (Config, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(f)
	data := buf.Bytes()

	config := make(Config, 0)
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing schema file: %w", err)
	}

	for i := range config {
		if config[i].Optionals == nil {
			config[i].Optionals = &optionals_default
		}

		for name, value := range config[i].Fields {
			if !value.Required && value.Default == VariableType("") {
				return nil, fmt.Errorf("error validating config: fileld '%s' requires default value set", name)
			}
		}
	}

	if err := validateSchema(data); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return config, nil
}
