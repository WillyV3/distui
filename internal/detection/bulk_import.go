package detection

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type DetectedDistribution struct {
	Name    string
	Type    string // "homebrew" or "npm"
	Version string
}

func DetectAllHomebrewFormulas(tap string) ([]DetectedDistribution, error) {
	if tap == "" {
		return nil, nil
	}

	cmd := exec.Command("brew", "tap-info", tap, "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var tapInfo []struct {
		FormulaNames []string `json:"formula_names"`
	}

	if err := json.Unmarshal(output, &tapInfo); err != nil {
		return nil, err
	}

	if len(tapInfo) == 0 || len(tapInfo[0].FormulaNames) == 0 {
		return nil, nil
	}

	var distributions []DetectedDistribution
	for _, fullFormulaName := range tapInfo[0].FormulaNames {
		parts := strings.Split(fullFormulaName, "/")
		formulaName := fullFormulaName
		if len(parts) > 0 {
			formulaName = parts[len(parts)-1]
		}

		info, err := VerifyHomebrewFormula(tap, formulaName)
		if err == nil && info.Exists {
			distributions = append(distributions, DetectedDistribution{
				Name:    formulaName,
				Type:    "homebrew",
				Version: info.Version,
			})
		}
	}

	return distributions, nil
}

func DetectAllNPMPackages(scope string) ([]DetectedDistribution, error) {
	if scope == "" {
		return nil, nil
	}

	var cmd *exec.Cmd
	if strings.HasPrefix(scope, "@") {
		cmd = exec.Command("npm", "search", scope, "--json")
	} else {
		cmd = exec.Command("npm", "search", "@"+scope, "--json")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := json.Unmarshal(output, &packages); err != nil {
		return nil, err
	}

	var distributions []DetectedDistribution
	for _, pkg := range packages {
		distributions = append(distributions, DetectedDistribution{
			Name:    pkg.Name,
			Type:    "npm",
			Version: pkg.Version,
		})
	}

	return distributions, nil
}
