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

	// Ensure git is initialized
	if !HasGitRepo() {
		cmd := exec.Command("git", "init")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	// Ensure we're on main branch (not master)
	exec.Command("git", "branch", "-M", "main").Run()

	// Check if we already have a remote - if so, remove it
	hasRemote := HasGitHubRemote()
	if hasRemote {
		exec.Command("git", "remote", "remove", "origin").Run()
	}

	// Create repo with gh CLI
	var repoFullName string
	if owner != "" {
		repoFullName = fmt.Sprintf("%s/%s", owner, repoName)
	} else {
		repoFullName = repoName
	}

	args := []string{"repo", "create", repoFullName}
	if isPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	// Add description if provided
	if description != "" {
		args = append(args, "--description", description)
	}

	// Create the repo
	cmd := exec.Command("gh", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub repo: %w", err)
	}

	// Add the remote manually (don't push yet - that's for release workflow)
	remoteURL := fmt.Sprintf("git@github.com:%s.git", repoFullName)
	cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
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