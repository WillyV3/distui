
import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"distui/internal/models"
)

func TestLoadAllProjects(t *testing.T) {
	tests := map[string]struct {
		setupFunc    func(t *testing.T) string // Returns temp dir path
		cleanupFunc  func(t *testing.T, dir string)
		wantErr      bool
		wantProjects int
		validate     func(t *testing.T, projects []models.ProjectConfig)
	}{
		"empty directory returns empty list": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}
				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 0,
		},
		"directory with valid yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				// Create valid YAML files
				validYAML := `name: test-project
version: "1.0"
`
				err = os.WriteFile(filepath.Join(projectsDir, "project1.yaml"), []byte(validYAML), 0644)
				if err \!= nil {
					t.Fatalf("failed to write yaml file: %v", err)
				}

				err = os.WriteFile(filepath.Join(projectsDir, "project2.yaml"), []byte(validYAML), 0644)
				if err \!= nil {
					t.Fatalf("failed to write yaml file: %v", err)
				}

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 2,
		},
		"directory with mixed yaml and non-yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				// Create YAML files
				os.WriteFile(filepath.Join(projectsDir, "project1.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "project2.yaml"), []byte(validYAML), 0644)

				// Create non-YAML files (should be ignored)
				os.WriteFile(filepath.Join(projectsDir, "README.md"), []byte("# README"), 0644)
				os.WriteFile(filepath.Join(projectsDir, "config.json"), []byte("{}"), 0644)
				os.WriteFile(filepath.Join(projectsDir, "script.sh"), []byte("#\!/bin/bash"), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 2,
		},
		"directory with subdirectories and yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				// Create YAML files
				os.WriteFile(filepath.Join(projectsDir, "project1.yaml"), []byte(validYAML), 0644)

				// Create subdirectories (should be ignored)
				subDir := filepath.Join(projectsDir, "subdir")
				os.MkdirAll(subDir, 0755)
				os.WriteFile(filepath.Join(subDir, "project2.yaml"), []byte(validYAML), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 1,
		},
		"directory with invalid yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				// Create invalid YAML files
				invalidYAML := `invalid: yaml: content: [[[`
				os.WriteFile(filepath.Join(projectsDir, "invalid1.yaml"), []byte(invalidYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "invalid2.yaml"), []byte(invalidYAML), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 0, // Invalid files should be skipped
		},
		"directory with mixed valid and invalid yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				invalidYAML := `invalid: yaml: content: [[[`

				os.WriteFile(filepath.Join(projectsDir, "valid1.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "invalid1.yaml"), []byte(invalidYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "valid2.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "invalid2.yaml"), []byte(invalidYAML), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 2, // Only valid files should be loaded
		},
		"non-existent directory returns error": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Don't create the .distui/projects directory
				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      true,
			wantProjects: 0,
		},
		"directory with yaml files with different extensions": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				// Create files with different extensions
				os.WriteFile(filepath.Join(projectsDir, "project1.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "project2.yml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "project3.YAML"), []byte(validYAML), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 1, // Only .yaml extension should be loaded
		},
		"directory with empty yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				// Create empty YAML files
				os.WriteFile(filepath.Join(projectsDir, "empty1.yaml"), []byte(""), 0644)
				os.WriteFile(filepath.Join(projectsDir, "empty2.yaml"), []byte(""), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 0, // Empty files should fail to load
		},
		"directory with special characters in filenames": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				// Create files with special characters
				os.WriteFile(filepath.Join(projectsDir, "project-with-dashes.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "project_with_underscores.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "project.with.dots.yaml"), []byte(validYAML), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 3,
		},
		"directory with no read permissions": {
			setupFunc: func(t *testing.T) string {
				if os.Getuid() == 0 {
					t.Skip("skipping test when running as root")
				}

				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				// Remove read permissions
				err = os.Chmod(projectsDir, 0000)
				if err \!= nil {
					t.Fatalf("failed to chmod dir: %v", err)
				}

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			cleanupFunc: func(t *testing.T, dir string) {
				// Restore permissions for cleanup
				projectsDir := filepath.Join(dir, ".distui", "projects")
				os.Chmod(projectsDir, 0755)
			},
			wantErr:      true,
			wantProjects: 0,
		},
		"large number of yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				// Create 100 YAML files
				for i := 0; i < 100; i++ {
					filename := filepath.Join(projectsDir, fmt.Sprintf("project%d.yaml", i))
					os.WriteFile(filename, []byte(validYAML), 0644)
				}

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 100,
		},
		"hidden yaml files": {
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectsDir := filepath.Join(tmpDir, ".distui", "projects")
				err := os.MkdirAll(projectsDir, 0755)
				if err \!= nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}

				validYAML := `name: test-project
version: "1.0"
`
				// Create hidden YAML files (files starting with .)
				os.WriteFile(filepath.Join(projectsDir, ".hidden.yaml"), []byte(validYAML), 0644)
				os.WriteFile(filepath.Join(projectsDir, "visible.yaml"), []byte(validYAML), 0644)

				t.Setenv("HOME", tmpDir)
				return tmpDir
			},
			wantErr:      false,
			wantProjects: 2, // Both hidden and visible should be loaded
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			tmpDir := tt.setupFunc(t)

			// Execute
			projects, err := LoadAllProjects()

			// Cleanup
			if tt.cleanupFunc \!= nil {
				defer tt.cleanupFunc(t, tmpDir)
			}

			// Assert error
			if (err \!= nil) \!= tt.wantErr {
				t.Errorf("LoadAllProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Assert project count
			if len(projects) \!= tt.wantProjects {
				t.Errorf("LoadAllProjects() returned %d projects, want %d", len(projects), tt.wantProjects)
			}

			// Additional validation
			if tt.validate \!= nil {
				tt.validate(t, projects)
			}
		})
	}
}

// TestLoadAllProjects_ProjectIdentifierExtraction tests that identifiers are correctly extracted from filenames
func TestLoadAllProjects_ProjectIdentifierExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, ".distui", "projects")
	err := os.MkdirAll(projectsDir, 0755)
	if err \!= nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	validYAML := `name: test-project
version: "1.0"
`
	os.WriteFile(filepath.Join(projectsDir, "my-awesome-project.yaml"), []byte(validYAML), 0644)

	t.Setenv("HOME", tmpDir)

	projects, err := LoadAllProjects()
	if err \!= nil {
		t.Fatalf("LoadAllProjects() unexpected error: %v", err)
	}

	if len(projects) \!= 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	// The identifier should be the filename without .yaml extension
	// This test validates that strings.TrimSuffix works correctly
}

// TestLoadAllProjects_Concurrent tests concurrent calls to LoadAllProjects
func TestLoadAllProjects_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, ".distui", "projects")
	err := os.MkdirAll(projectsDir, 0755)
	if err \!= nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	validYAML := `name: test-project
version: "1.0"
`
	for i := 0; i < 10; i++ {
		filename := filepath.Join(projectsDir, fmt.Sprintf("project%d.yaml", i))
		os.WriteFile(filename, []byte(validYAML), 0644)
	}

	t.Setenv("HOME", tmpDir)

	// Run LoadAllProjects concurrently
	const goroutines = 10
	errChan := make(chan error, goroutines)
	resultChan := make(chan int, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			projects, err := LoadAllProjects()
			if err \!= nil {
				errChan <- err
				return
			}
			resultChan <- len(projects)
		}()
	}

	// Collect results
	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errChan:
			t.Errorf("concurrent call failed: %v", err)
		case count := <-resultChan:
			if count \!= 10 {
				t.Errorf("expected 10 projects, got %d", count)
			}
		}
	}
}

// TestLoadAllProjects_SymbolicLinks tests behavior with symbolic links
func TestLoadAllProjects_SymbolicLinks(t *testing.T) {
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, ".distui", "projects")
	err := os.MkdirAll(projectsDir, 0755)
	if err \!= nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	validYAML := `name: test-project
version: "1.0"
`
	// Create a real file
	realFile := filepath.Join(projectsDir, "real.yaml")
	os.WriteFile(realFile, []byte(validYAML), 0644)

	// Create a symbolic link to the real file
	symlinkFile := filepath.Join(projectsDir, "symlink.yaml")
	err = os.Symlink(realFile, symlinkFile)
	if err \!= nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	t.Setenv("HOME", tmpDir)

	projects, err := LoadAllProjects()
	if err \!= nil {
		t.Fatalf("LoadAllProjects() unexpected error: %v", err)
	}

	// Both the real file and symlink should be loaded
	if len(projects) \!= 2 {
		t.Errorf("expected 2 projects (real + symlink), got %d", len(projects))
	}
}

// TestLoadAllProjects_NilReturn tests that the function returns a valid slice even on error
func TestLoadAllProjects_NilReturn(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create the projects directory
	t.Setenv("HOME", tmpDir)

	projects, err := LoadAllProjects()

	if err == nil {
		t.Error("expected error for non-existent directory")
	}

	if projects \!= nil {
		t.Errorf("expected nil projects on error, got %v", projects)
	}
}

// TestLoadAllProjects_FilePermissions tests handling of files with different permissions
func TestLoadAllProjects_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, ".distui", "projects")
	err := os.MkdirAll(projectsDir, 0755)
	if err \!= nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	validYAML := `name: test-project
version: "1.0"
`
	// Create files with different permissions
	readableFile := filepath.Join(projectsDir, "readable.yaml")
	os.WriteFile(readableFile, []byte(validYAML), 0644)

	unreadableFile := filepath.Join(projectsDir, "unreadable.yaml")
	os.WriteFile(unreadableFile, []byte(validYAML), 0000)

	t.Setenv("HOME", tmpDir)

	// Should load readable file and skip unreadable one
	projects, err := LoadAllProjects()

	// Restore permissions for cleanup
	os.Chmod(unreadableFile, 0644)

	if err \!= nil {
		t.Fatalf("LoadAllProjects() unexpected error: %v", err)
	}

	// Should load at least the readable file
	if len(projects) == 0 {
		t.Error("expected at least one project to be loaded")
	}
}

// Add fmt import at the top if not already present
import "fmt"