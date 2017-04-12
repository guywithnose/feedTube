package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/guywithnose/commandBuilder"
	"github.com/urfave/cli"
)

type youtubeItem struct {
	feeds.Item
	Filename string
}

func handleVideos(
	c *cli.Context,
	videos <-chan *youtubeItem,
	info *feeds.Feed,
	outputFolder,
	xmlFile,
	baseURL string,
	cmdBuilder commandBuilder.Builder,
) error {
	filter := c.String("filter")

	rssVideos := make([]*youtubeItem, 0)
	for video := range videos {
		if (filter != "" && !strings.Contains(video.Title, filter)) || video.Id == "" {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, video.Filename)); os.IsNotExist(err) {
			downloadVideo(outputFolder, video.Id, video.Filename, cmdBuilder, c.App.Writer)
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

	rss := buildRssFile(rssVideos, outputFolder, baseURL, info)

	return ioutil.WriteFile(xmlFile, []byte(rss), 0644)
}

func cleanupUnrelatedFiles(videos []*youtubeItem, outputFolder string, writer io.Writer) error {
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
		"--audio-quality",
		"0",
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

func buildRssFile(items []*youtubeItem, outputFolder, baseURL string, feed *feeds.Feed) string {
	feed.Updated = time.Now()
	feed.Link = &feeds.Link{}

	for _, item := range items {
		var length int64
		fileInfo, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, item.Filename))
		if err == nil {
			length = fileInfo.Size()
		}

		item.Link = &feeds.Link{
			Href:   fmt.Sprintf("%s/%s.mp3", baseURL, item.Filename),
			Type:   "audio/mpeg",
			Length: strconv.FormatInt(length, 10),
		}

		feed.Items = append(feed.Items, &item.Item)
	}

	xml, _ := feed.ToRss()
	return xml
}
