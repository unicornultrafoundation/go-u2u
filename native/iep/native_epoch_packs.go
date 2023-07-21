package iep

import (
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/native/ier"
)

type LlrEpochPack struct {
	Votes  []native.LlrSignedEpochVote
	Record ier.LlrIdxFullEpochRecord
}
