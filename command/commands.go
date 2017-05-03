package command

import (
	"github.com/guywithnose/runner"
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
	cli.BoolFlag{
		Name: "cleanupUnrelatedFiles",
		Usage: "Delete any files in the output folder not related to the current feed. " +
			"This can be useful if you are maintaining a playlist and you want to remove old files when you remove them from the playlist. " +
			"DO NOT use this if multiple feeds are sharing an output folder.",
	},
	cli.StringFlag{
		Name:  "overrideTitle, t",
		Usage: "Manually set the feed title",
	},
}

// Commands defines the commands that can be called on hostBuilder
var Commands = []cli.Command{
	{
		Name:         "channel",
		Aliases:      []string{"c"},
		Usage:        "Builds your rss file from a youtube channel",
		Action:       CmdChannel(runner.Real{}),
		BashComplete: Completion,
		Flags: append(
			flags,
			cli.StringFlag{
				Name:  "after, a",
				Usage: "Only process videos after a given date",
			},
		),
	},
	{
		Name:         "playlist",
		Aliases:      []string{"p"},
		Usage:        "Builds your rss file from a youtube playlist",
		Action:       CmdPlaylist(runner.Real{}),
		BashComplete: Completion,
		Flags:        flags,
	},
}
