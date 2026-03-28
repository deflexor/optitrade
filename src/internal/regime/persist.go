package regime

import (
	"context"
	"errors"

	"github.com/dfr/optitrade/src/internal/state"
)

// PersistIfChanged inserts regime_state when the effective label or classifier version
// changed versus the latest row (or when the database is empty).
func PersistIfChanged(ctx context.Context, repo state.RegimeRepository, eval Evaluation, effectiveAtMs int64) error {
	if repo == nil {
		return errors.New("nil regime repository")
	}
	latest, err := repo.LatestRegimeState(ctx)
	if err != nil && !errors.Is(err, state.ErrNoRegimeState) {
		return err
	}
	if latest != nil &&
		latest.Label == string(eval.Outcome.Label) &&
		latest.ClassifierVersion == eval.Outcome.ClassifierVersion {
		return nil
	}
	return repo.InsertRegimeState(ctx, &state.RegimeState{
		EffectiveAt:       effectiveAtMs,
		Label:             string(eval.Outcome.Label),
		ClassifierVersion: eval.Outcome.ClassifierVersion,
		InputsDigest:      eval.InputsDigest,
	})
}
