package main

import (
	"path"
	"reflect"
	"strings"
	"text/template"
)

var tmplNames = []string{"struct.tmpl", "constructor.tmpl", "optional.tmpl"}

func prepareTemplates() (map[string]*template.Template, error) {
	var err error
	tmpls := make(map[string]*template.Template)
	funcs := template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
		"IntSub":  func(a, b int) int { return a - b },
		"Join":    func(sep string, str ...string) string { return strings.Join(str, sep) },
		"Last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"Title": strings.Title,
	}

	for _, v := range tmplNames {
		if tmpls[v], err = template.New(v).Funcs(funcs).ParseFiles(path.Join("templates", v)); err != nil {
			return nil, err
		}
	}

	return tmpls, nil
}
