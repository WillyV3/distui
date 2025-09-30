package executor

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/models"
)

func RunTests(ctx context.Context, projectPath string) tea.Cmd {
	return RunCommandStreaming(ctx, "go", []string{"test", "./..."}, projectPath)
}

func CheckTestsExist(projectPath string) bool {
	cmd := RunCommandStreaming(context.Background(), "go", []string{"list", "./..."}, projectPath)
	msg := cmd()

	if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
		return completeMsg.ExitCode == 0
	}

	return false
}

func RunTestsWithCoverage(ctx context.Context, projectPath string) tea.Cmd {
	return RunCommandStreaming(ctx, "go", []string{"test", "-cover", "./..."}, projectPath)
}

func RunSpecificTest(ctx context.Context, projectPath string, testName string) tea.Cmd {
	return RunCommandStreaming(ctx, "go", []string{"test", "-run", testName, "./..."}, projectPath)
}

func ValidateProject(projectPath string) error {
	cmd := RunCommandStreaming(context.Background(), "go", []string{"mod", "verify"}, projectPath)
	msg := cmd()

	if completeMsg, ok := msg.(models.CommandCompleteMsg); ok {
		if completeMsg.ExitCode != 0 {
			return fmt.Errorf("go mod verify failed: %w", completeMsg.Error)
		}
	}

	return nil
}