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
	channelInfo := channelResponse{Items: []channelItem{}}
	bytes, _ := json.Marshal(channelInfo)
	responses["/channels?key=fakeApiKey&part=snippet&forUsername=awesome"] = string(bytes)
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
	searchPage1 := videoSearchResponse{
		Items: []map[string]interface{}{
			{
				"snippet": map[string]string{
					"title":       "t",
					"description": "d",
					"publishedAt": "2007-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&publishedAfter=2006-07-07T00%3A00%3A00Z&type=video"] = string(bytes)
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
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("after", "99-99-99", "doc")
	cb := &commandBuilder.Test{}
	assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), "Could not parse after date: parsing time \"99-99-99\": month out of range")
	assert.Equal(t, []*commandBuilder.ExpectedCommand(nil), cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelServerError(t *testing.T) {
	runChannelErrorTest(
		t,
		"/channels?key=fakeApiKey&part=snippet&forUsername=awesome",
		"httpError",
		"Could not connect to %s/channels?key=fakeApiKey&part=snippet&forUsername=awesome: "+
			"Get htp://notarealhostname.foo: unsupported protocol scheme \"htp\"",
	)
}

func TestCmdChannelYoutubeChannelError(t *testing.T) {
	runChannelErrorTest(
		t,
		"/channels?key=fakeApiKey&part=snippet&forUsername=awesome",
		"error",
		"EOF",
	)
}

func TestCmdChannelYoutubeSearchPage1Error(t *testing.T) {
	runChannelErrorTest(
		t,
		"/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video",
		"error",
		"EOF",
	)
}

func TestCmdChannelYoutubeSearchPage1HttpError(t *testing.T) {
	runChannelErrorTest(
		t,
		"/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video",
		"httpError",
		"Could not connect to %s/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video: "+
			"Get htp://notarealhostname.foo: unsupported protocol scheme \"htp\"",
	)
}

func runChannelErrorTest(t *testing.T, url, errorType, expectedError string) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestChannelServerOverrideResponse(url, errorType)
	defer ts.Close()
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{}
	if errorType == "httpError" {
		assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), fmt.Sprintf(expectedError, ts.URL))
	} else {
		assert.EqualError(t, CmdChannel(cb)(cli.NewContext(app, set, nil)), expectedError)
	}
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeSearchPage2Error(t *testing.T) {
	runChannelErrorTest(
		t,
		"/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video&pageToken=page2",
		"error",
		"EOF",
	)
}

func TestCmdChannelYoutubeSearchInvalidVideos(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultChannelResponses()
	searchPage1 := videoSearchResponse{
		Items: []map[string]interface{}{
			{
				"snippet": map[string]string{
					"title":       "t2",
					"description": "d2",
					"publishedAt": "2006-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#notvideo",
					"videoId": "vId2",
				},
			},
			{
				"snippet": map[string]string{
					"description": "d2",
					"publishedAt": "2006-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId2",
				},
			},
			{
				"snippet": map[string]string{
					"title":       "t2",
					"publishedAt": "2006-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId2",
				},
			},
			{
				"snippet": map[string]string{
					"title":       "t2",
					"description": "d2",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId2",
				},
			},
			{
				"snippet": map[string]string{
					"title":       "t2",
					"description": "d2",
					"publishedAt": "2006-01-02",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId2",
				},
			},
			{
				"snippet": map[string]string{
					"title":       "t2",
					"description": "d2",
					"publishedAt": "2006-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "",
				},
			},
			{
				"snippet": map[string]string{
					"title":       "t",
					"description": "d",
					"publishedAt": "2007-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video"] = string(bytes)
	ts := getTestServer(responses)
	youtubeAPIURLBase = ts.URL
	defer ts.Close()
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
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
	searchPage1 := videoSearchResponse{
		NextPageToken: "page2",
		Items: []map[string]interface{}{
			{
				"snippet": map[string]string{
					"title":       "t",
					"description": "d",
					"publishedAt": "2007-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId1",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video"] = string(bytes)

	searchPage2 := videoSearchResponse{
		NextPageToken: "",
		Items: []map[string]interface{}{
			{
				"snippet": map[string]string{
					"title":       "t2",
					"description": "d2",
					"publishedAt": "2006-01-02T15:04:05Z",
				},
				"id": map[string]string{
					"kind":    "youtube#video",
					"videoId": "vId2",
				},
			},
		},
	}
	bytes, _ = json.Marshal(searchPage2)
	responses["/search?channelId=id&key=fakeApiKey&maxResults=50&part=snippet&type=video&pageToken=page2"] = string(bytes)

	channelInfo := channelResponse{
		Items: []channelItem{
			{
				Snippet: map[string]interface{}{
					"title":       "t",
					"description": "d",
				},
				ID: "id",
			},
		},
	}
	bytes, _ = json.Marshal(channelInfo)
	responses["/channels?key=fakeApiKey&part=snippet&forUsername=awesome"] = string(bytes)
	return responses
}
