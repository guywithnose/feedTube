package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/guywithnose/commandBuilder"
	"github.com/urfave/cli"
)

// CmdBuild builds the hostfile from a configuration file
func handleVideos(c *cli.Context, videos []Video, feedInfo *FeedInfo, outputFolder, xmlFile, baseURL string, cmdBuilder commandBuilder.Builder) error {
	filter := c.String("filter")

	if filter != "" {
		videos = filterVideos(videos, filter)
	}

	downloadVideos(videos, outputFolder, cmdBuilder, c.App.Writer)

	if xmlFile == "" {
		return nil
	}

	if baseURL == "" {
		return cli.NewExitError("You must specify an baseURL", 1)
	}

	rss := buildRssFile(videos, baseURL, *feedInfo)

	return ioutil.WriteFile(xmlFile, rss, 0644)
}

func downloadVideos(videos []Video, outputFolder string, cmdBuilder commandBuilder.Builder, w io.Writer) {
	for _, video := range videos {
		if video.ID == "" {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, video.Filename)); os.IsNotExist(err) {
			downloadVideo(outputFolder, video.ID, video.Filename, cmdBuilder, w)
		}
	}
}

func filterVideos(videos []Video, filter string) []Video {
	filteredVideos := make([]Video, 0, len(videos))
	for _, video := range videos {
		if strings.Contains(video.Title, filter) {
			filteredVideos = append(filteredVideos, video)
		}
	}

	return filteredVideos
}

func downloadVideo(outputFolder, videoID, fileName string, cmdBuilder commandBuilder.Builder, w io.Writer) {
	params := []string{
		"/usr/bin/youtube-dl",
		"-x",
		"--audio-format",
		"mp3",
		"-o",
		fmt.Sprintf("%s/%s.%%(ext)s", outputFolder, fileName),
		fmt.Sprintf("https://youtu.be/%s", videoID),
	}
	cmd := cmdBuilder.CreateCommand(
		"",
		params...,
	)
	out, err := cmd.CombinedOutput()
	fmt.Fprintln(w, string(out))
	if err != nil {
		fmt.Fprintf(w, "Could not download %s: %v\nParams: '%s'\n", fileName, err, strings.Join(params, "' '"))
	}
}
