package command_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/guywithnose/commandBuilder"
	"github.com/guywithnose/feedTube/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMLBuilder(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	xmlFileName := fmt.Sprintf("%s/xmlFile", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"foo\nDuration: 02:13:45.22, startskskdjdk\ndkskskd",
				0,
			),
		},
	}
	xmlBuilder := command.NewXMLBuilder(
		cb,
		xmlFileName,
		outputFolder,
		"http://foo.com",
		"feedTube v0.3.0 (github.com/guywithnose/feedTube)",
		&command.ChannelInfo{
			Title:       "t",
			Description: "d",
			Link:        "https://www.youtube.com/channel/awesomeChannelId",
			Thumbnail:   "https://images.com/thumb.jpg",
		},
	)
	err = xmlBuilder.BuildRss(
		[]*command.VideoData{
			{
				GUID:        "vId1",
				Link:        "https://youtu.be/vId1",
				Title:       "t",
				Description: "d https://youtu.be/vId1",
				FileName:    "t-vId1",
				Image:       "https://images.com/vid1Thumb.jpg",
				PubDate:     time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
			},
			{
				GUID:        "vId2",
				Link:        "https://youtu.be/vId2",
				Title:       "t2",
				Description: "d2 https://youtu.be/vId2",
				FileName:    "t2-vId2",
				Image:       "https://images.com/thumb.jpg",
				PubDate:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			},
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(xmlFileName)
	require.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXML := getExpectedChannelXML(xmlLines[8:10])
	expectedXML = append(expectedXML[:22], append([]string{`      <itunes:duration>02:13:45</itunes:duration>`}, expectedXML[22:]...)...)
	assert.Equal(t, expectedXML, xmlLines)
}

func TestXMLBuilderGetDurationFailure(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	xmlFileName := fmt.Sprintf("%s/xmlFile", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"",
				1,
			),
		},
	}
	xmlBuilder := command.NewXMLBuilder(
		cb,
		xmlFileName,
		outputFolder,
		"http://foo.com",
		"feedTube v0.3.0 (github.com/guywithnose/feedTube)",
		&command.ChannelInfo{
			Title:       "t",
			Description: "d",
			Link:        "https://www.youtube.com/channel/awesomeChannelId",
			Thumbnail:   "https://images.com/thumb.jpg",
		},
	)
	err = xmlBuilder.BuildRss(
		[]*command.VideoData{
			{
				GUID:        "vId1",
				Link:        "https://youtu.be/vId1",
				Title:       "t",
				Description: "d https://youtu.be/vId1",
				FileName:    "t-vId1",
				Image:       "https://images.com/vid1Thumb.jpg",
				PubDate:     time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
			},
			{
				GUID:        "vId2",
				Link:        "https://youtu.be/vId2",
				Title:       "t2",
				Description: "d2 https://youtu.be/vId2",
				FileName:    "t2-vId2",
				Image:       "https://images.com/thumb.jpg",
				PubDate:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			},
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	xmlBytes, err := ioutil.ReadFile(xmlFileName)
	require.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXML := getExpectedChannelXML(xmlLines[8:10])
	assert.Equal(t, expectedXML, xmlLines)
}

func TestXMLBuilderInvalidtDuration(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	xmlFileName := fmt.Sprintf("%s/xmlFile", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"foo\nDuration: 02:1E:45.22, startskskdjdk\ndkskskd",
				0,
			),
		},
	}
	xmlBuilder := command.NewXMLBuilder(
		cb,
		xmlFileName,
		outputFolder,
		"http://foo.com",
		"feedTube v0.3.0 (github.com/guywithnose/feedTube)",
		&command.ChannelInfo{
			Title:       "t",
			Description: "d",
			Link:        "https://www.youtube.com/channel/awesomeChannelId",
			Thumbnail:   "https://images.com/thumb.jpg",
		},
	)
	err = xmlBuilder.BuildRss(
		[]*command.VideoData{
			{
				GUID:        "vId1",
				Link:        "https://youtu.be/vId1",
				Title:       "t",
				Description: "d https://youtu.be/vId1",
				FileName:    "t-vId1",
				Image:       "https://images.com/vid1Thumb.jpg",
				PubDate:     time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
			},
			{
				GUID:        "vId2",
				Link:        "https://youtu.be/vId2",
				Title:       "t2",
				Description: "d2 https://youtu.be/vId2",
				FileName:    "t2-vId2",
				Image:       "https://images.com/thumb.jpg",
				PubDate:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			},
		},
	)
	assert.Nil(t, err)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(xmlFileName)
	require.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXML := getExpectedChannelXML(xmlLines[8:10])
	assert.Equal(t, expectedXML, xmlLines)
}

func TestXMLBuilderInvalidtXmlFile(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	xmlFileName := filepath.Join(outputFolder, "notadir", "xmlFile")
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"foo\nDuration: 02:1E:45.22, startskskdjdk\ndkskskd",
				0,
			),
		},
	}
	xmlBuilder := command.NewXMLBuilder(
		cb,
		xmlFileName,
		outputFolder,
		"http://foo.com",
		"feedTube v0.3.0 (github.com/guywithnose/feedTube)",
		&command.ChannelInfo{
			Title:       "t",
			Description: "d",
			Link:        "https://www.youtube.com/channel/awesomeChannelId",
			Thumbnail:   "https://images.com/thumb.jpg",
		},
	)
	err = xmlBuilder.BuildRss(
		[]*command.VideoData{
			{
				GUID:        "vId1",
				Link:        "https://youtu.be/vId1",
				Title:       "t",
				Description: "d https://youtu.be/vId1",
				FileName:    "t-vId1",
				Image:       "https://images.com/vid1Thumb.jpg",
				PubDate:     time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
			},
			{
				GUID:        "vId2",
				Link:        "https://youtu.be/vId2",
				Title:       "t2",
				Description: "d2 https://youtu.be/vId2",
				FileName:    "t2-vId2",
				Image:       "https://images.com/thumb.jpg",
				PubDate:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			},
		},
	)
	assert.EqualError(t, err, "open /tmp/testFeedTube/notadir/xmlFile: no such file or directory")
}

func TestXMLBuilderInvalidVideo(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	xmlFileName := fmt.Sprintf("%s/xmlFile", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"foo\nDuration: 02:13:45.22, startskskdjdk\ndkskskd",
				0,
			),
		},
	}
	xmlBuilder := command.NewXMLBuilder(
		cb,
		xmlFileName,
		outputFolder,
		"http://foo.com",
		"feedTube v0.3.0 (github.com/guywithnose/feedTube)",
		&command.ChannelInfo{
			Title:       "t",
			Description: "d",
			Link:        "https://www.youtube.com/channel/awesomeChannelId",
			Thumbnail:   "https://images.com/thumb.jpg",
		},
	)
	err = xmlBuilder.BuildRss(
		[]*command.VideoData{
			{
				GUID:        "vId1",
				Link:        "https://youtu.be/vId1",
				Title:       "",
				Description: "d https://youtu.be/vId1",
				FileName:    "t-vId1",
				Image:       "https://images.com/vid1Thumb.jpg",
				PubDate:     time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
			},
			{
				GUID:        "vId2",
				Link:        "https://youtu.be/vId2",
				Title:       "t2",
				Description: "d2 https://youtu.be/vId2",
				FileName:    "t2-vId2",
				Image:       "https://images.com/thumb.jpg",
				PubDate:     time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			},
		},
	)
	assert.EqualError(t, err, "could not parse item to xml: Title and Description are reuired")
}
