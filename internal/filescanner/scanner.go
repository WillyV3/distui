package filescanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"distui/internal/models"
	"github.com/muesli/gitcha"
)

func ScanRepository(root string) (*models.CleanupScanResult, error) {
	startTime := time.Now()
	result := &models.CleanupScanResult{
		MediaFiles:   []models.FlaggedFile{},
		ExcessDocs:   []models.FlaggedFile{},
		DevArtifacts: []models.FlaggedFile{},
		ScannedAt:    startTime,
	}

	_, err := gitcha.GitRepoForPath(root)
	if err == nil {
		if err := scanTrackedMedia(root, result); err != nil {
			return nil, fmt.Errorf("scanning tracked media: %w", err)
		}

		if err := scanTrackedDocs(root, result); err != nil {
			return nil, fmt.Errorf("scanning tracked docs: %w", err)
		}
	}

	if err := scanUntrackedArtifacts(root, result); err != nil {
		return nil, fmt.Errorf("scanning untracked artifacts: %w", err)
	}

	result.ScanDuration = time.Since(startTime)
	result.TotalSizeBytes = calculateTotalSize(result)

	return result, nil
}

func scanTrackedMedia(root string, result *models.CleanupScanResult) error {
	mediaPatterns := []string{
		"*.mp4", "*.mov", "*.avi", "*.mkv", "*.flv", "*.wmv",
		"*.wav", "*.mp3", "*.flac", "*.aac", "*.ogg",
		"*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp", "*.svg",
	}

	ch, err := gitcha.FindFiles(root, mediaPatterns)
	if err != nil {
		return err
	}

	for file := range ch {
		basename := filepath.Base(file.Path)
		if isIconOrLogo(basename) {
			continue
		}

		info, _ := os.Stat(file.Path)
		result.MediaFiles = append(result.MediaFiles, models.FlaggedFile{
			Path:            file.Path,
			IssueType:       "media",
			SizeBytes:       info.Size(),
			SuggestedAction: "delete",
			FlaggedAt:       time.Now(),
		})
	}

	return nil
}

func scanTrackedDocs(root string, result *models.CleanupScanResult) error {
	ch, err := gitcha.FindFilesExcept(root,
		[]string{"*.md", "*.markdown"},
		[]string{"README.md", "README.markdown", "readme.md"},
	)
	if err != nil {
		return err
	}

	for file := range ch {
		info, _ := os.Stat(file.Path)
		result.ExcessDocs = append(result.ExcessDocs, models.FlaggedFile{
			Path:            file.Path,
			IssueType:       "excess-docs",
			SizeBytes:       info.Size(),
			SuggestedAction: "archive",
			FlaggedAt:       time.Now(),
		})
	}

	docPatterns := []string{"*.pdf", "*.doc", "*.docx", "*.ppt", "*.pptx"}
	ch, err = gitcha.FindFiles(root, docPatterns)
	if err != nil {
		return err
	}

	for file := range ch {
		info, _ := os.Stat(file.Path)
		result.ExcessDocs = append(result.ExcessDocs, models.FlaggedFile{
			Path:            file.Path,
			IssueType:       "excess-docs",
			SizeBytes:       info.Size(),
			SuggestedAction: "archive",
			FlaggedAt:       time.Now(),
		})
	}

	return nil
}

func scanUntrackedArtifacts(root string, result *models.CleanupScanResult) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == ".distui-archive" {
				return filepath.SkipDir
			}
			return nil
		}

		basename := d.Name()
		ext := filepath.Ext(basename)

		if basename == ".DS_Store" || basename == "Thumbs.db" || basename == "desktop.ini" {
			info, _ := d.Info()
			result.DevArtifacts = append(result.DevArtifacts, models.FlaggedFile{
				Path:            path,
				IssueType:       "dev-artifact",
				SizeBytes:       info.Size(),
				SuggestedAction: "ignore",
				FlaggedAt:       time.Now(),
			})
			return nil
		}

		if ext == ".log" || ext == ".tmp" || ext == ".temp" ||
			ext == ".swp" || ext == ".swo" {
			info, _ := d.Info()
			result.DevArtifacts = append(result.DevArtifacts, models.FlaggedFile{
				Path:            path,
				IssueType:       "dev-artifact",
				SizeBytes:       info.Size(),
				SuggestedAction: "ignore",
				FlaggedAt:       time.Now(),
			})
		}

		return nil
	})
}

func isIconOrLogo(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.Contains(lower, "icon") ||
		strings.Contains(lower, "logo") ||
		lower == "favicon.ico"
}

func calculateTotalSize(result *models.CleanupScanResult) int64 {
	var total int64
	for _, f := range result.MediaFiles {
		total += f.SizeBytes
	}
	for _, f := range result.ExcessDocs {
		total += f.SizeBytes
	}
	for _, f := range result.DevArtifacts {
		total += f.SizeBytes
	}
	return total
}
