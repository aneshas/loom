package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func main() {
	if err := generateTemplates(); err != nil {
		fmt.Printf("Error generating templates: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Templates generated successfully!")
}

func generateTemplates() error {
	sourceDir := "example"
	templateDir := "cmd/loom/templates"

	// Remove existing template directory
	if err := os.RemoveAll(templateDir); err != nil {
		return fmt.Errorf("failed to remove existing templates: %w", err)
	}

	// Create template directory
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Walk through source directory and generate templates
	return filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip database-related files and directories
		if shouldSkipDB(path) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Calculate relative path from source
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Calculate target path in templates
		targetPath := filepath.Join(templateDir, transformPath(relPath))

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0o755)
		}

		// Process file
		return processFile(path, targetPath)
	})
}

func shouldSkipDB(path string) bool {
	dbPaths := []string{
		"internal/db",
		"cmd/seed",
	}

	for _, dbPath := range dbPaths {
		if strings.Contains(path, dbPath) {
			return true
		}
	}

	// Skip specific DB-related files
	if strings.HasSuffix(path, ".sql") {
		return true
	}

	return false
}

func transformPath(path string) string {
	// Transform specific paths

	// this too
	path = strings.ReplaceAll(path, "cmd/helloapp", "cmd/app")
	path = strings.ReplaceAll(path, "helloapp", "{{.AppName}}")

	// Add .tmpl extension to template files
	if shouldTemplate(path) {
		path = path + ".tmpl"
	}

	return path
}

func shouldTemplate(path string) bool {
	// Files that should be templated
	templateExts := []string{
		".go", ".mod", ".yaml", ".yml", ".templ", ".html", ".js", ".css", ".toml",
	}

	ext := filepath.Ext(path)

	if slices.Contains(templateExts, ext) {
		return true
	}

	return false
}

func processFile(sourcePath, targetPath string) error {
	// Read source file
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sourcePath, err)
	}

	// Apply replacements if this is a template file
	if strings.HasSuffix(targetPath, ".tmpl") {
		content = []byte(applyReplacements(string(content)))
	}

	// Write target file
	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", targetPath, err)
	}

	return nil
}

func applyReplacements(content string) string {
	replacements := map[string]string{
		// Module name
		"github.com/aneshas/helloapp": "{{.ModuleName}}",

		// Package and app names
		"helloapp": "{{.AppName}}",
		"Helloapp": "{{.AppNameTitle}}",

		"./cmd/helloapp": "./cmd/{{.AppName}}",

		// Config struct and package
		"github.com/aneshas/helloapp/config":   "{{.ModuleName}}/config",
		"github.com/aneshas/helloapp/internal": "{{.ModuleName}}/internal",
		"github.com/aneshas/helloapp/web":      "{{.ModuleName}}/web",
	}

	// Apply direct replacements
	for old, new := range replacements {
		content = strings.ReplaceAll(content, old, new)
	}

	// Remove DB-related content
	content = removeDBContent(content)

	// Apply regex replacements for more complex patterns
	regexReplacements := map[string]string{
		// Import paths
		`"([^"]*/)helloapp(/[^"]*)"`: `"${1}{{.AppName}}${2}"`,
		// Package declarations
		`package helloapp\b`: `package {{.AppName}}`,
	}

	for pattern, replacement := range regexReplacements {
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, replacement)
	}

	return content
}

// TODO - this should be in template transformations

func removeDBContent(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skipNext := false

	for _, line := range lines {
		// Skip DB-related imports
		if strings.Contains(line, "internal/db") ||
			strings.Contains(line, "github.com/mattn/go-sqlite3") ||
			strings.Contains(line, "github.com/aarondl/sqlboiler") ||
			strings.Contains(line, "github.com/friendsofgo/errors") ||
			strings.Contains(line, "github.com/aarondl/null") ||
			strings.Contains(line, "github.com/aarondl/strmangle") {
			continue
		}

		// Skip replace directive for development /
		if strings.Contains(line, "replace github.com/aneshas/loom") && os.Getenv("LOOM_DEV") == "" {
			continue
		}

		// Skip DB-related code blocks
		if strings.Contains(line, "cfg.DBConn()") ||
			strings.Contains(line, "conn, err :=") ||
			strings.Contains(line, "user.NewRegister(conn)") ||
			strings.Contains(line, "db.NewUserStore(conn)") ||
			strings.Contains(line, "var userStore user.Store") {
			skipNext = true
			continue
		}

		if skipNext && strings.TrimSpace(line) == "" {
			skipNext = false
			continue
		}

		if strings.Contains(line, "check(err)") && skipNext {
			skipNext = false
			continue
		}

		if strings.Contains(line, "loom.Add(deps, userStore)") ||
			strings.Contains(line, "loom.Add(deps, user.NewRegister") {
			skipNext = false
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
