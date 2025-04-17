package consensus

import (
	"errors"

	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/dag"
)

type eventFilterFn func(event dag.Event) bool

// dfsSubgraph iterates all the events which are observed by head, and accepted by a filter.
// filter MAY BE called twice for the same event.
func (p *Orderer) dfsSubgraph(head hash.Event, filter eventFilterFn) error {
	stack := make(hash.EventsStack, 0, 300)

	for pwalk := &head; pwalk != nil; pwalk = stack.Pop() {
		walk := *pwalk

		event := p.input.GetEvent(walk)
		if event == nil {
			return errors.New("event not found " + walk.String())
		}

		// filter
		if !filter(event) {
			continue
		}

		// memorize parents
		for _, parent := range event.Parents() {
			stack.Push(parent)
		}
	}

	return nil
}
