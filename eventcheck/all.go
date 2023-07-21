package eventcheck

import (
	"github.com/unicornultrafoundation/go-u2u/eventcheck/basiccheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/epochcheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/gaspowercheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/heavycheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/parentscheck"
	"github.com/unicornultrafoundation/go-u2u/native"
)

// Checkers is collection of all the checkers
type Checkers struct {
	Basiccheck    *basiccheck.Checker
	Epochcheck    *epochcheck.Checker
	Parentscheck  *parentscheck.Checker
	Gaspowercheck *gaspowercheck.Checker
	Heavycheck    *heavycheck.Checker
}

// Validate runs all the checks except Poset-related
func (v *Checkers) Validate(e native.EventPayloadI, parents native.EventIs) error {
	if err := v.Basiccheck.Validate(e); err != nil {
		return err
	}
	if err := v.Epochcheck.Validate(e); err != nil {
		return err
	}
	if err := v.Parentscheck.Validate(e, parents); err != nil {
		return err
	}
	var selfParent native.EventI
	if e.SelfParent() != nil {
		selfParent = parents[0]
	}
	if err := v.Gaspowercheck.Validate(e, selfParent); err != nil {
		return err
	}
	if err := v.Heavycheck.ValidateEvent(e); err != nil {
		return err
	}
	return nil
}
