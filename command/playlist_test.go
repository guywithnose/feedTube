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
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		"<rss version=\"2.0\">",
		"  <channel>",
		"    <title>playlistTitle</title>",
		"    <description>playlistDescription</description>",
		xmlLines[4],
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
				"/usr/bin/youtube-dl -x --audio-format mp3 -o /tmp/testFeedTube/t2-vId2.%(ext)s https://youtu.be/vId2",
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
		"<rss version=\"2.0\">",
		"  <channel>",
		"    <title>playlistTitle</title>",
		"    <description>playlistDescription</description>",
		xmlLines[4],
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
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
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
			"Params: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '-o' '/tmp/testFeedTube/t2-vId2.%(ext)s' 'https://youtu.be/vId2'",
			"",
		},
		strings.Split(writer.String(), "\n"),
	)
}

func TestCmdPlaylistUsage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "Usage: \"feedTube playlist {playlistID}\"")
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
	playlistInfo := channelResponse{Items: []channelItem{}}
	bytes, _ := json.Marshal(playlistInfo)
	responses["/playlists?key=fakeApiKey&part=snippet&id=awesome&maxResults=1"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	youtubeAPIURLBase = ts.URL
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "playlist awesome not found")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistServerError(t *testing.T) {
	runPlaylistErrorTest(
		t,
		"/playlists?key=fakeApiKey&part=snippet&id=awesome&maxResults=1",
		"httpError",
		"Could not connect to %s/playlists?key=fakeApiKey&part=snippet&id=awesome&maxResults=1: "+
			"Get htp://notarealhostname.foo: unsupported protocol scheme \"htp\"",
	)
}

func TestCmdPlaylistYoutubePlaylistError(t *testing.T) {
	runPlaylistErrorTest(
		t,
		"/playlists?key=fakeApiKey&part=snippet&id=awesome&maxResults=1",
		"error",
		"EOF",
	)
}

func TestCmdPlaylistYoutubeSearchPage1Error(t *testing.T) {
	runPlaylistErrorTest(
		t,
		"/playlistItems?key=fakeApiKey&maxResults=50&part=snippet&playlistId=awesome",
		"error",
		"EOF",
	)
}

func TestCmdPlaylistYoutubeSearchPage1HttpError(t *testing.T) {
	runPlaylistErrorTest(
		t,
		"/playlistItems?key=fakeApiKey&maxResults=50&part=snippet&playlistId=awesome",
		"httpError",
		"Could not connect to %s/playlistItems?key=fakeApiKey&maxResults=50&part=snippet&playlistId=awesome: "+
			"Get htp://notarealhostname.foo: unsupported protocol scheme \"htp\"",
	)
}

func TestCmdPlaylistYoutubeSearchPage2Error(t *testing.T) {
	runPlaylistErrorTest(
		t,
		"/playlistItems?key=fakeApiKey&maxResults=50&part=snippet&playlistId=awesome&pageToken=page2",
		"error",
		"EOF",
	)
}

func runPlaylistErrorTest(t *testing.T, url, errorType, expectedError string) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestPlaylistServerOverrideResponse(url, errorType)
	defer ts.Close()
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	if errorType == "httpError" {
		assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), fmt.Sprintf(expectedError, ts.URL))
	} else {
		assert.EqualError(t, CmdPlaylist(cb)(cli.NewContext(app, set, nil)), expectedError)
	}
	assert.Equal(t, []error(nil), cb.Errors)
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
	playlistInfo := channelResponse{
		Items: []channelItem{
			{
				Snippet: map[string]interface{}{
					"title":       "playlistTitle",
					"description": "playlistDescription",
				},
			},
		},
	}
	bytes, _ := json.Marshal(playlistInfo)
	responses["/playlists?key=fakeApiKey&part=snippet&id=awesome&maxResults=1"] = string(bytes)

	playlistVideosPage1 := videoSearchResponse{
		NextPageToken: "page2",
		Items: []map[string]interface{}{
			{
				"snippet": map[string]interface{}{
					"title":       "t",
					"description": "d",
					"publishedAt": "2007-01-02T15:04:05Z",
					"resourceId": map[string]string{
						"videoId": "vId1",
					},
				},
				"id": "someId",
			},
		},
	}
	bytes, _ = json.Marshal(playlistVideosPage1)
	responses["/playlistItems?key=fakeApiKey&maxResults=50&part=snippet&playlistId=awesome"] = string(bytes)

	playlistVideosPage2 := videoSearchResponse{
		Items: []map[string]interface{}{
			{
				"snippet": map[string]interface{}{
					"title":       "t2",
					"description": "d2",
					"publishedAt": "2006-01-02T15:04:05Z",
					"resourceId": map[string]string{
						"videoId": "vId2",
					},
				},
				"id": "someId",
			},
		},
	}
	bytes, _ = json.Marshal(playlistVideosPage2)
	responses["/playlistItems?key=fakeApiKey&maxResults=50&part=snippet&playlistId=awesome&pageToken=page2"] = string(bytes)

	return responses
}
