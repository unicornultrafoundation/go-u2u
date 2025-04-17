package wmedian

import "github.com/unicornultrafoundation/go-u2u/consensus/native/pos"

type WeightedValue interface {
	Weight() pos.Weight
}

func Of(values []WeightedValue, stop pos.Weight) WeightedValue {
	// Calculate weighted median
	var curWeight pos.Weight
	for _, value := range values {
		curWeight += value.Weight()
		if curWeight >= stop {
			return value
		}
	}
	panic("invalid median")
}
