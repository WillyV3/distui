package executor

import (
	"context"
	"os/exec"
	"time"
)

type ReleaseExecutor struct {
	projectPath string
	config      ReleaseConfig
}

type ReleaseConfig struct {
	Version   string
	SkipTests bool
	AutoPush  bool
}

type ExecutionResult struct {
	Success  bool
	Output   []string
	Error    error
	Duration time.Duration
}

type ExecutionStep struct {
	Name    string
	Command string
	Args    []string
}

func NewReleaseExecutor(projectPath string, config ReleaseConfig) *ReleaseExecutor {
	return &ReleaseExecutor{
		projectPath: projectPath,
		config:      config,
	}
}

func (r *ReleaseExecutor) Execute(ctx context.Context) (*ExecutionResult, error) {
	startTime := time.Now()

	if !r.config.SkipTests {
		if err := r.runTests(ctx); err != nil {
			return &ExecutionResult{Success: false, Error: err, Duration: time.Since(startTime)}, err
		}
	}

	if err := r.buildRelease(ctx); err != nil {
		return &ExecutionResult{Success: false, Error: err, Duration: time.Since(startTime)}, err
	}

	if err := r.createTag(ctx); err != nil {
		return &ExecutionResult{Success: false, Error: err, Duration: time.Since(startTime)}, err
	}

	if r.config.AutoPush {
		if err := r.pushTag(ctx); err != nil {
			return &ExecutionResult{Success: false, Error: err, Duration: time.Since(startTime)}, err
		}
	}

	return &ExecutionResult{Success: true, Duration: time.Since(startTime)}, nil
}

func (r *ReleaseExecutor) runTests(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", "test", "./...")
	cmd.Dir = r.projectPath
	return cmd.Run()
}

func (r *ReleaseExecutor) buildRelease(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "dist/")
	cmd.Dir = r.projectPath
	return cmd.Run()
}

func (r *ReleaseExecutor) createTag(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "tag", r.config.Version)
	cmd.Dir = r.projectPath
	return cmd.Run()
}

func (r *ReleaseExecutor) pushTag(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "push", "origin", r.config.Version)
	cmd.Dir = r.projectPath
	return cmd.Run()
}