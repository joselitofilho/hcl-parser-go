package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/joselitofilho/hcl-parser-go/internal/fmtcolor"
)

// Resource represents a Terraform resource.
type Resource struct {
	Type       string
	Name       string
	Labels     []string
	Attributes map[string]any
}

// Module represents a Terraform module.
type Module struct {
	Source     string
	Labels     []string
	Attributes map[string]any
}

// Local represents a Terraform local value.
type Local struct {
	Attributes map[string]any
}

// Variable represents a Terraform variable.
type Variable struct {
	Attributes map[string]any
}

// Config represents the Terraform configuration.
type Config struct {
	Resources []*Resource
	Modules   []*Module
	Variables []*Variable
	Locals    []*Local
}

// Parse parses Terraform configuration files in specified directories and files.
func Parse(directories, files []string) (*Config, error) {
	config := &Config{}
	hclParser := hclparse.NewParser()

	for _, directory := range directories {
		err := parseDirectory(directory, hclParser, config)
		if err != nil {
			return nil, err
		}
	}

	for _, file := range files {
		err := parseSingleFile(file, hclParser, config)
		if err != nil {
			return config, err
		}
	}

	return config, nil
}

// parseDirectory parses all Terraform configuration files in a directory.
func parseDirectory(directory string, hclParser *hclparse.Parser, config *Config) error {
	return filepath.Walk(directory, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory: %w", err)
		}

		if info.IsDir() || strings.Contains(file, ".terraform/") || filepath.Ext(file) != ".tf" {
			return nil
		}

		return parseSingleFile(file, hclParser, config)
	})
}

// parseSingleFile parses a single Terraform configuration file.
func parseSingleFile(file string, hclParser *hclparse.Parser, config *Config) error {
	parsedConfig, err := parseHCLFile(file, hclParser)
	if err != nil {
		return fmt.Errorf("error parsing HCL file: %w", err)
	}

	config.Modules = append(config.Modules, parsedConfig.Modules...)
	config.Resources = append(config.Resources, parsedConfig.Resources...)
	config.Locals = append(config.Locals, parsedConfig.Locals...)

	return nil
}

// parseHCLFile parses a single HCL file and extracts its configurations.
func parseHCLFile(file string, parser *hclparse.Parser) (*Config, error) {
	if filepath.Ext(file) == ".tf" {
		_, err := os.Stat(file)
		if !os.IsNotExist(err) {
			file, diags := parser.ParseHCLFile(file)
			if diags.HasErrors() {
				return nil, fmt.Errorf("failed to load config file %s: %s", file, diags.Errs())
			}

			return parseConfig(file), nil
		}
	}

	return &Config{}, nil
}

// parseConfig extracts configurations from a parsed HCL file.
func parseConfig(file *hcl.File) *Config {
	resources := make([]*Resource, 0)
	modules := make([]*Module, 0)
	locals := make([]*Local, 0)

	for _, block := range file.Body.(*hclsyntax.Body).Blocks {
		switch block.Type {
		case "module":
			modules = append(modules, parseModule(block))
		case "resource":
			resources = append(resources, parseResource(block))
		case "locals":
			locals = append(locals, parseLocals(block))
		}
	}

	return &Config{Resources: resources, Modules: modules, Locals: locals}
}

// parseModule extracts module configuration from a block.
func parseModule(block *hclsyntax.Block) *Module {
	module := &Module{
		Labels:     block.Labels,
		Attributes: map[string]any{},
	}

	for _, attribute := range block.Body.Attributes {
		value := evaluateExpression(attribute.Expr)
		module.Attributes[attribute.Name] = value

		if attribute.Name == "source" {
			module.Source = value.(string)
		}
	}

	return module
}

// parseResource extracts resource configuration from a block.
func parseResource(block *hclsyntax.Block) *Resource {
	resource := &Resource{
		Type:       block.Labels[0],
		Name:       block.Labels[1],
		Labels:     block.Labels,
		Attributes: map[string]any{},
	}

	for _, attribute := range block.Body.Attributes {
		value := evaluateExpression(attribute.Expr)
		resource.Attributes[attribute.Name] = value
	}

	for _, bodyBlock := range block.Body.Blocks {
		parseResourcesFromBlock(bodyBlock, resource)
	}

	return resource
}

// parseResourcesFromBlock extracts environment configuration from a block.
func parseResourcesFromBlock(bodyBlock *hclsyntax.Block, resource *Resource) {
	blType := bodyBlock.Type
	if _, ok := resource.Attributes[blType]; !ok {
		resource.Attributes[blType] = map[string]any{}
	}

	data := resource.Attributes[blType].(map[string]any)

	for _, attribute := range bodyBlock.Body.Attributes {
		data[attribute.Name] = evaluateExpression(attribute.Expr)
	}
}

// parseLocals extracts local configuration from a block.
func parseLocals(block *hclsyntax.Block) *Local {
	local := &Local{
		Attributes: map[string]any{},
	}

	for _, attribute := range block.Body.Attributes {
		value := evaluateExpression(attribute.Expr)
		local.Attributes[attribute.Name] = value
	}

	return local
}

func buildVarExpressions(traversal hcl.Traversal) string {
	varExp := make([]string, 0, len(traversal))

	for _, v := range traversal {
		switch v := v.(type) {
		case hcl.TraverseRoot:
			if v.Name != "" {
				varExp = append(varExp, v.Name)
			}
		case hcl.TraverseAttr:
			if v.Name != "" {
				varExp = append(varExp, v.Name)
			}
		}
	}

	return strings.Join(varExp, ".")
}

func convertValueToString(val cty.Value) string {
	switch val.Type() {
	case cty.Number:
		return val.AsBigFloat().String()
	case cty.String:
		return val.AsString()
	case cty.Bool:
		var v bool
		_ = gocty.FromCtyValue(val, &v)

		return fmt.Sprintf("%v", v)
	default:
		fmtcolor.Yellow.Println("unsupported type:", val.Type().GoString())
		return ""
	}
}

// evaluateExpression evaluates the HCL expression and returns its value as a string or map[string]string.
func evaluateExpression(expr hcl.Expression) any {
	resultString := ""
	resultMap := map[string]any{}

	switch expr := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		resultString += buildVarExpressions(expr.Traversal)
	case *hclsyntax.LiteralValueExpr:
		resultString += convertValueToString(expr.Val)
	case *hclsyntax.TemplateExpr:
		parts := expr.Parts
		for _, part := range parts {
			resultString += evaluateExpression(part).(string)
		}
	case *hclsyntax.TupleConsExpr:
		for _, elem := range expr.Exprs {
			resultString += evaluateExpression(elem).(string) + ","
		}
	case *hclsyntax.ObjectConsKeyExpr:
		resultString += evaluateExpression(expr.Wrapped).(string)
	case *hclsyntax.ObjectConsExpr:
		for i := range expr.Items {
			item := expr.Items[i]

			resultMap[evaluateExpression(item.KeyExpr).(string)] = evaluateExpression(item.ValueExpr)
		}

		return resultMap
	case *hclsyntax.IndexExpr:
		resultString += evaluateExpression(expr.Collection).(string)
	case *hclsyntax.FunctionCallExpr:
		resultString += evaluateFunctionExpression(expr)
	default:
		fmtcolor.Yellow.Println("unsupported expr:", expr)
	}

	return resultString
}

func evaluateFunctionExpression(expr *hclsyntax.FunctionCallExpr) string {
	var args string

	for i := range expr.Args {
		exp := evaluateExpression(expr.Args[i])

		// TODO: Implement other cases

		switch exp := exp.(type) {
		case string:
			args += exp
		case map[string]any:
			var values string

			for k, v := range exp {
				values += k

				switch v := v.(type) {
				case string:
					values += ":" + v
				default:
					fmtcolor.Yellow.Println("unsupported function arg value:", expr)
				}
			}

			args = fmt.Sprintf("%s{%s}", args, values)
		default:
			fmtcolor.Yellow.Println("unsupported function arg:", expr)
		}
	}

	return fmt.Sprintf("%s(%s)", expr.Name, args)
}
