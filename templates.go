package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"path"
	"strings"
	"text/template"
)

var (
	tmplNames  = []string{"struct.tmpl", "constructor.tmpl", "optional.tmpl"}
	headerTmpl = "header.tmpl"
)

//go:embed templates/*
var tmplFiles embed.FS

// prepareTemplates initializes and prepares a map of template.Template objects
// for code generation. It sets up custom template functions and parses template
// files from the embedded filesystem.
func prepareTemplates() (map[string]*template.Template, error) {
	var err error
	tmpls := make(map[string]*template.Template)
	funcs := template.FuncMap{
		"ToLower":     strings.ToLower,
		"Join":        func(sep string, str ...string) string { return strings.Join(str, sep) },
		"Title":       strings.Title,
		"Dict":        dict,
		"ConvertType": ConvertType,
		"UnrefBool":   func(p *bool) bool { return *p },
		"ToSnakeCase": toSnakeCase,
	}

	for _, v := range tmplNames {
		if tmpls[v], err = template.New(v).Funcs(funcs).ParseFS(tmplFiles, path.Join("templates", v)); err != nil {
			return nil, err
		}
	}

	return tmpls, nil
}

// genFromSchema generates Go code from the provided schema using the prepared
// templates. It applies the templates to each schema item and formats the
// resulting code.
func genFromSchema(tmpls map[string]*template.Template, schema Schema, ah bool, packageName string) ([]byte, error) {
	genBuf := bytes.NewBuffer([]byte{})
	genBuf.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	if ah {
		addHeader(genBuf)
	}

	for _, s := range schema {
		for _, t := range tmplNames {
			err := tmpls[t].Execute(genBuf, s)
			if err != nil {
				return nil, fmt.Errorf("error executing template: %w", err)
			}
		}
	}

	formatted, err := format.Source(genBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error formatting source: %w", err)
	}

	return formatted, nil
}

// addHeader reads the header template file and writes its content to the
// provided buffer.
func addHeader(buf *bytes.Buffer) {
	header, _ := tmplFiles.ReadFile(path.Join("templates", headerTmpl))
	buf.Write(header)
	return
}

// dict is a helper function that creates a map from a list of key-value pairs.
// It's used in templates to create dynamic map structures.
func dict(elem ...any) map[string]interface{} {
	if len(elem)%2 != 0 {
		panic("dict: odd number of elements")
	}
	d := make(map[string]interface{}, len(elem)/2)
	for i := 0; i < len(elem); i += 2 {
		key, ok := elem[i].(string)
		if !ok {
			panic("dict: key not a string")
		}
		d[key] = elem[i+1]
	}
	return d
}

// ConvertType converts the input to a suitable type for template rendering.
// It returns integers as-is, wraps strings in quotes, and leaves other types unchanged.
func ConvertType(t any) any {
	switch i := t.(type) {
	case int:
		return i
	case string:
		return fmt.Sprintf("%q", i)
	default:
		return t
	}
}

// toSnakeCase converts a string to snake_case format.
// It handles various input cases, including camelCase and PascalCase.
func toSnakeCase(input string) string {
	if len(input) < 2 {
		return strings.ToLower(input)
	}

	input_runes := []rune(input)
	camel_result := ""
	uadded := false // underscore added?
	for i := 0; i < len(input_runes)-1; i++ {
		crl := strings.ToLower(string(input_runes[i]))
		cd := caseDiff(input_runes[i], input_runes[i+1])

		if cd == 1 && !uadded && i > 0 {
			camel_result = fmt.Sprint(camel_result, "_", crl)
			uadded = true
			continue
		}

		if cd == -1 && !uadded {
			camel_result = fmt.Sprint(camel_result, crl, "_")
			uadded = true
			continue
		}

		camel_result = fmt.Sprint(camel_result, crl)
		if input_runes[i] == '_' {
			uadded = true
		} else {
			uadded = false
		}
	}

	camel_result = fmt.Sprint(camel_result, strings.ToLower(string(input_runes[len(input_runes)-1])))
	return camel_result
}

// caseDiff compares two runes and returns:
//
//	 1 if the first is uppercase and the second is lowercase,
//	-1 if the first is lowercase and the second is uppercase,
//	 0 otherwise (including when either rune is an underscore).
func caseDiff(a, b rune) int {
	// if any of runes is underscore - cound pair as equal case
	if a == '_' || b == '_' {
		return 0
	}

	aIsUpper, bIsUpper := ('A' <= a && a <= 'Z'), ('A' <= b && b <= 'Z')
	if aIsUpper && !bIsUpper {
		return 1
	}
	if !aIsUpper && bIsUpper {
		return -1
	}

	return 0
}
