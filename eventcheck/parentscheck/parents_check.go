package parentscheck

import (
	"errors"

	base "github.com/unicornultrafoundation/go-u2u/consensus/eventcheck/parentscheck"

	"github.com/unicornultrafoundation/go-u2u/native"
)

var (
	ErrPastTime = errors.New("event has lower claimed time than self-parent")
)

// Checker which require only parents list + current epoch info
type Checker struct {
	base *base.Checker
}

// New validator which performs checks, which require known the parents
func New() *Checker {
	return &Checker{
		base: &base.Checker{},
	}
}

// Validate event
func (v *Checker) Validate(e native.EventI, parents native.EventIs) error {
	if err := v.base.Validate(e, parents.Bases()); err != nil {
		return err
	}

	if e.SelfParent() != nil {
		selfParent := parents[0]
		if !e.IsSelfParent(selfParent.ID()) {
			// sanity check, self-parent is always first, it's how it's stored
			return base.ErrWrongSelfParent
		}
		// selfParent time
		if e.CreationTime() <= selfParent.CreationTime() {
			return ErrPastTime
		}
	}

	return nil
}
