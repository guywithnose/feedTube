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
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
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
	assert.Equal(t, getExpectedXML(xmlLines[5:7]), xmlLines)
}

func TestCmdChannelById(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, set.Parse([]string{"awesomeChannelId"}))
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
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
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
	assert.Equal(t, getExpectedXML(xmlLines[5:7]), xmlLines)
}

func TestCmdChannelNoRedownload(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	err := ioutil.WriteFile(fmt.Sprintf("%s/t-vId1.mp3", outputFolder), []byte("123"), 0777)
	assert.Nil(t, err)
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
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
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(
		t,
		[]string{
			`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">`,
			`  <channel>`,
			`    <title>t</title>`,
			`    <link></link>`,
			`    <description>d</description>`,
			xmlLines[5],
			xmlLines[6],
			`    <item>`,
			`      <title>t</title>`,
			`      <link>http://foo.com/t-vId1.mp3</link>`,
			`      <description>d https://youtu.be/vId1</description>`,
			`      <enclosure url="http://foo.com/t-vId1.mp3" length="3" type="audio/mpeg"></enclosure>`,
			`      <guid>vId1</guid>`,
			`      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>`,
			`    </item>`,
			`    <item>`,
			`      <title>t2</title>`,
			`      <link>http://foo.com/t2-vId2.mp3</link>`,
			`      <description>d2 https://youtu.be/vId2</description>`,
			`      <enclosure url="http://foo.com/t2-vId2.mp3" length="0" type="audio/mpeg"></enclosure>`,
			`      <guid>vId2</guid>`,
			`      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>`,
			`    </item>`,
			`  </channel>`,
			`</rss>`,
		},
		xmlLines,
	)
}

func TestCmdChannelCleanup(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	unrelatedFile := fmt.Sprintf("%s/unrelated.mp3", outputFolder)
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	_, err = os.Create(unrelatedFile)
	assert.Nil(t, err)
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.Bool("cleanupUnrelatedFiles", true, "doc")
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
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
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	assert.Equal(t, "Removing file: /tmp/testFeedTube/unrelated.mp3\n", errWriter.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedXML(xmlLines[5:7]), xmlLines)
	_, err = os.Stat(unrelatedFile)
	assert.True(t, os.IsNotExist(err), "Unrelated file was not removed")
	_, err = os.Stat(relatedFile)
	assert.False(t, os.IsNotExist(err), "Related file was removed")
}

func TestCmdChannelCleanupDoesNotRemoveDirectoriesWithFiles(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	unrelatedFile := fmt.Sprintf("%s/unrelated", outputFolder)
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	err = os.Mkdir(unrelatedFile, 0777)
	assert.Nil(t, err)
	_, err = os.Create(fmt.Sprintf("%s/foo", unrelatedFile))
	assert.Nil(t, err)
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.Bool("cleanupUnrelatedFiles", true, "doc")
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
				"video 2 output",
				1,
			),
		},
	}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "Could not remove unrelated file: remove /tmp/testFeedTube/unrelated: directory not empty")
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		[]string{
			"video 2 output",
			"Could not download t2-vId2: exit status 1",
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	assert.Equal(t, "Removing file: /tmp/testFeedTube/unrelated\n", errWriter.String())
	_, err = os.Stat(unrelatedFile)
	assert.False(t, os.IsNotExist(err), "Unrelated file was removed")
	_, err = os.Stat(relatedFile)
	assert.False(t, os.IsNotExist(err), "Related file was removed")
}

func TestCmdChannelUsage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), `Usage: "feedTube channel {channelName|channelId}"`)
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
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "Channel ID awesome not found: Channel awesome not found")
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
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&publishedAfter=2006-07-07T00%3A00%3A00Z&type=video"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("after", "07-07-06", "doc")
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
	assert.Nil(t, CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "video 1 output\n", writer.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">`,
		`  <channel>`,
		`    <title>t</title>`,
		`    <link></link>`,
		`    <description>d</description>`,
		xmlLines[5],
		xmlLines[6],
		`    <item>`,
		`      <title>t</title>`,
		`      <link>http://foo.com/t-vId1.mp3</link>`,
		`      <description>d https://youtu.be/vId1</description>`,
		`      <enclosure url="http://foo.com/t-vId1.mp3" length="0" type="audio/mpeg"></enclosure>`,
		`      <guid>vId1</guid>`,
		`      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>`,
		`    </item>`,
		`  </channel>`,
		`</rss>`,
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
	assert.EqualError(
		t,
		CmdChannel(cb)(cli.NewContext(app, set, nil)),
		"Channel ID awesome not found: Channel request failed: googleapi: got HTTP response code 500 with body: ",
	)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeChannelIdError(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestChannelServerOverrideResponse("/channels?alt=json&id=awesomeChannelId&key=fakeApiKey&part=snippet", "error")
	defer ts.Close()
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, set.Parse([]string{"awesomeChannelId"}))
	assert.EqualError(
		t,
		CmdChannel(cb)(cli.NewContext(app, set, nil)),
		"Channel request failed: googleapi: got HTTP response code 500 with body: : Channel awesomeChannelId not found",
	)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeSearchPage1Error(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"Search request failed: googleapi: got HTTP response code 500 with body: \n",
		&commandBuilder.Test{},
		CmdChannel,
	)
}

func TestCmdChannelYoutubeSearchPage2Error(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&pageToken=page2&part=snippet&type=video", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"Search request failed: googleapi: got HTTP response code 500 with body: \n",
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
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video"] = string(bytes)
	ts := getTestServer(responses)
	youtubeAPIURLBase = ts.URL
	defer ts.Close()
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
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video"] = string(bytes)

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
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&pageToken=page2&part=snippet&type=video"] = string(bytes)

	channelInfo := youtube.ChannelListResponse{
		Items: []*youtube.Channel{
			{
				Snippet: &youtube.ChannelSnippet{
					Title:       "t",
					Description: "d",
				},
				Id: "awesomeChannelId",
			},
		},
	}
	bytes, _ = json.Marshal(channelInfo)
	responses["/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	responses["/channels?alt=json&id=awesomeChannelId&key=fakeApiKey&part=snippet"] = string(bytes)

	channelIDInfo := youtube.ChannelListResponse{Items: []*youtube.Channel{}}
	bytes, _ = json.Marshal(channelIDInfo)
	responses["/channels?alt=json&id=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	responses["/channels?alt=json&forUsername=awesomeChannelId&key=fakeApiKey&part=snippet"] = string(bytes)
	return responses
}

func getExpectedXML(dateLine []string) []string {
	return []string{
		`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0">`,
		`  <channel>`,
		`    <title>t</title>`,
		`    <link></link>`,
		`    <description>d</description>`,
		dateLine[0],
		dateLine[1],
		`    <item>`,
		`      <title>t</title>`,
		`      <link>http://foo.com/t-vId1.mp3</link>`,
		`      <description>d https://youtu.be/vId1</description>`,
		`      <enclosure url="http://foo.com/t-vId1.mp3" length="0" type="audio/mpeg"></enclosure>`,
		`      <guid>vId1</guid>`,
		`      <pubDate>Tue, 02 Jan 2007 15:04:05 +0000</pubDate>`,
		`    </item>`,
		`    <item>`,
		`      <title>t2</title>`,
		`      <link>http://foo.com/t2-vId2.mp3</link>`,
		`      <description>d2 https://youtu.be/vId2</description>`,
		`      <enclosure url="http://foo.com/t2-vId2.mp3" length="0" type="audio/mpeg"></enclosure>`,
		`      <guid>vId2</guid>`,
		`      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>`,
		`    </item>`,
		`  </channel>`,
		`</rss>`,
	}
}
