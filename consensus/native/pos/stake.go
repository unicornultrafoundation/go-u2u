package pos

import (
	"math"

	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
)

type (
	// Weight amount.
	Weight uint32
)

type (
	// WeightCounterProvider providers weight counter.
	WeightCounterProvider func() *WeightCounter

	// WeightCounter counts weights.
	WeightCounter struct {
		validators Validators
		already    []bool // idx.Validator -> bool

		quorum Weight
		sum    Weight
	}
)

// NewCounter constructor.
func (vv Validators) NewCounter() *WeightCounter {
	return newWeightCounter(vv)
}

func newWeightCounter(vv Validators) *WeightCounter {
	return &WeightCounter{
		validators: vv,
		quorum:     vv.Quorum(),
		already:    make([]bool, vv.Len()),
		sum:        0,
	}
}

// Count validator and return true if it hadn't counted before.
func (s *WeightCounter) Count(v idx.ValidatorID) bool {
	validatorIdx := s.validators.GetIdx(v)
	return s.CountByIdx(validatorIdx)
}

// CountByIdx validator and return true if it hadn't counted before.
func (s *WeightCounter) CountByIdx(validatorIdx idx.Validator) bool {
	if s.already[validatorIdx] {
		return false
	}
	s.already[validatorIdx] = true
	w := s.validators.GetWeightByIdx(validatorIdx)
	if uint64(s.sum)+uint64(w) > math.MaxUint32 {
		return false
	}
	s.sum += w
	return true
}

// HasQuorum achieved.
func (s *WeightCounter) HasQuorum() bool {
	return s.sum >= s.quorum
}

// Sum of counted weights.
func (s *WeightCounter) Sum() Weight {
	return s.sum
}

// NumCounted of validators
func (s *WeightCounter) NumCounted() int {
	num := 0
	for _, counted := range s.already {
		if counted {
			num++
		}
	}
	return num
}
