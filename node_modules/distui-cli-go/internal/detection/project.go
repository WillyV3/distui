package detection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"distui/internal/models"
	"golang.org/x/mod/modfile"
)

func DetectProject(path string) (*models.ProjectInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path provided")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	goModPath := filepath.Join(absPath, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return nil, fmt.Errorf("go.mod not found at %s", absPath)
	}

	moduleInfo, err := parseGoModule(goModPath)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	repoInfo := detectGitRepository(absPath)

	binaryName := extractBinaryName(moduleInfo.Name)
	identifier := sanitizeIdentifier(moduleInfo.Name)
	now := time.Now()

	return &models.ProjectInfo{
		Identifier:   identifier,
		Path:         absPath,
		LastAccessed: &now,
		DetectedAt:   &now,
		Repository:   repoInfo,
		Module:       moduleInfo,
		Binary: &models.BinaryInfo{
			Name:       binaryName,
			BuildFlags: []string{},
		},
	}, nil
}

func parseGoModule(goModPath string) (*models.ModuleInfo, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}

	// Try the official parser first
	modFile, err := modfile.Parse(goModPath, data, nil)
	if err == nil {
		version := getCurrentVersion(filepath.Dir(goModPath))
		if version == "" {
			version = "v0.0.1"
		}
		return &models.ModuleInfo{
			Name:    modFile.Module.Mod.Path,
			Version: version,
		}, nil
	}

	// Fallback to simple parsing if official parser fails
	lines := strings.Split(string(data), "\n")
	moduleName := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			break
		}
	}

	if moduleName == "" {
		return nil, fmt.Errorf("no module declaration found in go.mod")
	}

	version := getCurrentVersion(filepath.Dir(goModPath))
	if version == "" {
		version = "v0.0.1"
	}

	return &models.ModuleInfo{
		Name:    moduleName,
		Version: version,
	}, nil
}

func detectGitRepository(path string) *models.RepositoryInfo {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return &models.RepositoryInfo{
			DefaultBranch: "main",
		}
	}

	remoteURL := strings.TrimSpace(string(output))
	owner, name := parseGitRemoteURL(remoteURL)

	defaultBranch := getDefaultBranch(path)

	return &models.RepositoryInfo{
		Owner:         owner,
		Name:          name,
		DefaultBranch: defaultBranch,
	}
}

func parseGitRemoteURL(remoteURL string) (owner, name string) {
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "git@github.com:"), "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	if strings.Contains(remoteURL, "github.com/") {
		idx := strings.Index(remoteURL, "github.com/")
		if idx >= 0 {
			parts := strings.Split(remoteURL[idx+11:], "/")
			if len(parts) >= 2 {
				return parts[0], parts[1]
			}
		}
	}

	return "", filepath.Base(remoteURL)
}

func getDefaultBranch(path string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = path
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		branch = strings.TrimPrefix(branch, "refs/remotes/origin/")
		if branch != "" {
			return branch
		}
	}

	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	output, err = cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "" && branch != "HEAD" {
			return branch
		}
	}

	return "main"
}

func getCurrentVersion(path string) string {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = path
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
		if version != "" {
			return version
		}
	}
	return ""
}

func extractBinaryName(modulePath string) string {
	parts := strings.Split(modulePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "app"
}

func sanitizeIdentifier(modulePath string) string {
	identifier := strings.ReplaceAll(modulePath, "/", "-")
	identifier = strings.ReplaceAll(identifier, ".", "-")
	identifier = strings.ReplaceAll(identifier, "_", "-")
	return identifier
}

func DetectGitHubUsingGH(path string) (*models.RepositoryInfo, error) {
	cmd := exec.Command("gh", "repo", "view", "--json", "owner,name")
	cmd.Dir = path

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh CLI command failed: %w", err)
	}

	var result struct {
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("parsing gh output: %w", err)
	}

	return &models.RepositoryInfo{
		Owner:         result.Owner.Login,
		Name:          result.Name,
		DefaultBranch: getDefaultBranch(path),
	}, nil
}

type UserEnvironment struct {
	GitName     string
	GitEmail    string
	GitHubUser  string
	HasGitConfig bool
	HasGHCLI     bool
}

func (u UserEnvironment) HasMinimalRequirements() bool {
	return u.HasGitConfig && u.GitName != "" && u.GitEmail != ""
}

func DetectUserEnvironment() (*UserEnvironment, error) {
	env := &UserEnvironment{}

	gitName, nameErr := exec.Command("git", "config", "--global", "user.name").Output()
	if nameErr == nil {
		env.GitName = strings.TrimSpace(string(gitName))
		env.HasGitConfig = true
	}

	gitEmail, emailErr := exec.Command("git", "config", "--global", "user.email").Output()
	if emailErr == nil {
		env.GitEmail = strings.TrimSpace(string(gitEmail))
	}

	ghUser, ghErr := exec.Command("gh", "auth", "status", "--hostname", "github.com").CombinedOutput()
	if ghErr == nil {
		output := string(ghUser)
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Logged in to github.com account") {
				// Line format: "✓ Logged in to github.com account williavs (keyring)"
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "account" && i+1 < len(parts) {
						username := parts[i+1]
						// Remove parentheses if present
						if idx := strings.Index(username, "("); idx > 0 {
							username = username[:idx]
						}
						env.GitHubUser = strings.TrimSpace(username)
						env.HasGHCLI = true
						return env, nil
					}
				}
			}
		}
	}

	return env, nil
}