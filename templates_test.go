package main

import (
	"testing"
	"text/template"
)

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

func TestGenFromSchema(t *testing.T) {
	tests := []struct {
		name        string
		tmpls       map[string]*template.Template
		schema      Schema
		ah          bool
		packageName string
		wantErr     bool
	}{
		{
			name: "Valid schema",
			tmpls: map[string]*template.Template{
				"struct.tmpl":      template.Must(template.New("struct.tmpl").Parse("type {{.StructName}} struct {}\n")),
				"constructor.tmpl": template.Must(template.New("constructor.tmpl").Parse("func New{{.StructName}}() *{{.StructName}} { return &{{.StructName}}{} }")),
				"optional.tmpl":    template.Must(template.New("optional.tmpl").Parse("")),
			},
			schema:      Schema{{StructName: "TestStruct"}},
			ah:          false,
			packageName: "testpackage",
			wantErr:     false,
		},
		{
			name: "Invalid template",
			tmpls: map[string]*template.Template{
				"struct.tmpl":      template.Must(template.New("struct.tmpl").Parse("type {{.InvalidField}} struct {}")),
				"constructor.tmpl": template.Must(template.New("constructor.tmpl").Parse("")),
				"optional.tmpl":    template.Must(template.New("optional.tmpl").Parse("")),
			},
			schema:      Schema{{StructName: "TestStruct"}},
			ah:          false,
			packageName: "testpackage",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := genFromSchema(tt.tmpls, tt.schema, tt.ah, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("genFromSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPrepareTemplates(t *testing.T) {
	tmpls, err := prepareTemplates()
	if err != nil {
		t.Fatalf("prepareTemplates() returned an error: %v", err)
	}

	// Check if all expected templates are present
	for _, name := range tmplNames {
		if _, ok := tmpls[name]; !ok {
			t.Errorf("Expected template %s not found in result", name)
		}
	}
}
