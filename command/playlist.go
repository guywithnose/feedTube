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

		outputFolder, apiKey, err := checkFlags(c)
		if err != nil {
			return err
		}

		playlistID := c.Args().Get(0)

		videos, channel, err := getVideosForPlaylist(apiKey, playlistID, c.App.ErrWriter)
		if err != nil {
			return err
		}

		xmlFileName := c.String("xmlFile")
		downloadedFiles, err := handleVideos(c, videos, channel, outputFolder, xmlFileName, c.String("baseURL"), cmdBuilder)
		if err != nil {
			return err
		}

		if c.Bool("cleanupUnrelatedFiles") {
			err := cleanupUnrelatedFiles(downloadedFiles, outputFolder, xmlFileName, c.App.ErrWriter)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
