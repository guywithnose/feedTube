package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/guywithnose/runner"
)

// Downloader downloads youtube videos
type Downloader struct {
	cmdBuilder   runner.Builder
	outputFolder string
	quality      string
}

// NewDownloader returns a new Downloader
func NewDownloader(cmdBuilder runner.Builder, outputFolder string, quality string) *Downloader {
	return &Downloader{
		cmdBuilder:   cmdBuilder,
		outputFolder: outputFolder,
		quality:      quality,
	}
}

// DownloadVideos downloads any items that are not already in outputfolder
func (downloader Downloader) DownloadVideos(items []*VideoData) error {
	for _, item := range items {
		if fileExists(getFileName(downloader.outputFolder, item)) {
			continue
		}

		err := downloader.downloadVideo(item.GUID, item.FileName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (downloader Downloader) downloadVideo(videoID, fileName string) error {
	params := []string{
		"/usr/bin/youtube-dl",
		"-x",
		"--audio-format",
		"mp3",
		"--audio-quality",
		downloader.quality,
		"-o",
		fmt.Sprintf("%s/%s.%%(ext)s", downloader.outputFolder, fileName),
		fmt.Sprintf("https://youtu.be/%s", videoID),
	}
	cmd := downloader.cmdBuilder.New("", params...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not download %s: %v\nParams: '%s': %s", fileName, err, strings.Join(params, "' '"), string(out))
	}

	return nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
