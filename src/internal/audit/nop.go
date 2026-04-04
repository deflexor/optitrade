package audit

import "context"

// NopDecisionLogger satisfies DecisionLogger without persisting (e.g. dashboard runner until audit DB is wired).
type NopDecisionLogger struct{}

func (NopDecisionLogger) LogDecision(_ context.Context, _ DecisionRecord) error { return nil }
