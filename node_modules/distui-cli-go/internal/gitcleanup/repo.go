package gitcleanup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type RepoStatus string

const (
	RepoStatusNoRepo    RepoStatus = "no_repo"
	RepoStatusNoRemote  RepoStatus = "no_remote"
	RepoStatusUnpushed  RepoStatus = "unpushed"
	RepoStatusDirty     RepoStatus = "dirty"
	RepoStatusClean     RepoStatus = "clean"
)

type RepoInfo struct {
	Status          RepoStatus
	GitExists       bool
	RemoteExists    bool
	RemoteURL       string
	Branch          string
	Owner           string
	RepoName        string
	FileStats       FileStats
	UnpushedCommits int
}

type FileStats struct {
	Modified int
	Added    int
	Deleted  int
	Untracked int
}

func CheckRepoState() (*RepoInfo, error) {
	info := &RepoInfo{}

	gitDir := filepath.Join(".", ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		info.Status = RepoStatusNoRepo
		return info, nil
	}
	info.GitExists = true

	branchCmd := exec.Command("git", "branch", "--show-current")
	if output, err := branchCmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(output))
	}

	remoteCmd := exec.Command("git", "remote", "get-url", "origin")
	if output, err := remoteCmd.Output(); err == nil {
		info.RemoteURL = strings.TrimSpace(string(output))
		info.RemoteExists = true

		if strings.Contains(info.RemoteURL, "github.com") {
			parts := strings.Split(info.RemoteURL, "/")
			if len(parts) >= 2 {
				info.RepoName = strings.TrimSuffix(parts[len(parts)-1], ".git")
				info.Owner = parts[len(parts)-2]
				if strings.Contains(info.Owner, ":") {
					info.Owner = strings.Split(info.Owner, ":")[1]
				}
			}
		}
	}

	if !info.RemoteExists {
		info.Status = RepoStatusNoRemote
		return info, nil
	}

	statusCmd := exec.Command("git", "status", "--porcelain")
	output, _ := statusCmd.Output()
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		status := line[:2]
		switch {
		case strings.Contains(status, "M"):
			info.FileStats.Modified++
		case strings.Contains(status, "A"):
			info.FileStats.Added++
		case strings.Contains(status, "D"):
			info.FileStats.Deleted++
		case status == "??":
			info.FileStats.Untracked++
		}
	}

	hasChanges := info.FileStats.Modified > 0 || info.FileStats.Added > 0 ||
		info.FileStats.Deleted > 0 || info.FileStats.Untracked > 0

	// Check for unpushed commits if we have a remote
	hasUnpushedCommits := false
	if info.RemoteExists {
		// Try with upstream first
		cmd := exec.Command("git", "rev-list", "--count", "@{upstream}..HEAD")
		output, err := cmd.Output()

		// If upstream not set, try origin/main or origin/master
		if err != nil {
			// Check if remote branch exists first
			checkCmd := exec.Command("git", "ls-remote", "--heads", "origin", "main")
			if checkOutput, checkErr := checkCmd.Output(); checkErr == nil && len(checkOutput) > 0 {
				cmd = exec.Command("git", "rev-list", "--count", "origin/main..HEAD")
				output, err = cmd.Output()
			} else {
				// Try master
				checkCmd = exec.Command("git", "ls-remote", "--heads", "origin", "master")
				if checkOutput, checkErr := checkCmd.Output(); checkErr == nil && len(checkOutput) > 0 {
					cmd = exec.Command("git", "rev-list", "--count", "origin/master..HEAD")
					output, err = cmd.Output()
				} else {
					// Remote exists but no branches - all local commits are unpushed
					cmd = exec.Command("git", "rev-list", "--count", "HEAD")
					output, err = cmd.Output()
				}
			}
		}

		if err == nil {
			countStr := strings.TrimSpace(string(output))
			if countStr != "0" && countStr != "" {
				hasUnpushedCommits = true
				var count int
				fmt.Sscanf(countStr, "%d", &count)
				info.UnpushedCommits = count
			}
		}
	}

	if hasChanges {
		info.Status = RepoStatusDirty
	} else if hasUnpushedCommits {
		info.Status = RepoStatusUnpushed
	} else {
		info.Status = RepoStatusClean
	}

	return info, nil
}

func (r RepoStatus) String() string {
	switch r {
	case RepoStatusNoRepo:
		return "Not a git repository"
	case RepoStatusNoRemote:
		return "No remote configured"
	case RepoStatusUnpushed:
		return "Changes not pushed"
	case RepoStatusDirty:
		return fmt.Sprintf("Uncommitted changes")
	case RepoStatusClean:
		return "Ready for release"
	default:
		return string(r)
	}
}