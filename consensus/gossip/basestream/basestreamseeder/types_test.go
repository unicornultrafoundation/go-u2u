package basestreamseeder

import (
	"bytes"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"

	"github.com/unicornultrafoundation/go-u2u/consensus/gossip/basestream"
	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/dag"
)

type testLocator struct {
	B []byte
}

func (l testLocator) Compare(b basestream.Locator) int {
	return bytes.Compare(l.B, b.(testLocator).B)
}

func (l testLocator) Inc() basestream.Locator {
	nextBn := new(big.Int).SetBytes(l.B)
	nextBn.Add(nextBn, common.Big1)
	return testLocator{
		B: nextBn.Bytes(),
	}
}

type testPayload struct {
	IDs    hash.Events
	Events dag.Events
	Size   uint64
}

func (p testPayload) AddEvent(id hash.Event, event dag.Event) {
	p.IDs = append(p.IDs, id)          // nolint:staticcheck
	p.Events = append(p.Events, event) // nolint:staticcheck
	p.Size += uint64(event.Size())     // nolint:staticcheck
}

func (p testPayload) Len() int {
	return len(p.IDs)
}

func (p testPayload) TotalSize() uint64 {
	return p.Size
}

func (p testPayload) TotalMemSize() int {
	return int(p.Size) + len(p.IDs)*128
}
