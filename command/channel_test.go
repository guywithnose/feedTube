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

	"github.com/guywithnose/commandBuilder"
	"github.com/guywithnose/feedTube/command"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCmdChannel(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := getBaseRunner()
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedChannelXML(xmlLines[8:10]), xmlLines)
}

func TestCmdChannelDownloadFailure(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := getBaseRunner()
	cb.ExpectedCommands[1] = commandBuilder.NewExpectedCommand(
		"",
		fmt.Sprintf("/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o %s/t2-vId2.%%(ext)s https://youtu.be/vId2", getOutputFolder()),
		"video 2 output",
		1,
	)
	assert.EqualError(
		t,
		command.CmdChannel(cb)(cli.NewContext(app, set, nil)),
		fmt.Sprintf(
			"could not download t2-vId2: exit status 1\nParams: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0'"+
				" '-o' '%s/t2-vId2.%%(ext)s' 'https://youtu.be/vId2': video 2 output",
			getOutputFolder(),
		),
	)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
}

func TestCmdChannelOverrideTitle(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := getBaseRunner()
	set.String("overrideTitle", "ovride", "doc")
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(
		t,
		[]string{
			`<?xml version="1.0" encoding="UTF-8"?>`,
			`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
			`  <channel>`,
			`    <title>ovride</title>`,
			`    <link>https://www.youtube.com/channel/awesomeChannelId</link>`,
			`    <description>d</description>`,
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
		},
		xmlLines,
	)
}

func TestCmdChannelFilter(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
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
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("filter", "t2", "doc")
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`  <channel>`,
		`    <title>t</title>`,
		`    <link>https://www.youtube.com/channel/awesomeChannelId</link>`,
		`    <description>d</description>`,
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

func TestCmdChannelInvalidXmlFile(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	set.String("baseURL", "http://foo.com", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	set.String("xmlFile", "/notadir/invalidFile", "doc")
	cb := getBaseRunner()
	assert.EqualError(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)), "open /notadir/invalidFile: no such file or directory")
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelById(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, set.Parse([]string{"awesomeChannelId"}))
	cb := getBaseRunner()
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedChannelXML(xmlLines[8:10]), xmlLines)
}

func TestCmdChannelNoRedownload(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	err := ioutil.WriteFile(fmt.Sprintf("%s/t-vId1.mp3", outputFolder), []byte("123"), 0777)
	assert.Nil(t, err)
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := getFfprobeRunner()
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(
		t,
		[]string{
			`<?xml version="1.0" encoding="UTF-8"?>`,
			`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
			`  <channel>`,
			`    <title>t</title>`,
			`    <link>https://www.youtube.com/channel/awesomeChannelId</link>`,
			`    <description>d</description>`,
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
			`      <enclosure url="http://foo.com/t-vId1.mp3" length="3" type="audio/mpeg"></enclosure>`,
			`      <itunes:image href="https://images.com/vid1Thumb.jpg"></itunes:image>`,
			`      <itunes:duration>02:13:45</itunes:duration>`,
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
		},
		xmlLines,
	)
}

func TestCmdChannelInvalidDuration(t *testing.T) {
	outputFolder := getOutputFolder()
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
	command.YoutubeAPIURLBase = ts.URL
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := getFfprobeRunner()
	cb.ExpectedCommands[1] = commandBuilder.NewExpectedCommand(
		"",
		fmt.Sprintf("/usr/bin/ffprobe %s/t-vId1.mp3", getOutputFolder()),
		"Duration: 02:1E:45.22, start",
		0,
	)
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, "", errWriter.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	assert.Equal(t, getExpectedChannelXML(xmlLines[8:10]), xmlLines)
}

func TestCmdChannelCleanup(t *testing.T) {
	outputFolder := getOutputFolder()
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
	command.YoutubeAPIURLBase = ts.URL
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.Bool("cleanupUnrelatedFiles", true, "doc")
	cb := getFfprobeRunner()
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, fmt.Sprintf("Removing file: %s/unrelated.mp3\n", getOutputFolder()), errWriter.String())
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXML := getExpectedChannelXML(xmlLines[8:10])
	expectedXML = append(expectedXML[:22], append([]string{`      <itunes:duration>02:13:45</itunes:duration>`}, expectedXML[22:]...)...)
	assert.Equal(t, expectedXML, xmlLines)
	_, err = os.Stat(unrelatedFile)
	assert.True(t, os.IsNotExist(err), "Unrelated file was not removed")
	_, err = os.Stat(relatedFile)
	assert.False(t, os.IsNotExist(err), "Related file was removed")
}

func TestCmdChannelCleanupDoesNotRemoveDirectoriesWithFiles(t *testing.T) {
	outputFolder := getOutputFolder()
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
	command.YoutubeAPIURLBase = ts.URL
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	set.Bool("cleanupUnrelatedFiles", true, "doc")
	cb := getFfprobeRunner()
	assert.EqualError(
		t,
		command.CmdChannel(cb)(cli.NewContext(app, set, nil)),
		fmt.Sprintf("could not remove unrelated file: remove %s/unrelated: directory not empty", getOutputFolder()),
	)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, fmt.Sprintf("Removing file: %s/unrelated\n", getOutputFolder()), errWriter.String())
	_, err = os.Stat(unrelatedFile)
	assert.False(t, os.IsNotExist(err), "Unrelated file was removed")
	_, err = os.Stat(relatedFile)
	assert.False(t, os.IsNotExist(err), "Related file was removed")
}

func TestCmdChannelUsage(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)), `Usage: "feedTube channel {channelName|channelId}"`)
}

func TestCmdChannelNoOutputFolder(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)), "You must specify an outputFolder")
}

func TestCmdChannelNoApiKey(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	outputFolder := getOutputFolder()
	set.String("outputFolder", outputFolder, "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, _, _ := appWithTestWriters()
	cb := &commandBuilder.Test{}
	assert.EqualError(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)), "You must specify an apiKey")
}

func TestCmdChannelInvalidChannelName(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultChannelResponses()
	channelInfo := youtube.ChannelListResponse{Items: []*youtube.Channel{}}
	bytes, _ := json.Marshal(channelInfo)
	responses["/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	cb := &commandBuilder.Test{}
	assert.EqualError(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)), "Channel ID awesome not found: Channel awesome not found")
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelAfter(t *testing.T) {
	outputFolder := getOutputFolder()
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
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vid1Thumb.jpg",
						},
					},
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
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("after", "07-07-06", "doc")
	cb := getBaseRunner()
	cb.ExpectedCommands = cb.ExpectedCommands[:1]
	assert.Nil(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
	xmlBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/xmlFile", outputFolder))
	assert.Nil(t, err)
	xmlLines := strings.Split(string(xmlBytes), "\n")
	expectedXMLLines := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`  <channel>`,
		`    <title>t</title>`,
		`    <link>https://www.youtube.com/channel/awesomeChannelId</link>`,
		`    <description>d</description>`,
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
		`  </channel>`,
		`</rss>`,
	}
	assert.Equal(t, expectedXMLLines, xmlLines)
}

func TestCmdChannelAfterInvalidDate(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	set.String("after", "99-99-99", "doc")
	cb := &commandBuilder.Test{}
	assert.EqualError(t, command.CmdChannel(cb)(cli.NewContext(app, set, nil)), "could not parse after date: parsing time \"99-99-99\": month out of range")
	assert.Equal(t, []*commandBuilder.ExpectedCommand(nil), cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeChannelError(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestChannelServerOverrideResponse("/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet", "error")
	defer ts.Close()
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(
		t,
		command.CmdChannel(cb)(cli.NewContext(app, set, nil)),
		"Channel ID awesome not found: Channel request failed: googleapi: got HTTP response code 500 with body: ",
	)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeChannelIdError(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	ts := getTestChannelServerOverrideResponse("/channels?alt=json&id=awesomeChannelId&key=fakeApiKey&part=snippet", "error")
	defer ts.Close()
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, set.Parse([]string{"awesomeChannelId"}))
	assert.EqualError(
		t,
		command.CmdChannel(cb)(cli.NewContext(app, set, nil)),
		"Channel request failed: googleapi: got HTTP response code 500 with body: : Channel awesomeChannelId not found",
	)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeSearchPage1Error(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video", "error")
	defer ts.Close()
	runErrorTest(
		t,
		"search request failed: googleapi: got HTTP response code 500 with body: ",
		&commandBuilder.Test{},
		command.CmdChannel,
	)
}

func TestCmdChannelYoutubeSearchPage2Error(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&pageToken=page2&part=snippet&type=video", "error")
	defer ts.Close()
	cb := getBaseRunner()
	cb.ExpectedCommands = cb.ExpectedCommands[:1]
	runErrorTest(
		t,
		"search request failed: googleapi: got HTTP response code 500 with body: ",
		cb,
		command.CmdChannel,
	)
}

func TestCmdChannelYoutubeSearchInvalidVideos(t *testing.T) {
	outputFolder := getOutputFolder()
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
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vid1Thumb.jpg",
						},
					},
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
	command.YoutubeAPIURLBase = ts.URL
	defer ts.Close()
	cb := &commandBuilder.Test{}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(
		t,
		command.CmdChannel(cb)(cli.NewContext(app, set, nil)),
		`search request failed: error parsing publish date on video vId1: parsing time "2006-01-02" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`,
	)
	assert.Equal(t, []*commandBuilder.ExpectedCommand(nil), cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func TestCmdChannelYoutubeSearchNoTitle(t *testing.T) {
	outputFolder := getOutputFolder()
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	responses := getDefaultChannelResponses()
	searchPage1 := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:       "",
					Description: "d2",
					PublishedAt: "2006-01-02T15:04:05Z",
				},
				Id: &youtube.ResourceId{
					VideoId: "vId2",
				},
			},
		},
	}
	bytes, _ := json.Marshal(searchPage1)
	responses["/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video"] = string(bytes)
	ts := getTestServer(responses)
	command.YoutubeAPIURLBase = ts.URL
	defer ts.Close()
	cb := &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				fmt.Sprintf("/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o %s/-vId2.%%(ext)s https://youtu.be/vId2", getOutputFolder()),
				"video 2 output",
				0,
			),
		},
	}
	app, _, _, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.EqualError(
		t,
		command.CmdChannel(cb)(cli.NewContext(app, set, nil)),
		`could not parse item to xml: Title and Description are reuired`,
	)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cb.ExpectedCommands)
	assert.Equal(t, []error(nil), cb.Errors)
}

func getBaseRunner() *commandBuilder.Test {
	return &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				fmt.Sprintf("/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o %s/t-vId1.%%(ext)s https://youtu.be/vId1", getOutputFolder()),
				"video 1 output",
				0,
			),
			commandBuilder.NewExpectedCommand(
				"",
				fmt.Sprintf("/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o %s/t2-vId2.%%(ext)s https://youtu.be/vId2", getOutputFolder()),
				"video 2 output",
				0,
			),
		},
	}
}

func getFfprobeRunner() *commandBuilder.Test {
	return &commandBuilder.Test{
		ExpectedCommands: []*commandBuilder.ExpectedCommand{
			commandBuilder.NewExpectedCommand(
				"",
				fmt.Sprintf("/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o %s/t2-vId2.%%(ext)s https://youtu.be/vId2", getOutputFolder()),
				"video 2 output",
				0,
			),
			commandBuilder.NewExpectedCommand(
				"",
				fmt.Sprintf("/usr/bin/ffprobe %s/t-vId1.mp3", getOutputFolder()),
				"Duration: 02:13:45.22, start",
				0,
			),
		},
	}
}

func getOutputFolder() string {
	return fmt.Sprintf("%s/feedTubeCommand", os.TempDir())
}
