package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"log"
	"path"
	"strings"
	"text/template"
)

var (
	tmplNames  = []string{"struct.tmpl", "constructor.tmpl", "optional.tmpl"}
	headerTmpl = "header.tmpl"
	footerTmpl = "footer.tmpl"
)

//go:embed templates/*
var tmplFiles embed.FS

func prepareTemplates() (map[string]*template.Template, error) {
	var err error
	tmpls := make(map[string]*template.Template)
	funcs := template.FuncMap{
		"ToLower": strings.ToLower,
		"Join":    func(sep string, str ...string) string { return strings.Join(str, sep) },
		"Title":   strings.Title,
		"Dict":    dict,
		"ConvertType": func(t any) any {
			switch i := t.(type) {
			case int:
				return i
			case string:
				return fmt.Sprintf("%q", i)
			default:
				return t
			}
		},
	}

	for _, v := range tmplNames {
		if tmpls[v], err = template.New(v).Funcs(funcs).ParseFS(tmplFiles, path.Join("templates", v)); err != nil {
			return nil, err
		}
	}

	return tmpls, nil
}

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
				log.Fatalf("error executing template: %s", err)
			}
		}
	}

	if ah {
		addFooter(genBuf)
	}

	formatted, err := format.Source(genBuf.Bytes())
	if err != nil {
		fmt.Println(genBuf.String())
		log.Fatalf("error formatting source: %s", err)
	}

	return formatted, nil
}

func addHeader(buf *bytes.Buffer) {
	header, _ := tmplFiles.ReadFile(path.Join("templates", headerTmpl))
	buf.Write(header)
	return
}

func addFooter(buf *bytes.Buffer) {
	header, _ := tmplFiles.ReadFile(path.Join("templates", footerTmpl))
	buf.Write(header)
	return
}

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
