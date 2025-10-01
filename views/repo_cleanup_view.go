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

	chromeLines := 10
	maxFileLines := m.Height - chromeLines
	if maxFileLines < 5 {
		maxFileLines = 5
	}

	var b strings.Builder
	b.WriteString("\n┌─ REPOSITORY CLEANUP ─────────────────────────\n")
	b.WriteString("│\n")

	totalSize := humanizeBytes(m.ScanResult.TotalSizeBytes)
	totalFiles := len(m.FlaggedFiles)

	b.WriteString(fmt.Sprintf("│  %s (%d files, %s)\n",
		mediaStyle.Render("Media Files"),
		len(m.ScanResult.MediaFiles),
		humanizeBytes(calculateGroupSize(m.ScanResult.MediaFiles))))

	linesUsed := 0
	currentIndex := 0
	visibleStart, visibleEnd := calculateVisibleRange(m.SelectedIndex, totalFiles, maxFileLines)

	currentIndex = renderGroupPaginated(&b, m.ScanResult.MediaFiles, currentIndex, m.SelectedIndex,
		visibleStart, visibleEnd, &linesUsed, maxFileLines)

	if linesUsed < maxFileLines {
		b.WriteString("│\n")
		linesUsed++
		b.WriteString(fmt.Sprintf("│  %s (%d files, %s)\n",
			docsStyle.Render("Excess Docs"),
			len(m.ScanResult.ExcessDocs),
			humanizeBytes(calculateGroupSize(m.ScanResult.ExcessDocs))))
		linesUsed++
		currentIndex = renderGroupPaginated(&b, m.ScanResult.ExcessDocs, currentIndex, m.SelectedIndex,
			visibleStart, visibleEnd, &linesUsed, maxFileLines)
	}

	if linesUsed < maxFileLines {
		b.WriteString("│\n")
		linesUsed++
		b.WriteString(fmt.Sprintf("│  %s (%d files, %s)\n",
			artifactsStyle.Render("Dev Artifacts"),
			len(m.ScanResult.DevArtifacts),
			humanizeBytes(calculateGroupSize(m.ScanResult.DevArtifacts))))
		linesUsed++
		renderGroupPaginated(&b, m.ScanResult.DevArtifacts, currentIndex, m.SelectedIndex,
			visibleStart, visibleEnd, &linesUsed, maxFileLines)
	}

	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("│  Total: %d files, %s\n", totalFiles, totalSize))
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

func calculateVisibleRange(selected, total, maxLines int) (int, int) {
	if total <= maxLines {
		return 0, total
	}

	start := selected - maxLines/2
	if start < 0 {
		start = 0
	}

	end := start + maxLines
	if end > total {
		end = total
		start = end - maxLines
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

func renderGroupPaginated(b *strings.Builder, files []models.FlaggedFile,
	startIndex, selectedIndex, visibleStart, visibleEnd int, linesUsed *int, maxLines int) int {

	currentIndex := startIndex
	for _, file := range files {
		if *linesUsed >= maxLines {
			break
		}

		if currentIndex >= visibleStart && currentIndex < visibleEnd {
			renderFile(b, currentIndex, file, selectedIndex)
			*linesUsed++
		}

		currentIndex++
	}

	return currentIndex
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
