package strategy

import (
	"fmt"

	"github.com/dfr/optitrade/execution/internal/config"
	"github.com/dfr/optitrade/execution/internal/regime"
)

// AllowedStructures resolves the playbook bucket for label and returns the
// policy allow-list. It errors when the list is empty (misconfiguration).
func AllowedStructures(policy *config.Policy, label regime.Label) ([]string, error) {
	if policy == nil {
		return nil, fmt.Errorf("policy is nil")
	}
	var pb config.Playbook
	switch label {
	case regime.LabelLow:
		pb = policy.Playbooks.Low
	case regime.LabelNormal:
		pb = policy.Playbooks.Normal
	case regime.LabelHigh:
		pb = policy.Playbooks.High
	default:
		return nil, fmt.Errorf("unknown regime label %q", label)
	}
	if len(pb.AllowedStructures) == 0 {
		return nil, fmt.Errorf("playbook %q has no allowed_structures", label)
	}
	out := append([]string(nil), pb.AllowedStructures...)
	return out, nil
}
