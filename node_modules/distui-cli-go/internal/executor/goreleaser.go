package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func CheckGoReleaserInstalled() bool {
	// First check PATH
	cmd := exec.Command("goreleaser", "--version")
	if err := cmd.Run(); err == nil {
		return true
	}

	// Check in ~/go/bin
	goreleaserPath := os.Getenv("HOME") + "/go/bin/goreleaser"
	cmd = exec.Command(goreleaserPath, "--version")
	return cmd.Run() == nil
}

func GetGitHubToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting GitHub token: %w", err)
	}

	// Trim whitespace and newlines from token
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("GitHub token is empty")
	}

	return token, nil
}

func RunGoReleaser(ctx context.Context, projectPath string, version string) tea.Cmd {
	return func() tea.Msg {
		if !CheckGoReleaserInstalled() {
			return fmt.Errorf("goreleaser not installed - install from https://goreleaser.com")
		}

		token, err := GetGitHubToken()
		if err != nil {
			return fmt.Errorf("getting GitHub token: %w", err)
		}
		// Ensure token is trimmed
		token = strings.TrimSpace(token)

		os.Setenv("GITHUB_TOKEN", token)

		// Try goreleaser in PATH first, then ~/go/bin
		goreleaserCmd := "goreleaser"
		if _, err := exec.LookPath("goreleaser"); err != nil {
			goreleaserCmd = os.Getenv("HOME") + "/go/bin/goreleaser"
		}

		return RunCommandStreaming(ctx, goreleaserCmd, []string{"release", "--clean"}, projectPath)()
	}
}

func RunGoReleaserWithOutput(ctx context.Context, projectPath string, version string, outputChan chan<- string) tea.Cmd {
	return func() tea.Msg {
		if !CheckGoReleaserInstalled() {
			return fmt.Errorf("goreleaser not installed - install from https://goreleaser.com")
		}

		token, err := GetGitHubToken()
		if err != nil {
			return fmt.Errorf("getting GitHub token: %w", err)
		}
		// Ensure token is trimmed
		token = strings.TrimSpace(token)

		// Try goreleaser in PATH first, then ~/go/bin
		goreleaserCmd := "goreleaser"
		if _, err := exec.LookPath("goreleaser"); err != nil {
			goreleaserCmd = os.Getenv("HOME") + "/go/bin/goreleaser"
		}

		// First, run a check to see if config has fatal errors (not deprecations)
		checkCmd := exec.Command(goreleaserCmd, "check")
		checkCmd.Dir = projectPath
		checkCmd.Env = append(os.Environ(), "GITHUB_TOKEN="+strings.TrimSpace(token))
		if checkOutput, checkErr := checkCmd.CombinedOutput(); checkErr != nil {
			outputStr := string(checkOutput)
			// Only fail on actual errors, not deprecation warnings
			if !strings.Contains(outputStr, "configuration is valid, but uses deprecated properties") {
				// Extract meaningful error from check output
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(strings.ToLower(line), "error") &&
					   !strings.Contains(line, "deprecated") &&
					   !strings.Contains(line, "configuration is valid") {
						return fmt.Errorf("configuration error: %s", strings.TrimSpace(line))
					}
					if strings.Contains(line, "✗") ||
					   strings.Contains(strings.ToLower(line), "invalid") {
						return fmt.Errorf("configuration error: %s", strings.TrimSpace(line))
					}
				}
			}
			// If we get here with checkErr but no specific error found,
			// it's probably just deprecation warnings - continue
		}

		// Create the actual release command
		cmd := exec.Command(goreleaserCmd, "release", "--clean")
		cmd.Dir = projectPath
		cmd.Env = append(os.Environ(), "GITHUB_TOKEN="+strings.TrimSpace(token))

		// Get stdout and stderr pipes
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("creating stdout pipe: %w", err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("creating stderr pipe: %w", err)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("starting goreleaser: %w", err)
		}

		// Collect all output including errors
		var allOutput []string
		var lastErrorLine string

		// Read and format stdout
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stdout.Read(buf)
				if n > 0 {
					lines := string(buf[:n])
					for _, line := range splitLines(lines) {
						if line != "" {
							allOutput = append(allOutput, line)
							// Track potential error lines
							if strings.Contains(strings.ToLower(line), "error") ||
							   strings.Contains(strings.ToLower(line), "failed") ||
							   strings.Contains(line, "✗") {
								lastErrorLine = line
							}
							formattedLine := formatGoReleaserOutput(line)
							if outputChan != nil {
								select {
								case outputChan <- formattedLine:
								default:
								}
							}
						}
					}
				}
				if err != nil {
					break
				}
			}
		}()

		// Read stderr and capture errors
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stderr.Read(buf)
				if n > 0 {
					lines := string(buf[:n])
					for _, line := range splitLines(lines) {
						if line != "" {
							allOutput = append(allOutput, line)
							// stderr is more likely to contain the actual error
							if line != "" {
								lastErrorLine = line
							}
							if outputChan != nil {
								select {
								case outputChan <- line:
								default:
								}
							}
						}
					}
				}
				if err != nil {
					break
				}
			}
		}()

		// Wait for command to complete
		err = cmd.Wait()
		if err != nil {
			// Find the most relevant error message
			errorMsg := "release failed"

			// Try to extract a meaningful error
			if lastErrorLine != "" {
				// Clean up common prefixes
				errorMsg = strings.TrimPrefix(lastErrorLine, "• ")
				errorMsg = strings.TrimPrefix(errorMsg, "✗ ")
				errorMsg = strings.TrimPrefix(errorMsg, "⨯ ")
				errorMsg = strings.TrimPrefix(errorMsg, "   ⨯ ")
				errorMsg = strings.TrimSpace(errorMsg)
			} else {
				// Look through all output for clues
				for i := len(allOutput) - 1; i >= 0 && i > len(allOutput)-10; i-- {
					if i < len(allOutput) {
						line := allOutput[i]
						if strings.Contains(strings.ToLower(line), "error") ||
						   strings.Contains(strings.ToLower(line), "failed") ||
						   strings.Contains(strings.ToLower(line), "release is already") {
							errorMsg = strings.TrimSpace(line)
							break
						}
					}
				}
			}

			return fmt.Errorf("%s", errorMsg)
		}

		return nil
	}
}

func splitLines(text string) []string {
	lines := []string{}
	current := ""
	for _, r := range text {
		if r == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else if r != '\r' {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func formatGoReleaserOutput(line string) string {
	// Pretty format GoReleaser output
	if strings.Contains(line, "• building") {
		return "• Building binaries..."
	}
	if strings.Contains(line, "• loading") {
		return "• Loading configuration..."
	}
	if strings.Contains(line, "• publishing") {
		return "• Publishing release..."
	}
	if strings.Contains(line, "• archiving") {
		return "• Creating archives..."
	}
	if strings.Contains(line, "• creating") {
		return "• Creating GitHub release..."
	}
	if strings.Contains(line, "• uploading") {
		return "• Uploading artifacts..."
	}
	if strings.Contains(line, "• announcing") {
		return "• Announcing release..."
	}
	if strings.Contains(line, "✓") || strings.Contains(line, "✔") {
		return "✓ " + strings.TrimSpace(line)
	}

	// Return cleaned line
	return strings.TrimSpace(line)
}

func ValidateGoReleaserConfig(projectPath string) error {
	if !CheckGoReleaserInstalled() {
		return fmt.Errorf("goreleaser not installed")
	}

	cmd := exec.Command("goreleaser", "check")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("goreleaser config invalid: %s", string(output))
	}

	return nil
}

func RunGoReleaserSnapshot(ctx context.Context, projectPath string) tea.Cmd {
	return func() tea.Msg {
		if !CheckGoReleaserInstalled() {
			return fmt.Errorf("goreleaser not installed")
		}

		return RunCommandStreaming(ctx, "goreleaser", []string{"release", "--snapshot", "--clean", "--skip=publish"}, projectPath)()
	}
}

func CheckGoReleaserConfigExists(projectPath string) bool {
	configFiles := []string{
		".goreleaser.yml",
		".goreleaser.yaml",
		"goreleaser.yml",
		"goreleaser.yaml",
	}

	for _, file := range configFiles {
		if _, err := os.Stat(projectPath + "/" + file); err == nil {
			return true
		}
	}

	return false
}