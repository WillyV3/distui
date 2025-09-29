package gitcleanup

import (
	"fmt"
	"os/exec"
	"strings"
)

// HasGitHubRemote checks if a GitHub remote exists
func HasGitHubRemote() bool {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	remoteURL := strings.TrimSpace(string(output))
	return strings.Contains(remoteURL, "github.com")
}

// HasGitRepo checks if we're in a git repository
func HasGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetRepoInfo extracts owner and repo name from git remote
func GetRepoInfo() (owner, repo string, err error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return "", "", fmt.Errorf("no git remote found")
	}

	remoteURL := strings.TrimSpace(string(output))

	// Handle both SSH and HTTPS formats
	// SSH: git@github.com:owner/repo.git
	// HTTPS: https://github.com/owner/repo.git

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		// SSH format
		path := strings.TrimPrefix(remoteURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	} else if strings.Contains(remoteURL, "github.com/") {
		// HTTPS format
		parts := strings.Split(remoteURL, "github.com/")
		if len(parts) == 2 {
			path := strings.TrimSuffix(parts[1], ".git")
			pathParts := strings.Split(path, "/")
			if len(pathParts) == 2 {
				return pathParts[0], pathParts[1], nil
			}
		}
	}

	return "", "", fmt.Errorf("could not parse GitHub remote URL: %s", remoteURL)
}

// CheckGitHubRepoExists checks if the repo exists on GitHub
func CheckGitHubRepoExists() bool {
	owner, repo, err := GetRepoInfo()
	if err != nil {
		return false
	}

	// Use gh CLI to check if repo exists
	cmd := exec.Command("gh", "repo", "view", fmt.Sprintf("%s/%s", owner, repo))
	err = cmd.Run()
	return err == nil
}

// GetDefaultRepoName returns the current directory name as default repo name
func GetDefaultRepoName() string {
	output, err := exec.Command("sh", "-c", "basename $(pwd)").Output()
	if err != nil {
		return "my-project"
	}
	return strings.TrimSpace(string(output))
}

// CreateGitHubRepo creates a new GitHub repository
func CreateGitHubRepo(isPrivate bool, customName string, description string, owner string) error {
	repoName := customName
	if repoName == "" {
		repoName = GetDefaultRepoName()
	}

	// Check if we already have a remote - if so, we need to handle it differently
	hasRemote := HasGitHubRemote()

	if hasRemote {
		// Remove existing remote first
		exec.Command("git", "remote", "remove", "origin").Run()
	}

	// Create repo with gh CLI
	var repoFullName string
	if owner != "" {
		repoFullName = fmt.Sprintf("%s/%s", owner, repoName)
	} else {
		repoFullName = repoName
	}

	args := []string{"repo", "create", repoFullName, "--source", "."}
	if isPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	// Add description if provided
	if description != "" {
		args = append(args, "--description", description)
	}

	// Add --push to set up remote and push
	args = append(args, "--push")

	cmd := exec.Command("gh", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub repo: %w", err)
	}

	return nil
}

// IsGHCLIAvailable checks if gh CLI is installed
func IsGHCLIAvailable() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

// IsAuthenticated checks if gh CLI is authenticated
func IsAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}