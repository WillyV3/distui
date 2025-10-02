package gitcleanup

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitFile represents a file with git status
type GitFile struct {
	Path     string
	Status   string       // M, A, D, ??
	Category FileCategory
}

// FileChange represents a simplified file change
type FileChange struct {
	Path        string
	Status      string
	StatusText  string // "modified", "new file", etc
	Icon        string // Visual indicator
}

// IsWorkingTreeClean returns true if there are no uncommitted changes
// Ignores .gitignore changes - safe to leave uncommitted
func IsWorkingTreeClean() bool {
	files, err := GetGitStatus()
	if err != nil {
		return false
	}

	// Filter out .gitignore - safe to leave uncommitted and proceed with release
	significantChanges := 0
	for _, f := range files {
		if f.Path != ".gitignore" {
			significantChanges++
		}
	}

	return significantChanges == 0
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

		// If path is a directory, expand it to individual files
		if strings.HasSuffix(path, "/") {
			expandedFiles, err := expandDirectory(path, status)
			if err == nil {
				files = append(files, expandedFiles...)
			}
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

func expandDirectory(dirPath, status string) ([]GitFile, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard", dirPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []GitFile
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		category := CategorizeFile(line)
		files = append(files, GitFile{
			Path:     line,
			Status:   status,
			Category: category,
		})
	}

	return files, nil
}

// GetFileChanges returns simplified, user-friendly file changes
func GetFileChanges() ([]FileChange, error) {
	var changes []FileChange

	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		status := strings.TrimSpace(line[:2])
		path := strings.TrimSpace(line[2:])

		// If path is a directory, expand it to individual files
		if strings.HasSuffix(path, "/") {
			expandedChanges, err := expandDirectoryToChanges(path, status)
			if err == nil {
				changes = append(changes, expandedChanges...)
			}
			continue
		}

		change := FileChange{
			Path:   path,
			Status: status,
		}

		switch status {
		case "M", " M", "MM":
			change.StatusText = "modified"
			change.Icon = "📝"
		case "A", " A":
			change.StatusText = "new file"
			change.Icon = "📄"
		case "D", " D":
			change.StatusText = "deleted"
			change.Icon = "🗑"
		case "??":
			change.StatusText = "untracked"
			change.Icon = "❓"
		case "R":
			change.StatusText = "renamed"
			change.Icon = "📋"
		default:
			change.StatusText = "changed"
			change.Icon = "📝"
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func expandDirectoryToChanges(dirPath, status string) ([]FileChange, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard", dirPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var changes []FileChange
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		change := FileChange{
			Path:       line,
			Status:     status,
			StatusText: "untracked",
			Icon:       "❓",
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// GetRepoStatus returns a simple string describing repository state
func GetRepoStatus() string {
	info, err := CheckRepoState()
	if err != nil {
		return "Error checking status"
	}

	switch info.Status {
	case RepoStatusClean:
		return "✅ Repository clean"
	case RepoStatusDirty:
		total := info.FileStats.Modified + info.FileStats.Added +
			info.FileStats.Deleted + info.FileStats.Untracked
		return fmt.Sprintf("🔴 %d uncommitted changes", total)
	case RepoStatusNoRemote:
		return "⚠️ No GitHub remote configured"
	case RepoStatusNoRepo:
		return "❌ Not a git repository"
	default:
		return string(info.Status)
	}
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