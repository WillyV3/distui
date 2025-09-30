package executor

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"os"
)

type NPMPublisher struct {
	projectPath string
	version     string
	packageName string
}

func NewNPMPublisher(projectPath, version, packageName string) *NPMPublisher {
	return &NPMPublisher{
		projectPath: projectPath,
		version:     version,
		packageName: packageName,
	}
}

func (n *NPMPublisher) CheckAuth() error {
	cmd := exec.Command("npm", "whoami")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("not authenticated to npm: %w", err)
	}
	return nil
}

func (n *NPMPublisher) UpdatePackageVersion() error {
	pkgPath := filepath.Join(n.projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("reading package.json: %w", err)
	}

	version := n.version
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	versionRegex := regexp.MustCompile(`("version"\s*:\s*)"[^"]*"`)
	updatedData := versionRegex.ReplaceAll(data, []byte(`$1"`+version+`"`))
	if err := os.WriteFile(pkgPath, updatedData, 0644); err != nil {
		return fmt.Errorf("writing updated package.json: %w", err)
	}

	return nil
}

func (n *NPMPublisher) CheckIfPublished() (bool, error) {
	version := n.version
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	cmd := exec.Command("npm", "view", fmt.Sprintf("%s@%s", n.packageName, version), "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "E404") {
			return false, nil
		}
		return false, fmt.Errorf("checking npm registry: %w", err)
	}

	return true, nil
}

func (n *NPMPublisher) Publish(outputChan chan<- string) error {
	published, err := n.CheckIfPublished()
	if err != nil {
		return fmt.Errorf("checking publish status: %w", err)
	}
	if published {
		msg := fmt.Sprintf("Version %s already published to NPM", n.version)
		outputChan <- msg
		return nil
	}

	cmd := exec.Command("npm", "publish", "--access", "public")
	cmd.Dir = n.projectPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if stdout.Len() > 0 {
		outputChan <- stdout.String()
	}
	if stderr.Len() > 0 {
		outputChan <- stderr.String()
	}

	if err != nil {
		return fmt.Errorf("npm publish failed: %w", err)
	}

	successMsg := fmt.Sprintf("✓ Successfully published %s@%s to NPM", n.packageName, n.version)
	outputChan <- successMsg
	return nil
}

func PublishToNPM(projectPath, version, packageName string, outputChan chan<- string) error {
	publisher := NewNPMPublisher(projectPath, version, packageName)

	if err := publisher.CheckAuth(); err != nil {
		return err
	}

	if err := publisher.UpdatePackageVersion(); err != nil {
		return err
	}

	if err := publisher.Publish(outputChan); err != nil {
		return err
	}

	return nil
}