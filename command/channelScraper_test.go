package command_test

import (
	"encoding/json"
	"testing"
	"time"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/guywithnose/feedTube/command"
	"github.com/stretchr/testify/assert"
)

func TestGetVideosForChannelName(t *testing.T) {
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	videoData, channelInfo, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesome", "")
	assert.Nil(t, err)
	assert.Equal(t, []*command.VideoData{&videoData1, &videoData2}, videoData)
	assert.Equal(t, &awesomeChannelInfo, channelInfo)
}

func TestGetVideosForChannelId(t *testing.T) {
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	videoData, channelInfo, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesomeChannelId", "")
	assert.Nil(t, err)
	assert.Equal(t, []*command.VideoData{&videoData1, &videoData2}, videoData)
	assert.Equal(t, &awesomeChannelInfo, channelInfo)
}

func TestGetVideosForChannelIdWithAfter(t *testing.T) {
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	videoData, channelInfo, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesomeChannelId", "07-07-06")
	assert.Nil(t, err)
	assert.Equal(t, []*command.VideoData{&videoData1}, videoData)
	assert.Equal(t, &awesomeChannelInfo, channelInfo)
}

func TestGetVideosForChannelIdWithInvalidAfter(t *testing.T) {
	ts := getTestServer(getDefaultChannelResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesomeChannelId", "07-a7-06")
	assert.EqualError(t, err, `could not parse after date: parsing time "07-a7-06" as "01-02-06": cannot parse "a7-06" as "02"`)
}

func TestChannelFailure(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/channels?alt=json&forUsername=awesome&key=fakeApiKey&part=snippet")
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesome", "")
	assert.EqualError(t, err, "Channel ID awesome not found: Channel request failed: googleapi: got HTTP response code 500 with body: ")
}

func TestChannelIdFailure(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/channels?alt=json&id=awesomeChannelId&key=fakeApiKey&part=snippet")
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesomeChannelId", "")
	assert.EqualError(t, err, "Channel request failed: googleapi: got HTTP response code 500 with body: : Channel awesomeChannelId not found")
}

func TestSearchPage1Failure(t *testing.T) {
	ts := getTestChannelServerOverrideResponse("/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&part=snippet&type=video")
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesome", "")
	assert.EqualError(t, err, "search request failed: googleapi: got HTTP response code 500 with body: ")
}

func TestSearchPage2Failure(t *testing.T) {
	ts := getTestChannelServerOverrideResponse(
		"/search?alt=json&channelId=awesomeChannelId&key=fakeApiKey&pageToken=page2&part=snippet&type=video",
	)
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesome", "")
	assert.EqualError(t, err, "search request failed: googleapi: got HTTP response code 500 with body: ")
}

func TestInvalidVideoData(t *testing.T) {
	responses := getDefaultChannelResponses()
	searchPage1 := youtube.SearchListResponse{
		Items: []*youtube.SearchResult{
			{
				Snippet: &youtube.SearchResultSnippet{
					Title:                "t2",
					Description:          "d2",
					PublishedAt:          "2006-01-02",
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
	ts := getTestServer(responses)
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewChannelScraper("fakeApiKey").GetVideosForChannel("awesome", "")
	assert.EqualError(
		t,
		err,
		`search request failed: error parsing publish date on video vId1: parsing time "2006-01-02" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`,
	)
}

var videoData1 = command.VideoData{
	GUID:        "vId1",
	Link:        "https://youtu.be/vId1",
	Title:       "t",
	Description: "d https://youtu.be/vId1",
	FileName:    "t-vId1",
	Image:       "https://images.com/vid1Thumb.jpg",
	PubDate:     time.Date(2007, time.January, 02, 15, 04, 05, 0, time.UTC),
}
var videoData2 = command.VideoData{
	GUID:        "vId2",
	Link:        "https://youtu.be/vId2",
	Title:       "t2",
	Description: "d2 https://youtu.be/vId2",
	FileName:    "t2-vId2",
	Image:       "",
	PubDate:     time.Date(2006, time.January, 02, 15, 04, 05, 0, time.UTC),
}
var awesomeChannelInfo = command.ChannelInfo{
	Title:       "t",
	Description: "d",
	Link:        "https://www.youtube.com/channel/awesomeChannelId",
	Thumbnail:   "https://images.com/thumb.jpg",
}
