package gitcleanup

import (
	"os"
	"path/filepath"
	"strings"

	"distui/internal/models"
)

// FileCategory represents how a file should be handled
type FileCategory string

const (
	CategoryAuto   FileCategory = "auto"   // Automatically commit (Go files)
	CategoryDocs   FileCategory = "docs"   // Ask user (documentation)
	CategoryIgnore FileCategory = "ignore" // Never commit (binaries)
	CategoryOther  FileCategory = "other"  // Unknown - ask user
)

// directoryContainsGoFiles checks if a directory contains any .go files
func directoryContainsGoFiles(dir string) bool {
	if dir == "." || dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}
	return false
}

// CategorizeFile determines the category of a file based on its path
// If projectConfig is provided and has custom rules enabled, uses those rules
// Otherwise falls back to default categorization
func CategorizeFile(path string) FileCategory {
	return CategorizeFileWithConfig(path, nil)
}

// CategorizeFileWithConfig determines the category using optional custom rules
func CategorizeFileWithConfig(path string, projectConfig *models.ProjectConfig) FileCategory {
	// Check if custom smart commit rules are enabled
	if projectConfig != nil && projectConfig.Config != nil &&
		projectConfig.Config.SmartCommit != nil &&
		projectConfig.Config.SmartCommit.UseCustomRules {
		category := CategorizeWithRules(path, projectConfig.Config.SmartCommit.Categories)
		if category != "other" {
			return mapCategoryToFileCategory(category)
		}
	}

	// Fall back to default categorization
	return categorizeWithDefaults(path)
}

func mapCategoryToFileCategory(category string) FileCategory {
	switch category {
	case "code", "config", "build":
		return CategoryAuto
	case "docs":
		return CategoryDocs
	default:
		return CategoryOther
	}
}

func categorizeWithDefaults(path string) FileCategory {
	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)
	dir := filepath.Dir(path)

	// Ignore patterns first (binaries and temp files)
	ignoreExtensions := []string{".out", ".exe", ".dll", ".so", ".dylib", ".test", ".a"}
	ignoreDirs := []string{"bin", "dist", "node_modules", "vendor"}
	ignoreFiles := []string{".DS_Store", "thumbs.db", "distui", "tuitemplate"}

	// Special check: exclude .git/ directory itself but allow .github/, .goreleaser.yaml, etc.
	if strings.HasPrefix(path, ".git/") || path == ".git" {
		return CategoryIgnore
	}

	// Check if path starts with dist/ or is dist directory
	if strings.HasPrefix(path, "dist/") || path == "dist" {
		return CategoryIgnore
	}

	for _, e := range ignoreExtensions {
		if ext == e {
			return CategoryIgnore
		}
	}
	for _, f := range ignoreFiles {
		if base == f {
			return CategoryIgnore
		}
	}
	for _, d := range ignoreDirs {
		if strings.Contains(dir, d) {
			return CategoryIgnore
		}
	}

	// Check for likely binaries without extension
	// If no extension and not a known text file, assume it's a binary
	if ext == "" && !strings.Contains(base, ".") {
		// Check if it looks like a Go binary name
		if base == "distui" || base == "tuitemplate" || strings.HasPrefix(base, "main") {
			return CategoryIgnore
		}
		// Unknown files without extension - probably binaries
		return CategoryIgnore
	}

	// Check if file's directory contains Go files
	hasGoFiles := directoryContainsGoFiles(dir)

	// Root-level project files (always auto-commit regardless of directory)
	rootProjectFiles := []string{"go.mod", "go.sum", "go.work", "go.work.sum", ".goreleaser.yaml", ".goreleaser.yml"}
	for _, f := range rootProjectFiles {
		if base == f {
			return CategoryAuto
		}
	}

	// Go files in Go directories (auto-commit)
	if ext == ".go" && hasGoFiles {
		return CategoryAuto
	}

	// Non-Go files in Go directories (ask user)
	if hasGoFiles {
		return CategoryOther
	}

	// Files in non-Go directories (ask user)
	// Documentation and config files
	docsExtensions := []string{".md", ".txt", ".json", ".yaml", ".yml", ".toml"}
	for _, e := range docsExtensions {
		if ext == e {
			return CategoryDocs
		}
	}

	// Everything else (ask user)
	return CategoryOther
}

// SuggestCommitMessage generates a smart commit message based on files
func SuggestCommitMessage(files []string) string {
	hasGoFiles := false
	hasDocFiles := false
	hasConfigFiles := false

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".go", ".mod", ".sum":
			hasGoFiles = true
		case ".md", ".txt":
			hasDocFiles = true
		case ".json", ".yaml", ".yml", ".toml":
			hasConfigFiles = true
		}
	}

	if hasGoFiles && hasDocFiles {
		return "Update code and documentation"
	} else if hasGoFiles {
		return "Update Go code and project files"
	} else if hasDocFiles {
		return "Update documentation"
	} else if hasConfigFiles {
		return "Update configuration files"
	}
	return "Update project files"
}