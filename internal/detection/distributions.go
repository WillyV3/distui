package detection

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type DistributionInfo struct {
	Version string
	Exists  bool
}

func VerifyHomebrewFormula(tap, formula string) (*DistributionInfo, error) {
	if tap == "" || formula == "" {
		return &DistributionInfo{Exists: false}, nil
	}

	fullName := tap + "/" + formula
	cmd := exec.Command("brew", "info", fullName, "--json=v2")
	output, err := cmd.Output()
	if err != nil {
		return &DistributionInfo{Exists: false}, nil
	}

	var result struct {
		Formulae []struct {
			Versions struct {
				Stable string `json:"stable"`
			} `json:"versions"`
		} `json:"formulae"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parsing brew info: %w", err)
	}

	if len(result.Formulae) == 0 {
		return &DistributionInfo{Exists: false}, nil
	}

	version := result.Formulae[0].Versions.Stable
	if version == "" {
		return &DistributionInfo{Exists: false}, nil
	}

	return &DistributionInfo{
		Version: ensureVPrefix(version),
		Exists:  true,
	}, nil
}

func VerifyNPMPackage(packageName string) (*DistributionInfo, error) {
	if packageName == "" {
		return &DistributionInfo{Exists: false}, nil
	}

	cmd := exec.Command("npm", "view", packageName, "version")
	output, err := cmd.Output()
	if err != nil {
		return &DistributionInfo{Exists: false}, nil
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return &DistributionInfo{Exists: false}, nil
	}

	return &DistributionInfo{
		Version: ensureVPrefix(version),
		Exists:  true,
	}, nil
}

func ensureVPrefix(version string) string {
	if version == "" {
		return ""
	}
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}
