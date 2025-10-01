package gitcleanup

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"distui/internal/models"
)

func MatchesPattern(path string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		matched, err := doublestar.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func MatchesExtension(path string, extensions []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false
	}

	for _, e := range extensions {
		if strings.EqualFold(ext, e) {
			return true
		}
	}
	return false
}

func CategorizeWithRules(path string, rules map[string]models.CategoryRules) string {
	for category, rule := range rules {
		if MatchesExtension(path, rule.Extensions) {
			return category
		}

		matched, err := MatchesPattern(path, rule.Patterns)
		if err == nil && matched {
			return category
		}
	}
	return "other"
}

func GetDefaultRules() map[string]models.CategoryRules {
	return map[string]models.CategoryRules{
		"config": {
			Extensions: []string{".yaml", ".yml", ".json", ".toml", ".ini", ".conf", ".env"},
			Patterns:   []string{"**/config/**", "**/configs/**", "**/.env*"},
		},
		"code": {
			Extensions: []string{".go", ".js", ".ts", ".py", ".rb", ".java", ".c", ".cpp", ".h", ".rs"},
			Patterns:   []string{"**/src/**", "**/lib/**", "**/pkg/**"},
		},
		"docs": {
			Extensions: []string{".md", ".txt", ".rst", ".adoc"},
			Patterns:   []string{"**/docs/**", "**/doc/**", "**/*.md"},
		},
		"build": {
			Extensions: []string{".mod", ".sum", ".lock"},
			Patterns:   []string{"**/.goreleaser*", "**/Makefile", "**/Dockerfile", "**/.github/**"},
		},
		"test": {
			Extensions: []string{"_test.go", ".test", ".spec.js", ".spec.ts"},
			Patterns:   []string{"**/test/**", "**/tests/**", "**/*_test.go"},
		},
		"assets": {
			Extensions: []string{".png", ".jpg", ".svg", ".ico", ".gif", ".woff", ".ttf", ".css"},
			Patterns:   []string{"**/assets/**", "**/static/**", "**/public/**"},
		},
		"data": {
			Extensions: []string{".sql", ".db", ".csv", ".xml"},
			Patterns:   []string{"**/data/**", "**/migrations/**"},
		},
	}
}
