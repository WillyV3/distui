package gitcleanup

import (
	"path/filepath"
	"strings"
)

// FileCategory represents how a file should be handled
type FileCategory string

const (
	CategoryAuto   FileCategory = "auto"   // Automatically commit (Go files)
	CategoryDocs   FileCategory = "docs"   // Ask user (documentation)
	CategoryIgnore FileCategory = "ignore" // Never commit (binaries)
	CategoryOther  FileCategory = "other"  // Unknown - ask user
)

// CategorizeFile determines the category of a file based on its path
func CategorizeFile(path string) FileCategory {
	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)
	dir := filepath.Dir(path)

	// Auto-commit patterns (Go code and project files)
	autoExtensions := []string{".go", ".mod", ".sum"}
	autoFiles := []string{"go.work", "go.work.sum"}

	for _, e := range autoExtensions {
		if ext == e {
			return CategoryAuto
		}
	}
	for _, f := range autoFiles {
		if base == f {
			return CategoryAuto
		}
	}

	// Documentation and config files (ask user)
	docsExtensions := []string{".md", ".txt", ".json", ".yaml", ".yml", ".toml"}
	for _, e := range docsExtensions {
		if ext == e {
			return CategoryDocs
		}
	}

	// Ignore patterns (binaries and temp files)
	ignoreExtensions := []string{".out", ".exe", ".dll", ".so", ".dylib", ".test"}
	ignoreDirs := []string{"bin", "dist", "node_modules", ".git", "vendor"}
	ignoreFiles := []string{".DS_Store", "thumbs.db"}

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

	// Everything else
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