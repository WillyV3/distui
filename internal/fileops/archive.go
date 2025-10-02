package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchiveCustomFiles moves specified files to .distui-backup/<timestamp>/ directory.
// Returns the backup path and any error.
func ArchiveCustomFiles(projectPath string, files []string) (string, error) {
	timestamp := time.Now().Format("20060102-150405.000000")
	backupDir := filepath.Join(projectPath, ".distui-backup", timestamp)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("creating backup dir: %w", err)
	}

	for _, file := range files {
		srcPath := filepath.Join(projectPath, file)
		dstPath := filepath.Join(backupDir, file)

		// Create parent directory if file is nested
		dstParent := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstParent, 0755); err != nil {
			return "", fmt.Errorf("creating backup parent dir for %s: %w", file, err)
		}

		if err := os.Rename(srcPath, dstPath); err != nil {
			return "", fmt.Errorf("archiving %s: %w", file, err)
		}
	}

	return backupDir, nil
}
