package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type FieldParameters struct {
	Type     string
	Required bool
}

type Fields map[string]FieldParameters

type Schema []struct {
	StructName string
	Fields     Fields
}

func main() {
	tmpls, err := prepareTemplates()
	if err != nil {
		log.Fatalf("error preparing templates: %s", err)
	}

	file := "example.json"
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Fatalf("error opening file: %s", err)
	}

	schema := make(Schema, 0)
	jsonParser := json.NewDecoder(f)
	if err := jsonParser.Decode(&schema); err != nil {
		log.Fatalf("error parsing json: %s", err)
	}
	genBuf := bytes.NewBuffer([]byte{})
	for _, s := range schema {
		for _, t := range tmplNames {
			err := tmpls[t].Execute(genBuf, s)
			if err != nil {
				log.Fatalf("error executing template: %s", err)
			}
		}
	}
	fmt.Println(genBuf.String())
}
