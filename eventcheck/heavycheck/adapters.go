package heavycheck

import (
	"github.com/unicornultrafoundation/go-helios/native/dag"

	"github.com/unicornultrafoundation/go-u2u/native"
)

type EventsOnly struct {
	*Checker
}

func (c *EventsOnly) Enqueue(e dag.Event, onValidated func(error)) error {
	return c.Checker.EnqueueEvent(e.(native.EventPayloadI), onValidated)
}

type BVsOnly struct {
	*Checker
}

func (c *BVsOnly) Enqueue(bvs native.LlrSignedBlockVotes, onValidated func(error)) error {
	return c.Checker.EnqueueBVs(bvs, onValidated)
}

type EVOnly struct {
	*Checker
}

func (c *EVOnly) Enqueue(ers native.LlrSignedEpochVote, onValidated func(error)) error {
	return c.Checker.EnqueueEV(ers, onValidated)
}
