package command_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/guywithnose/feedTube/command"
	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func removeFile(t *testing.T, fileName string) {
	assert.Nil(t, os.RemoveAll(fileName))
}

func TestHelperProcess(*testing.T) {
	runner.ErrorCodeHelper()
}

func getTestServer(responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, ok := responses[r.URL.String()]
		if !ok {
			panic(r.URL.String())
		}

		if response == "error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, response)
	}))
}

func appWithTestWriters() (*cli.App, *bytes.Buffer, *bytes.Buffer) {
	app := cli.NewApp()
	writer := new(bytes.Buffer)
	errWriter := new(bytes.Buffer)
	app.Writer = writer
	app.ErrWriter = errWriter
	return app, writer, errWriter
}

func getBaseAppAndFlagSet(t *testing.T, outputFolder string) (*cli.App, *bytes.Buffer, *bytes.Buffer, *flag.FlagSet) {
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	xmlFile := fmt.Sprintf("%s/xmlFile", outputFolder)
	set.String("xmlFile", xmlFile, "doc")
	set.String("baseURL", "http://foo.com", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, writer, errorWriter := appWithTestWriters()
	return app, writer, errorWriter, set
}

func runErrorTest(
	t *testing.T,
	expectedError string,
	cb *runner.Test,
	cmdFunc func(runner.Builder) func(*cli.Context) error,
) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(t, cmdFunc(cb)(cli.NewContext(app, set, nil)), expectedError)
	assert.Equal(t, []error(nil), cb.Errors)
}

func getDefaultChannelResponses() map[string]string {
	responses := map[string]string{}
	searchPage1 := youtube.SearchListResponse{
		NextPageToken: "page2",
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-03T15:04:05Z",
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vidLiveThumb.jpg",
						},
					},
					LiveBroadcastContent: "live",
				},
				Id: &youtube.ResourceId{
					VideoId: "vIdLive",
				},
			},
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-02T15:04:05Z",
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vid1Thumb.jpg",
						},
					},
					LiveBroadcastContent: "none",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video"] = string(bytes)

	searchPage2 := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:                "t2",
					Description:          "d2",
					PublishedAt:          "2006-01-02T15:04:05Z",
					LiveBroadcastContent: "none",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId2",
				},
			},
		},
	}
	bytes, _ = json.Marshal(searchPage2)
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&pageToken=page2&part=snippet&type=video"] = string(bytes)

	channelInfo := youtube.ChannelListResponse{
		Items: []*youtube.Channel{
			{
				Snippet: &youtube.ChannelSnippet{
					Title:       "t",
					Description: "d",
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/thumb.jpg",
						},
					},
				},
				Id: "awesomeChannelId",
			},
		},
	}
	bytes, _ = json.Marshal(channelInfo)
	responses["/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	responses["/channels?alt=json&id=awesomeChannelId&key=fakeApiKey&part=snippet"] = string(bytes)

	searchPage1WithAfter := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-02T15:04:05Z",
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vid1Thumb.jpg",
						},
					},
					LiveBroadcastContent: "none",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId1",
				},
			},
		},
	}
	bytes, _ = json.Marshal(searchPage1WithAfter)
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&publishedAfter=2006-07-07T00%3A00%3A00Z&type=video"] =
		string(bytes)

	channelIDInfo := youtube.ChannelListResponse{Items: []*youtube.Channel{}}
	bytes, _ = json.Marshal(channelIDInfo)
	responses["/channels?alt=json&id=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	responses["/channels?alt=json&forUsername=awesomeChannelId&key=fakeApiKey&part=snippet"] = string(bytes)
	return responses
}

func getTestChannelServerOverrideResponse(URL string) *httptest.Server {
	responses := getDefaultChannelResponses()
	responses[URL] = "error"
	server := getTestServer(responses)
	command.YoutubeAPIURLBase = server.URL
	return server
}

func getExpectedChannelXML(dateLine []string) []string {
	return []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`  <channel>`,
		`    <title>t</title>`,
		`    <link>https://www.youtube.com/channel/awesomeChannelId</link>`,
		`    <description>d</description>`,
		fmt.Sprintf(`    <generator>feedTube v%s (github.com/guywithnose/feedTube)</generator>`, command.Version),
		`    <language>en-us</language>`,
		dateLine[0],
		dateLine[1],
		`    <image>`,
		`      <url>https://images.com/thumb.jpg</url>`,
		`    </image>`,
		`    <itunes:image href="https://images.com/thumb.jpg"></itunes:image>`,
		`    <item>`,
		`      <guid>vId1</guid>`,
		`      <title>t</title>`,
		`      <link>https://youtu.be/vId1</link>`,
		`      <description>d https://youtu.be/vId1</description>`,
		`      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>`,
		`      <enclosure url="http://foo.com/t-vId1.mp3" length="0" type="audio/mpeg"></enclosure>`,
		`      <itunes:image href="https://images.com/vid1Thumb.jpg"></itunes:image>`,
		`    </item>`,
		`    <item>`,
		`      <guid>vId2</guid>`,
		`      <title>t2</title>`,
		`      <link>https://youtu.be/vId2</link>`,
		`      <description>d2 https://youtu.be/vId2</description>`,
		`      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>`,
		`      <enclosure url="http://foo.com/t2-vId2.mp3" length="0" type="audio/mpeg"></enclosure>`,
		`      <itunes:image href="https://images.com/thumb.jpg"></itunes:image>`,
		`    </item>`,
		`  </channel>`,
		`</rss>`,
	}
}
