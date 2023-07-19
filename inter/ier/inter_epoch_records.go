package ier

import (
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/inter/idx"

	"github.com/unicornultrafoundation/go-u2u/inter/iblockproc"
)

type LlrFullEpochRecord struct {
	BlockState iblockproc.BlockState
	EpochState iblockproc.EpochState
}

type LlrIdxFullEpochRecord struct {
	LlrFullEpochRecord
	Idx idx.Epoch
}

func (er LlrFullEpochRecord) Hash() hash.Hash {
	return hash.Of(er.BlockState.Hash().Bytes(), er.EpochState.Hash().Bytes())
}
