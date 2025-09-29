package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func CheckGoReleaserInstalled() bool {
	cmd := exec.Command("goreleaser", "--version")
	err := cmd.Run()
	return err == nil
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

		return RunCommandStreaming(ctx, "goreleaser", []string{"release", "--clean"}, projectPath)()
	}
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