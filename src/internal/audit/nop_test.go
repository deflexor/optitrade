package audit

import (
	"context"
	"testing"
)

func TestNopDecisionLogger(t *testing.T) {
	var l NopDecisionLogger
	err := l.LogDecision(context.Background(), DecisionRecord{CorrelationID: "x"})
	if err != nil {
		t.Fatalf("LogDecision: %v", err)
	}
}
