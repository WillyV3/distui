package executor

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"distui/internal/models"
)

func RunCommandStreaming(ctx context.Context, name string, args []string, dir string) tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()

		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = dir

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return models.CommandCompleteMsg{
				ExitCode: -1,
				Error:    fmt.Errorf("creating stdout pipe: %w", err),
				Duration: time.Since(startTime),
			}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return models.CommandCompleteMsg{
				ExitCode: -1,
				Error:    fmt.Errorf("creating stderr pipe: %w", err),
				Duration: time.Since(startTime),
			}
		}

		if err := cmd.Start(); err != nil {
			return models.CommandCompleteMsg{
				ExitCode: -1,
				Error:    fmt.Errorf("starting command: %w", err),
				Duration: time.Since(startTime),
			}
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
			}
		}()

		err = cmd.Wait()
		duration := time.Since(startTime)

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}

		return models.CommandCompleteMsg{
			ExitCode: exitCode,
			Error:    err,
			Duration: duration,
		}
	}
}

func RunCommandCapture(ctx context.Context, name string, args []string, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}