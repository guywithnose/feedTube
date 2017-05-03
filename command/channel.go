package command

import (
	"github.com/guywithnose/runner"
	"github.com/urfave/cli"
)

// CmdChannel builds an rss feed from a youtube channel
func CmdChannel(cmdBuilder runner.Builder) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() != 1 {
			return cli.NewExitError("Usage: \"feedTube channel {channelName|channelId}\"", 1)
		}

		err := checkFlags(c)
		if err != nil {
			return err
		}

		channelName := c.Args().Get(0)
		items, channelInfo, err := NewChannelScraper(c.String("apiKey")).GetVideosForChannel(channelName, c.String("after"))
		if err != nil {
			return err
		}

		if c.String("overrideTitle") != "" {
			channelInfo.Title = c.String("overrideTitle")
		}

		return Build(c, cmdBuilder, items, channelInfo)
	}
}
