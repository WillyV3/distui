package executor

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type HomebrewUpdateResult struct {
	Success      bool
	FormulaPath  string
	TapPath      string
	CommitHash   string
	Error        error
}

func UpdateHomebrewTap(ctx context.Context, projectName string, version string, tapRepo string, repoOwner string, repoName string) tea.Cmd {
	return func() tea.Msg {
		// Log for debugging
		fmt.Printf("UpdateHomebrewTap called: project=%s, version=%s, tapRepo=%s, owner=%s, repo=%s\n",
			projectName, version, tapRepo, repoOwner, repoName)

		// Wait for GitHub to process the tag
		fmt.Println("Waiting 10 seconds for GitHub to process the tag...")
		time.Sleep(10 * time.Second)

		tarballURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", repoOwner, repoName, version)
		fmt.Printf("Tarball URL: %s\n", tarballURL)

		sha256sum, err := downloadAndCalculateSHA256(tarballURL)
		if err != nil {
			return HomebrewUpdateResult{
				Success: false,
				Error:   fmt.Errorf("calculating SHA256 for %s: %w", tarballURL, err),
			}
		}
		fmt.Printf("SHA256: %s\n", sha256sum)

		// tapRepo is like "willyv3/homebrew-tap", convert to local path
		homeDir := os.Getenv("HOME")
		tapPath := filepath.Join(homeDir, "homebrew-tap")
		formulaPath := filepath.Join(tapPath, "Formula", projectName+".rb")

		// Check if formula exists, if not create it first
		if _, err := os.Stat(formulaPath); os.IsNotExist(err) {
			// Create initial formula with the actual SHA256
			if err := CreateInitialFormulaWithSHA(projectName, projectName, repoOwner, repoName, version, sha256sum, tapPath); err != nil {
				return HomebrewUpdateResult{
					Success: false,
					Error:   fmt.Errorf("creating initial formula: %w", err),
				}
			}
		}

		if err := updateFormulaFile(formulaPath, version, tarballURL, sha256sum); err != nil {
			return HomebrewUpdateResult{
				Success: false,
				Error:   fmt.Errorf("updating formula: %w", err),
			}
		}

		if err := commitAndPushFormula(ctx, tapPath, projectName, version); err != nil {
			return HomebrewUpdateResult{
				Success:     false,
				FormulaPath: formulaPath,
				TapPath:     tapPath,
				Error:       fmt.Errorf("committing changes: %w", err),
			}
		}

		return HomebrewUpdateResult{
			Success:     true,
			FormulaPath: formulaPath,
			TapPath:     tapPath,
		}
	}
}

func downloadAndCalculateSHA256(url string) (string, error) {
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		fmt.Printf("Download attempt %d/5...\n", attempt)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("downloading tarball: %w", err)
			if attempt < 5 {
				waitTime := time.Duration(attempt*3) * time.Second
				fmt.Printf("Download failed, waiting %v before retry...\n", waitTime)
				time.Sleep(waitTime)
				continue
			}
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("download failed with status: %d", resp.StatusCode)
			if attempt < 5 {
				waitTime := time.Duration(attempt*3) * time.Second
				fmt.Printf("Download failed with status %d, waiting %v before retry...\n", resp.StatusCode, waitTime)
				time.Sleep(waitTime)
				continue
			}
			break
		}

		hash := sha256.New()
		if _, err := io.Copy(hash, resp.Body); err != nil {
			lastErr = fmt.Errorf("calculating hash: %w", err)
			continue
		}

		return fmt.Sprintf("%x", hash.Sum(nil)), nil
	}

	return "", fmt.Errorf("failed after 5 attempts: %w", lastErr)
}

func updateFormulaFile(formulaPath string, version string, url string, sha256sum string) error {
	content, err := os.ReadFile(formulaPath)
	if err != nil {
		return fmt.Errorf("reading formula: %w", err)
	}

	formulaContent := string(content)

	versionPattern := regexp.MustCompile(`version\s+"[^"]*"`)
	formulaContent = versionPattern.ReplaceAllString(formulaContent, fmt.Sprintf(`version "%s"`, version))

	urlPattern := regexp.MustCompile(`url\s+"[^"]*"`)
	formulaContent = urlPattern.ReplaceAllString(formulaContent, fmt.Sprintf(`url "%s"`, url))

	sha256Pattern := regexp.MustCompile(`sha256\s+"[^"]*"`)
	formulaContent = sha256Pattern.ReplaceAllString(formulaContent, fmt.Sprintf(`sha256 "%s"`, sha256sum))

	if err := os.WriteFile(formulaPath, []byte(formulaContent), 0644); err != nil {
		return fmt.Errorf("writing formula: %w", err)
	}

	return nil
}

func commitAndPushFormula(ctx context.Context, tapPath string, projectName string, version string) error {
	// Get current branch
	currentBranch, err := RunCommandCapture(ctx, "git", []string{"branch", "--show-current"}, tapPath)
	if err != nil {
		currentBranch = "main" // default
	} else {
		currentBranch = strings.TrimSpace(currentBranch)
	}
	fmt.Printf("Current branch: %s\n", currentBranch)

	// Add the formula file
	formulaFile := fmt.Sprintf("Formula/%s.rb", projectName)
	_, err = RunCommandCapture(ctx, "git", []string{"add", formulaFile}, tapPath)
	if err != nil {
		fmt.Printf("Git add failed: %v\n", err)
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit with message
	commitMsg := fmt.Sprintf("Update %s to %s\n\nGenerated by distui", projectName, version)
	output, err := RunCommandCapture(ctx, "git", []string{"commit", "-m", commitMsg}, tapPath)
	if err != nil {
		// Check if there's nothing to commit
		if strings.Contains(output, "nothing to commit") || strings.Contains(err.Error(), "nothing to commit") {
			fmt.Println("No changes to commit - formula may already be up to date")
			return nil
		}
		fmt.Printf("Git commit failed: %v\n", err)
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Try to push to current branch first, then try main/master as fallback
	fmt.Printf("Pushing to origin/%s...\n", currentBranch)
	_, err = RunCommandCapture(ctx, "git", []string{"push", "origin", currentBranch}, tapPath)
	if err != nil {
		// Try main branch
		fmt.Println("Failed to push to current branch, trying main...")
		_, err = RunCommandCapture(ctx, "git", []string{"push", "origin", "main"}, tapPath)
		if err != nil {
			// Try master branch
			fmt.Println("Failed to push to main, trying master...")
			_, err = RunCommandCapture(ctx, "git", []string{"push", "origin", "master"}, tapPath)
			if err != nil {
				fmt.Printf("Git push failed: %v\n", err)
				return fmt.Errorf("git push failed: %w", err)
			}
		}
	}

	fmt.Println("Successfully pushed changes to GitHub")
	return nil
}

func CreateInitialFormula(projectName string, description string, repoOwner string, repoName string, version string, tapPath string) error {
	return CreateInitialFormulaWithSHA(projectName, description, repoOwner, repoName, version, "REPLACE_WITH_ACTUAL_SHA256", tapPath)
}

func CreateInitialFormulaWithSHA(projectName string, description string, repoOwner string, repoName string, version string, sha256sum string, tapPath string) error {
	formulaTemplate := fmt.Sprintf(`class %s < Formula
  desc "%s"
  homepage "https://github.com/%s/%s"
  version "%s"
  url "https://github.com/%s/%s/archive/refs/tags/%s.tar.gz"
  sha256 "%s"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"%s"
  end

  test do
    system "#{bin}/%s", "--version"
  end
end
`, strings.Title(projectName), description, repoOwner, repoName, version, repoOwner, repoName, version, sha256sum, projectName, projectName)

	formulaPath := filepath.Join(tapPath, "Formula", projectName+".rb")
	formulaDir := filepath.Dir(formulaPath)

	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		return fmt.Errorf("creating Formula directory: %w", err)
	}

	if err := os.WriteFile(formulaPath, []byte(formulaTemplate), 0644); err != nil {
		return fmt.Errorf("writing formula: %w", err)
	}

	return nil
}