package command_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/guywithnose/feedTube/command"
	"github.com/stretchr/testify/assert"
)

func TestGetVideosForPlaylist(t *testing.T) {
	ts := getTestServer(getDefaultPlaylistResponses())
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	videoData, channelInfo, err := command.NewPlaylistScraper("fakeApiKey").GetVideosForPlaylist("awesome")
	assert.Nil(t, err)
	assert.Equal(t, []*command.VideoData{&videoData1, &videoData2}, videoData)
	assert.Equal(t, &awesomePlaylistInfo, channelInfo)
}

func TestPlaylistRequestFailure(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet")
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewPlaylistScraper("fakeApiKey").GetVideosForPlaylist("awesome")
	assert.EqualError(
		t,
		err,
		`Playlist request failed: googleapi: got HTTP response code 500 with body: `,
	)
}

func TestPlaylistZeroResults(t *testing.T) {
	responses := getDefaultPlaylistResponses()
	playlistInfo := youtube.PlaylistListResponse{Items: []*youtube.Playlist{}}
	bytes, _ := json.Marshal(playlistInfo)
	responses["/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet"] = string(bytes)
	ts := getTestServer(responses)
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewPlaylistScraper("fakeApiKey").GetVideosForPlaylist("awesome")
	assert.EqualError(
		t,
		err,
		`playlist awesome not found`,
	)
}

func TestPage1Failure(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet")
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewPlaylistScraper("fakeApiKey").GetVideosForPlaylist("awesome")
	assert.EqualError(
		t,
		err,
		`Playlist request failed: googleapi: got HTTP response code 500 with body: `,
	)
}

func TestPage2Failure(t *testing.T) {
	ts := getTestPlaylistServerOverrideResponse("/playlists?alt=json&id=awesome&key=fakeApiKey&part=snippet")
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewPlaylistScraper("fakeApiKey").GetVideosForPlaylist("awesome")
	assert.EqualError(
		t,
		err,
		`Playlist request failed: googleapi: got HTTP response code 500 with body: `,
	)
}

func TestInvalidVideo(t *testing.T) {
	responses := getDefaultPlaylistResponses()
	playlistVideosPage1 := youtube.PlaylistItemListResponse{
		Items: []*youtube.PlaylistItem{
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
	defer ts.Close()
	command.YoutubeAPIURLBase = ts.URL
	_, _, err := command.NewPlaylistScraper("fakeApiKey").GetVideosForPlaylist("awesome")
	assert.EqualError(
		t,
		err,
		`playlist items request failed: error parsing publish date on video vId2: parsing time "2006-01-02" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"`,
	)
}

var awesomePlaylistInfo = command.ChannelInfo{
	Title:       "playlistTitle",
	Description: "playlistDescription",
	Link:        "https://www.youtube.com/playlist?list=awesome",
	Thumbnail:   "https://images.com/thumb.jpg",
}

func getDefaultPlaylistResponses() map[string]string {
	responses := map[string]string{}
	playlistInfo := youtube.PlaylistListResponse{
		Items: []*youtube.Playlist{
			{
				Snippet: &youtube.PlaylistSnippet{
					Title:       "playlistTitle",
					Description: "playlistDescription",
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/thumb.jpg",
						},
					},
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
					Thumbnails: &youtube.ThumbnailDetails{
						Default: &youtube.Thumbnail{
							Url: "https://images.com/vid1Thumb.jpg",
						},
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

func getTestPlaylistServerOverrideResponse(URL string) *httptest.Server {
	responses := getDefaultPlaylistResponses()
	responses[URL] = "error"
	server := getTestServer(responses)
	command.YoutubeAPIURLBase = server.URL
	return server
}
