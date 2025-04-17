package hash

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"strings"

	"github.com/unicornultrafoundation/go-u2u/common"

	"github.com/unicornultrafoundation/go-u2u/common/bigendian"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
)

type (
	// Event is a unique identifier of event.
	// It is a hash of Event.
	Event Hash

	// OrderedEvents is a sortable slice of event hashes.
	OrderedEvents []Event

	// Events is a slice of event hashes.
	Events []Event

	EventsStack []Event

	// EventsSet provides additional methods of event hash index.
	EventsSet map[Event]struct{}
)

var (
	// ZeroEvent is a hash of virtual initial event.
	ZeroEvent = Event{}
)

/*
 * Event methods:
 */

// Bytes returns value as byte slice.
func (h Event) Bytes() []byte {
	return (Hash)(h).Bytes()
}

// Big converts a hash to a big integer.
func (h *Event) Big() *big.Int {
	return (*Hash)(h).Big()
}

// setBytes converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func (h *Event) SetBytes(raw []byte) {
	(*Hash)(h).SetBytes(raw)
}

// BytesToEvent converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToEvent(b []byte) Event {
	return Event(FromBytes(b))
}

// FromBytes converts bytes to hash.
// If b is larger than len(h), b will be cropped from the left.
func FromBytes(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// HexToEventHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToEventHash(s string) Event {
	return Event(HexToHash(s))
}

// Hex converts an event hash to a hex string.
func (h Event) Hex() string {
	return Hash(h).Hex()
}

// Lamport returns [4:8] bytes, which store event's Lamport.
func (h Event) Lamport() idx.Lamport {
	return idx.BytesToLamport(h[4:8])
}

// Epoch returns [0:4] bytes, which store event's Epoch.
func (h Event) Epoch() idx.Epoch {
	return idx.BytesToEpoch(h[0:4])
}

// String returns human readable string representation.
func (h Event) String() string {
	return h.ShortID(3)
}

// FullID returns human readable string representation with no information loss.
func (h Event) FullID() string {
	return h.ShortID(32 - 4 - 4)
}

// ShortID returns human readable ID representation, suitable for API calls.
func (h Event) ShortID(precision int) string {
	if name := GetEventName(h); len(name) > 0 {
		return name
	}
	// last bytes, because first are occupied by epoch and lamport
	return fmt.Sprintf("%d:%d:%s", h.Epoch(), h.Lamport(), common.Bytes2Hex(h[8:8+precision]))
}

// IsZero returns true if hash is empty.
func (h *Event) IsZero() bool {
	return *h == Event{}
}

/*
 * EventsSet methods:
 */

// NewEventsSet makes event hash index.
func NewEventsSet(h ...Event) EventsSet {
	hh := EventsSet{}
	hh.Add(h...)
	return hh
}

// Copy copies events to a new structure.
func (hh EventsSet) Copy() EventsSet {
	ee := make(EventsSet, len(hh))
	for k, v := range hh {
		ee[k] = v
	}

	return ee
}

// String returns human readable string representation.
func (hh EventsSet) String() string {
	ss := make([]string, 0, len(hh))
	for h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Slice returns whole index as slice.
func (hh EventsSet) Slice() Events {
	arr := make(Events, len(hh))
	i := 0
	for h := range hh {
		arr[i] = h
		i++
	}
	return arr
}

// Add appends hash to the index.
func (hh EventsSet) Add(hash ...Event) {
	for _, h := range hash {
		hh[h] = struct{}{}
	}
}

// Erase erase hash from the index.
func (hh EventsSet) Erase(hash ...Event) {
	for _, h := range hash {
		delete(hh, h)
	}
}

// Contains returns true if hash is in.
func (hh EventsSet) Contains(hash Event) bool {
	_, ok := hh[hash]
	return ok
}

/*
 * Events methods:
 */

// NewEvents makes event hash slice.
func NewEvents(h ...Event) Events {
	hh := Events{}
	hh.Add(h...)
	return hh
}

// Copy copies events to a new structure.
func (hh Events) Copy() Events {
	return append(Events(nil), hh...)
}

// String returns human readable string representation.
func (hh Events) String() string {
	ss := make([]string, 0, len(hh))
	for _, h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Set returns whole index as a EventsSet.
func (hh Events) Set() EventsSet {
	set := make(EventsSet, len(hh))
	for _, h := range hh {
		set[h] = struct{}{}
	}
	return set
}

// Add appends hash to the slice.
func (hh *Events) Add(hash ...Event) {
	*hh = append(*hh, hash...)
}

/*
 * EventsStack methods:
 */

// Push event ID on top
func (s *EventsStack) Push(v Event) {
	*s = append(*s, v)
}

// PushAll event IDs on top
func (s *EventsStack) PushAll(vv Events) {
	*s = append(*s, vv...)
}

// Pop event ID from top. Erases element.
func (s *EventsStack) Pop() *Event {
	l := len(*s)
	if l == 0 {
		return nil
	}

	res := &(*s)[l-1]
	*s = (*s)[:l-1]

	return res
}

/*
 * OrderedEvents methods:
 */

// String returns string representation.
func (hh OrderedEvents) String() string {
	buf := &strings.Builder{}

	out := func(s string) {
		if _, err := buf.WriteString(s); err != nil {
			panic(err)
		}
	}

	out("[")
	for _, h := range hh {
		out(h.String() + ", ")
	}
	out("]")

	return buf.String()
}

func (hh OrderedEvents) Len() int      { return len(hh) }
func (hh OrderedEvents) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }
func (hh OrderedEvents) Less(i, j int) bool {
	return bytes.Compare(hh[i].Bytes(), hh[j].Bytes()) < 0
}

// ByEpochAndLamport sorts events by epoch first, by lamport second, by ID third
func (hh OrderedEvents) ByEpochAndLamport() {
	sort.Sort(hh)
}

// Of returns hash of data
func Of(data ...[]byte) (hash Hash) {
	d := sha256.New()
	for _, b := range data {
		_, err := d.Write(b)
		if err != nil {
			panic(err)
		}
	}
	d.Sum(hash[:0])
	return hash
}

/*
 * Utils:
 */

// FakePeer generates random fake peer id for testing purpose.
func FakePeer() idx.ValidatorID {
	return idx.BytesToValidatorID(FakeHash().Bytes()[:4])
}

// FakeEpoch gives fixed value of fake epoch for testing purpose.
func FakeEpoch() idx.Epoch {
	return 123456
}

// FakeEvent generates random fake event hash with the same epoch for testing purpose.
func FakeEvent() (h Event) {
	_, err := rand.Read(h[:]) // nolint:gosec
	if err != nil {
		panic(err)
	}
	copy(h[0:4], bigendian.Uint32ToBytes(uint32(FakeEpoch())))
	return
}

// FakeEvents generates random hashes of fake event with the same epoch for testing purpose.
func FakeEvents(n int) Events {
	res := Events{}
	for i := 0; i < n; i++ {
		res.Add(FakeEvent())
	}
	return res
}
