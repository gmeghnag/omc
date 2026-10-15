package upgrade

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/gmeghnag/omc/vars"
)

func TestCheckManagedInstall(t *testing.T) {
	tests := []struct {
		name          string
		installMethod string
		wantManaged   bool
		wantContains  string
	}{
		{
			name:          "not set, omc upgrades itself",
			installMethod: "",
			wantManaged:   false,
		},
		{
			name:          "blanks only, omc upgrades itself",
			installMethod: "   ",
			wantManaged:   false,
		},
		{
			name:          "homebrew",
			installMethod: "homebrew",
			wantManaged:   true,
			wantContains:  "brew upgrade omc",
		},
		{
			name:          "brew alias",
			installMethod: "brew",
			wantManaged:   true,
			wantContains:  "brew upgrade omc",
		},
		{
			name:          "case and spaces are ignored",
			installMethod: " Homebrew ",
			wantManaged:   true,
			wantContains:  "brew upgrade omc",
		},
		{
			name:          "any other package manager",
			installMethod: "rpm",
			wantManaged:   true,
			wantContains:  "installed via rpm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			managed := checkManagedInstall(&out, tc.installMethod)
			if managed != tc.wantManaged {
				t.Fatalf("checkManagedInstall(%q) = %v, want %v", tc.installMethod, managed, tc.wantManaged)
			}
			if !managed && out.Len() != 0 {
				t.Fatalf("checkManagedInstall(%q) printed %q, want nothing", tc.installMethod, out.String())
			}
			if tc.wantContains != "" && !strings.Contains(out.String(), tc.wantContains) {
				t.Errorf("checkManagedInstall(%q) printed %q, want it to contain %q", tc.installMethod, out.String(), tc.wantContains)
			}
		})
	}
}

// TestUpgradeBinaryHomebrewExits runs upgradeBinary in a subprocess (it calls
// os.Exit) to check that a homebrew install is warned about and no upgrade is
// attempted.
func TestUpgradeBinaryHomebrewExits(t *testing.T) {
	if os.Getenv("OMC_TEST_UPGRADE_SUBPROCESS") == "1" {
		vars.InstallMethod = "homebrew"
		DesiredVersion = "latest"
		upgradeBinary("gmeghnag/omc")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestUpgradeBinaryHomebrewExits")
	cmd.Env = append(os.Environ(), "OMC_TEST_UPGRADE_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("upgradeBinary exited with %v, want exit status 1\nOutput: %s", err, string(output))
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("upgradeBinary exited with status %d, want 1\nOutput: %s", exitErr.ExitCode(), string(output))
	}
	if !strings.Contains(string(output), "brew upgrade omc") {
		t.Errorf("upgradeBinary output = %q, want it to contain %q", string(output), "brew upgrade omc")
	}
}
