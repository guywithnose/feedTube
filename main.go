package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/guywithnose/feedTube/command"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = command.Name
	app.Version = fmt.Sprintf("%s-%s", command.Version, runtime.Version())
	app.Author = "Robert Bittle"
	app.Email = "guywithnose@gmail.com"
	app.Usage = "feedTube channel channelName"

	app.Commands = command.Commands
	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
		os.Exit(2)
	}

	app.EnableBashCompletion = true
	app.BashComplete = command.RootCompletion
	app.ErrWriter = os.Stderr

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
