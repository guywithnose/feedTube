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
func handleVideos(c *cli.Context, videos <-chan *Video, feedInfo *FeedInfo, outputFolder, xmlFile, baseURL string, cmdBuilder commandBuilder.Builder) error {
	filter := c.String("filter")

	rssVideos := make([]*Video, 0)
	for video := range videos {
		if (filter != "" && !strings.Contains(video.Title, filter)) || video.ID == "" {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, video.Filename)); os.IsNotExist(err) {
			downloadVideo(outputFolder, video.ID, video.Filename, cmdBuilder, c.App.Writer)
		}

		rssVideos = append(rssVideos, video)
	}

	if c.Bool("cleanupUnrelatedFiles") {
		err := cleanupUnrelatedFiles(rssVideos, outputFolder, c.App.ErrWriter)
		if err != nil {
			return err
		}
	}

	if xmlFile == "" {
		return nil
	}

	if baseURL == "" {
		return cli.NewExitError("You must specify an baseURL", 1)
	}

	rss := buildRssFile(rssVideos, baseURL, *feedInfo)

	return ioutil.WriteFile(xmlFile, rss, 0644)
}

func cleanupUnrelatedFiles(videos []*Video, outputFolder string, writer io.Writer) error {
	dir, _ := os.Open(outputFolder)
	files, _ := dir.Readdir(-1)

fileLoop:
	for _, file := range files {
		for _, video := range videos {
			if file.Name() == fmt.Sprintf("%s.mp3", video.Filename) {
				continue fileLoop
			}

		}

		filePath := (fmt.Sprintf("%s/%s", outputFolder, file.Name()))
		fmt.Fprintf(writer, "Removing file: %s\n", filePath)
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("Could not remove unrelated file: %v", err)
		}
	}

	return nil
}

func downloadVideo(outputFolder, videoID, fileName string, cmdBuilder commandBuilder.Builder, w io.Writer) bool {
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
		return false
	}

	return true
}
