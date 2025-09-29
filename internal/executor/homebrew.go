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

	tea "github.com/charmbracelet/bubbletea"
)

type HomebrewUpdateResult struct {
	Success      bool
	FormulaPath  string
	TapPath      string
	CommitHash   string
	Error        error
}

func UpdateHomebrewTap(ctx context.Context, projectName string, version string, tapPath string, repoOwner string, repoName string) tea.Cmd {
	return func() tea.Msg {
		tarballURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", repoOwner, repoName, version)

		sha256sum, err := downloadAndCalculateSHA256(tarballURL)
		if err != nil {
			return HomebrewUpdateResult{
				Success: false,
				Error:   fmt.Errorf("calculating SHA256: %w", err),
			}
		}

		formulaPath := filepath.Join(tapPath, "Formula", projectName+".rb")

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
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("downloading tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, resp.Body); err != nil {
		return "", fmt.Errorf("calculating hash: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
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
	commitMsg := fmt.Sprintf("Update %s to %s", projectName, version)

	cmdAdd := RunCommandStreaming(ctx, "git", []string{"add", "."}, tapPath)
	if msg := cmdAdd(); msg != nil {
		return fmt.Errorf("git add failed")
	}

	cmdCommit := RunCommandStreaming(ctx, "git", []string{"commit", "-m", commitMsg}, tapPath)
	if msg := cmdCommit(); msg != nil {
		return fmt.Errorf("git commit failed")
	}

	cmdPush := RunCommandStreaming(ctx, "git", []string{"push"}, tapPath)
	if msg := cmdPush(); msg != nil {
		return fmt.Errorf("git push failed")
	}

	return nil
}

func CreateInitialFormula(projectName string, description string, repoOwner string, repoName string, version string, tapPath string) error {
	formulaTemplate := fmt.Sprintf(`class %s < Formula
  desc "%s"
  homepage "https://github.com/%s/%s"
  version "%s"
  url "https://github.com/%s/%s/archive/refs/tags/#{version}.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"%s"
  end

  test do
    system "#{bin}/%s", "--version"
  end
end
`, strings.Title(projectName), description, repoOwner, repoName, version, repoOwner, repoName, projectName, projectName)

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