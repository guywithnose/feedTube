package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/guywithnose/commandBuilder"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCmdPlaylist(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				1,
			),
		},
	}
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			"video 1 output",
			"video 2 output",
			"Could not download t2-vId2: exit status 1",
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">`,
		`  <channel>`,
		`    <title>playlistTitle</title>`,
		`    <link></link>`,
		`    <description>playlistDescription</description>`,
		xmlLines[5],
		xmlLines[6],
		`    <item>`,
		`      <title>t</title>`,
		`      <link>http://foo.com/t-vId1.mp3</link>`,
		`      <description>d</description>`,
		`      <guid>vId1</guid>`,
		`      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>`,
		`    </item>`,
		`    <item>`,
		`      <title>t2</title>`,
		`      <link>http://foo.com/t2-vId2.mp3</link>`,
		`      <description>d2</description>`,
		`      <guid>vId2</guid>`,
		`      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>`,
		`    </item>`,
		`  </channel>`,
		`</rss>`,
	}
	assert.Equal(t, expectedXMLLines, xmlLines)
}

func TestCmdPlaylistFilter(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
		},
	}
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("filter", "t2", "doc")
	assert.Nil(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "video 2 output\n", writer.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">`,
		`  <channel>`,
		`    <title>playlistTitle</title>`,
		`    <link></link>`,
		`    <description>playlistDescription</description>`,
		xmlLines[5],
		xmlLines[6],
		`    <item>`,
		`      <title>t2</title>`,
		`      <link>http://foo.com/t2-vId2.mp3</link>`,
		`      <description>d2</description>`,
		`      <guid>vId2</guid>`,
		`      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>`,
		`    </item>`,
		`  </channel>`,
		`</rss>`,
	}
	assert.Equal(t, expectedXMLLines, xmlLines)
}

func TestCmdPlaylistNoBaseUrl(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				1,
			),
		},
	}
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	xmlFile := fmt.Sprintf("%s/xmlFile", outputFolder)
	set.String("xmlFile", xmlFile, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, writer, _ := appWithTestWriters()
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "You must specify an baseURL")
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			"video 1 output",
			"video 2 output",
			"Could not download t2-vId2: exit status 1",
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
}

func TestCmdPlaylistNoXmlFile(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				1,
			),
		},
	}
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, writer, _ := appWithTestWriters()
	assert.Nil(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			"video 1 output",
			"video 2 output",
			"Could not download t2-vId2: exit status 1",
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
}

func TestCmdPlaylistUsage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), `Usage: "feedTube playlist {playlistID}"`)
}

func TestCmdPlaylistNoOutputFolder(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "You must specify an outputFolder")
}

func TestCmdPlaylistNoApiKey(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	set.String("outputFolder", outputFolder, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "You must specify an apiKey")
}

func TestCmdPlaylistInvalidPlaylistID(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultPlaylistResponses()
	playlistInfo := youtube.PlaylistListResponse{Items: []*youtube.Playlist{}}
	bytes, _ := json.Marshal(playlistInfo)
	responses["/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "playlist awesome not found")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistYoutubePlaylistError(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestPlaylistServerOverrideResponse("/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet", "error")
	defer ts.Close()
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "Playlist request failed: googleapi: got HTTP response code 500 with body: ")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistYoutubeSearchPage1Error(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlistItems?alt=json&key=fakeApiKey&part=snippet&playlistId=awesome", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"Playlist items request failed: googleapi: got HTTP response code 500 with body: \n",
		&commandBuilder.Test{},
		CmdPlaylist,
	)
}

func TestCmdPlaylistYoutubeSearchPage2Error(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlistItems?alt=json&key=fakeApiKey&pageToken=page2&part=snippet&playlistId=awesome", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"Playlist items request failed: googleapi: got HTTP response code 500 with body: \n",
		&commandBuilder.Test{
			ExpectedCommands: []*commandBuilder.ExpectedCommand{
				commandBuilder.NewExpectedCommand(
					"",
					"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
					"video 1 output",
					0,
				),
			},
		},
		CmdPlaylist,
	)
}

func TestCmdPlaylistYoutubeSearchInvalidVideos(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultPlaylistResponses()
	playlistVideosPage1 := youtube.PlaylistItemListResponse{
		Items: []*youtube.PlaylistItem{
			{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-02T15:04:05Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "vId1",
					},
				},
			},
			{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "t2",
					Description: "d2",
					PublishedAt: "2006-01-02T15:04:05Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "",
					},
				},
			},
			{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "t2",
					Description: "d2",
					PublishedAt: "2006-01-02",
					ResourceId: &youtube.ResourceId{
						VideoId: "vId2",
					},
				},
			},
		},
	}
	bytes, _ := json.Marshal(playlistVideosPage1)
	responses["/playlistItems?alt=json&key=fakeApiKey&part=snippet&playlistId=awesome"] = string(bytes)

	ts := getTestServer(responses)
	youtubeAPIURLBase = ts.URL
	defer ts.Close()
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
		},
	}
	assert.Nil(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "video 1 output\n", writer.String())
}

func getTestPlaylistServerOverrideResponse(URL, response string) *httptest.Server {
	responses := getDefaultPlaylistResponses()
	responses[URL] = response
	server := getTestServer(responses)
	youtubeAPIURLBase = server.URL
	return server
}

func getDefaultPlaylistResponses() map[string]string {
	responses := map[string]string{}
	playlistInfo := youtube.PlaylistListResponse{
		Items: []*youtube.Playlist{
			{
				Snippet: &youtube.PlaylistSnippet{
					Title:       "playlistTitle",
					Description: "playlistDescription",
				},
			},
		},
	}
	bytes, _ := json.Marshal(playlistInfo)
	responses["/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet"] = string(bytes)

	playlistVideosPage1 := youtube.PlaylistItemListResponse{
		NextPageToken: "page2",
		Items: []*youtube.PlaylistItem{
			{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-02T15:04:05Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "vId1",
					},
				},
			},
		},
	}
	bytes, _ = json.Marshal(playlistVideosPage1)
	responses["/playlistItems?alt=json&key=fakeApiKey&part=snippet&playlistId=awesome"] = string(bytes)

	playlistVideosPage2 := youtube.PlaylistItemListResponse{
		Items: []*youtube.PlaylistItem{
			{
				Snippet: &youtube.PlaylistItemSnippet{
					Title:       "t2",
					Description: "d2",
					PublishedAt: "2006-01-02T15:04:05Z",
					ResourceId: &youtube.ResourceId{
						VideoId: "vId2",
					},
				},
			},
		},
	}
	bytes, _ = json.Marshal(playlistVideosPage2)
	responses["/playlistItems?alt=json&key=fakeApiKey&pageToken=page2&part=snippet&playlistId=awesome"] = string(bytes)

	return responses
}
