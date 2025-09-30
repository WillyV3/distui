package executor

import (
	"fmt"
	"os/exec"
	"strings"
)

type NPMNameStatus string

const (
	NPMNameAvailable   NPMNameStatus = "available"
	NPMNameUnavailable NPMNameStatus = "unavailable"
	NPMNameChecking    NPMNameStatus = "checking"
	NPMNameError       NPMNameStatus = "error"
)

type NPMNameCheckResult struct {
	Name        string
	Status      NPMNameStatus
	Error       string
	Suggestions []string
}

func CheckNPMNameAvailable(packageName, username string) (bool, string, error) {
	if packageName == "" {
		return false, "", fmt.Errorf("package name cannot be empty")
	}

	// Get current NPM user
	whoamiCmd := exec.Command("npm", "whoami")
	whoamiOutput, whoamiErr := whoamiCmd.CombinedOutput()
	npmUsername := ""
	if whoamiErr == nil {
		npmUsername = strings.TrimSpace(string(whoamiOutput))
	}

	// Check exact match
	cmd := exec.Command("npm", "view", packageName, "maintainers", "--json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// E404 means exact package doesn't exist
		if strings.Contains(string(output), "E404") || strings.Contains(err.Error(), "404") {
			// Check for similar names by generating variations and checking each
			variations := generateNameVariations(packageName)
			for _, variation := range variations {
				checkCmd := exec.Command("npm", "view", variation, "name")
				checkOutput, checkErr := checkCmd.CombinedOutput()

				if checkErr == nil || !strings.Contains(string(checkOutput), "E404") {
					// Found a similar package that exists
					return false, "", fmt.Errorf("similar package exists: %s", variation)
				}
			}

			return true, "", nil
		}
		return false, "", fmt.Errorf("checking npm registry: %w", err)
	}

	// Package exists - check if current NPM user owns it
	outputStr := strings.ToLower(string(output))
	if npmUsername != "" && strings.Contains(outputStr, strings.ToLower(npmUsername)) {
		// User owns this package
		return true, npmUsername, nil
	}

	// Package exists but owned by someone else
	return false, "", nil
}

func generateNameVariations(packageName string) []string {
	variations := []string{}

	// NPM considers names too similar if they differ only by hyphens, underscores, or case
	// Generate common variations to check
	lower := strings.ToLower(packageName)

	// If name already has separators, swap them
	if strings.Contains(lower, "-") {
		variations = append(variations, strings.ReplaceAll(lower, "-", "_"))
		variations = append(variations, strings.ReplaceAll(lower, "-", ""))
	}
	if strings.Contains(lower, "_") {
		variations = append(variations, strings.ReplaceAll(lower, "_", "-"))
		variations = append(variations, strings.ReplaceAll(lower, "_", ""))
	}

	// Check variations with hyphens between all word boundaries
	// For "distui" -> "dist-ui", "di-stui", etc.
	chars := []rune(lower)
	for i := 1; i < len(chars); i++ {
		variation := string(chars[:i]) + "-" + string(chars[i:])
		variations = append(variations, variation)
	}

	// Check with underscores
	for i := 1; i < len(chars); i++ {
		variation := string(chars[:i]) + "_" + string(chars[i:])
		variations = append(variations, variation)
	}

	// Check uppercase/title case variations
	if lower != packageName {
		variations = append(variations, lower)
	}
	variations = append(variations, strings.ToUpper(packageName))

	return variations
}

func GenerateNPMNameSuggestions(packageName, username string) []string {
	suggestions := []string{}

	if username != "" {
		suggestions = append(suggestions, fmt.Sprintf("@%s/%s", username, packageName))
	}

	suggestions = append(suggestions,
		packageName+"-cli",
		packageName+"-tool",
		packageName+"-release",
		packageName+"-dist",
	)

	return suggestions
}

func CheckNPMName(packageName, username string) NPMNameCheckResult {
	result := NPMNameCheckResult{
		Name:   packageName,
		Status: NPMNameChecking,
	}

	available, owner, err := CheckNPMNameAvailable(packageName, username)
	if err != nil {
		// Check if this is a similarity conflict (not a real error)
		if strings.Contains(err.Error(), "similar package exists:") {
			result.Status = NPMNameUnavailable
			result.Error = err.Error()
			result.Suggestions = GenerateNPMNameSuggestions(packageName, username)
			return result
		}
		// Real error (network, npm not installed, etc)
		result.Status = NPMNameError
		result.Error = err.Error()
		return result
	}

	if available {
		if owner != "" {
			// Package exists and user owns it
			result.Status = NPMNameAvailable
			result.Error = fmt.Sprintf("You own this package (%s)", owner)
		} else {
			// Package doesn't exist - truly available
			result.Status = NPMNameAvailable
		}
		return result
	}

	// Package exists but owned by someone else
	result.Status = NPMNameUnavailable
	result.Suggestions = GenerateNPMNameSuggestions(packageName, username)
	return result
}