package launcher

import (
	"os"
)

// AppControl is a struct of channels facilitating the interaction of a test
// harness with a u2u application instance.
type AppControl struct {
	// Upon a successful start of the u2u node, the node ID is sent to this
	// channel. The channel is closed when the process stops.
	NodeIdAnnouncement chan<- string
	// Upon a successful start of the u2u node, the HTTP port used by the HTTP
	// server is sent to this channel. The channel is closed when the process stops.
	HttpPortAnnouncement chan<- string
	// The process is stopped by sending a message through this channel, or by
	// closing it.
	Shutdown <-chan struct{}
}

func Run() error {
	return RunWithControl(nil)
}

// RunWithControl starts u2u with optional control channels for testing.
// If control is nil, it behaves like the regular Run() function.
func RunWithControl(control *AppControl) error {
	initApp()
	initAppHelp()

	// If we have control channels, set them up for use during execution
	if control != nil {
		appControl = control
	}

	return app.Run(os.Args)
}

// Global variable to hold the AppControl for access during node operation
var appControl *AppControl
