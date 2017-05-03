package command

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/guywithnose/runner"
)

// XMLBuilder handles building RSS XML
type XMLBuilder struct {
	cmdBuilder   runner.Builder
	xmlFileName  string
	outputFolder string
	baseURL      string
	generator    string
	feed         *podcast.Podcast
}

// NewXMLBuilder returns a new XMLBuilder
func NewXMLBuilder(cmdBuilder runner.Builder, xmlFileName, outputFolder, baseURL, generator string, channelInfo *ChannelInfo) *XMLBuilder {
	now := time.Now()
	feed := podcast.New(channelInfo.Title, channelInfo.Link, channelInfo.Description, &now, &now)
	if channelInfo.Thumbnail != "" {
		feed.AddImage(channelInfo.Thumbnail)
	}

	return &XMLBuilder{
		cmdBuilder:   cmdBuilder,
		xmlFileName:  xmlFileName,
		outputFolder: outputFolder,
		baseURL:      baseURL,
		generator:    generator,
		feed:         &feed,
	}
}

// BuildRss builds an RSS XML feed from an list of VideoData
func (xmlBuilder XMLBuilder) BuildRss(items []*VideoData) error {
	xmlBuilder.appendDataToFeed()
	its := xmlBuilder.buildItems(items)
	return xmlBuilder.buildXML(its)
}

func (xmlBuilder XMLBuilder) appendDataToFeed() {
	now := time.Now()
	xmlBuilder.feed.AddLastBuildDate(&now)
	xmlBuilder.feed.Generator = xmlBuilder.generator
}

func (xmlBuilder XMLBuilder) buildItems(items []*VideoData) []*podcast.Item {
	its := make([]*podcast.Item, 0, len(items))
	for _, item := range items {
		it := &podcast.Item{
			GUID:        item.GUID,
			Link:        item.Link,
			Title:       item.Title,
			Description: item.Description,
		}

		it.AddImage(item.Image)
		it.AddPubDate(&item.PubDate)

		length, err := getFileSize(getFileName(xmlBuilder.outputFolder, item))
		// Fail silently since duration is not important
		if err == nil {
			duration, durationErr := xmlBuilder.getFileDuration(item)
			if durationErr == nil {
				it.IDuration = duration
			}
		}

		it.AddEnclosure(xmlBuilder.getFileURL(item), podcast.MP3, length)

		its = append(its, it)
	}

	return its
}

func getFileSize(fileName string) (int64, error) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (xmlBuilder XMLBuilder) getFileDuration(item *VideoData) (string, error) {
	cmd := xmlBuilder.cmdBuilder.New("", "/usr/bin/ffprobe", getFileName(xmlBuilder.outputFolder, item))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return parseDurationOutput(out)
}

func parseDurationOutput(out []byte) (string, error) {
	durationRegex, _ := regexp.Compile(`.*Duration: (\d\d:\d\d:\d\d)\.\d\d, start.*`)
	matches := durationRegex.FindSubmatch(out)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse duration from output: %s", out)
	}

	return string(matches[1]), nil
}

func (xmlBuilder XMLBuilder) getFileURL(item *VideoData) string {
	return fmt.Sprintf("%s/%s.mp3", xmlBuilder.baseURL, item.FileName)
}

func (xmlBuilder XMLBuilder) buildXML(items []*podcast.Item) error {
	for _, item := range items {
		err := xmlBuilder.addItemToFeed(item)
		if err != nil {
			return err
		}
	}

	return xmlBuilder.writeToFile()
}

func (xmlBuilder XMLBuilder) addItemToFeed(item *podcast.Item) error {
	numItems, err := xmlBuilder.feed.AddItem(*item)
	if err != nil {
		return fmt.Errorf("could not parse item to xml: %v", err)
	}

	// podcast library sets the GUID as the link
	xmlBuilder.feed.Items[numItems-1].GUID = item.GUID
	return nil
}

func (xmlBuilder XMLBuilder) writeToFile() error {
	xmlFile, err := os.Create(xmlBuilder.xmlFileName)
	if err != nil {
		return err
	}

	return xmlBuilder.feed.Encode(xmlFile)
}
