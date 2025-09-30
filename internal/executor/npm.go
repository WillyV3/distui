package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type NPMPublisher struct {
	projectPath string
	version     string
	packageName string
	debug       bool
}

func NewNPMPublisher(projectPath, version, packageName string) *NPMPublisher {
	return &NPMPublisher{
		projectPath: projectPath,
		version:     version,
		packageName: packageName,
		debug:       os.Getenv("DISTUI_NPM_DEBUG") != "",
	}
}

func (n *NPMPublisher) debugLog(format string, args ...interface{}) {
	if n.debug {
		fmt.Printf("[NPM DEBUG] "+format+"\n", args...)
	}
}

func (n *NPMPublisher) CheckAuth() error {
	n.debugLog("Checking NPM authentication...")
	cmd := exec.Command("npm", "whoami")
	output, err := cmd.CombinedOutput()
	if err != nil {
		n.debugLog("Auth check failed: %s", string(output))
		return fmt.Errorf("not authenticated to npm: %w", err)
	}
	n.debugLog("Authenticated as: %s", strings.TrimSpace(string(output)))
	return nil
}

func (n *NPMPublisher) UpdatePackageVersion() error {
	n.debugLog("Updating package.json version to %s", n.version)

	pkgPath := filepath.Join(n.projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("reading package.json: %w", err)
	}

	// Strip 'v' prefix if present
	version := n.version
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	n.debugLog("Setting version field to: %s", version)

	// Use regex to replace version value while preserving exact formatting/spacing
	// Matches: "version": "any.version.number" and replaces just the version value
	versionRegex := regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)
	updatedData := versionRegex.ReplaceAll(data, []byte(`$1"`+version+`"`))
	if err := os.WriteFile(pkgPath, updatedData, 0644); err != nil {
		return fmt.Errorf("writing updated package.json: %w", err)
	}

	n.debugLog("package.json version updated successfully")
	return nil
}

func (n *NPMPublisher) CheckIfPublished() (bool, error) {
	n.debugLog("Checking if %s@%s is already published...", n.packageName, n.version)

	version := n.version
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	cmd := exec.Command("npm", "view", fmt.Sprintf("%s@%s", n.packageName, version), "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "E404") {
			n.debugLog("Version %s not found in registry (good, we can publish)", version)
			return false, nil
		}
		n.debugLog("Error checking npm registry: %s", string(output))
		return false, fmt.Errorf("checking npm registry: %w", err)
	}

	n.debugLog("Version %s already exists in registry", version)
	return true, nil
}

func (n *NPMPublisher) Publish(outputChan chan<- string) error {
	n.debugLog("Starting NPM publish process...")

	// Check if already published
	published, err := n.CheckIfPublished()
	if err != nil {
		return fmt.Errorf("checking publish status: %w", err)
	}
	if published {
		msg := fmt.Sprintf("Version %s already published to NPM", n.version)
		n.debugLog(msg)
		outputChan <- msg
		return nil
	}

	// Run npm publish
	cmd := exec.Command("npm", "publish", "--access", "public")
	cmd.Dir = n.projectPath

	n.debugLog("Running: npm publish --access public in %s", n.projectPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Send output to channel
	if stdout.Len() > 0 {
		outputChan <- stdout.String()
		n.debugLog("STDOUT: %s", stdout.String())
	}
	if stderr.Len() > 0 {
		outputChan <- stderr.String()
		n.debugLog("STDERR: %s", stderr.String())
	}

	if err != nil {
		return fmt.Errorf("npm publish failed: %w", err)
	}

	outputChan <- fmt.Sprintf("✓ Successfully published %s@%s to NPM", n.packageName, n.version)
	return nil
}

func PublishToNPM(projectPath, version, packageName string, outputChan chan<- string) error {
	publisher := NewNPMPublisher(projectPath, version, packageName)

	// Check auth first
	if err := publisher.CheckAuth(); err != nil {
		return err
	}

	// Update package.json version
	if err := publisher.UpdatePackageVersion(); err != nil {
		return err
	}

	// Publish
	return publisher.Publish(outputChan)
}