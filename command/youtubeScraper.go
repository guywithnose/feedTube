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

	"github.com/eduncan911/podcast"
	"github.com/kennygrant/sanitize"
)

var youtubeAPIURLBase = "https://www.googleapis.com/youtube/v3/"

// getVideosForChannel returns an array of all the youtube video ids on a channel
func getVideosForChannel(apiKey, channelName, after string, writer io.Writer) (<-chan *youtubeItem, *podcast.Podcast, error) {
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

		item := &youtubeItem{
			Item: podcast.Item{
				GUID:        result.Id.VideoId,
				Link:        fmt.Sprintf("https://youtu.be/%s", result.Id.VideoId),
				Title:       result.Snippet.Title,
				Description: fmt.Sprintf("%s https://youtu.be/%s", result.Snippet.Description, result.Id.VideoId),
			},
			Filename: fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(result.Snippet.Title), " ", "-", -1), result.Id.VideoId),
		}
		if result.Snippet.Thumbnails != nil && result.Snippet.Thumbnails.Default != nil {
			item.AddImage(result.Snippet.Thumbnails.Default.Url)
		}

		item.AddPubDate(&publishedTime)
		videos <- item
	}
}

func parsePlaylistItems(results []*youtube.PlaylistItem, videos chan<- *youtubeItem, errWriter io.Writer) {
	for _, result := range results {
		publishedTime, err := time.Parse(time.RFC3339, result.Snippet.PublishedAt)
		if err != nil {
			fmt.Fprintf(errWriter, "error parsing publish date on video %s: %v\n", result.Snippet.ResourceId.VideoId, err)
			continue
		}

		item := &youtubeItem{
			Item: podcast.Item{
				GUID:        result.Snippet.ResourceId.VideoId,
				Link:        fmt.Sprintf("https://youtu.be/%s", result.Snippet.ResourceId.VideoId),
				Title:       result.Snippet.Title,
				Description: fmt.Sprintf("%s https://youtu.be/%s", result.Snippet.Description, result.Snippet.ResourceId.VideoId),
			},
			Filename: fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(result.Snippet.Title), " ", "-", -1), result.Snippet.ResourceId.VideoId),
		}
		if result.Snippet.Thumbnails != nil && result.Snippet.Thumbnails.Default != nil {
			item.AddImage(result.Snippet.Thumbnails.Default.Url)
		}

		item.AddPubDate(&publishedTime)
		videos <- item
	}
}

// getVideosForPlaylist returns an array of all the youtube video ids in a playlist
func getVideosForPlaylist(apiKey, playlistID string, writer io.Writer) (<-chan *youtubeItem, *podcast.Podcast, error) {
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

func getChannelInfo(apiKey, channelID string) (string, *podcast.Podcast, error) {
	channel, idErr := getChannelByID(apiKey, channelID)
	if idErr != nil {
		var err error
		channel, err = getChannelByName(apiKey, channelID)
		if err != nil {
			return "", nil, fmt.Errorf("%v: %v", idErr, err)
		}
	}

	now := time.Now()
	info := podcast.New(channel.Snippet.Title, fmt.Sprintf("https://www.youtube.com/channel/%s", channel.Id), channel.Snippet.Description, &now, &now)
	if channel.Snippet.Thumbnails != nil && channel.Snippet.Thumbnails.Default != nil {
		info.AddImage(channel.Snippet.Thumbnails.Default.Url)
	}

	return channel.Id, &info, nil
}

func getChannelByName(apiKey, channelName string) (*youtube.Channel, error) {
	youtubeService := getYoutubeService(apiKey)
	listCall := youtubeService.Channels.List("snippet").ForUsername(channelName)
	items, err := makeChannelRequest(listCall)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("Channel %s not found", channelName)
	}

	return items[0], nil
}

func getChannelByID(apiKey, channelName string) (*youtube.Channel, error) {
	youtubeService := getYoutubeService(apiKey)
	listCall := youtubeService.Channels.List("snippet").Id(channelName)
	items, err := makeChannelRequest(listCall)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("Channel ID %s not found", channelName)
	}

	return items[0], nil
}

func makeChannelRequest(listCall *youtube.ChannelsListCall) ([]*youtube.Channel, error) {
	resp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("Channel request failed: %v", err)
	}

	return resp.Items, nil
}

func getPlaylistInfo(apiKey, playlistID string) (*podcast.Podcast, error) {
	youtubeService := getYoutubeService(apiKey)
	listCall := youtubeService.Playlists.List("snippet").Id(playlistID)
	resp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("Playlist request failed: %v", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("playlist %s not found", playlistID)
	}

	now := time.Now()
	feed := podcast.New(
		resp.Items[0].Snippet.Title,
		fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistID),
		resp.Items[0].Snippet.Description,
		&now,
		&now,
	)
	if resp.Items[0].Snippet.Thumbnails != nil && resp.Items[0].Snippet.Thumbnails.Default != nil {
		feed.AddImage(resp.Items[0].Snippet.Thumbnails.Default.Url)
	}
	return &feed, nil
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
