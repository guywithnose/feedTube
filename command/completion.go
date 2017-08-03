package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// Completion handles bash completion for the commands
func Completion(c *cli.Context) {
	lastParam := os.Args[len(os.Args)-2]
	noCompletionFlags := []string{"--apiKey", "--filter", "--baseURL", "--after"}
	if ContainsString(lastParam, noCompletionFlags) {
		return
	}

	fileCompletionFlags := []string{"--outputFolder", "--xmlFile"}
	if ContainsString(lastParam, fileCompletionFlags) {
		fmt.Fprintln(c.App.Writer, "fileCompletion")
		return
	}

	if len(os.Args) > 2 {
		completeFlags(c)
	}
}

func completeFlags(c *cli.Context) {
	for _, flag := range c.App.Command(os.Args[1]).Flags {
		name := strings.Split(flag.GetName(), ",")[0]
		if !c.IsSet(name) {
			fmt.Fprintf(c.App.Writer, "--%s\n", name)
		}
	}
}

// RootCompletion prints the list of root commands as the root completion method
// This is similar to the default method, but it excludes aliases
func RootCompletion(c *cli.Context) {
	for _, command := range c.App.Commands {
		if command.Hidden {
			continue
		}

		fmt.Fprintf(c.App.Writer, "%s:%s\n", command.Name, command.Usage)
	}
}
