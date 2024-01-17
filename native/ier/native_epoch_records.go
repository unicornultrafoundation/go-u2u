package ier

import (
	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"

	"github.com/unicornultrafoundation/go-u2u/native/iblockproc"
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
