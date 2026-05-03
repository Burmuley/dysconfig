package main

import (
	"strings"
	"testing"
	"text/template"
)

// --- caseDiff ---

func TestCaseDiff(t *testing.T) {
	tests := []struct {
		a, b rune
		want int
	}{
		{'A', 'a', 1},  // upper followed by lower
		{'a', 'A', -1}, // lower followed by upper
		{'a', 'b', 0},  // both lower
		{'A', 'B', 0},  // both upper
		{'_', 'a', 0},  // underscore as first
		{'a', '_', 0},  // underscore as second
		{'_', '_', 0},  // both underscore
	}
	for _, tt := range tests {
		got := caseDiff(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("caseDiff(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// --- ConvertType ---

func TestConvertType(t *testing.T) {
	if got := ConvertType(42); got != 42 {
		t.Errorf("int: got %v, want 42", got)
	}
	if got, want := ConvertType("hello"), `"hello"`; got != want {
		t.Errorf("string: got %v, want %v", got, want)
	}
	if got := ConvertType(true); got != true {
		t.Errorf("bool: got %v, want true", got)
	}
	if got := ConvertType(3.14); got != 3.14 {
		t.Errorf("float: got %v, want 3.14", got)
	}
}

// --- dict ---

func TestDict(t *testing.T) {
	t.Run("creates map from pairs", func(t *testing.T) {
		m := dict("a", 1, "b", "two")
		if m["a"] != 1 || m["b"] != "two" {
			t.Errorf("unexpected map: %v", m)
		}
	})

	t.Run("empty args returns empty map", func(t *testing.T) {
		m := dict()
		if len(m) != 0 {
			t.Errorf("expected empty map, got %v", m)
		}
	})

	t.Run("panics on odd number of arguments", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic for odd args")
			}
		}()
		dict("only_key") //nolint:staticcheck
	})

	t.Run("panics on non-string key", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected panic for non-string key")
			}
		}()
		dict(42, "value") //nolint:staticcheck
	})
}

// --- toSnakeCase ---

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"alreadySnakeCase", "already_snake_case"},
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"aA", "a_a"},
		{"snake_case", "snake_case"},
		{"snake_Case", "snake_case"},
		{"APIResponse", "api_response"},
		{"UserID", "user_id"},
		{"HTML", "html"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := toSnakeCase(test.input)
			if result != test.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

// --- prepareTemplates ---

func TestPrepareTemplates(t *testing.T) {
	tmpls, err := prepareTemplates()
	if err != nil {
		t.Fatalf("prepareTemplates() returned an error: %v", err)
	}
	for _, name := range tmplNames {
		if _, ok := tmpls[name]; !ok {
			t.Errorf("expected template %s not found", name)
		}
	}
}

// --- genFromConfig (mock templates) ---

func TestGenFromConfig(t *testing.T) {
	tests := []struct {
		name        string
		tmpls       map[string]*template.Template
		config      Config
		ah          bool
		packageName string
		wantErr     bool
	}{
		{
			name: "valid schema produces formatted output",
			tmpls: map[string]*template.Template{
				"struct.tmpl":      template.Must(template.New("struct.tmpl").Parse("type {{.StructName}} struct {}\n")),
				"constructor.tmpl": template.Must(template.New("constructor.tmpl").Parse("func New{{.StructName}}() *{{.StructName}} { return &{{.StructName}}{} }\n")),
				"optional.tmpl":    template.Must(template.New("optional.tmpl").Parse("")),
			},
			config:      Config{{StructName: "TestStruct"}},
			ah:          false,
			packageName: "testpackage",
			wantErr:     false,
		},
		{
			name: "template referencing unknown field produces invalid Go and fails",
			tmpls: map[string]*template.Template{
				"struct.tmpl":      template.Must(template.New("struct.tmpl").Parse("type {{.InvalidField}} struct {}")),
				"constructor.tmpl": template.Must(template.New("constructor.tmpl").Parse("")),
				"optional.tmpl":    template.Must(template.New("optional.tmpl").Parse("")),
			},
			config:      Config{{StructName: "TestStruct"}},
			ah:          false,
			packageName: "testpackage",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := genFromConfig(tt.tmpls, tt.config, tt.ah, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("genFromConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- genFromConfig (real embedded templates) ---

// TestGenFromConfigOutput validates the actual Go code produced by the embedded templates.
func TestGenFromConfigOutput(t *testing.T) {
	tmpls, err := prepareTemplates()
	if err != nil {
		t.Fatalf("prepareTemplates() error: %v", err)
	}

	trueVal, falseVal := true, false

	gen := func(t *testing.T, cfg Config, header bool) string {
		t.Helper()
		out, err := genFromConfig(tmpls, cfg, header, "testpkg")
		if err != nil {
			t.Fatalf("genFromConfig() error: %v", err)
		}
		return string(out)
	}

	assertContains := func(t *testing.T, out string, fragments ...string) {
		t.Helper()
		for _, f := range fragments {
			if !strings.Contains(out, f) {
				t.Errorf("output missing %q\nfull output:\n%s", f, out)
			}
		}
	}

	assertAbsent := func(t *testing.T, out, fragment string) {
		t.Helper()
		if strings.Contains(out, fragment) {
			t.Errorf("output should not contain %q\nfull output:\n%s", fragment, out)
		}
	}

	t.Run("package declaration", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Pkg",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields:     Fields{"X": {Type: "int", Required: true}},
		}}, false)
		if !strings.HasPrefix(out, "package testpkg") {
			t.Errorf("output should begin with package declaration:\n%s", out)
		}
	})

	t.Run("header comment included when requested", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "H",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields:     Fields{"X": {Type: "int", Required: true}},
		}}, true)
		assertContains(t, out, "DO NOT EDIT")
	})

	t.Run("struct definition emitted", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Simple",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields:     Fields{"Name": {Type: "string", Required: true}},
		}}, false)
		assertContains(t, out,
			"type Simple struct {",
			"Name string",
		)
	})

	t.Run("no json tags when json_tags is false", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "NoTags",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields:     Fields{"Name": {Type: "string", Required: true}},
		}}, false)
		assertAbsent(t, out, "json:")
	})

	t.Run("json tag for required field uses snake_case with required marker", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Tagged",
			JsonTags:   true,
			Optionals:  &falseVal,
			Fields:     Fields{"HostName": {Type: "string", Required: true}},
		}}, false)
		assertContains(t, out, `json:"host_name,required"`)
	})

	t.Run("json tag for optional field has no required marker", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Tagged",
			JsonTags:   true,
			Optionals:  &falseVal,
			Fields:     Fields{"Port": {Type: "int", Required: false, Default: "8080"}},
		}}, false)
		assertContains(t, out, `json:"port"`)
		assertAbsent(t, out, "required")
	})

	t.Run("override_json_tag replaces field name in tag", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Tagged",
			JsonTags:   true,
			Optionals:  &falseVal,
			Fields:     Fields{"DbType": {Type: "string", Required: true, OverrideJsonTag: "type"}},
		}}, false)
		assertContains(t, out, `json:"type,required"`)
		assertAbsent(t, out, `json:"db_type`)
	})

	t.Run("extra tags appended after json tag", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Env",
			JsonTags:   true,
			Optionals:  &falseVal,
			Fields:     Fields{"Addr": {Type: "string", Required: true, Tags: []string{`env:"DB_ADDR"`}}},
		}}, false)
		assertContains(t, out, `env:"DB_ADDR"`)
	})

	t.Run("constructor has required fields as params", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Ctor",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields:     Fields{"Addr": {Type: "string", Required: true}},
		}}, false)
		assertContains(t, out,
			"func NewCtor(addr string) *Ctor {",
			"Addr: addr,",
		)
	})

	t.Run("constructor with no required fields takes no params", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Ctor",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields:     Fields{"Port": {Type: "int", Required: false, Default: "8080"}},
		}}, false)
		assertContains(t, out, "func NewCtor() *Ctor {")
	})

	t.Run("constructor sets optional fields to their defaults", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "Ctor",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields: Fields{
				"Req":     {Type: "string", Required: true},
				"Timeout": {Type: "int", Required: false, Default: "30"},
				"Debug":   {Type: "bool", Required: false, Default: "false"},
				"Prefix":  {Type: "string", Required: false, Default: `"app"`},
				"Ptr":     {Type: "*string", Required: false, Default: "nil"},
			},
		}}, false)
		// go/format aligns composite-literal colons, so check field name and value separately
		assertContains(t, out,
			"Timeout:", "30,",
			"Debug:", "false,",
			"Prefix:", `"app",`,
			"Ptr:", "nil,",
		)
	})

	t.Run("With* methods generated when optionals true", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "WithOpts",
			JsonTags:   false,
			Optionals:  &trueVal,
			Fields: Fields{
				"Host": {Type: "string", Required: true},
				"Port": {Type: "int", Required: false, Default: "8080"},
			},
		}}, false)
		assertContains(t, out,
			"func (c *WithOpts) WithPort(value int) *WithOpts {",
			"c.Port = value",
			"return c",
		)
	})

	t.Run("With* methods not generated when optionals false", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "NoOpts",
			JsonTags:   false,
			Optionals:  &falseVal,
			Fields: Fields{
				"Host": {Type: "string", Required: true},
				"Port": {Type: "int", Required: false, Default: "8080"},
			},
		}}, false)
		assertAbsent(t, out, "func (c *NoOpts) With")
	})

	t.Run("required field has no With* method", func(t *testing.T) {
		out := gen(t, Config{{
			StructName: "WithOpts",
			JsonTags:   false,
			Optionals:  &trueVal,
			Fields: Fields{
				"Host": {Type: "string", Required: true},
				"Port": {Type: "int", Required: false, Default: "9090"},
			},
		}}, false)
		assertAbsent(t, out, "WithHost")
	})

	t.Run("multiple structs all generated", func(t *testing.T) {
		out := gen(t, Config{
			{
				StructName: "Alpha",
				JsonTags:   false,
				Optionals:  &falseVal,
				Fields:     Fields{"X": {Type: "int", Required: true}},
			},
			{
				StructName: "Beta",
				JsonTags:   false,
				Optionals:  &trueVal,
				Fields: Fields{
					"Y": {Type: "string", Required: true},
					"Z": {Type: "bool", Required: false, Default: "true"},
				},
			},
		}, false)
		assertContains(t, out,
			"type Alpha struct {",
			"type Beta struct {",
			"func NewAlpha(",
			"func NewBeta(",
			"func (c *Beta) WithZ(value bool) *Beta {",
		)
	})
}
