package election

import (
	"errors"
)

// Chooses the decided "yes" roots with the greatest weight amount.
// This root serves as a "checkpoint" within DAG, as it's guaranteed to be final and consistent unless more than 1/3W are Byzantine.
// Other validators will come to the same Event not later than current highest frame + 2.
func (el *Election) chooseEvent() (*Res, error) {
	// iterate until Yes root is met, which will be Event. I.e. not necessarily all the roots must be decided
	for _, validator := range el.validators.SortedIDs() {
		vote, ok := el.decidedRoots[validator]
		if !ok {
			return nil, nil // not decided
		}
		if vote.yes {
			return &Res{
				Frame: el.frameToDecide,
				Event: vote.observedRoot,
			}, nil
		}
	}
	return nil, errors.New("all the roots are decided as 'no', which is possible only if more than 1/3W are Byzantine")
}
