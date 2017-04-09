package main

import (
	"fmt"
	"os"

	"github.com/guywithnose/commandBuilder"
	"github.com/guywithnose/feedTube/command"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "apiKey, k",
		Usage:  "The youtube api key",
		EnvVar: "YOUTUBE_APIKEY",
	},
	cli.StringFlag{
		Name:  "filter, f",
		Usage: "Only include videos whose titles contain a filter",
	},
	cli.StringFlag{
		Name:  "outputFolder, o",
		Usage: "The folder to save the audio files",
	},
	cli.StringFlag{
		Name:  "xmlFile, x",
		Usage: "The output rss file",
	},
	cli.StringFlag{
		Name:  "baseURL, b",
		Usage: "The base URL to access the output folder",
	},
}

// Commands defines the commands that can be called on hostBuilder
var Commands = []cli.Command{
	{
		Name:    "channel",
		Aliases: []string{"c"},
		Usage:   "Builds your rss file from a youtube channel",
		Action:  command.CmdChannel(commandBuilder.Real{}),
		Flags: append(
			flags,
			cli.StringFlag{
				Name:  "after",
				Usage: "Only process videos after a given date",
			},
		),
	},
	{
		Name:    "playlist",
		Aliases: []string{"p"},
		Usage:   "Builds your rss file from a youtube playlist",
		Action:  command.CmdPlaylist(commandBuilder.Real{}),
		Flags:   flags,
	},
}

// CommandNotFound runs when hostBuilder is invoked with an invalid command
func CommandNotFound(c *cli.Context, command string) {
	fmt.Fprintf(c.App.Writer, "%s: '%s' is not a %s command. See '%s --help'.", c.App.Name, command, c.App.Name, c.App.Name)
	os.Exit(2)
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
