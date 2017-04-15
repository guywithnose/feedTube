package command

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/guywithnose/commandBuilder"
	"github.com/urfave/cli"
)

type youtubeItem struct {
	podcast.Item
	Filename string
}

func handleVideos(
	c *cli.Context,
	videos <-chan *youtubeItem,
	feed *podcast.Podcast,
	outputFolder,
	xmlFileName,
	baseURL string,
	cmdBuilder commandBuilder.Builder,
) ([]string, error) {
	canBuildXML := false
	if xmlFileName != "" {
		if baseURL == "" {
			return nil, cli.NewExitError("You must specify an baseURL", 1)
		}

		canBuildXML = true
	}

	filter := c.String("filter")
	now := time.Now()
	feed.AddLastBuildDate(&now)
	feed.Generator = ""

	downloadedFiles := make([]string, 0, len(videos))
	for item := range videos {
		if (filter != "" && !strings.Contains(item.Title, filter)) || item.GUID == "" {
			continue
		}

		handleItem(outputFolder, baseURL, item, feed, canBuildXML, cmdBuilder, c.App.Writer, c.App.ErrWriter)
		downloadedFiles = append(downloadedFiles, fmt.Sprintf("%s.mp3", item.Filename))
	}

	if canBuildXML {
		xmlFile, err := os.Create(xmlFileName)
		if err != nil {
			return nil, err
		}

		return downloadedFiles, feed.Encode(xmlFile)
	}

	return downloadedFiles, nil
}

func handleItem(
	outputFolder,
	baseURL string,
	item *youtubeItem,
	feed *podcast.Podcast,
	canBuildXML bool,
	cmdBuilder commandBuilder.Builder,
	writer io.Writer,
	errWriter io.Writer,
) {
	if _, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, item.Filename)); os.IsNotExist(err) {
		downloadVideo(outputFolder, item.GUID, item.Filename, cmdBuilder, writer)
	}

	if canBuildXML {
		var length int64
		fileInfo, err := os.Stat(fmt.Sprintf("%s/%s.mp3", outputFolder, item.Filename))
		if err == nil {
			length = fileInfo.Size()
		}
		link := fmt.Sprintf("%s/%s.mp3", baseURL, item.Filename)

		item.AddEnclosure(link, podcast.MP3, length)

		numItems, err := feed.AddItem(item.Item)
		if err != nil {
			fmt.Fprintf(errWriter, "Could not parse item to xml: %v", err)
		}

		// podcast library wants to set the GUID as the link
		feed.Items[numItems-1].GUID = item.Item.GUID
	}
}

func cleanupUnrelatedFiles(downloadedFiles []string, outputFolder, xmlFileName string, writer io.Writer) error {
	dir, _ := os.Open(outputFolder)
	files, _ := dir.Readdir(-1)
	// TODO get absolute path of xmlFileName

fileLoop:
	for _, file := range files {
		for _, downloadedFile := range downloadedFiles {
			if file.Name() == downloadedFile || fmt.Sprintf("%s/%s", outputFolder, file.Name()) == xmlFileName {
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
