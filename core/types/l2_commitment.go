// Copyright 2024 The go-u2u Authors
// This file is part of the go-u2u library.
//
// The go-u2u library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-u2u library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-u2u library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"github.com/unicornultrafoundation/go-u2u/common"
)

//go:generate gencodec -type L2StateCommitment -out gen_l2_commitment.go

// L2StateCommitment represents a commitment to an L2 rollup state.
// This allows U2U to act as a settlement layer for Layer 2 rollups.
type L2StateCommitment struct {
	RollupID    uint64      `json:"rollupId"    gencodec:"required"` // L2 rollup identifier
	StateRoot   common.Hash `json:"stateRoot"   gencodec:"required"` // L2 state root hash
	BlockNumber uint64      `json:"blockNumber" gencodec:"required"` // L2 block number
	Timestamp   uint64      `json:"timestamp"   gencodec:"required"` // Commitment timestamp
	DataCommit  common.Hash `json:"dataCommit"  gencodec:"required"` // Data availability commitment hash
	ProofType   uint8       `json:"proofType"   gencodec:"required"` // Proof type (0=optimistic, 1=zk)
}

// L2StateCommitmentList is a list of L2 state commitments.
type L2StateCommitmentList []L2StateCommitment

// ProofType constants for L2 rollups
const (
	OptimisticRollupProof uint8 = 0 // Optimistic rollup with fraud proofs
	ZKRollupProof         uint8 = 1 // Zero-knowledge rollup with validity proofs
)
