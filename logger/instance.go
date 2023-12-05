package logger

import (
	"github.com/unicornultrafoundation/go-u2u/log"
)

type Instance struct {
	Log log.Logger
}

func New(name ...string) Instance {
	if len(name) == 0 {
		return Instance{
			Log: log.New(),
		}
	}
	return Instance{
		Log: log.New("module", name[0]),
	}
}
