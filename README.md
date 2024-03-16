# Dysfunctional options pattern in Go

This code generator is inspired by the article [Dysfunctional options pattern in Go](https://rednafi.com/go/dysfunctional_options_pattern/) by [Redowan Delowar](https://rednafi.com/).

Now you can define configuration struct(s) schema in a tiny JSON file and use this generator to make up Go code for you.

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

## Configuration schema example

```json
[
  {
    "struct_name": "DatabaseConfig",
    "json_tags": true,
    "fields": {
      "Address": {
        "type": "string",
        "required": true
      },
      "Aogin": {
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
	Address      string `json:"Address,required"`
	Aogin        string `json:"Aogin"`
	DatabaseName string `json:"DatabaseName,required"`
	Password     string `json:"Password"`
}

func NewDatabaseConfig(address string, databasename string) *DatabaseConfig {
	return &DatabaseConfig{
		Address:      address,
		Aogin:        "root",
		DatabaseName: databasename,
		Password:     "secret_password",
	}
}

func (c *DatabaseConfig) WithAogin(value string) *DatabaseConfig {
	c.Aogin = value
	return c
}

func (c *DatabaseConfig) WithPassword(value string) *DatabaseConfig {
	c.Password = value
	return c
}
```
