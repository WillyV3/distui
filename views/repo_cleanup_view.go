package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"distui/handlers"
	"distui/internal/models"
)

var (
	mediaStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	docsStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	artifactsStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cleanupSelectedStyle  = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("12"))
)

func RenderRepoCleanup(m handlers.RepoCleanupModel) string {
	if m.Scanning {
		return fmt.Sprintf("\n%s Scanning repository...\n", m.ScanSpinner.View())
	}

	if m.ScanResult == nil {
		return "\nNo scan results available. Press 'r' to scan.\n"
	}

	var b strings.Builder

	b.WriteString("\n┌─ REPOSITORY CLEANUP ─────────────────────────\n")
	b.WriteString("│\n")

	mediaCount := len(m.ScanResult.MediaFiles)
	docsCount := len(m.ScanResult.ExcessDocs)
	artifactsCount := len(m.ScanResult.DevArtifacts)

	totalSize := humanizeBytes(m.ScanResult.TotalSizeBytes)

	b.WriteString(fmt.Sprintf("│  %s (%d files, %s)\n",
		mediaStyle.Render("Media Files"),
		mediaCount,
		humanizeBytes(calculateGroupSize(m.ScanResult.MediaFiles))))

	currentIndex := 0
	for _, file := range m.ScanResult.MediaFiles {
		renderFile(&b, currentIndex, file, m.SelectedIndex)
		currentIndex++
	}

	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("│  %s (%d files, %s)\n",
		docsStyle.Render("Excess Docs"),
		docsCount,
		humanizeBytes(calculateGroupSize(m.ScanResult.ExcessDocs))))

	for _, file := range m.ScanResult.ExcessDocs {
		renderFile(&b, currentIndex, file, m.SelectedIndex)
		currentIndex++
	}

	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("│  %s (%d files, %s)\n",
		artifactsStyle.Render("Dev Artifacts"),
		artifactsCount,
		humanizeBytes(calculateGroupSize(m.ScanResult.DevArtifacts))))

	for _, file := range m.ScanResult.DevArtifacts {
		renderFile(&b, currentIndex, file, m.SelectedIndex)
		currentIndex++
	}

	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("│  Total: %d files, %s\n", len(m.FlaggedFiles), totalSize))
	b.WriteString(fmt.Sprintf("│  Scan duration: %s\n", m.ScanResult.ScanDuration))
	b.WriteString("│\n")
	b.WriteString("│  [d] Delete  [i] Ignore  [a] Archive  [r] Re-scan  [Esc] Cancel\n")
	b.WriteString("└──────────────────────────────────────────────\n")

	return b.String()
}

func renderFile(b *strings.Builder, index int, file models.FlaggedFile, selectedIndex int) {
	prefix := "  "
	line := fmt.Sprintf("│  %s%s (%s) - %s\n", prefix, file.Path, humanizeBytes(file.SizeBytes), file.SuggestedAction)

	if index == selectedIndex {
		prefix = "> "
		line = fmt.Sprintf("│  %s%s (%s) - %s", prefix, file.Path, humanizeBytes(file.SizeBytes), file.SuggestedAction)
		b.WriteString(cleanupSelectedStyle.Render(line) + "\n")
	} else {
		b.WriteString(line)
	}
}

func calculateGroupSize(files []models.FlaggedFile) int64 {
	var total int64
	for _, file := range files {
		total += file.SizeBytes
	}
	return total
}

func humanizeBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}
