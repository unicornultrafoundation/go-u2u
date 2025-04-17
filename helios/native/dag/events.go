package dag

import (
	"strings"

	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
)

// Events is a ordered slice of events.
type Events []Event

// String returns human readable representation.
func (ee Events) String() string {
	ss := make([]string, len(ee))
	for i := 0; i < len(ee); i++ {
		ss[i] = ee[i].String()
	}
	return strings.Join(ss, " ")
}

func (ee Events) Metric() (metric Metric) {
	metric.Num = idx.Event(len(ee))
	for _, e := range ee {
		metric.Size += uint64(e.Size())
	}
	return metric
}

func (ee Events) IDs() hash.Events {
	ids := make(hash.Events, len(ee))
	for i, e := range ee {
		ids[i] = e.ID()
	}
	return ids
}
