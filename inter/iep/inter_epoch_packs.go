package iep

import (
	"github.com/unicornultrafoundation/go-u2u/inter"
	"github.com/unicornultrafoundation/go-u2u/inter/ier"
)

type LlrEpochPack struct {
	Votes  []inter.LlrSignedEpochVote
	Record ier.LlrIdxFullEpochRecord
}
