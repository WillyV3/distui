package gitcleanup

import (
	"os/exec"
	"strings"
)

// GitFile represents a file with git status
type GitFile struct {
	Path     string
	Status   string       // M, A, D, ??
	Category FileCategory
}

// GetGitStatus returns all modified and untracked files
func GetGitStatus() ([]GitFile, error) {
	var files []GitFile

	// Get git status in porcelain format
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse status and path
		status := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[2:])

		// Skip if no path
		if path == "" {
			continue
		}

		// Categorize the file
		category := CategorizeFile(path)

		files = append(files, GitFile{
			Path:     path,
			Status:   status,
			Category: category,
		})
	}

	return files, nil
}

// HasUncommittedChanges checks if there are any uncommitted changes
func HasUncommittedChanges() bool {
	// Check for modified files
	cmd := exec.Command("git", "diff", "--quiet")
	if err := cmd.Run(); err != nil {
		return true // Changes exist
	}

	// Check for staged files
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	if err := cmd.Run(); err != nil {
		return true // Staged changes exist
	}

	// Check for untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
}