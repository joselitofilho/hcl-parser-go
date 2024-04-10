<div align="center">

# hcl-parser-go

[![GitHub tag](https://img.shields.io/github/release/joselitofilho/hcl-parser-go?include_prereleases=&sort=semver&color=2ea44f&style=for-the-badge)](https://github.com/joselitofilho/hcl-parser-go/releases/)
[![Go Report Card](https://goreportcard.com/badge/github.com/joselitofilho/hcl-parser-go?style=for-the-badge)](https://goreportcard.com/report/github.com/joselitofilho/hcl-parser-go)
[![Code coverage](https://img.shields.io/badge/Coverage-90.1%25-2ea44f?style=for-the-badge)](#)

[![Made with Golang](https://img.shields.io/badge/Golang-1.21.6-blue?logo=go&logoColor=white&style=for-the-badge)](https://go.dev "Go to Golang homepage")

[![BuyMeACoffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-ffdd00?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://www.buymeacoffee.com/joselitofilho)

</div>

# Overview

This is a GoLang library designed to parse Terraform configuration files written in HashiCorp Configuration Language (HCL). 
It allows extracting resources, modules, variables, and locals defined in these configuration files.

## How to Use

```bash
$ go get github.com/joselitofilho/hcl-parser-go@latest
```

## Key Features

1. **Terraform File Parsing**: Implement robust parsing functionality to extract data from Terraform files efficiently.

## Example Usage

`lambda.tf`:

```hcl
resource "aws_lambda_function" "my_receiver_lambda" {
  filename      = "./artifacts/my_receiver.zip"
  function_name = "my_receiver"
  description   = "myReceiver lambda"
  role          = aws_iam_role.execute_lambda.arn
  handler       = "my_receiver"

  source_code_hash = filebase64sha256("./artifacts/my_receiver.zip")

  runtime = "go1.x"

  environment {
    variables = {
      TRACE                        = "1"
      TRACE_ENTITIES               = "Y"
      TIME_LOCATION                = "UTC"
      MY_STREAM_KINESIS_STREAM_URL = aws_kinesis_stream.my_stream_kinesis.name
    }
  }
}

// myReceiver SQS trigger rule for lambda
resource "aws_lambda_event_source_mapping" "my_receiver_lambda_sqs_trigger" {
  event_source_arn = aws_sqs_queue.source_sqs.arn
  function_name    = aws_lambda_function.my_receiver_lambda.arn
  batch_size       = 1
  enabled          = true
}

```

`main.go`:

```Go
package main

import (
	"fmt"

	hcl "github.com/joselitofilho/hcl-parser-go/pkg/parser/config"
)

func main() {
    directories := []string{}
	files := []string{"lambda.tf"}

	// Parse Terraform configurations
	config, err := hcl.Parse(directories, files)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print resources
	fmt.Println("Resources:")
	for _, resource := range config.Resources {
		fmt.Printf("Type: %s, Name: %s\n", resource.Type, resource.Name)
		for key, value := range resource.Attributes {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	// Print modules
	fmt.Println("\nModules:")
	for _, module := range config.Modules {
		fmt.Printf("Source: %s\n", module.Source)
		for key, value := range module.Attributes {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	// Print variables
	fmt.Println("\nVariables:")
	for _, variable := range config.Variables {
		for key, value := range variable.Attributes {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	// Print locals
	fmt.Println("\nLocals:")
	for _, local := range config.Locals {
		for key, value := range local.Attributes {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}
}

```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, feel free to create an 
[issue][issues] or submit a pull request. Your contribution is much appreciated. See [Contributing](CONTRIBUTING.md).

[![open - Contributing](https://img.shields.io/badge/open-contributing-blue?style=for-the-badge)](CONTRIBUTING.md "Go to contributing")

## License

This project is licensed under the [MIT License](LICENSE).

[diagrams]: https://app.diagrams.net/
[issues]: https://github.com/joselitofilho/hcl-parser-go/issues