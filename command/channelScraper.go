package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	youtube "google.golang.org/api/youtube/v3"

	"github.com/kennygrant/sanitize"
)

//TODO better name

// ChannelInfo contains the metadata for a channel
type ChannelInfo struct {
	Title       string
	Description string
	Link        string
	Thumbnail   string
}

// ChannelScraper retrieves data about youtube videos
type ChannelScraper struct {
	youtubeService *youtube.Service
}

// NewChannelScraper returns a YoutubeScraper
func NewChannelScraper(apiKey string) *ChannelScraper {
	youtubeService := getYoutubeService(apiKey)
	return &ChannelScraper{youtubeService: youtubeService}
}

// GetVideosForChannel returns an array of all the youtube video ids on a channel
func (scraper ChannelScraper) GetVideosForChannel(channelName, after string) ([]*VideoData, *ChannelInfo, error) {
	channelID, info, err := scraper.getChannelInfo(channelName)
	if err != nil {
		return nil, nil, err
	}

	listCall, err := scraper.buildSearchListCall(channelID, after)
	if err != nil {
		return nil, nil, err
	}

	items := make([]*VideoData, 0)
	err = listCall.Pages(context.Background(), func(resp *youtube.SearchListResponse) error {
		videoPage, pageErr := parseSearchResults(resp.Items)
		if pageErr != nil {
			return pageErr
		}

		items = append(items, videoPage...)
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("search request failed: %v", err)
	}

	return items, info, nil
}

func (scraper ChannelScraper) buildSearchListCall(channelID, after string) (*youtube.SearchListCall, error) {
	listCall := scraper.youtubeService.Search.List("snippet").ChannelId(channelID).Type("video")
	if after != "" {
		afterTime, timeParseErr := time.Parse("01-02-06", after)
		if timeParseErr != nil {
			return nil, fmt.Errorf("could not parse after date: %v", timeParseErr)
		}

		listCall = listCall.PublishedAfter(afterTime.Format(time.RFC3339))
	}

	return listCall, nil
}

func parseSearchResults(results []*youtube.SearchResult) ([]*VideoData, error) {
	items := make([]*VideoData, 0, len(results))
	for _, result := range results {
		publishedTime, err := time.Parse(time.RFC3339, result.Snippet.PublishedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing publish date on video %s: %v", result.Id.VideoId, err)
		}

		if result.Snippet.LiveBroadcastContent != "none" {
			continue
		}

		item := &VideoData{
			GUID:        result.Id.VideoId,
			Link:        fmt.Sprintf("https://youtu.be/%s", result.Id.VideoId),
			Title:       result.Snippet.Title,
			Description: fmt.Sprintf("%s https://youtu.be/%s", result.Snippet.Description, result.Id.VideoId),
			FileName:    fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(result.Snippet.Title), " ", "-", -1), result.Id.VideoId),
			PubDate:     publishedTime,
		}

		if result.Snippet.Thumbnails != nil && result.Snippet.Thumbnails.Default != nil {
			item.Image = result.Snippet.Thumbnails.Default.Url
		}

		items = append(items, item)
	}

	return items, nil
}

func (scraper ChannelScraper) getChannelInfo(channelID string) (string, *ChannelInfo, error) {
	channel, idErr := scraper.getChannelByID(channelID)
	if idErr != nil {
		var err error
		channel, err = scraper.getChannelByName(channelID)
		if err != nil {
			return "", nil, fmt.Errorf("%v: %v", idErr, err)
		}
	}

	info := &ChannelInfo{
		Title:       channel.Snippet.Title,
		Link:        fmt.Sprintf("https://www.youtube.com/channel/%s", channel.Id),
		Description: channel.Snippet.Description,
	}
	if channel.Snippet.Thumbnails != nil && channel.Snippet.Thumbnails.Default != nil {
		info.Thumbnail = channel.Snippet.Thumbnails.Default.Url
	}

	return channel.Id, info, nil
}

func (scraper ChannelScraper) getChannelByName(channelName string) (*youtube.Channel, error) {
	listCall := scraper.youtubeService.Channels.List("snippet").ForUsername(channelName)
	items, err := makeChannelRequest(listCall)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("Channel %s not found", channelName)
	}

	return items[0], nil
}

func (scraper ChannelScraper) getChannelByID(channelName string) (*youtube.Channel, error) {
	listCall := scraper.youtubeService.Channels.List("snippet").Id(channelName)
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
