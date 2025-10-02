package fileops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveCustomFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		".goreleaser.yaml",
		"package.json",
	}

	for _, file := range testFiles {
		content := "test content for " + file
		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Archive the files
	backupPath, err := ArchiveCustomFiles(tmpDir, testFiles)
	if err != nil {
		t.Fatalf("ArchiveCustomFiles failed: %v", err)
	}

	// Verify backup directory was created
	if !strings.Contains(backupPath, ".distui-backup") {
		t.Errorf("Expected backup path to contain .distui-backup, got %s", backupPath)
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup directory was not created: %s", backupPath)
	}

	// Verify files were moved to backup
	for _, file := range testFiles {
		backupFile := filepath.Join(backupPath, file)
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			t.Errorf("File %s was not archived to %s", file, backupFile)
		}

		// Verify original file no longer exists
		originalFile := filepath.Join(tmpDir, file)
		if _, err := os.Stat(originalFile); !os.IsNotExist(err) {
			t.Errorf("Original file %s still exists after archive", originalFile)
		}
	}

	// Verify archived files have correct content
	archivedContent, err := os.ReadFile(filepath.Join(backupPath, ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("Failed to read archived file: %v", err)
	}

	expectedContent := "test content for .goreleaser.yaml"
	if string(archivedContent) != expectedContent {
		t.Errorf("Archived content mismatch. Expected %q, got %q", expectedContent, string(archivedContent))
	}
}

func TestArchiveCustomFiles_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to archive empty list
	backupPath, err := ArchiveCustomFiles(tmpDir, []string{})
	if err != nil {
		t.Fatalf("ArchiveCustomFiles with empty list failed: %v", err)
	}

	// Backup directory should still be created
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup directory was not created for empty list")
	}
}

func TestArchiveCustomFiles_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to archive nonexistent file
	_, err := ArchiveCustomFiles(tmpDir, []string{"nonexistent.yaml"})
	if err == nil {
		t.Error("Expected error when archiving nonexistent file, got nil")
	}
}
