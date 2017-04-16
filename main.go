package main

import (
	"os"

	"github.com/guywithnose/feedTube/command"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = command.Name
	app.Version = command.Version
	app.Author = "Robert Bittle"
	app.Email = "guywithnose@gmail.com"
	app.Usage = "feedTube build channelName"

	app.Commands = Commands
	app.CommandNotFound = CommandNotFound
	app.EnableBashCompletion = true
	app.BashComplete = RootCompletion
	app.ErrWriter = os.Stderr

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
