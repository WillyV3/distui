package fsearch

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func FuzzySearchDirectories(query string, maxResults int) []string {
	if len(query) < 2 {
		return []string{}
	}

	homeDir, _ := os.UserHomeDir()

	fdCmd := exec.Command("fd",
		"--type", "d",
		"--max-depth", "4",
		"--base-directory", homeDir,
		"--absolute-path",
		"--exclude", "node_modules",
		"--exclude", ".git",
		"--exclude", "Library",
		"--exclude", ".cache")

	fdOutput, err := fdCmd.Output()
	if err != nil {
		return []string{}
	}

	fzfCmd := exec.Command("fzf", "--filter", query, "-i")
	fzfCmd.Stdin = bytes.NewReader(fdOutput)

	fzfOutput, err := fzfCmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(fzfOutput)), "\n")
	results := []string{}
	for _, line := range lines {
		if line != "" && len(results) < maxResults {
			results = append(results, line)
		}
	}

	return results
}
