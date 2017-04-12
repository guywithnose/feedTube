package command

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"

	"github.com/gorilla/feeds"
	"github.com/kennygrant/sanitize"
)

var youtubeAPIURLBase = "https://www.googleapis.com/youtube/v3/"

// getVideosForChannel returns an array of all the youtube video ids on a channel
func getVideosForChannel(apiKey, channelName, after string, writer io.Writer) (<-chan *youtubeItem, *feeds.Feed, error) {
	channelID, info, err := getChannelInfo(apiKey, channelName)
	if err != nil {
		return nil, nil, err
	}

	videos := make(chan *youtubeItem, 10)
	go func() {
		defer close(videos)
		youtubeService := getYoutubeService(apiKey)
		listCall := youtubeService.Search.List("snippet").ChannelId(channelID).Type("video")
		if after != "" {
			afterTime, timeParseErr := time.Parse("01-02-06", after)
			if timeParseErr != nil {
				fmt.Fprintf(writer, "Could not parse after date: %v\n", timeParseErr)
				return
			}

			listCall = listCall.PublishedAfter(afterTime.Format(time.RFC3339))
		}

		err := listCall.Pages(context.Background(), func(resp *youtube.SearchListResponse) error {
			parseSearchResults(resp.Items, videos, writer)
			return nil
		})

		if err != nil {
			fmt.Fprintf(writer, "Search request failed: %v\n", err)
			return
		}
	}()

	return videos, info, nil
}

func parseSearchResults(results []*youtube.SearchResult, videos chan<- *youtubeItem, errWriter io.Writer) {
	for _, result := range results {
		publishedTime, err := time.Parse(time.RFC3339, result.Snippet.PublishedAt)
		if err != nil {
			fmt.Fprintf(errWriter, "error parsing publish date on video %s: %v\n", result.Id.VideoId, err)
			continue
		}

		videos <- &youtubeItem{
			Item: feeds.Item{
				Id:          result.Id.VideoId,
				Title:       result.Snippet.Title,
				Description: result.Snippet.Description,
				Created:     publishedTime,
			},
			Filename: fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(result.Snippet.Title), " ", "-", -1), result.Id.VideoId),
		}
	}
}

func parsePlaylistItems(results []*youtube.PlaylistItem, videos chan<- *youtubeItem, errWriter io.Writer) {
	for _, result := range results {
		publishedTime, err := time.Parse(time.RFC3339, result.Snippet.PublishedAt)
		if err != nil {
			fmt.Fprintf(errWriter, "error parsing publish date on video %s: %v\n", result.Snippet.ResourceId.VideoId, err)
			continue
		}

		videos <- &youtubeItem{
			Item: feeds.Item{
				Id:          result.Snippet.ResourceId.VideoId,
				Title:       result.Snippet.Title,
				Description: result.Snippet.Description,
				Created:     publishedTime,
			},
			Filename: fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(result.Snippet.Title), " ", "-", -1), result.Snippet.ResourceId.VideoId),
		}
	}
}

// getVideosForPlaylist returns an array of all the youtube video ids in a playlist
func getVideosForPlaylist(apiKey, playlistID string, writer io.Writer) (<-chan *youtubeItem, *feeds.Feed, error) {
	info, err := getPlaylistInfo(apiKey, playlistID)
	if err != nil {
		return nil, nil, err
	}

	videos := make(chan *youtubeItem, 10)
	go func() {
		defer close(videos)
		youtubeService := getYoutubeService(apiKey)
		listCall := youtubeService.PlaylistItems.List("snippet").PlaylistId(playlistID)
		err = listCall.Pages(context.Background(), func(resp *youtube.PlaylistItemListResponse) error {
			parsePlaylistItems(resp.Items, videos, writer)
			return nil
		})

		if err != nil {
			fmt.Fprintf(writer, "Playlist items request failed: %v\n", err)
			return
		}
	}()

	return videos, info, nil
}

func getChannelInfo(apiKey, channel string) (string, *feeds.Feed, error) {
	id, info, err := getChannelByID(apiKey, channel)
	if err != nil {
		id, info, err = getChannelByName(apiKey, channel)
	}

	return id, info, err
}

func getChannelByName(apiKey, channelName string) (string, *feeds.Feed, error) {
	youtubeService := getYoutubeService(apiKey)
	listCall := youtubeService.Channels.List("snippet").ForUsername(channelName)
	items, err := makeChannelRequest(listCall)
	if err != nil {
		return "", nil, err
	}

	if len(items) == 0 {
		return "", nil, fmt.Errorf("channel %s not found", channelName)
	}

	return items[0].Id, &feeds.Feed{Title: items[0].Snippet.Title, Description: items[0].Snippet.Description}, nil
}

func getChannelByID(apiKey, channelName string) (string, *feeds.Feed, error) {
	youtubeService := getYoutubeService(apiKey)
	listCall := youtubeService.Channels.List("snippet").Id(channelName)
	items, err := makeChannelRequest(listCall)
	if err != nil {
		return "", nil, err
	}

	if len(items) == 0 {
		return "", nil, fmt.Errorf("channel %s not found", channelName)
	}

	return items[0].Id, &feeds.Feed{Title: items[0].Snippet.Title, Description: items[0].Snippet.Description}, nil
}

func makeChannelRequest(listCall *youtube.ChannelsListCall) ([]*youtube.Channel, error) {
	resp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("Channel request failed: %v", err)
	}

	return resp.Items, nil
}

func getPlaylistInfo(apiKey, playlistID string) (*feeds.Feed, error) {
	youtubeService := getYoutubeService(apiKey)
	listCall := youtubeService.Playlists.List("snippet").Id(playlistID)
	resp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("Playlist request failed: %v", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("playlist %s not found", playlistID)
	}

	return &feeds.Feed{Title: resp.Items[0].Snippet.Title, Description: resp.Items[0].Snippet.Description}, nil
}

func getYoutubeService(apiKey string) *youtube.Service {
	client := &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	}
	service, _ := youtube.New(client)
	// Only for testing
	service.BasePath = youtubeAPIURLBase
	return service
}
