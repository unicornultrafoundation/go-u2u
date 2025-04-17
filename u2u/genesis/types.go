package genesis

import (
	"github.com/unicornultrafoundation/go-u2u/helios/hash"

	"github.com/unicornultrafoundation/go-u2u/native/ibr"
	"github.com/unicornultrafoundation/go-u2u/native/ier"
)

type (
	Hashes map[string]hash.Hash
	Header struct {
		GenesisID   hash.Hash
		NetworkID   uint64
		NetworkName string
	}
	Blocks interface {
		ForEach(fn func(ibr.LlrIdxFullBlockRecord) bool)
	}
	Epochs interface {
		ForEach(fn func(ier.LlrIdxFullEpochRecord) bool)
	}
	EvmItems interface {
		ForEach(fn func(key, value []byte) bool)
	}
	Genesis struct {
		Header

		Blocks      Blocks
		Epochs      Epochs
		RawEvmItems EvmItems
	}
)

func (hh Hashes) Includes(hh2 Hashes) bool {
	for n, h := range hh {
		if hh2[n] != h {
			return false
		}
	}
	return true
}

func (hh Hashes) Equal(hh2 Hashes) bool {
	return hh.Includes(hh2) && hh2.Includes(hh)
}

func (h Header) Equal(h2 Header) bool {
	return h == h2
}
