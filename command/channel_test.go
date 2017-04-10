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

func TestCmdChannel(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				1,
			),
		},
	}
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			"video 1 output",
			"video 2 output",
			"Could not download t2-vId2: exit status 1",
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedXML(xmlLines[4]), xmlLines)
}

func TestCmdChannelNoRedownload(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	_, err := os.Create(fmt.Sprintf("%s/t-vId1.mp3", outputFolder))
	assert.Nil(t, err)
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				1,
			),
		},
	}
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			"video 2 output",
			"Could not download t2-vId2: exit status 1",
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedXML(xmlLines[4]), xmlLines)
}

func TestCmdChannelUsage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "Usage: \"feedTube channel {channelName}\"")
}

func TestCmdChannelNoOutputFolder(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "You must specify an outputFolder")
}

func TestCmdChannelNoApiKey(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	set.String("outputFolder", outputFolder, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "You must specify an apiKey")
}

func TestCmdChannelInvalidChannelName(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultChannelResponses()
	channelInfo := youtube.ChannelListResponse{Items: []*youtube.Channel{}}
	bytes, _ := json.Marshal(channelInfo)
	responses["/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "channel awesome not found")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelAfter(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultChannelResponses()
	searchPage1 := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-02T15:04:05Z",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?alt=json&channelId=id&key=fakeApiKey&part=snippet&publishedAfter=2006-07-07T00%3A00%3A00Z&type=video"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("after", "07-07-06", "doc")
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
		},
	}
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "video 1 output\n", writer.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		"<rss version=\"2.0\">",
		"  <channel>",
		"    <title>t</title>",
		"    <description>d</description>",
		xmlLines[4],
		"    <item>",
		"      <title>t</title>",
		"      <description>d</description>",
		"      <guid>vId1</guid>",
		"      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>",
		"      <enclosure url=\"http://foo.com/t-vId1.mp3\" type=\"audio/mpeg\"></enclosure>",
		"    </item>",
		"  </channel>",
		"</rss>",
	}
	assert.Equal(t, expectedXMLLines, xmlLines)
}

func TestCmdChannelAfterInvalidDate(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("after", "99-99-99", "doc")
	cb := &commandBuilder.Test{}
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand(nil), cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "Could not parse after date: parsing time \"99-99-99\": month out of range\n", errWriter.String())
}

func TestCmdChannelYoutubeChannelError(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestChannelServerOverrideResponse("/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet", "error")
	defer ts.Close()
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "Channel request failed: googleapi: got HTTP response code 500 with body: ")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeSearchPage1Error(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=id&key=fakeApiKey&part=snippet&type=video", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"Search request failed: googleapi: got HTTP response code 500 with body: \n",
		&commandBuilder.Test{},
		CmdChannel,
	)
}

func TestCmdChannelYoutubeSearchPage2Error(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=id&key=fakeApiKey&pageToken=page2&part=snippet&type=video", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"Search request failed: googleapi: got HTTP response code 500 with body: \n",
		&commandBuilder.Test{
			ExpectedCommands: []*commandBuilder.ExpectedCommand{
				commandBuilder.NewExpectedCommand(
					"",
					"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
					"video 1 output",
					0,
				),
			},
		},
		CmdChannel,
	)
}

func TestCmdChannelYoutubeSearchInvalidVideos(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultChannelResponses()
	searchPage1 := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t",
					Description: "d",
					PublishedAt: "2007-01-02T15:04:05Z",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId1",
				},
			},
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t2",
					Description: "d2",
					PublishedAt: "2006-01-02T15:04:05Z",
				},
				Id: &youtube.ResourceId{
					VideoId: "",
				},
			},
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t2",
					Description: "d2",
					PublishedAt: "2006-01-02",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?alt=json&channelId=id&key=fakeApiKey&part=snippet&type=video"] = string(bytes)
	ts := getTestServer(responses)
	youtubeAPIURLBase = ts.URL
	defer ts.Close()
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t-vId1.%(ext)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
		},
	}
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "video 1 output\n", writer.String())
}

func getTestChannelServerOverrideResponse(URL, response string) *httptest.Server {
	responses := getDefaultChannelResponses()
	responses[URL] = response
	server := getTestServer(responses)
	youtubeAPIURLBase = server.URL
	return server
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
					PublishedAt: "2007-01-02T15:04:05Z",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?alt=json&channelId=id&key=fakeApiKey&part=snippet&type=video"] = string(bytes)

	searchPage2 := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "t2",
					Description: "d2",
					PublishedAt: "2006-01-02T15:04:05Z",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId2",
				},
			},
		},
	}
	bytes, _ = json.Marshal(searchPage2)
	responses["/search?alt=json&channelId=id&key=fakeApiKey&pageToken=page2&part=snippet&type=video"] = string(bytes)

	channelInfo := youtube.ChannelListResponse{
		Items: []*youtube.Channel{
			{
				Snippet: &youtube.ChannelSnippet{
					Title:       "t",
					Description: "d",
				},
				Id: "id",
			},
		},
	}
	bytes, _ = json.Marshal(channelInfo)
	responses["/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	return responses
}

func getExpectedXML(dateLine string) []string {
	return []string{
		"<rss version=\"2.0\">",
		"  <channel>",
		"    <title>t</title>",
		"    <description>d</description>",
		dateLine,
		"    <item>",
		"      <title>t</title>",
		"      <description>d</description>",
		"      <guid>vId1</guid>",
		"      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>",
		"      <enclosure url=\"http://foo.com/t-vId1.mp3\" type=\"audio/mpeg\"></enclosure>",
		"    </item>",
		"    <item>",
		"      <title>t2</title>",
		"      <description>d2</description>",
		"      <guid>vId2</guid>",
		"      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>",
		"      <enclosure url=\"http://foo.com/t2-vId2.mp3\" type=\"audio/mpeg\"></enclosure>",
		"    </item>",
		"  </channel>",
		"</rss>",
	}
}
