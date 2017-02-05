package command

import (
	"github.com/guywithnose/feedTube/youtubeScraper"
	"github.com/urfave/cli"
)

// CmdPlaylist builds an rss feed from a youtube playlist
func CmdPlaylist(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.NewExitError("Usage: \"feedTube playlist {playlistID}\"", 1)
	}

	apiKey := c.String("apiKey")
	if apiKey == "" {
		return cli.NewExitError("You must specify an apiKey", 1)
	}
	playlistID := c.Args().Get(0)

	videos, channel, err := youtubeScraper.GetVideosForPlaylist(apiKey, playlistID, c.App.ErrWriter)
	if err != nil {
		return err
	}

	return handleVideos(c, videos, channel)
}
