package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/guywithnose/feedTube/rssBuilder"
	"github.com/urfave/cli"
)

// CmdBuild builds the hostfile from a configuration file
func handleVideos(c *cli.Context, videos []rssBuilder.Video, feedInfo *rssBuilder.FeedInfo) error {
	filter := c.String("filter")
	after := c.String("after")
	outputFolder := c.String("outputFolder")
	if outputFolder == "" {
		return cli.NewExitError("You must specify an outputFolder", 1)
	}

	if filter != "" {
		videos = filterVideos(videos, filter)
	}

	if after != "" {
		afterTime, err := time.Parse("01-02-06", after)
		if err != nil {
			return err
		}

		videos = videosAfter(videos, afterTime)
	}

	downloadVideos(videos, outputFolder)

	xmlFile := c.String("xmlFile")
	if xmlFile == "" {
		return nil
	}

	baseURL := c.String("baseURL")
	if baseURL == "" {
		return cli.NewExitError("You must specify an baseURL", 1)
	}

	rss, err := rssBuilder.BuildRssFile(videos, baseURL, *feedInfo)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(xmlFile, rss, 0644)
}

func downloadVideos(videos []rssBuilder.Video, outputFolder string) {
	for _, video := range videos {
		if video.ID == "" {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, video.Filename)); os.IsNotExist(err) {
			downloadVideo(outputFolder, video.ID, video.Filename)
		}
	}
}

func filterVideos(videos []rssBuilder.Video, filter string) []rssBuilder.Video {
	filteredVideos := make([]rssBuilder.Video, 0, len(videos))
	for _, video := range videos {
		if strings.Contains(video.Title, filter) {
			filteredVideos = append(filteredVideos, video)
		}
	}

	return filteredVideos
}

func videosAfter(videos []rssBuilder.Video, after time.Time) []rssBuilder.Video {
	filteredVideos := make([]rssBuilder.Video, 0, len(videos))
	for _, video := range videos {
		if video.Published.After(after) {
			filteredVideos = append(filteredVideos, video)
		}
	}

	return filteredVideos
}

func downloadVideo(outputFolder, videoID, fileName string) {
	out, err := exec.Command("/usr/bin/youtube-dl", "-x", "--audio-format", "mp3", "-o", fmt.Sprintf("%s/%s.%%(ext)s", outputFolder, fileName), videoID).Output()
	log.Println(string(out))
	log.Println(err)
}
