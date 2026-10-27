package launcher

import "os"

// Ctrl+C goes to all processes of the console, so claude gets it. Windows cannot send other signals.
var (
	ignored   = []os.Signal{os.Interrupt}
	forwarded []os.Signal
)
