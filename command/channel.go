package command

import (
	"github.com/guywithnose/commandBuilder"
	"github.com/urfave/cli"
)

// CmdChannel builds an rss feed from a youtube channel
func CmdChannel(cmdBuilder commandBuilder.Builder) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.NArg() != 1 {
			return cli.NewExitError("Usage: \"feedTube channel {channelName|channelId}\"", 1)
		}

		outputFolder, apiKey, err := checkFlags(c)
		if err != nil {
			return err
		}

		channelName := c.Args().Get(0)

		videos, feedInfo, err := getVideosForChannel(apiKey, channelName, c.String("after"), c.App.ErrWriter)
		if err != nil {
			return err
		}

		return handleVideos(c, videos, feedInfo, outputFolder, c.String("xmlFile"), c.String("baseURL"), cmdBuilder)
	}
}

func checkFlags(c *cli.Context) (string, string, error) {
	outputFolder := c.String("outputFolder")
	if outputFolder == "" {
		return "", "", cli.NewExitError("You must specify an outputFolder", 1)
	}

	apiKey := c.String("apiKey")
	if apiKey == "" {
		return "", "", cli.NewExitError("You must specify an apiKey", 1)
	}

	return outputFolder, apiKey, nil
}
