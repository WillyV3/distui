package main

import (
	_ "embed"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//go:embed ascii-art.txt
var asciiArt string

// renderASCIIArt renders the distui ASCII art left-aligned and styled
// maxLines controls animation: 0 = not started (show nothing), -1 = complete (show all), >0 = show that many lines
// flashCount controls color flash: 0-8 cycles through teal/orange/purple
func renderASCIIArt(terminalWidth int, maxLines int, flashCount int) string {
	// Clean the ASCII art (trim trailing spaces from each line)
	lines := strings.Split(strings.TrimSpace(asciiArt), "\n")

	// Determine how many lines to show
	linesToShow := 0
	if maxLines == -1 || maxLines >= len(lines) {
		// Animation complete or not animated - show all
		linesToShow = len(lines)
	} else if maxLines > 0 {
		// Animation in progress - show partial
		linesToShow = maxLines
	}
	// maxLines == 0 means not started, show nothing

	// If nothing to show, return empty
	if linesToShow == 0 {
		return ""
	}

	// Determine color based on flash count (teal -> orange -> purple cycle)
	var artColor string
	if flashCount > 0 {
		colorIndex := flashCount % 3
		switch colorIndex {
		case 1:
			artColor = "51"  // Bright teal/cyan
		case 2:
			artColor = "208" // Bright orange
		case 0:
			artColor = "141" // Purple
		}
	} else {
		artColor = "117" // Default cyan
	}

	// Style for the ASCII art with color cycling
	artStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(artColor)).
		Bold(true)

	// Build left-aligned output
	var output strings.Builder
	output.WriteString("\n")

	for i := 0; i < linesToShow; i++ {
		output.WriteString(artStyle.Render(lines[i]))
		output.WriteString("\n")
	}

	output.WriteString("\n")

	return output.String()
}
