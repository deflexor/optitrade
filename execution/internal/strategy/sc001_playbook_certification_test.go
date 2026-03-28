package strategy

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/regime"
)

func policyExamplePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "config", "examples", "policy.example.json")
}

// SC-001 (spec): every playbook-allowed structure token builds defined-risk legs (certification corpus).
func TestSC001_PlaybookAllowedStructuresAreDefinedRisk(t *testing.T) {
	t.Parallel()
	pol, err := config.LoadFile(policyExamplePath(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range []regime.Label{regime.LabelLow, regime.LabelNormal, regime.LabelHigh} {
		allowed, err := AllowedStructures(pol, label)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range allowed {
			legs, err := BuildLegsForStructure(name)
			if err != nil {
				t.Fatalf("regime=%s structure=%q: %v", label, name, err)
			}
			if err := ValidateDefinedRisk(legs); err != nil {
				t.Fatalf("regime=%s structure=%q legs=%+v: %v", label, name, legs, err)
			}
		}
	}
}
