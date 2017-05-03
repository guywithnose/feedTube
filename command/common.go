package command

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/guywithnose/commandBuilder"
	"github.com/urfave/cli"
)

func checkFlags(c *cli.Context) error {
	if c.String("outputFolder") == "" {
		return cli.NewExitError("You must specify an outputFolder", 1)
	}

	if c.String("apiKey") == "" {
		return cli.NewExitError("You must specify an apiKey", 1)
	}

	if c.String("xmlFile") != "" && c.String("baseURL") == "" {
		return cli.NewExitError("You must specify an baseURL", 1)
	}

	return nil
}

func getGenerator() string {
	return fmt.Sprintf("%s v%s (github.com/guywithnose/feedTube)", Name, Version)
}

func getRelatedFiles(items []*VideoData, xmlFile, outputFolder string) []string {
	relatedFiles := make([]string, 0, len(items)+1)
	absoluteXMLFileName, err := filepath.Abs(xmlFile)
	if err == nil {
		relatedFiles = append(relatedFiles, absoluteXMLFileName)
	}

	for _, item := range items {
		filePath, err := filepath.Abs(fmt.Sprintf("%s/%s.mp3", outputFolder, item.FileName))
		if err == nil {
			relatedFiles = append(relatedFiles, filePath)
		}
	}

	return relatedFiles
}

func filterItems(filter string, items []*VideoData) []*VideoData {
	filteredItems := make([]*VideoData, 0, len(items))
	for _, item := range items {
		if (filter == "" || strings.Contains(item.Title, filter)) && item.GUID != "" {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

// Build downloads the videos and builds the feed XML
func Build(c *cli.Context, cmdBuilder commandBuilder.Builder, items []*VideoData, info *ChannelInfo) error {
	if c.String("filter") != "" {
		items = filterItems(c.String("filter"), items)
	}

	err := NewDownloader(cmdBuilder, c.String("outputFolder")).DownloadVideos(items)
	if err != nil {
		return err
	}

	if c.String("xmlFile") != "" {
		err := NewXMLBuilder(cmdBuilder, c.String("xmlFile"), c.String("outputFolder"), c.String("baseURL"), getGenerator(), info).BuildRss(items)
		if err != nil {
			return err
		}
	}

	if c.Bool("cleanupUnrelatedFiles") {
		relatedFiles := getRelatedFiles(items, c.String("xmlFile"), c.String("outputFolder"))
		return NewDirectoryCleaner(c.String("outputFolder")).CleanupUnrelatedFiles(relatedFiles, c.App.ErrWriter)
	}

	return nil
}

// ContainsString searches a string slice to see if it contains a given string
func ContainsString(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}

func getFileName(outputFolder string, item *VideoData) string {
	return fmt.Sprintf("%s/%s.mp3", outputFolder, item.FileName)
}
