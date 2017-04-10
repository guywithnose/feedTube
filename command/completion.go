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
	if lastParam == "--apiKey" || lastParam == "--filter" || lastParam == "--baseURL" || lastParam == "--after" {
		return
	}

	if lastParam == "--outputFolder" || lastParam == "--xmlFile" {
		fmt.Fprintln(c.App.Writer, "fileCompletion")
		return
	}

	if len(os.Args) > 2 {
		for _, flag := range c.App.Command(os.Args[1]).Flags {
			name := strings.Split(flag.GetName(), ",")[0]
			if !c.IsSet(name) {
				fmt.Fprintf(c.App.Writer, "--%s\n", name)
			}
		}
	}

	return
}
