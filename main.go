package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// getOutput returns an io.Writer based on the given output destination.
// If 'o' is "stdout", it returns os.Stdout. Otherwise, it creates and returns a file with the given name.
// It returns an error if there's an issue creating the file.
func getOutput(o string) (io.Writer, error) {
	if o == "stdout" {
		return os.Stdout, nil
	}
	f, err := os.Create(o)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	return f, nil
}

func main() {
	// setup and parse command line arguments
	configFile := flag.String("config", "config.json", "configuration file")
	outputDst := flag.String("output", "stdout", "output file name")
	packageName := flag.String("package", "main", "package name for generated files")
	addHeaderFooter := flag.Bool("headers", true, "add header to generated files")
	flag.Parse()

	// get output destination
	output, err := getOutput(*outputDst)
	if err != nil {
		log.Fatalf("error getting output destination: %s", err)
	}

	// prepare templates
	tmpls, err := prepareTemplates()
	if err != nil {
		log.Fatalf("error preparing templates: %s", err)
	}
	// parse schema
	schema, err := parseConfig(*configFile)
	if err != nil {
		log.Fatalf("error parsing config: %s", err)
	}
	// generate code
	formatted, err := genFromConfig(tmpls, schema, *addHeaderFooter, *packageName)
	if err != nil {
		log.Fatalf("error generating templates: %s", err)
	}
	fmt.Fprint(output, string(formatted))
}
