package filescanner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DeleteFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("deleting file: %w", err)
	}
	return nil
}

func AddToGitignore(path string) error {
	gitignorePath := ".gitignore"

	existingLines, err := readGitignore(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading .gitignore: %w", err)
	}

	pattern := filepath.Base(path)
	if strings.HasPrefix(pattern, ".") {
		pattern = "*" + pattern
	} else {
		ext := filepath.Ext(path)
		if ext != "" {
			pattern = "*" + ext
		}
	}

	for _, line := range existingLines {
		if strings.TrimSpace(line) == pattern {
			return nil
		}
	}

	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening .gitignore: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(pattern + "\n"); err != nil {
		return fmt.Errorf("writing to .gitignore: %w", err)
	}

	return nil
}

func ArchiveFile(path string) error {
	timestamp := time.Now().Format("2006-01-02-150405")
	archiveRoot := filepath.Join(".distui-archive", timestamp)

	if err := os.MkdirAll(archiveRoot, 0755); err != nil {
		return fmt.Errorf("creating archive directory: %w", err)
	}

	rel, err := filepath.Rel(".", path)
	if err != nil {
		rel = filepath.Base(path)
	}

	archivePath := filepath.Join(archiveRoot, rel)
	archiveDir := filepath.Dir(archivePath)

	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("creating archive subdirectory: %w", err)
	}

	if err := os.Rename(path, archivePath); err != nil {
		if err := copyFile(path, archivePath); err != nil {
			return fmt.Errorf("copying file to archive: %w", err)
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing original file: %w", err)
		}
	}

	return nil
}

func readGitignore(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}
