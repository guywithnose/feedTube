package youtubeScraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/guywithnose/feedTube/rssBuilder"
	"github.com/kennygrant/sanitize"
)

type videoSearchResponse struct {
	NextPageToken string                   `json:"nextPageToken"`
	Items         []map[string]interface{} `json:"items"`
}

type channelResponse struct {
	Items []channelItem `json:"items"`
}

type channelItem struct {
	ID      string                 `json:"id"`
	Snippet map[string]interface{} `json:"snippet"`
}

// GetVideosForChannel returns an array of all the youtube video ids on a channel
func GetVideosForChannel(apiKey, channelName string, writer io.Writer) ([]rssBuilder.Video, *rssBuilder.FeedInfo, error) {
	channelID, info, err := getChannelIDFromName(apiKey, channelName)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?key=%s&part=snippet&channelId=%s&maxResults=50&type=video", apiKey, channelID)
	nextPageToken, videos, err := runRequest("", url, writer)
	if err != nil {
		return nil, nil, err
	}

	var newVideos []rssBuilder.Video
	for nextPageToken != "" {
		nextPageToken, newVideos, err = runRequest(nextPageToken, url, writer)
		if err != nil {
			return nil, nil, err
		}

		videos = append(videos, newVideos...)
	}

	return videos, info, nil
}

// GetVideosForPlaylist returns an array of all the youtube video ids in a playlist
func GetVideosForPlaylist(apiKey, playlistID string, writer io.Writer) ([]rssBuilder.Video, *rssBuilder.FeedInfo, error) {
	info, err := getPlaylistInfo(apiKey, playlistID)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?key=%s&part=snippet&playlistId=%s&maxResults=50", apiKey, playlistID)
	nextPageToken, videos, err := runRequest("", url, writer)
	if err != nil {
		return nil, nil, err
	}

	var newVideos []rssBuilder.Video
	for nextPageToken != "" {
		nextPageToken, newVideos, err = runRequest(nextPageToken, url, writer)
		if err != nil {
			return nil, nil, err
		}

		videos = append(videos, newVideos...)
	}

	return videos, info, nil
}

func getChannelIDFromName(apiKey, channelName string) (string, *rssBuilder.FeedInfo, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?key=%s&part=snippet&forUsername=%s", apiKey, channelName)
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}

	var body channelResponse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", nil, err
	}

	if len(body.Items) == 0 {
		return "", nil, fmt.Errorf("channel %s not found", channelName)
	}

	item := body.Items[0]
	title, _ := item.Snippet["title"].(string)
	description, _ := item.Snippet["description"].(string)

	return item.ID, &rssBuilder.FeedInfo{Title: title, Description: description}, nil
}

func getPlaylistInfo(apiKey, playlistID string) (*rssBuilder.FeedInfo, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlists?key=%s&part=snippet&id=%s&maxResults=1", apiKey, playlistID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var body channelResponse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	if len(body.Items) == 0 {
		fmt.Println(url)
		return nil, fmt.Errorf("playlist %s not found", playlistID)
	}

	item := body.Items[0]
	title, _ := item.Snippet["title"].(string)
	description, _ := item.Snippet["description"].(string)

	return &rssBuilder.FeedInfo{Title: title, Description: description}, nil
}

func runRequest(pageToken, url string, writer io.Writer) (string, []rssBuilder.Video, error) {
	if pageToken != "" {
		url = fmt.Sprintf(url+"&pageToken=%s", pageToken)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}

	var body videoSearchResponse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", nil, err
	}

	videos := make([]rssBuilder.Video, 0, 50)
	for _, item := range body.Items {
		video, err := parseVideoItem(item)
		if err != nil {
			fmt.Fprintln(writer, err)
			continue
		}

		videos = append(videos, *video)
	}

	return body.NextPageToken, videos, nil
}

func parseVideoItem(item map[string]interface{}) (*rssBuilder.Video, error) {
	snippet := item["snippet"].(map[string]interface{})
	title, ok := snippet["title"].(string)
	rootID := item["id"]
	var id map[string]interface{}
	switch rid := rootID.(type) {
	case map[string]interface{}:
		id = rid
		if rid["kind"] != "youtube#video" {
			return nil, errors.New("Not a video item")
		}
	case string:
		id = snippet["resourceId"].(map[string]interface{})
	}

	if !ok {
		return nil, fmt.Errorf("title not set on video %s", id)
	}

	description, ok := snippet["description"].(string)
	if !ok {
		return nil, fmt.Errorf("description not set on video %s", id)
	}

	publishedAt, ok := snippet["publishedAt"].(string)
	if !ok {
		return nil, fmt.Errorf("published Date not set on video %s", id)
	}

	publishedTime, err := time.Parse(time.RFC3339, publishedAt)
	if err != nil {
		return nil, fmt.Errorf("error parsing publidh date on video %s", id)
	}

	return &rssBuilder.Video{
		ID:          id["videoId"].(string),
		Title:       title,
		Description: description,
		Published:   publishedTime,
		Filename:    fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(title), " ", "-", -1), id["videoId"].(string)),
	}, nil
}
