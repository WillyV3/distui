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
func renderASCIIArt(terminalWidth int, maxLines int) string {
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

	// Style for the ASCII art - vibrant cyan/blue gradient
	artStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
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
