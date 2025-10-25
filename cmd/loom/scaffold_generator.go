package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

// runScaffoldCommand generates a model file with standard Go types from a SQLBoiler model
func runScaffoldCommand(modelName string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Path to the SQLBoiler model file
	modelPath := filepath.Join(cwd, "internal", "db", "model", strings.ToLower(modelName)+"s.go")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try singular form
		modelPath = filepath.Join(cwd, "internal", "db", "model", strings.ToLower(modelName)+".go")
		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			return fmt.Errorf("model file not found for %s in internal/db/model", modelName)
		}
	}

	// Parse the SQLBoiler model file
	fields, err := parseModelFile(modelPath, modelName)
	if err != nil {
		return fmt.Errorf("failed to parse model file: %w", err)
	}

	// Generate the output file
	outputDir := filepath.Join(cwd, "web", "controller")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, toSnakeCase(modelName)+".go")
	if err := generateModelFile(outputPath, modelName, fields); err != nil {
		return fmt.Errorf("failed to generate model file: %w", err)
	}

	ctx := context.Background()

	cmd := exec.CommandContext(ctx, "go", "fmt", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go fmt: %w", err)
	}

	return nil
}

// ModelField represents a field in the model
type ModelField struct {
	Name       string
	Type       string
	IsNullable bool
	JSONTag    string
}

func (f ModelField) IsTimestamp() bool {
	return f.Name == "CreatedAt" || f.Name == "UpdatedAt"
}

// parseModelFile parses a SQLBoiler model file and extracts field information
func parseModelFile(filePath, modelName string) ([]ModelField, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var fields []ModelField

	// Find the model struct
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != modelName {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Parse each field
		for _, field := range structType.Fields.List {
			if len(field.Names) == 0 {
				continue
			}

			fieldName := field.Names[0].Name

			// Skip internal fields (R, L)
			if fieldName == "R" || fieldName == "L" {
				continue
			}

			// Extract type information
			fieldType, isNullable := extractFieldType(field.Type)

			// Extract JSON tag
			jsonTag := ""
			if field.Tag != nil {
				tag := field.Tag.Value
				jsonTag = extractJSONTag(tag)
			}

			fields = append(fields, ModelField{
				Name:       fieldName,
				Type:       fieldType,
				IsNullable: isNullable,
				JSONTag:    jsonTag,
			})
		}

		return false
	})

	if len(fields) == 0 {
		return nil, fmt.Errorf("model %s not found in file", modelName)
	}

	return fields, nil
}

// extractFieldType extracts the Go type from an AST expression
func extractFieldType(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple types like string, int, bool, etc.
		return t.Name, false
	case *ast.SelectorExpr:
		// Types like null.String, null.Int64, time.Time
		pkg := t.X.(*ast.Ident).Name
		typeName := t.Sel.Name

		if pkg == "null" {
			// Convert null types to standard Go types
			return convertNullType(typeName), true
		} else if pkg == "time" && typeName == "Time" {
			return "time.Time", false
		}
		return pkg + "." + typeName, false
	case *ast.StarExpr:
		// Pointer types
		innerType, _ := extractFieldType(t.X)
		return innerType, true
	default:
		return "interface{}", false
	}
}

// convertNullType converts SQLBoiler null types to standard Go types
func convertNullType(nullType string) string {
	switch nullType {
	case "String":
		return "string"
	case "Int":
		return "int"
	case "Int8":
		return "int8"
	case "Int16":
		return "int16"
	case "Int32":
		return "int32"
	case "Int64":
		return "int64"
	case "Uint":
		return "uint"
	case "Uint8":
		return "uint8"
	case "Uint16":
		return "uint16"
	case "Uint32":
		return "uint32"
	case "Uint64":
		return "uint64"
	case "Float32":
		return "float32"
	case "Float64":
		return "float64"
	case "Bool":
		return "bool"
	case "Time":
		return "time.Time"
	case "Bytes":
		return "[]byte"
	default:
		return "any"
	}
}

// extractJSONTag extracts the JSON tag value from a struct tag
func extractJSONTag(tag string) string {
	// Remove backticks
	tag = strings.Trim(tag, "`")

	// Find json tag
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "json:") {
			jsonValue := strings.TrimPrefix(part, "json:")
			jsonValue = strings.Trim(jsonValue, `"`)
			// Remove omitempty and other options
			jsonValue = strings.Split(jsonValue, ",")[0]
			return jsonValue
		}
	}

	return ""
}

// generateModelFile generates the model file with standard Go types
func generateModelFile(outputPath, modelName string, fields []ModelField) error {
	var sb strings.Builder

	// Package declaration
	sb.WriteString("package controller\n\n")

	// Imports
	needsTime := false
	for _, field := range fields {
		if field.Type == "time.Time" {
			needsTime = true
			break
		}
	}

	if needsTime {
		sb.WriteString("import (\n")
		sb.WriteString("\t\"time\"\n\n")
		sb.WriteString("\t\"github.com/aarondl/null/v8\"\n")
		sb.WriteString("\t\"<module_path>/internal/db/model\"\n")
		sb.WriteString(")\n\n")
	} else {
		sb.WriteString("import (\n")
		sb.WriteString("\t\"github.com/aarondl/null/v8\"\n")
		sb.WriteString("\t\"<module_path>/internal/db/model\"\n")
		sb.WriteString(")\n\n")
	}

	// Model struct
	sb.WriteString(fmt.Sprintf("type %s struct {\n", modelName))
	for _, field := range fields {
		goType := field.Type
		if field.IsNullable {
			goType = "*" + goType
		}

		// Build struct tags
		tags := []string{}
		if field.JSONTag != "" {
			tags = append(tags, fmt.Sprintf(`json:"%s"`, field.JSONTag))
		}
		tags = append(tags, fmt.Sprintf(`form:"%s"`, field.JSONTag))

		// Add validate tag for non-nullable fields
		if !field.IsNullable && !field.IsTimestamp() {
			tags = append(tags, `validate:"required"`)
		}

		tagString := strings.Join(tags, " ")
		sb.WriteString(fmt.Sprintf("\t%s %s `%s`\n", field.Name, goType, tagString))
	}
	sb.WriteString("}\n\n")

	// Generate ToDB mapper
	sb.WriteString(fmt.Sprintf("// %sToDB converts %s to model.%s\n", modelName, modelName, modelName))
	sb.WriteString(fmt.Sprintf("func (m *%s) ToDB() *model.%s {\n", modelName, modelName))
	sb.WriteString(fmt.Sprintf("\treturn &model.%s{\n", modelName))
	for _, field := range fields {
		if field.IsNullable {
			// Convert pointer to null type
			sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", field.Name, convertToNullType(field.Type, "m."+field.Name)))
		} else {
			// Direct assignment
			sb.WriteString(fmt.Sprintf("\t\t%s: m.%s,\n", field.Name, field.Name))
		}
	}
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Generate FromDB mapper
	sb.WriteString(fmt.Sprintf("// %sFromDB converts model.%s to %s\n", modelName, modelName, modelName))
	sb.WriteString(fmt.Sprintf("func %sFromDB(m *model.%s) *%s {\n", modelName, modelName, modelName))
	sb.WriteString(fmt.Sprintf("\treturn &%s{\n", modelName))
	for _, field := range fields {
		if field.IsNullable {
			// Convert null type to pointer
			sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", field.Name, convertFromNullType(field.Type, "m."+field.Name)))
		} else {
			// Direct assignment
			sb.WriteString(fmt.Sprintf("\t\t%s: m.%s,\n", field.Name, field.Name))
		}
	}
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	// Write to file
	content := sb.String()

	// Try to get module path from go.mod
	modulePath, err := getModulePath()
	if err == nil {
		content = strings.ReplaceAll(content, "<module_path>", modulePath)
	}

	return os.WriteFile(outputPath, []byte(content), 0o644)
}

// convertToNullType generates code to convert a pointer to a null type
func convertToNullType(goType, fieldAccess string) string {
	switch goType {
	case "string":
		return fmt.Sprintf("null.StringFromPtr(%s)", fieldAccess)
	case "int":
		return fmt.Sprintf("null.IntFromPtr(%s)", fieldAccess)
	case "int8":
		return fmt.Sprintf("null.Int8FromPtr(%s)", fieldAccess)
	case "int16":
		return fmt.Sprintf("null.Int16FromPtr(%s)", fieldAccess)
	case "int32":
		return fmt.Sprintf("null.Int32FromPtr(%s)", fieldAccess)
	case "int64":
		return fmt.Sprintf("null.Int64FromPtr(%s)", fieldAccess)
	case "uint":
		return fmt.Sprintf("null.UintFromPtr(%s)", fieldAccess)
	case "uint8":
		return fmt.Sprintf("null.Uint8FromPtr(%s)", fieldAccess)
	case "uint16":
		return fmt.Sprintf("null.Uint16FromPtr(%s)", fieldAccess)
	case "uint32":
		return fmt.Sprintf("null.Uint32FromPtr(%s)", fieldAccess)
	case "uint64":
		return fmt.Sprintf("null.Uint64FromPtr(%s)", fieldAccess)
	case "float32":
		return fmt.Sprintf("null.Float32FromPtr(%s)", fieldAccess)
	case "float64":
		return fmt.Sprintf("null.Float64FromPtr(%s)", fieldAccess)
	case "bool":
		return fmt.Sprintf("null.BoolFromPtr(%s)", fieldAccess)
	case "time.Time":
		return fmt.Sprintf("null.TimeFromPtr(%s)", fieldAccess)
	case "[]byte":
		return fmt.Sprintf("null.BytesFromPtr(%s)", fieldAccess)
	default:
		return fieldAccess
	}
}

// convertFromNullType generates code to convert a null type to a pointer
func convertFromNullType(goType, fieldAccess string) string {
	switch goType {
	case "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "bool", "time.Time", "[]byte":
		return fmt.Sprintf("%s.Ptr()", fieldAccess)
	default:
		return fieldAccess
	}
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// getModulePath reads the module path from go.mod
func getModulePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	goModPath := filepath.Join(cwd, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	return "", fmt.Errorf("module path not found in go.mod")
}
