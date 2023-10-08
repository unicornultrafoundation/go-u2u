package logger

import (
	"testing"

	"github.com/unicornultrafoundation/go-u2u/libs/log"
)

// SetTestMode sets test mode.
func SetTestMode(t testing.TB) {
	log.Root().SetHandler(
		log.CallerStackHandler("%v", TestHandler(t, log.LogfmtFormat())))
}

// TestHandler writes into test log.
func TestHandler(t testing.TB, fmtr log.Format) log.Handler {
	return log.FuncHandler(func(r *log.Record) error {
		t.Log(string(fmtr.Format(r)))
		return nil
	})
}
