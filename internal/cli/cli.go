// Package cli selects a dx command or the arguments to pass to claude.
package cli

// Command is the result of Parse.
type Command struct {
	// Name is the dx command. It is empty when the arguments go to claude.
	Name string
	// Args are the arguments of the dx command, or the arguments for claude.
	Args []string
}

// commands maps the first argument to a dx command name.
var commands = map[string]string{
	"--help":    "help",
	"-h":        "help",
	"--version": "version",
	"update":    "update",
	"cache":     "cache",
	"doctor":    "doctor",
	"report":    "report",
	"mcp":       "mcp",
	"hook":      "hook",
}

// Parse selects the command from the first argument only (ADR-003).
// The argument "--" passes all later arguments to claude.
func Parse(args []string) Command {
	if len(args) == 0 {
		return Command{}
	}
	if args[0] == "--" {
		return Command{Args: args[1:]}
	}
	name, ok := commands[args[0]]
	if !ok {
		return Command{Args: args}
	}
	return Command{Name: name, Args: args[1:]}
}
