package gitcleanup

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommitFiles stages and commits the specified files
func CommitFiles(files []GitFile, message string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files to commit")
	}

	// Stage files
	for _, file := range files {
		cmd := exec.Command("git", "add", file.Path)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add %s: %w", file.Path, err)
		}
	}

	// Commit with message
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// AddToGitignore adds files to .gitignore
func AddToGitignore(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// Open or create .gitignore
	file, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .gitignore: %w", err)
	}
	defer file.Close()

	// Check existing entries to avoid duplicates
	existingContent, _ := os.ReadFile(".gitignore")
	existing := strings.Split(string(existingContent), "\n")
	existingMap := make(map[string]bool)
	for _, line := range existing {
		existingMap[strings.TrimSpace(line)] = true
	}

	// Add new entries
	for _, path := range paths {
		if !existingMap[path] {
			if _, err := file.WriteString(path + "\n"); err != nil {
				return fmt.Errorf("failed to write to .gitignore: %w", err)
			}
		}
	}

	return nil
}

// ExecuteSmartCommit performs the smart commit based on file categorization
func ExecuteSmartCommit(items []CleanupItem) (string, error) {
	var toCommit []GitFile
	var toIgnore []string
	var filePaths []string

	// Process items based on their actions
	for _, item := range items {
		switch item.Action {
		case "commit":
			toCommit = append(toCommit, GitFile{
				Path:     item.Path,
				Status:   item.Status,
				Category: FileCategory(item.Category),
			})
			filePaths = append(filePaths, item.Path)
		case "ignore":
			toIgnore = append(toIgnore, item.Path)
		}
	}

	// Add files to .gitignore if needed
	if len(toIgnore) > 0 {
		// Untrack files that are already tracked
		for _, path := range toIgnore {
			// Mark as unchanged to stop tracking changes
			cmd := exec.Command("git", "update-index", "--assume-unchanged", path)
			cmd.Run() // Ignore errors - file might not be tracked
		}

		if err := AddToGitignore(toIgnore); err != nil {
			return "", fmt.Errorf("failed to update .gitignore: %w", err)
		}
		// Also stage .gitignore
		cmd := exec.Command("git", "add", ".gitignore")
		cmd.Run()
	}

	// Generate commit message
	commitMsg := SuggestCommitMessage(filePaths)

	// Commit files
	if len(toCommit) > 0 {
		if err := CommitFiles(toCommit, commitMsg); err != nil {
			return "", err
		}
		return commitMsg, nil
	}

	return "", fmt.Errorf("no files to commit")
}

// CleanupItem represents an item in the cleanup list (matches handler type)
type CleanupItem struct {
	Path     string
	Status   string
	Category string
	Action   string
}