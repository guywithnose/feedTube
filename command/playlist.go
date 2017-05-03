package command

import (
	"github.com/guywithnose/commandBuilder"
	"github.com/urfave/cli"
)

// CmdPlaylist builds an rss feed from a youtube playlist
func CmdPlaylist(cmdBuilder commandBuilder.Builder) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() != 1 {
			return cli.NewExitError("Usage: \"feedTube playlist {playlistID}\"", 1)
		}

		err := checkFlags(c)
		if err != nil {
			return err
		}

		playlistID := c.Args().Get(0)
		items, playlistInfo, err := NewPlaylistScraper(c.String("apiKey")).GetVideosForPlaylist(playlistID)
		if err != nil {
			return err
		}

		if c.String("overrideTitle") != "" {
			playlistInfo.Title = c.String("overrideTitle")
		}

		return Build(c, cmdBuilder, items, playlistInfo)
	}
}
