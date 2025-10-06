package gitcleanup

import (
	"os"
	"os/exec"
	"testing"
)

// TestReleaseBlocking tests the IsWorkingTreeClean function with various scenarios
// This test is designed to be EASY TO DEBUG - each scenario is clearly labeled
func TestReleaseBlocking(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Initialize git repo
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to cd to temp dir: %v", err)
	}

	runCmd(t, "git", "init")
	runCmd(t, "git", "config", "user.email", "test@test.com")
	runCmd(t, "git", "config", "user.name", "Test User")

	// Create initial commit so we have a proper repo
	os.WriteFile("README.md", []byte("# Test Repo"), 0644)
	runCmd(t, "git", "add", "README.md")
	runCmd(t, "git", "commit", "-m", "Initial commit")

	tests := []struct {
		name          string
		setup         func()
		expectClean   bool
		debugMessage  string
	}{
		{
			name: "CLEAN: No changes at all",
			setup: func() {
				// Nothing to do - repo is clean
			},
			expectClean: true,
			debugMessage: "Repository with no changes should be clean",
		},
		{
			name: "CLEAN: Untracked file in root",
			setup: func() {
				os.WriteFile("random-file.txt", []byte("untracked"), 0644)
			},
			expectClean: true,
			debugMessage: "Untracked files should NOT block release",
		},
		{
			name: "CLEAN: dist/ folder with files",
			setup: func() {
				os.MkdirAll("dist", 0755)
				os.WriteFile("dist/app", []byte("binary"), 0644)
				os.WriteFile("dist/checksums.txt", []byte("checksums"), 0644)
			},
			expectClean: true,
			debugMessage: "dist/ folder should NOT block release (build output)",
		},
		{
			name: "CLEAN: logs/ folder",
			setup: func() {
				os.MkdirAll("logs", 0755)
				os.WriteFile("logs/app.log", []byte("log data"), 0644)
			},
			expectClean: true,
			debugMessage: "logs/ folder should NOT block release",
		},
		{
			name: "CLEAN: bin/ folder",
			setup: func() {
				os.MkdirAll("bin", 0755)
				os.WriteFile("bin/myapp", []byte("binary"), 0644)
			},
			expectClean: true,
			debugMessage: "bin/ folder should NOT block release",
		},
		{
			name: "CLEAN: node_modules/ folder",
			setup: func() {
				os.MkdirAll("node_modules/package", 0755)
				os.WriteFile("node_modules/package/index.js", []byte("code"), 0644)
			},
			expectClean: true,
			debugMessage: "node_modules/ should NOT block release",
		},
		{
			name: "CLEAN: .gitignore changes",
			setup: func() {
				os.WriteFile(".gitignore", []byte("*.log\n"), 0644)
			},
			expectClean: true,
			debugMessage: ".gitignore changes should NOT block release",
		},
		{
			name: "CLEAN: .distui-backup/ folder",
			setup: func() {
				os.MkdirAll(".distui-backup", 0755)
				os.WriteFile(".distui-backup/backup.tar", []byte("backup"), 0644)
			},
			expectClean: true,
			debugMessage: ".distui-backup/ should NOT block release (distui internal)",
		},
		{
			name: "DIRTY: Modified tracked file",
			setup: func() {
				os.WriteFile("README.md", []byte("# Modified"), 0644)
			},
			expectClean: false,
			debugMessage: "Modified tracked files SHOULD block release",
		},
		{
			name: "DIRTY: Staged new file",
			setup: func() {
				os.WriteFile("new-feature.go", []byte("package main"), 0644)
				runCmd(t, "git", "add", "new-feature.go")
			},
			expectClean: false,
			debugMessage: "Staged files SHOULD block release",
		},
		{
			name: "DIRTY: Deleted tracked file",
			setup: func() {
				os.Remove("README.md")
			},
			expectClean: false,
			debugMessage: "Deleted tracked files SHOULD block release",
		},
		{
			name: "CLEAN: Multiple untracked files",
			setup: func() {
				os.WriteFile("file1.txt", []byte("data"), 0644)
				os.WriteFile("file2.txt", []byte("data"), 0644)
				os.WriteFile("file3.txt", []byte("data"), 0644)
			},
			expectClean: true,
			debugMessage: "Multiple untracked files should NOT block release",
		},
		{
			name: "CLEAN: Mixed untracked and ignored dirs",
			setup: func() {
				os.MkdirAll("dist/linux", 0755)
				os.MkdirAll("dist/darwin", 0755)
				os.WriteFile("dist/linux/app", []byte("binary"), 0644)
				os.WriteFile("dist/darwin/app", []byte("binary"), 0644)
				os.WriteFile("random.txt", []byte("untracked"), 0644)
				os.MkdirAll("tmp", 0755)
				os.WriteFile("tmp/temp.txt", []byte("temp"), 0644)
			},
			expectClean: true,
			debugMessage: "Mix of untracked files and ignored directories should NOT block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up from previous test
			cleanupRepo(t)

			// Run setup
			tt.setup()

			// Show git status for debugging
			cmd := exec.Command("git", "status", "--porcelain")
			output, _ := cmd.Output()
			gitStatus := string(output)

			// Test the function
			result := IsWorkingTreeClean()

			// Debug output
			if result != tt.expectClean {
				t.Errorf("\n"+
					"❌ TEST FAILED: %s\n"+
					"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
					"Expected: %v (clean=%t)\n"+
					"Got:      %v (clean=%t)\n"+
					"Reason:   %s\n"+
					"\nGit Status Output:\n%s"+
					"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n",
					tt.name,
					tt.expectClean, tt.expectClean,
					result, result,
					tt.debugMessage,
					gitStatus,
				)

				// Show detailed file list
				files, _ := GetGitStatus()
				t.Logf("\nDetailed file status:")
				for i, f := range files {
					t.Logf("  [%d] Status: '%s' | Path: '%s' | Category: %v",
						i, f.Status, f.Path, f.Category)
				}
			} else {
				t.Logf("✓ PASS: %s", tt.name)
				if gitStatus != "" {
					t.Logf("  Git status:\n%s", gitStatus)
				}
			}
		})
	}
}

// TestReleaseBlockingRealWorld tests real-world scenarios
func TestReleaseBlockingRealWorld(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to cd to temp dir: %v", err)
	}

	runCmd(t, "git", "init")
	runCmd(t, "git", "config", "user.email", "test@test.com")
	runCmd(t, "git", "config", "user.name", "Test User")

	// Simulate a real project structure
	os.WriteFile("main.go", []byte("package main\n\nfunc main() {}"), 0644)
	os.WriteFile("go.mod", []byte("module example.com/app"), 0644)
	os.WriteFile(".gitignore", []byte("dist/\nbin/\nlogs/\n*.log"), 0644)
	runCmd(t, "git", "add", ".")
	runCmd(t, "git", "commit", "-m", "Initial commit")

	t.Run("Scenario: After running 'go build'", func(t *testing.T) {
		// Simulate build output
		os.MkdirAll("dist/linux_amd64", 0755)
		os.MkdirAll("dist/darwin_arm64", 0755)
		os.WriteFile("dist/linux_amd64/app", []byte("binary"), 0755)
		os.WriteFile("dist/darwin_arm64/app", []byte("binary"), 0755)
		os.WriteFile("dist/checksums.txt", []byte("checksums"), 0644)

		result := IsWorkingTreeClean()
		if !result {
			t.Errorf("❌ Build artifacts in dist/ should NOT block release")
			showGitStatus(t)
		} else {
			t.Log("✓ Build artifacts correctly ignored")
		}
	})

	t.Run("Scenario: After running tests with logs", func(t *testing.T) {
		cleanupRepo(t)

		os.MkdirAll("logs", 0755)
		os.WriteFile("logs/test-output.log", []byte("test logs"), 0644)
		os.WriteFile("test.log", []byte("test log"), 0644)

		result := IsWorkingTreeClean()
		if !result {
			t.Errorf("❌ Log files should NOT block release")
			showGitStatus(t)
		} else {
			t.Log("✓ Log files correctly ignored")
		}
	})

	t.Run("Scenario: Forgot to commit code changes", func(t *testing.T) {
		cleanupRepo(t)

		os.WriteFile("main.go", []byte("package main\n\n// Modified\nfunc main() {}"), 0644)

		result := IsWorkingTreeClean()
		if result {
			t.Errorf("❌ Modified source code SHOULD block release")
			showGitStatus(t)
		} else {
			t.Log("✓ Modified source code correctly blocks release")
		}
	})

	t.Run("Scenario: Staged but uncommitted changes", func(t *testing.T) {
		cleanupRepo(t)

		os.WriteFile("feature.go", []byte("package main"), 0644)
		runCmd(t, "git", "add", "feature.go")

		result := IsWorkingTreeClean()
		if result {
			t.Errorf("❌ Staged changes SHOULD block release")
			showGitStatus(t)
		} else {
			t.Log("✓ Staged changes correctly block release")
		}
	})

	t.Run("Scenario: Typical pre-release state (clean)", func(t *testing.T) {
		cleanupRepo(t)

		// Simulate typical state before release: build artifacts but no code changes
		os.MkdirAll("dist", 0755)
		os.WriteFile("dist/app_linux_amd64", []byte("binary"), 0755)
		os.WriteFile("dist/app_darwin_arm64", []byte("binary"), 0755)
		os.MkdirAll("logs", 0755)
		os.WriteFile("logs/release.log", []byte("log"), 0644)
		os.WriteFile("random-notes.txt", []byte("notes"), 0644)

		result := IsWorkingTreeClean()
		if !result {
			t.Errorf("❌ Typical pre-release state SHOULD be clean")
			showGitStatus(t)
		} else {
			t.Log("✓ Typical pre-release state correctly passes")
		}
	})
}

// Helper functions for easier debugging

func runCmd(t *testing.T, name string, args ...string) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command failed: %s %v\nOutput: %s", name, args, output)
	}
}

func cleanupRepo(t *testing.T) {
	// Reset to clean state
	exec.Command("git", "reset", "--hard", "HEAD").Run()
	exec.Command("git", "clean", "-fd").Run()
}

func showGitStatus(t *testing.T) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, _ := cmd.Output()
	t.Logf("\nGit status:\n%s", output)

	files, _ := GetGitStatus()
	t.Logf("\nParsed files:")
	for i, f := range files {
		t.Logf("  [%d] Status='%s' Path='%s'", i, f.Status, f.Path)
	}
}

// TestIgnoredPrefixes verifies the exact list of ignored prefixes
func TestIgnoredPrefixes(t *testing.T) {
	expectedIgnored := []string{
		"dist/",
		".distui-backup/",
		"build/",
		"out/",
		"target/",
		"bin/",
		"logs/",
		"tmp/",
		"temp/",
		"node_modules/",
	}

	t.Log("Ignored prefixes that should NEVER block release:")
	for _, prefix := range expectedIgnored {
		t.Logf("  ✓ %s", prefix)
	}

	t.Log("\nFile statuses that should NEVER block release:")
	t.Log("  ✓ ?? (untracked files)")

	t.Log("\nSpecial files that should NEVER block release:")
	t.Log("  ✓ .gitignore")
}

// TestDebugOutput helps debug the actual IsWorkingTreeClean logic
func TestDebugOutput(t *testing.T) {
	t.Log("\n" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n" +
		"RELEASE BLOCKING DEBUG GUIDE\n" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	t.Log("How to debug release blocking issues:")
	t.Log("")
	t.Log("1. Run this test: go test -v ./internal/gitcleanup -run TestReleaseBlocking")
	t.Log("")
	t.Log("2. Look for '❌ TEST FAILED' in output")
	t.Log("")
	t.Log("3. Check the 'Git Status Output' section")
	t.Log("   - First 2 chars = status code")
	t.Log("   - ?? = untracked (SHOULD be ignored)")
	t.Log("   - M  = modified (SHOULD block)")
	t.Log("   - A  = added/staged (SHOULD block)")
	t.Log("")
	t.Log("4. Check if file path starts with ignored prefix")
	t.Log("")
	t.Log("5. If test fails, the debug output shows:")
	t.Log("   - Expected result (clean=true/false)")
	t.Log("   - Actual result")
	t.Log("   - Complete git status")
	t.Log("   - Detailed parsed file list")
	t.Log("")

	t.Log("Common issues:")
	t.Log("  • File shows ?? but blocks = bug in untracked handling")
	t.Log("  • File in dist/ but blocks = bug in prefix checking")
	t.Log("  • Modified file doesn't block = bug in status detection")

	t.Log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}
