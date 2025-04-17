package types

import "github.com/unicornultrafoundation/go-u2u/helios/hash"

// Block is a part of an ordered chain of batches of events.
type Block struct {
	Event    hash.Event
	Cheaters Cheaters
}
