package main

import (
	_ "embed"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

//go:embed ascii-art.txt
var asciiArt string

// renderASCIIArt renders the distui ASCII art left-aligned and styled
func renderASCIIArt(terminalWidth int) string {
	// Clean the ASCII art (trim trailing spaces from each line)
	lines := strings.Split(strings.TrimSpace(asciiArt), "\n")

	// Style for the ASCII art - vibrant cyan/blue gradient
	artStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("117")).
		Bold(true)

	// Build left-aligned output
	var output strings.Builder
	output.WriteString("\n")

	for _, line := range lines {
		output.WriteString(artStyle.Render(line))
		output.WriteString("\n")
	}

	output.WriteString("\n")

	return output.String()
}
