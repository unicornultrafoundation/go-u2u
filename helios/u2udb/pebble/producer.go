package pebble

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
)

type Producer struct {
	datadir         string
	getCacheFdLimit func(string) (int, int)
}

// NewProducer of Pebble db.
func NewProducer(datadir string, getCacheFdLimit func(string) (int, int)) u2udb.IterableDBProducer {
	return &Producer{
		datadir:         datadir,
		getCacheFdLimit: getCacheFdLimit,
	}
}

// Names of existing databases.
func (p *Producer) Names() []string {
	files, err := os.ReadDir(p.datadir)
	if err != nil {
		panic(err)
	}

	names := make([]string, 0, len(files))
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		names = append(names, f.Name())
	}
	return names
}

// OpenDB or create db with name.
func (p *Producer) OpenDB(name string) (u2udb.Store, error) {
	// Validate name parameter
	if !u2udb.IsValidDatabaseName(name) {
		return nil, errors.New("invalid database name")
	}
	path := p.resolvePath(name)

	err := os.MkdirAll(path, 0700)
	if err != nil {
		return nil, err
	}

	onDrop := func() {
		_ = os.RemoveAll(path)
	}

	cache, fdlimit := p.getCacheFdLimit(name)
	db, err := New(path, cache, fdlimit, nil, onDrop)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (p *Producer) resolvePath(name string) string {
	return filepath.Join(p.datadir, name)
}
