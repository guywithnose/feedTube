package command_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/guywithnose/feedTube/command"
	"github.com/guywithnose/runner"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCmdPlaylist(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{
		ExpectedCommands: []*runner.ExpectedCommand{
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%\\(ext\\)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%\\(ext\\)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
		},
	}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedPlaylistXML(xmlLines[8:10]), xmlLines)
}

func TestCmdPlaylistDownloadFailure(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := getBaseRunner()
	cb.ExpectedCommands[1] = runner.NewExpectedCommand(
		"",
		fmt.Sprintf("/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o %s/t2-vId2.%%\\(ext\\)s https://youtu.be/vId2", getOutputFolder()),
		"video 2 output",
		1,
	)
	assert.EqualError(
		t,
		command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)),
		fmt.Sprintf(
			"could not download t2-vId2: exit status 1\nParams: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0'"+
				" '-o' '%s/t2-vId2.%%(ext)s' 'https://youtu.be/vId2': video 2 output",
			getOutputFolder(),
		),
	)
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistOverrideTitle(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{
		ExpectedCommands: []*runner.ExpectedCommand{
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%\\(ext\\)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%\\(ext\\)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
		},
	}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("overrideTitle", "ovride", "doc")
	assert.Nil(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`  <channel>`,
		`    <title>ovride</title>`,
		`    <link>https://www.youtube.com/playlist?list=awesome</link>`,
		`    <description>playlistDescription</description>`,
		fmt.Sprintf(`    <generator>feedTube v%s (github.com/guywithnose/feedTube)</generator>`, command.Version),
		`    <language>en-us</language>`,
		xmlLines[8],
		xmlLines[9],
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
	assert.Equal(t, expectedXMLLines, xmlLines)
}

func TestCmdPlaylistCleanup(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	unrelatedFile := fmt.Sprintf("%s/unrelated.mp3", outputFolder)
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	_, err = os.Create(unrelatedFile)
	assert.Nil(t, err)
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{
		ExpectedCommands: []*runner.ExpectedCommand{
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%\\(ext\\)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
			runner.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"Duration: 02:13:45.22, start",
				1,
			),
		},
	}
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.Bool("cleanupUnrelatedFiles", true, "doc")
	assert.Nil(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(
		t,
		"Removing file: /tmp/testFeedTube/unrelated.mp3\n",
		errWriter.String(),
	)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedPlaylistXML(xmlLines[8:10]), xmlLines)
	_, err = os.Stat(unrelatedFile)
	assert.True(t, os.IsNotExist(err), "Unrelated file was not removed")
	_, err = os.Stat(relatedFile)
	assert.False(t, os.IsNotExist(err), "Related file was removed")
}

func TestCmdPlaylistCleanupDoesNotRemoveDirectoriesWithFiles(t *testing.T) {
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
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{
		ExpectedCommands: []*runner.ExpectedCommand{
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%\\(ext\\)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
			runner.NewExpectedCommand(
				"",
				"/usr/bin/ffprobe /tmp/testFeedTube/t-vId1.mp3",
				"Duration: 02:13:45.22, start",
				0,
			),
		},
	}
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.Bool("cleanupUnrelatedFiles", true, "doc")
	assert.EqualError(
		t,
		command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)),
		"could not remove unrelated file: remove /tmp/testFeedTube/unrelated: directory not empty",
	)
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "Removing file: /tmp/testFeedTube/unrelated\n", errWriter.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXML := getExpectedPlaylistXML(xmlLines[8:10])
	expectedXML = append(expectedXML[:22], append([]string{`      <itunes:duration>02:13:45</itunes:duration>`}, expectedXML[22:]...)...)
	assert.Equal(t, expectedXML, xmlLines)
	_, err = os.Stat(unrelatedFile)
	assert.False(t, os.IsNotExist(err), "Unrelated file was not removed")
	_, err = os.Stat(relatedFile)
	assert.False(t, os.IsNotExist(err), "Related file was removed")
}

func TestCmdPlaylistFilter(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{
		ExpectedCommands: []*runner.ExpectedCommand{
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%\\(ext\\)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
		},
	}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("filter", "t2", "doc")
	assert.Nil(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`  <channel>`,
		`    <title>playlistTitle</title>`,
		`    <link>https://www.youtube.com/playlist?list=awesome</link>`,
		`    <description>playlistDescription</description>`,
		fmt.Sprintf(`    <generator>feedTube v%s (github.com/guywithnose/feedTube)</generator>`, command.Version),
		`    <language>en-us</language>`,
		xmlLines[8],
		xmlLines[9],
		`    <image>`,
		`      <url>https://images.com/thumb.jpg</url>`,
		`    </image>`,
		`    <itunes:image href="https://images.com/thumb.jpg"></itunes:image>`,
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
	assert.Equal(t, expectedXMLLines, xmlLines)
}

func TestCmdPlaylistNoBaseUrl(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{}
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	xmlFile := fmt.Sprintf("%s/xmlFile", outputFolder)
	set.String("xmlFile", xmlFile, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, writer, _ := appWithTestWriters()
	assert.EqualError(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "You must specify an baseURL")
	assert.Equal(t, []*runner.ExpectedCommand(nil), cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "", writer.String())
}

func TestCmdPlaylistNoXmlFile(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{
		ExpectedCommands: []*runner.ExpectedCommand{
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%\\(ext\\)s https://youtu.be/vId1",
				"video 1 output",
				0,
			),
			runner.NewExpectedCommand(
				"",
				"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t2-vId2.%\\(ext\\)s https://youtu.be/vId2",
				"video 2 output",
				0,
			),
		},
	}
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	assert.Nil(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*runner.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistUsage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, _, _ := appWithTestWriters()
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)), `Usage: "feedTube playlist {playlistID}"`)
}

func TestCmdPlaylistNoOutputFolder(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "You must specify an outputFolder")
}

func TestCmdPlaylistNoApiKey(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	set.String("outputFolder", outputFolder, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "You must specify an apiKey")
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
	command.YoutubeAPIURLBase = ts.URL
	cb := &runner.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "playlist awesome not found")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistYoutubePlaylistError(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestPlaylistServerOverrideResponse("/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet", "error")
	defer ts.Close()
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &runner.Test{}
	assert.EqualError(t, command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)), "Playlist request failed: googleapi: got HTTP response code 500 with body: ")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdPlaylistYoutubeSearchPage1Error(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlistItems?alt=json&key=fakeApiKey&part=snippet&playlistId=awesome", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"playlist items request failed: googleapi: got HTTP response code 500 with body: ",
		&runner.Test{},
		command.CmdPlaylist,
	)
}

func TestCmdPlaylistYoutubeSearchPage2Error(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlistItems?alt=json&key=fakeApiKey&pageToken=page2&part=snippet&playlistId=awesome", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"playlist items request failed: googleapi: got HTTP response code 500 with body: ",
		&runner.Test{
			ExpectedCommands: []*runner.ExpectedCommand{
				runner.NewExpectedCommand(
					"",
					"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/t-vId1.%\\(ext\\)s https://youtu.be/vId1",
					"video 1 output",
					0,
				),
			},
		},
		command.CmdPlaylist,
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
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vid1Thumb.jpg",
						},
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
	command.YoutubeAPIURLBase = ts.URL
	defer ts.Close()
	app, writer, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &runner.Test{}
	assert.EqualError(
		t,
		command.CmdPlaylist(cb)(cli.NewContext(app, set, nil)),
		`playlist items request failed: error parsing publish date on video vId2: parsing time "2006-01-02" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`,
	)
	assert.Equal(t, []*runner.ExpectedCommand(nil), cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "", writer.String())
}

func getExpectedPlaylistXML(dateLine []string) []string {
	return []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`  <channel>`,
		`    <title>playlistTitle</title>`,
		`    <link>https://www.youtube.com/playlist?list=awesome</link>`,
		`    <description>playlistDescription</description>`,
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
