package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/kennygrant/sanitize"
)

// PlaylistScraper retrieves data about youtube videos
type PlaylistScraper struct {
	youtubeService *youtube.Service
}

// NewPlaylistScraper returns a YoutubeScraper
func NewPlaylistScraper(apiKey string) *PlaylistScraper {
	youtubeService := getYoutubeService(apiKey)
	return &PlaylistScraper{youtubeService: youtubeService}
}

func parsePlaylistItems(results []*youtube.PlaylistItem) ([]*VideoData, error) {
	items := make([]*VideoData, 0, len(results))
	for _, result := range results {
		publishedTime, err := time.Parse(time.RFC3339, result.Snippet.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing publish date on video %s: %v", result.Snippet.ResourceId.VideoId, err)
		}

		item := &VideoData{
			GUID:        result.Snippet.ResourceId.VideoId,
			Link:        fmt.Sprintf("https://youtu.be/%s", result.Snippet.ResourceId.VideoId),
			Title:       result.Snippet.Title,
			Description: fmt.Sprintf("%s https://youtu.be/%s", result.Snippet.Description, result.Snippet.ResourceId.VideoId),
			FileName:    fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(result.Snippet.Title), " ", "-", -1), result.Snippet.ResourceId.VideoId),
			PubDate:     publishedTime,
		}

		if result.Snippet.Thumbnails != nil && result.Snippet.Thumbnails.Default != nil {
			item.Image = result.Snippet.Thumbnails.Default.Url
		}

		items = append(items, item)
	}

	return items, nil
}

// GetVideosForPlaylist returns an array of all the youtube video ids in a playlist
func (scraper PlaylistScraper) GetVideosForPlaylist(playlistID string) ([]*VideoData, *ChannelInfo, error) {
	info, err := scraper.getPlaylistInfo(playlistID)
	if err != nil {
		return nil, nil, err
	}

	items := make([]*VideoData, 0)
	listCall := scraper.youtubeService.PlaylistItems.List("snippet").PlaylistId(playlistID)
	err = listCall.Pages(context.Background(), func(resp *youtube.PlaylistItemListResponse) error {
		videoPage, pageErr := parsePlaylistItems(resp.Items)
		if pageErr != nil {
			return pageErr
		}

		items = append(items, videoPage...)
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("playlist items request failed: %v", err)
	}

	return items, info, nil
}

func (scraper PlaylistScraper) getPlaylistInfo(playlistID string) (*ChannelInfo, error) {
	listCall := scraper.youtubeService.Playlists.List("snippet").Id(playlistID)
	resp, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("Playlist request failed: %v", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("playlist %s not found", playlistID)
	}

	feed := &ChannelInfo{
		Title:       resp.Items[0].Snippet.Title,
		Link:        fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistID),
		Description: resp.Items[0].Snippet.Description,
	}
	if resp.Items[0].Snippet.Thumbnails != nil && resp.Items[0].Snippet.Thumbnails.Default != nil {
		feed.Thumbnail = resp.Items[0].Snippet.Thumbnails.Default.Url
	}
	return feed, nil
}
