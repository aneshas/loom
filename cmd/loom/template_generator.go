package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

// TemplateData holds the variables used in template generation
type TemplateData struct {
	AppName      string // "myapp"
	ModuleName   string // "github.com/user/myapp" or "myapp"
	AppNameTitle string // "Myapp"
}

// GenerateProject creates a new project from embedded templates
func GenerateProject(appName, moduleName, targetDir string) error {
	// Validate app name
	if appName == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	// Validate module name
	if moduleName == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Create template data
	data := TemplateData{
		AppName:      appName,
		ModuleName:   moduleName,
		AppNameTitle: strings.Title(appName),
	}

	// Create target directory
	projectPath := filepath.Join(targetDir, appName)
	if err := os.MkdirAll(projectPath, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Check if directory already has content
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to read project directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("directory %s already exists and is not empty", projectPath)
	}

	// Walk through embedded templates and generate files
	return fs.WalkDir(templatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root templates directory
		if path == "templates" {
			return nil
		}

		// Calculate relative path from templates root
		relPath, err := filepath.Rel("templates", path)
		if err != nil {
			return err
		}

		// Transform path (replace {{.AppName}} placeholders in paths)
		targetPath := transformTemplatePath(relPath, data)
		fullTargetPath := filepath.Join(projectPath, targetPath)

		if d.IsDir() {
			// Skip creating empty cmd/app directory since we transform it to cmd/{appName}
			if targetPath == "cmd/app" {
				return nil
			}
			// Create directory
			return os.MkdirAll(fullTargetPath, 0o755)
		}

		// Process file
		return processTemplateFile(path, fullTargetPath, data)
	})
}

func transformTemplatePath(path string, data TemplateData) string {
	// Replace "app" with actual app name in cmd directory
	if strings.HasPrefix(path, "cmd/app/") {
		path = strings.Replace(path, "cmd/app/", "cmd/"+data.AppName+"/", 1)
	}

	// Replace placeholders in the path itself
	path = strings.ReplaceAll(path, "{{.AppName}}", data.AppName)

	// Remove .tmpl extension
	// if strings.HasSuffix(path, ".tmpl") {
	path = strings.TrimSuffix(path, ".tmpl")
	// }

	return path
}

func processTemplateFile(templatePath, targetPath string, data TemplateData) error {
	// Read template file
	content, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
	}

	// If this is a template file and should be processed as a template
	if strings.HasSuffix(templatePath, ".tmpl") && shouldProcessAsTemplate(templatePath) {
		tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
		}

		// Create target file
		file, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}
		defer file.Close()

		// Execute template
		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
		}
	} else {
		// Copy file directly (possibly after string replacements for non-Go-template files)
		contentStr := string(content)
		if strings.HasSuffix(templatePath, ".tmpl") {
			// Apply simple string replacements instead of template processing
			contentStr = applyStringReplacements(contentStr, data)
		}

		if err := os.WriteFile(targetPath, []byte(contentStr), 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetPath, err)
		}
	}

	return nil
}

func shouldProcessAsTemplate(templatePath string) bool {
	// Only process certain file types as Go templates
	processAsTemplate := []string{
		".go.tmpl",
		".mod.tmpl",
		".yaml.tmpl",
		".yml.tmpl",
	}

	for _, ext := range processAsTemplate {
		if strings.HasSuffix(templatePath, ext) {
			return true
		}
	}

	return false
}

func applyStringReplacements(content string, data TemplateData) string {
	// Simple string replacements for files that contain template syntax
	content = strings.ReplaceAll(content, "{{.ModuleName}}", data.ModuleName)
	content = strings.ReplaceAll(content, "{{.AppName}}", data.AppName)
	content = strings.ReplaceAll(content, "{{.AppNameTitle}}", data.AppNameTitle)

	return content
}
