package consensus

import "github.com/unicornultrafoundation/go-u2u/helios/native/idx"

const dsKey = "d"

// SetLastDecidedState save LastDecidedState.
// LastDecidedState is seldom read; so no cache.
func (s *Store) SetLastDecidedState(v *LastDecidedState) {
	s.cache.LastDecidedState = v

	s.set(s.table.LastDecidedState, []byte(dsKey), v)
}

// GetLastDecidedState returns stored LastDecidedState.
// State is seldom read; so no cache.
func (s *Store) GetLastDecidedState() *LastDecidedState {
	if s.cache.LastDecidedState != nil {
		return s.cache.LastDecidedState
	}

	w, exists := s.get(s.table.LastDecidedState, []byte(dsKey), &LastDecidedState{}).(*LastDecidedState)
	if !exists {
		s.crit(ErrNoGenesis)
	}

	s.cache.LastDecidedState = w
	return w
}

func (s *Store) GetLastDecidedFrame() idx.Frame {
	return s.GetLastDecidedState().LastDecidedFrame
}
