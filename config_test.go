package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- VariableType.UnmarshalJSON ---

func TestVariableTypeUnmarshalJSON(t *testing.T) {
	scalars := []struct {
		name  string
		input string
		want  string
	}{
		{"null", `null`, "nil"},
		{"integer", `42`, "42"},
		{"negative integer", `-7`, "-7"},
		{"float", `3.14`, "3.14"},
		{"string", `"hello"`, `"hello"`},
		{"empty string", `""`, `""`},
		{"bool true", `true`, "true"},
		{"bool false", `false`, "false"},
	}

	for _, tt := range scalars {
		t.Run(tt.name, func(t *testing.T) {
			var vt VariableType
			if err := json.Unmarshal([]byte(tt.input), &vt); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(vt) != tt.want {
				t.Errorf("got %q, want %q", string(vt), tt.want)
			}
		})
	}

	t.Run("map contains keys and values", func(t *testing.T) {
		var vt VariableType
		if err := json.Unmarshal([]byte(`{"key": 99}`), &vt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(vt)
		if !strings.Contains(s, `"key"`) || !strings.Contains(s, "99") {
			t.Errorf("map output %q missing expected content", s)
		}
	})

	t.Run("array contains elements", func(t *testing.T) {
		var vt VariableType
		if err := json.Unmarshal([]byte(`["foo", "bar"]`), &vt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(vt)
		if !strings.Contains(s, `"foo"`) || !strings.Contains(s, `"bar"`) {
			t.Errorf("array output %q missing expected content", s)
		}
	})

	t.Run("nested map values are recursively converted", func(t *testing.T) {
		var vt VariableType
		if err := json.Unmarshal([]byte(`{"n": 1, "s": "v"}`), &vt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s := string(vt)
		if !strings.Contains(s, `"n"`) || !strings.Contains(s, "1") {
			t.Errorf("map output %q missing numeric entry", s)
		}
		if !strings.Contains(s, `"s"`) || !strings.Contains(s, `"v"`) {
			t.Errorf("map output %q missing string entry", s)
		}
	})
}

// --- validateSchema ---

func TestValidateSchema(t *testing.T) {
	passes := []struct {
		name string
		data string
	}{
		{"empty array", `[]`},
		{"minimal valid entry", `[{"struct_name":"A","fields":{"F":{"type":"string"}}}]`},
		{"required field with type", `[{"struct_name":"A","fields":{"F":{"type":"int","required":true}}}]`},
		{"all optional properties present", `[{"struct_name":"A","json_tags":true,"optionals":false,"fields":{"F":{"type":"string","required":true,"tags":["env:X"],"override_json_tag":"y"}}}]`},
		{"multiple structs", `[{"struct_name":"A","fields":{"F":{"type":"string"}}},{"struct_name":"B","fields":{"G":{"type":"int"}}}]`},
	}
	for _, tt := range passes {
		t.Run("valid/"+tt.name, func(t *testing.T) {
			if err := validateSchema([]byte(tt.data)); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	fails := []struct {
		name string
		data string
	}{
		{"missing struct_name", `[{"fields":{"F":{"type":"string"}}}]`},
		{"missing fields", `[{"struct_name":"A"}]`},
		{"field missing type", `[{"struct_name":"A","fields":{"F":{"required":true}}}]`},
		{"malformed JSON", `{bad`},
		{"not an array", `{"struct_name":"A","fields":{}}`},
	}
	for _, tt := range fails {
		t.Run("invalid/"+tt.name, func(t *testing.T) {
			if err := validateSchema([]byte(tt.data)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// --- parseConfig ---

func writeTempSchema(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseConfig(t *testing.T) {
	t.Run("required field parses correctly", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"Name":{"type":"string","required":true}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg) != 1 || cfg[0].StructName != "Foo" {
			t.Errorf("unexpected result: %+v", cfg)
		}
		f := cfg[0].Fields["Name"]
		if f.Type != "string" || !f.Required {
			t.Errorf("unexpected field params: %+v", f)
		}
	})

	t.Run("optional field with default parses correctly", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"Timeout":{"type":"int","required":false,"default":30}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f := cfg[0].Fields["Timeout"]; string(f.Default) != "30" {
			t.Errorf("unexpected default value: %q", string(f.Default))
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := parseConfig(filepath.Join(t.TempDir(), "missing.json"))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		p := writeTempSchema(t, `{invalid`)
		if _, err := parseConfig(p); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("optional field without default is rejected", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"Name":{"type":"string","required":false}}}]`)
		_, err := parseConfig(p)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "Name") {
			t.Errorf("error should mention field name, got: %v", err)
		}
	})

	t.Run("field missing type fails schema validation", func(t *testing.T) {
		// required:true bypasses the default check but schema validation catches missing type
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"X":{"required":true}}}]`)
		if _, err := parseConfig(p); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("optionals defaults to true when absent", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"X":{"type":"string","required":true}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatal(err)
		}
		if cfg[0].Optionals == nil || !*cfg[0].Optionals {
			t.Error("expected Optionals to default to true")
		}
	})

	t.Run("optionals explicit false is preserved", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","optionals":false,"fields":{"X":{"type":"string","required":true}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatal(err)
		}
		if cfg[0].Optionals == nil || *cfg[0].Optionals {
			t.Error("expected Optionals to be false")
		}
	})

	t.Run("multiple structs all parsed", func(t *testing.T) {
		p := writeTempSchema(t, `[
			{"struct_name":"Foo","fields":{"A":{"type":"string","required":true}}},
			{"struct_name":"Bar","fields":{"B":{"type":"int","required":true}}}
		]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg) != 2 {
			t.Errorf("expected 2 structs, got %d", len(cfg))
		}
		names := map[string]bool{cfg[0].StructName: true, cfg[1].StructName: true}
		if !names["Foo"] || !names["Bar"] {
			t.Errorf("expected Foo and Bar, got %v", names)
		}
	})

	t.Run("all default value types accepted", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{
			"A":{"type":"string","required":false,"default":"hello"},
			"B":{"type":"int","required":false,"default":42},
			"C":{"type":"bool","required":false,"default":false},
			"D":{"type":"*string","required":false,"default":null}
		}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		checks := map[string]string{
			"A": `"hello"`,
			"B": "42",
			"C": "false",
			"D": "nil",
		}
		for field, want := range checks {
			if got := string(cfg[0].Fields[field].Default); got != want {
				t.Errorf("field %s: default = %q, want %q", field, got, want)
			}
		}
	})

	t.Run("json_tags field parsed correctly", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","json_tags":true,"fields":{"X":{"type":"string","required":true}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatal(err)
		}
		if !cfg[0].JsonTags {
			t.Error("expected JsonTags to be true")
		}
	})

	t.Run("override_json_tag parsed correctly", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"DbType":{"type":"string","required":true,"override_json_tag":"type"}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatal(err)
		}
		if cfg[0].Fields["DbType"].OverrideJsonTag != "type" {
			t.Errorf("unexpected OverrideJsonTag: %q", cfg[0].Fields["DbType"].OverrideJsonTag)
		}
	})

	t.Run("tags list parsed correctly", func(t *testing.T) {
		p := writeTempSchema(t, `[{"struct_name":"Foo","fields":{"Addr":{"type":"string","required":true,"tags":["env:\"DB_ADDR\""]}}}]`)
		cfg, err := parseConfig(p)
		if err != nil {
			t.Fatal(err)
		}
		tags := cfg[0].Fields["Addr"].Tags
		if len(tags) != 1 || tags[0] != `env:"DB_ADDR"` {
			t.Errorf("unexpected tags: %v", tags)
		}
	})
}
