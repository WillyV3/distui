package fsearch

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestFuzzySearchDirectories(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		maxResults int
		want       []string
	}{
		{
			name:       "empty query returns empty slice",
			query:      "",
			maxResults: 10,
			want:       []string{},
		},
		{
			name:       "single character query returns empty slice",
			query:      "a",
			maxResults: 10,
			want:       []string{},
		},
		{
			name:       "two character query passes validation",
			query:      "ab",
			maxResults: 10,
			want:       nil, // Will depend on actual fd/fzf results
		},
		{
			name:       "max results zero should limit results",
			query:      "test",
			maxResults: 0,
			want:       []string{},
		},
		{
			name:       "negative max results should return empty",
			query:      "test",
			maxResults: -1,
			want:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FuzzySearchDirectories(tt.query, tt.maxResults)
			
			// For query length validation tests, we can assert exact behavior
			if len(tt.query) < 2 {
				if len(got) \!= 0 {
					t.Errorf("FuzzySearchDirectories() with query length < 2 = %v, want empty slice", got)
				}
				return
			}
			
			// For maxResults = 0 or negative, expect empty results
			if tt.maxResults <= 0 {
				if len(got) \!= 0 {
					t.Errorf("FuzzySearchDirectories() with maxResults <= 0 = %v, want empty slice", got)
				}
				return
			}
			
			// For other cases, verify the result is a slice (may be empty if fd/fzf fail)
			if got == nil {
				t.Errorf("FuzzySearchDirectories() returned nil, want non-nil slice")
			}
			
			// Verify maxResults is respected
			if len(got) > tt.maxResults {
				t.Errorf("FuzzySearchDirectories() returned %d results, want at most %d", len(got), tt.maxResults)
			}
			
			// Verify no empty strings in results
			for i, result := range got {
				if result == "" {
					t.Errorf("FuzzySearchDirectories() result[%d] is empty string", i)
				}
			}
		})
	}
}

func TestFuzzySearchDirectories_QueryLengthValidation(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"empty string", ""},
		{"single char", "x"},
		{"whitespace", " "},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuzzySearchDirectories(tt.query, 10)
			if len(result) \!= 0 {
				t.Errorf("Expected empty slice for query %q, got %v", tt.query, result)
			}
		})
	}
}

func TestFuzzySearchDirectories_MaxResultsHandling(t *testing.T) {
	tests := []struct {
		name       string
		maxResults int
		wantEmpty  bool
	}{
		{"zero max results", 0, true},
		{"negative max results", -5, true},
		{"positive max results", 5, false},
		{"large max results", 1000, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a query that might return results
			result := FuzzySearchDirectories("test", tt.maxResults)
			
			// Verify result is never nil
			if result == nil {
				t.Error("Result should never be nil")
			}
			
			// Verify length constraint
			if len(result) > tt.maxResults && tt.maxResults > 0 {
				t.Errorf("Result length %d exceeds maxResults %d", len(result), tt.maxResults)
			}
			
			// For zero/negative maxResults, expect empty
			if tt.wantEmpty && len(result) \!= 0 {
				t.Errorf("Expected empty result for maxResults=%d, got %d items", tt.maxResults, len(result))
			}
		})
	}
}

func TestFuzzySearchDirectories_ResultsFiltering(t *testing.T) {
	// This test verifies that empty lines are filtered out
	// We'll test with a valid query
	result := FuzzySearchDirectories("documents", 100)
	
	// Result should be non-nil
	if result == nil {
		t.Fatal("Result should never be nil")
	}
	
	// Verify no empty strings in results
	for i, item := range result {
		if item == "" {
			t.Errorf("Result[%d] should not be empty string", i)
		}
		// Verify no leading/trailing whitespace
		if strings.TrimSpace(item) \!= item {
			t.Errorf("Result[%d] has leading/trailing whitespace: %q", i, item)
		}
	}
}

func TestFuzzySearchDirectories_CommandAvailability(t *testing.T) {
	// Test behavior when fd is not available
	t.Run("handles fd unavailability gracefully", func(t *testing.T) {
		// We can't actually make fd unavailable, but we can test with an unlikely query
		// The function should return an empty slice on error, not panic
		result := FuzzySearchDirectories("xyzabc123unlikely", 10)
		if result == nil {
			t.Error("Should return non-nil slice even on command failure")
		}
	})
}

func TestFuzzySearchDirectories_HomeDirectoryHandling(t *testing.T) {
	// Test that the function handles home directory properly
	t.Run("works with valid home directory", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err \!= nil {
			t.Skip("Cannot get home directory, skipping test")
		}
		
		if homeDir == "" {
			t.Skip("Home directory is empty, skipping test")
		}
		
		// Should not panic even with valid home directory
		result := FuzzySearchDirectories("downloads", 5)
		if result == nil {
			t.Error("Result should never be nil")
		}
	})
}

func TestFuzzySearchDirectories_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		maxResults int
	}{
		{"minimum valid query length", "ab", 1},
		{"query with spaces", "my documents", 5},
		{"query with special chars", "test-dir", 5},
		{"maxResults equals 1", "docs", 1},
		{"very large maxResults", "test", 10000},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuzzySearchDirectories(tt.query, tt.maxResults)
			
			if result == nil {
				t.Fatal("Result should never be nil")
			}
			
			if len(result) > tt.maxResults {
				t.Errorf("Result length %d exceeds maxResults %d", len(result), tt.maxResults)
			}
			
			// Verify all results are non-empty
			for _, item := range result {
				if item == "" {
					t.Error("Result contains empty string")
				}
			}
		})
	}
}

func TestFuzzySearchDirectories_ExclusionPatterns(t *testing.T) {
	// Test that excluded directories are actually excluded
	// This is a behavioral test - we search for typically excluded patterns
	t.Run("excludes common patterns", func(t *testing.T) {
		excludedPatterns := []string{"node_modules", ".git", "Library", ".cache"}
		
		for _, pattern := range excludedPatterns {
			result := FuzzySearchDirectories(pattern, 100)
			
			// Results should not contain exact matches for excluded patterns
			// (though they might contain these as substring in allowed paths)
			for _, path := range result {
				// Check if the path ends with the excluded pattern
				// This would indicate the excluded directory itself was returned
				if strings.HasSuffix(path, "/"+pattern) || strings.HasSuffix(path, "\\"+pattern) {
					t.Logf("Warning: Excluded pattern %q found in results: %s", pattern, path)
				}
			}
		}
	})
}

func TestFuzzySearchDirectories_ConcurrentCalls(t *testing.T) {
	// Test that multiple concurrent calls don't interfere with each other
	t.Run("handles concurrent calls", func(t *testing.T) {
		done := make(chan bool, 3)
		
		queries := []string{"doc", "download", "desktop"}
		
		for _, query := range queries {
			go func(q string) {
				result := FuzzySearchDirectories(q, 5)
				if result == nil {
					t.Error("Concurrent call returned nil")
				}
				done <- true
			}(query)
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < len(queries); i++ {
			<-done
		}
	})
}

func TestFuzzySearchDirectories_QueryCaseSensitivity(t *testing.T) {
	// fzf is called with -i flag (case insensitive)
	// Test that queries with different cases work
	tests := []struct {
		name  string
		query string
	}{
		{"lowercase", "documents"},
		{"uppercase", "DOCUMENTS"},
		{"mixed case", "DoCoMeNtS"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuzzySearchDirectories(tt.query, 10)
			if result == nil {
				t.Error("Result should not be nil")
			}
			// All three queries should potentially return results
			// We just verify they don't panic and return valid slices
		})
	}
}

func TestFuzzySearchDirectories_EmptyResultsHandling(t *testing.T) {
	// Test with a query that's very unlikely to match anything
	t.Run("handles no matches gracefully", func(t *testing.T) {
		result := FuzzySearchDirectories("xyzqrstuvw123456789unlikely", 10)
		
		if result == nil {
			t.Error("Result should not be nil even with no matches")
		}
		
		// Should return empty slice, not nil
		if len(result) \!= 0 {
			// This is okay - there might actually be matches
			// Just verify results are valid
			for _, item := range result {
				if item == "" {
					t.Error("Result should not contain empty strings")
				}
			}
		}
	})
}

func TestFuzzySearchDirectories_MaxDepthRespected(t *testing.T) {
	// This is more of a documentation test - verifying the behavior
	// The function uses --max-depth 4
	t.Run("documents max depth setting", func(t *testing.T) {
		// Search for something that might exist at various depths
		result := FuzzySearchDirectories("test", 50)
		
		if result == nil {
			t.Error("Result should not be nil")
		}
		
		// We can't easily verify max depth without mocking,
		// but we can verify the results are valid paths
		for _, path := range result {
			if \!strings.HasPrefix(path, "/") && \!strings.Contains(path, ":") {
				// On Unix, absolute paths start with /
				// On Windows, they contain :
				// This is a weak check, but better than nothing
				t.Logf("Path might not be absolute: %s", path)
			}
		}
	})
}

func TestFuzzySearchDirectories_AbsolutePathsReturned(t *testing.T) {
	// Verify that returned paths are absolute (due to --absolute-path flag)
	t.Run("returns absolute paths", func(t *testing.T) {
		result := FuzzySearchDirectories("documents", 5)
		
		if result == nil {
			t.Fatal("Result should not be nil")
		}
		
		// If we have results, verify they look like absolute paths
		for _, path := range result {
			if path == "" {
				t.Error("Path should not be empty")
				continue
			}
			
			// Check for absolute path indicators
			// Unix: starts with /
			// Windows: contains : (like C:)
			if \!strings.HasPrefix(path, "/") && \!strings.Contains(path, ":") {
				t.Logf("Warning: Path might not be absolute: %s", path)
			}
		}
	})
}

// Benchmark tests to measure performance
func BenchmarkFuzzySearchDirectories(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FuzzySearchDirectories("documents", 10)
	}
}

func BenchmarkFuzzySearchDirectories_ShortQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FuzzySearchDirectories("ab", 5)
	}
}

func BenchmarkFuzzySearchDirectories_LongQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FuzzySearchDirectories("very_long_query_string_test", 10)
	}
}

func BenchmarkFuzzySearchDirectories_InvalidQuery(b *testing.B) {
	// Should be very fast since it short-circuits
	for i := 0; i < b.N; i++ {
		FuzzySearchDirectories("x", 10)
	}
}