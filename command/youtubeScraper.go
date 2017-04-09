package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
)

var youtubeAPIURLBase = "https://www.googleapis.com/youtube/v3"

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

// getVideosForChannel returns an array of all the youtube video ids on a channel
func getVideosForChannel(apiKey, channelName, after string, writer io.Writer) ([]Video, *FeedInfo, error) {
	channelID, info, err := getChannelIDFromName(apiKey, channelName)
	if err != nil {
		return nil, nil, err
	}

	searchParams := url.Values{}
	searchParams["key"] = []string{apiKey}
	searchParams["part"] = []string{"snippet"}
	searchParams["channelId"] = []string{channelID}
	searchParams["maxResults"] = []string{"50"}
	searchParams["type"] = []string{"video"}
	if after != "" {
		afterTime, timeParseErr := time.Parse("01-02-06", after)
		if timeParseErr != nil {
			return nil, nil, fmt.Errorf("Could not parse after date: %v", timeParseErr)
		}

		searchParams["publishedAfter"] = []string{afterTime.Format(time.RFC3339)}
	}

	url := fmt.Sprintf("%s/search?%s", youtubeAPIURLBase, searchParams.Encode())
	nextPageToken, videos, err := runRequest("", url, writer)
	if err != nil {
		return nil, nil, err
	}

	var newVideos []Video
	for nextPageToken != "" {
		nextPageToken, newVideos, err = runRequest(nextPageToken, url, writer)
		if err != nil {
			return nil, nil, err
		}

		videos = append(videos, newVideos...)
	}

	return videos, info, nil
}

// getVideosForPlaylist returns an array of all the youtube video ids in a playlist
func getVideosForPlaylist(apiKey, playlistID string, writer io.Writer) ([]Video, *FeedInfo, error) {
	info, err := getPlaylistInfo(apiKey, playlistID)
	if err != nil {
		return nil, nil, err
	}

	searchParams := url.Values{}
	searchParams["key"] = []string{apiKey}
	searchParams["part"] = []string{"snippet"}
	searchParams["playlistId"] = []string{playlistID}
	searchParams["maxResults"] = []string{"50"}

	url := fmt.Sprintf("%s/playlistItems?%s", youtubeAPIURLBase, searchParams.Encode())
	nextPageToken, videos, err := runRequest("", url, writer)
	if err != nil {
		return nil, nil, err
	}

	var newVideos []Video
	for nextPageToken != "" {
		nextPageToken, newVideos, err = runRequest(nextPageToken, url, writer)
		if err != nil {
			return nil, nil, err
		}

		videos = append(videos, newVideos...)
	}

	return videos, info, nil
}

func getChannelIDFromName(apiKey, channelName string) (string, *FeedInfo, error) {
	url := fmt.Sprintf("%s/channels?key=%s&part=snippet&forUsername=%s", youtubeAPIURLBase, apiKey, channelName)
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, fmt.Errorf("Could not connect to %s: %v", url, err)
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

	return item.ID, &FeedInfo{Title: title, Description: description}, nil
}

func getPlaylistInfo(apiKey, playlistID string) (*FeedInfo, error) {
	url := fmt.Sprintf("%s/playlists?key=%s&part=snippet&id=%s&maxResults=1", youtubeAPIURLBase, apiKey, playlistID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to %s: %v", url, err)
	}

	var body channelResponse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	if len(body.Items) == 0 {
		return nil, fmt.Errorf("playlist %s not found", playlistID)
	}

	item := body.Items[0]
	title, _ := item.Snippet["title"].(string)
	description, _ := item.Snippet["description"].(string)

	return &FeedInfo{Title: title, Description: description}, nil
}

func runRequest(pageToken, url string, writer io.Writer) (string, []Video, error) {
	if pageToken != "" {
		url = fmt.Sprintf(url+"&pageToken=%s", pageToken)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", nil, fmt.Errorf("Could not connect to %s: %v", url, err)
	}

	defer resp.Body.Close()
	var body videoSearchResponse

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", nil, err
	}

	videos := make([]Video, 0, 50)
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

func parseVideoItem(item map[string]interface{}) (*Video, error) {
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
		return nil, fmt.Errorf("error parsing publish date on video %s", id)
	}

	return &Video{
		ID:          id["videoId"].(string),
		Title:       title,
		Description: description,
		Published:   publishedTime,
		Filename:    fmt.Sprintf("%s-%s", strings.Replace(sanitize.BaseName(title), " ", "-", -1), id["videoId"].(string)),
	}, nil
}
