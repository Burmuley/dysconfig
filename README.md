# Dysfunctional options pattern in Go (Golang)

This code generator is inspired by the article [Dysfunctional options pattern in Go](https://rednafi.com/go/dysfunctional_options_pattern/) by [Redowan Delowar](https://rednafi.com/).

Now you can define configuration struct(s) schema in a tiny JSON file and use this generator to make up Go code for you.

Table of Contents
=================

* [Installation](#installation)
* [Command line parameters](#command-line-parameters)
* [go:generate example](#gogenerate-example)
* [Configuration schema](#configuration-schema)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)

## Installation

```shell
go install -u github.com/burmuley/dysconfig@latest
```

## Command line parameters

* `-schema` - path to the JSON file with the schema; default: `config_schema.json`
* `-output` - path to the output file; default: `stdout`
* `-package` - package name; default: `main`
* `-headers` - add header and footer wrapping the generated output; default: `true`

## go:generate example

```go
//go:generate dysconfig -schema=config_schema.json -output=config.go -package=testoutput
```

## Configuration schema

More information on the configuration schema see in [schema.json](schema.json) file.

Basically JSON should contain a list of data structures you want to generate.
Thus you can store all configuration data structures definitions in a single JSON file, or you can decouple it into multiple if needed.

Each data structure definition include the following easy to grasp key-value pairs:
* `struct_name` - (**required**) the name of the data structure will be used in the Go code
* `fields` - (**required**) dictionary of data structure fields (parameters) definitions, where key is the name of the field
* `json_tags` - (**optional**) flag to enable/disable automatic JSON tags generation for each field in the `fields` list; default: `false`
* `optionals` - (**optional**) flag to enable/disable automatic optional fields helper methods generation (those `With*` functions); default: `true`

Each field value in `fields` dictionary also has these key-values you can use to adjust the generated Go code:
* `type` - (**required**) string name of the type you want this fiueld to be, can be any value even if it is not included in the configuration schema (like when you have a data type defined somewhere else in your code)
* `default` - (**optional**) string containing default value for the field to be included in the constructor function
* `required` - (**optional**) flag determines whether the filed is required or optional
* `tags` - (**optional**) list of strings containing extra tags to add to the field

```json
[
  {
    "struct_name": "DatabaseConfig",
    "json_tags": true,
    "optionals": true,
    "fields": {
      "Address": {
        "type": "string",
        "required": true,
        "tags": ["env:DATABASE_ADDRESS"]
      },
      "Login": {
        "type": "string",
        "required": false,
        "default": "root"
      },
      "Password": {
        "type": "string",
        "required": false,
        "default": "secret_password"
      },
      "DatabaseName": {
        "type": "string",
        "required": true
      }
    }
  }
]
```

After running the generator, you will get the following code:

```shell
dysconfig -schema example.json -output config.go -package testoutput
```

```go
package testoutput

type DatabaseConfig struct {
	Address      string `json:"Address,required" env:DATABASE_ADDRESS`
	Login        string `json:"Login"`
	DatabaseName string `json:"DatabaseName,required"`
	Password     string `json:"Password"`
}

func NewDatabaseConfig(address string, databasename string) *DatabaseConfig {
	return &DatabaseConfig{
		Address:      address,
		Login:        "root",
		DatabaseName: databasename,
		Password:     "secret_password",
	}
}

func (c *DatabaseConfig) WithLogin(value string) *DatabaseConfig {
	c.Login = value
	return c
}

func (c *DatabaseConfig) WithPassword(value string) *DatabaseConfig {
	c.Password = value
	return c
}
```

If you try to set `optionals` parameter to `false`, then result will look like:

```go
package testoutput

type DatabaseConfig struct {
	Address      string `json:"Address,required" env:DATABASE_ADDRESS`
	Login        string `json:"Login"`
	DatabaseName string `json:"DatabaseName,required"`
	Password     string `json:"Password"`
}

func NewDatabaseConfig(address string, databasename string) *DatabaseConfig {
	return &DatabaseConfig{
		Address:      address,
		Login:        "root",
		DatabaseName: databasename,
		Password:     "secret_password",
	}
}
```
