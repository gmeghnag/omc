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

// runUpgradeSubprocess re-runs the current test binary with the given
// DesiredVersion and a managed install, so upgradeBinary's os.Exit can be
// observed from the outside.
func runUpgradeSubprocess(t *testing.T, testName, desiredVersion, versionTag string) (string, int) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	cmd.Env = append(os.Environ(),
		"OMC_TEST_UPGRADE_SUBPROCESS=1",
		"OMC_TEST_DESIRED_VERSION="+desiredVersion,
		"OMC_TEST_VERSION_TAG="+versionTag,
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), 0
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("upgradeBinary exited with %v, want an exit status\nOutput: %s", err, string(output))
	}
	return string(output), exitErr.ExitCode()
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

// TestUpgradeBinaryManagedInstallStillValidates checks that a managed install
// still runs the read-only validation of --to: only replacing the binary
// belongs to the package manager, and turning these away made the
// "check them by running \"omc upgrade\"" hint point at a command that refused
// to run.
func TestUpgradeBinaryManagedInstallStillValidates(t *testing.T) {
	if os.Getenv("OMC_TEST_UPGRADE_SUBPROCESS") == "1" {
		vars.InstallMethod = "homebrew"
		DesiredVersion = os.Getenv("OMC_TEST_DESIRED_VERSION")
		vars.OMCVersionTag = os.Getenv("OMC_TEST_VERSION_TAG")
		upgradeBinary("gmeghnag/omc")
		return
	}

	t.Run("invalid --to", func(t *testing.T) {
		output, code := runUpgradeSubprocess(t, "TestUpgradeBinaryManagedInstallStillValidates", "not-a-version", "v2.0.0")
		if code != 1 {
			t.Errorf("exit status = %d, want 1\nOutput: %s", code, output)
		}
		if !strings.Contains(output, "must be a semantic version") {
			t.Errorf("output = %q, want the --to validation to run", output)
		}
		if strings.Contains(output, "brew upgrade omc") {
			t.Errorf("output = %q, the managed-install warning must not preempt the validation", output)
		}
	})

	t.Run("downgrade refused", func(t *testing.T) {
		output, code := runUpgradeSubprocess(t, "TestUpgradeBinaryManagedInstallStillValidates", "v1.0.0", "v2.0.0")
		if code != 1 {
			t.Errorf("exit status = %d, want 1\nOutput: %s", code, output)
		}
		if !strings.Contains(output, "not one of the available updates") {
			t.Errorf("output = %q, want the downgrade to be refused", output)
		}
		if strings.Contains(output, "brew upgrade omc") {
			t.Errorf("output = %q, the managed-install warning must not preempt the validation", output)
		}
	})

	t.Run("a real upgrade still defers", func(t *testing.T) {
		output, code := runUpgradeSubprocess(t, "TestUpgradeBinaryManagedInstallStillValidates", "latest", "v2.0.0")
		if code != 1 {
			t.Errorf("exit status = %d, want 1\nOutput: %s", code, output)
		}
		if !strings.Contains(output, "brew upgrade omc") {
			t.Errorf("output = %q, want it to defer to homebrew", output)
		}
	})
}
