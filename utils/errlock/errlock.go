package errlock

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/utils/caution"
)

var datadir string

// Check if errlock is written
func Check() {
	locked, reason, eLockPath, err := read(datadir)
	if err != nil {
		// This is a user-facing error, so we want to provide a clear message.
		//nolint:staticcheck // ST1005: allow capitalized error message and punctuation
		utils.Fatalf("Node isn't allowed to start due to an error reading"+
			" the lock file %s.\n Please fix the issue. Error message:\n%s",
			eLockPath, err)
	}

	if locked {
		// This is a user-facing error, so we want to provide a clear message.
		//nolint:staticcheck // ST1005: allow capitalized error message and punctuation
		utils.Fatalf("Node isn't allowed to start due to a previous error."+
			" Please fix the issue and then delete file \"%s\". Error message:\n%s",
			eLockPath, reason)
	}
}

// SetDefaultDatadir for errlock files
func SetDefaultDatadir(dir string) {
	datadir = dir
}

// Permanent error
func Permanent(err error) {
	eLockPath, _ := write(datadir, err.Error())
	// This is a user-facing error, so we want to provide a clear message.
	//nolint:staticcheck // ST1005: allow capitalized error message and punctuation
	utils.Fatalf("Node is permanently stopping due to an issue. Please fix"+
		" the issue and then delete file \"%s\". Error message:\n%s",
		eLockPath, err.Error())
}

func readAll(reader io.Reader, max int) ([]byte, error) {
	buf := make([]byte, max)
	consumed := 0
	for {
		n, err := reader.Read(buf[consumed:])
		consumed += n
		if consumed == len(buf) || err == io.EOF {
			return buf[:consumed], nil
		}
		if err != nil {
			return nil, err
		}
	}
}

// read errlock file
func read(dir string) (bool, string, string, error) {
	eLockPath := path.Join(dir, "errlock")

	data, err := os.Open(eLockPath)
	if err != nil {
		return false, "", eLockPath, err
	}
	defer caution.CloseAndReportError(&err, data, "Failed to close errlock file")

	// read no more than N bytes
	maxFileLen := 5000
	eLockBytes, err := readAll(data, maxFileLen)
	if err != nil {
		return true, "", eLockPath, fmt.Errorf("failed to read lock file %v: %w", eLockPath, err)
	}
	return true, string(eLockBytes), eLockPath, nil
}

// write errlock file
func write(dir string, eLockStr string) (string, error) {
	eLockPath := path.Join(dir, "errlock")

	return eLockPath, ioutil.WriteFile(eLockPath, []byte(eLockStr), 0666) // assume no custom encoding needed
}
