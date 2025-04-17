package eventcheck

import (
	"github.com/unicornultrafoundation/go-u2u/consensus/eventcheck/basiccheck"
	"github.com/unicornultrafoundation/go-u2u/consensus/eventcheck/epochcheck"
	"github.com/unicornultrafoundation/go-u2u/consensus/eventcheck/parentscheck"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/dag"
)

// Checkers is collection of all the checkers
type Checkers struct {
	Basiccheck   *basiccheck.Checker
	Epochcheck   *epochcheck.Checker
	Parentscheck *parentscheck.Checker
}

// Validate runs all the checks except Lachesis-related
func (v *Checkers) Validate(e dag.Event, parents dag.Events) error {
	if err := v.Basiccheck.Validate(e); err != nil {
		return err
	}
	if err := v.Epochcheck.Validate(e); err != nil {
		return err
	}
	if err := v.Parentscheck.Validate(e, parents); err != nil {
		return err
	}
	return nil
}
