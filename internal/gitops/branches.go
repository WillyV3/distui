package gitops

import (
	"fmt"
	"os/exec"
	"strings"

	"distui/internal/models"
)

func ListBranches() ([]models.BranchInfo, error) {
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)|%(upstream:short)|%(HEAD)", "refs/heads")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []models.BranchInfo

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		branch := models.BranchInfo{
			Name:           parts[0],
			TrackingBranch: parts[1],
			IsCurrent:      parts[2] == "*",
			AheadCount:     0,
			BehindCount:    0,
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func PushToBranch(branch string) error {
	cmd := exec.Command("git", "push", "origin", fmt.Sprintf("HEAD:refs/heads/%s", branch))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pushing to branch %s: %w\nOutput: %s", branch, err, string(output))
	}

	return nil
}
