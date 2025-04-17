package tdag

import (
	"github.com/unicornultrafoundation/go-u2u/rlp"

	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
)

type TestEventMarshaling struct {
	Epoch idx.Epoch
	Seq   idx.Event

	Frame idx.Frame

	Creator idx.ValidatorID

	Parents hash.Events

	Lamport idx.Lamport

	ID   hash.Event
	Name string
}

// EventToBytes serializes events
func (e *TestEvent) Bytes() []byte {
	b, _ := rlp.EncodeToBytes(&TestEventMarshaling{
		Epoch:   e.Epoch(),
		Seq:     e.Seq(),
		Frame:   e.Frame(),
		Creator: e.Creator(),
		Parents: e.Parents(),
		Lamport: e.Lamport(),
		ID:      e.ID(),
		Name:    e.Name,
	})
	return b
}
