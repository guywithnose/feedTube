package command

import (
	"github.com/guywithnose/feedTube/youtubeScraper"
	"github.com/urfave/cli"
)

// CmdChannel builds an rss feed from a youtube channel
func CmdChannel(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.NewExitError("Usage: \"feedTube channel {channelName}\"", 1)
	}

	apiKey := c.String("apiKey")
	if apiKey == "" {
		return cli.NewExitError("You must specify an apiKey", 1)
	}
	channelName := c.Args().Get(0)

	videos, feedInfo, err := youtubeScraper.GetVideosForChannel(apiKey, channelName, c.App.ErrWriter)
	if err != nil {
		return err
	}

	return handleVideos(c, videos, feedInfo)
}
