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

	token := string(output)
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

		os.Setenv("GITHUB_TOKEN", token)

		// Try goreleaser in PATH first, then ~/go/bin
		goreleaserCmd := "goreleaser"
		if _, err := exec.LookPath("goreleaser"); err != nil {
			goreleaserCmd = os.Getenv("HOME") + "/go/bin/goreleaser"
		}

		// Create the command
		cmd := exec.Command(goreleaserCmd, "release", "--clean")
		cmd.Dir = projectPath
		cmd.Env = append(os.Environ(), "GITHUB_TOKEN="+token)

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

		// Read and format output
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stdout.Read(buf)
				if n > 0 {
					lines := string(buf[:n])
					for _, line := range splitLines(lines) {
						if line != "" {
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

		// Read stderr too
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stderr.Read(buf)
				if n > 0 {
					lines := string(buf[:n])
					for _, line := range splitLines(lines) {
						if line != "" && outputChan != nil {
							select {
							case outputChan <- line:
							default:
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
			return fmt.Errorf("goreleaser failed: %w", err)
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