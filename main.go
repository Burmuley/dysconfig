package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func getOutput(o string) io.Writer {
	if o == "stdout" {
		return os.Stdout
	}
	f, err := os.Create(o)
	if err != nil {
		log.Fatalf("error creating file: %s", err)
	}
	return f
}

func main() {
	// setup and parse command line arguments
	schemaFile := flag.String("schema", "config_schema.json", "configuration schema file")
	outputDst := flag.String("output", "stdout", "output file name")
	packageName := flag.String("package", "main", "package name for generated files")
	addHeaderFooter := flag.Bool("headers", true, "add header to generated files")
	flag.Parse()

	// get output destination
	output := getOutput(*outputDst)

	// prepare templates
	tmpls, err := prepareTemplates()
	if err != nil {
		log.Fatalf("error preparing templates: %s", err)
	}
	// parse schema
	schema, err := parseSchema(*schemaFile)
	if err != nil {
		log.Fatalf("error parsing schema: %s", err)
	}
	// generate code
	formatted, err := genFromSchema(tmpls, schema, *addHeaderFooter, *packageName)
	fmt.Fprint(output, string(formatted))
}
