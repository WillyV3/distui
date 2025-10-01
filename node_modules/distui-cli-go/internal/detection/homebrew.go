package detection

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TapInfo struct {
	Path     string
	RepoURL  string
	Formulas []string
	Exists   bool
}

func DetectHomebrewTap(username string) (*TapInfo, error) {
	if username == "" {
		return nil, fmt.Errorf("username required")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	locations := []string{
		filepath.Join(homeDir, "homebrew-tap"),
		filepath.Join(homeDir, "repos", "homebrew-tap"),
		filepath.Join(homeDir, ".homebrew-tap"),
	}

	for _, loc := range locations {
		if info, exists := checkTapLocation(loc); exists {
			return info, nil
		}
	}

	tapFromGH := findTapWithGH(username)
	if tapFromGH != nil {
		return tapFromGH, nil
	}

	return &TapInfo{
		Path:     filepath.Join(homeDir, "homebrew-tap"),
		RepoURL:  fmt.Sprintf("https://github.com/%s/homebrew-tap", username),
		Formulas: []string{},
		Exists:   false,
	}, nil
}

func checkTapLocation(path string) (*TapInfo, bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, false
	}

	formulaDir := filepath.Join(path, "Formula")
	formulas := []string{}

	if entries, err := os.ReadDir(formulaDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rb") {
				formulas = append(formulas, strings.TrimSuffix(entry.Name(), ".rb"))
			}
		}
	}

	repoURL := getRemoteURL(path)

	return &TapInfo{
		Path:     path,
		RepoURL:  repoURL,
		Formulas: formulas,
		Exists:   true,
	}, true
}

func getRemoteURL(path string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func findTapWithGH(username string) *TapInfo {
	cmd := exec.Command("gh", "repo", "list", username, "--json", "name,url", "--limit", "100")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "homebrew-") {
			return nil
		}
	}

	return nil
}