package batched

import "github.com/unicornultrafoundation/go-u2u/helios/u2udb"

// Store is a wrapper which translates every Put/Delete op into a batch
type Store struct {
	u2udb.Store
	batch u2udb.Batch
}

func Wrap(s u2udb.Store) *Store {
	return &Store{
		Store: s,
		batch: s.NewBatch(),
	}
}

func (s *Store) Write() error {
	return s.batch.Write()
}

func (s *Store) Reset() {
	s.batch.Reset()
}

func (s *Store) Replay(w u2udb.Writer) error {
	return s.batch.Replay(w)
}

func (s *Store) Flush() error {
	err := s.batch.Write()
	if err != nil {
		return err
	}
	s.batch.Reset()
	return nil
}

func (s *Store) MayFlush() (bool, error) {
	if s.batch.ValueSize() <= u2udb.IdealBatchSize {
		return false, nil
	}
	return true, s.Flush()
}

// Put inserts the given value into the batch and may flush the batch.
func (s *Store) Put(key []byte, value []byte) error {
	if _, err := s.MayFlush(); err != nil {
		return err
	}
	return s.batch.Put(key, value)
}

// Delete places removal of the given value into the batch and may flush the batch.
func (s *Store) Delete(key []byte) error {
	if _, err := s.MayFlush(); err != nil {
		return err
	}
	return s.batch.Delete(key)
}

func (s *Store) Close() error {
	if err := s.Flush(); err != nil {
		return err
	}
	return s.Store.Close()
}
