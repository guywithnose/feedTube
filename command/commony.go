package command

import (
	"net/http"
	"time"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

// YoutubeAPIURLBase is the base url to use for the youtube API.  This should only be changed for tests.
var YoutubeAPIURLBase = "https://www.googleapis.com/youtube/v3/"

// VideoData holds all the data that we scrape from YouTube
type VideoData struct {
	GUID        string
	Link        string
	Title       string
	Description string
	FileName    string
	Image       string
	PubDate     time.Time
}

func getYoutubeService(apiKey string) *youtube.Service {
	client := &http.Client{
		Transport: &transport.APIKey{Key: apiKey},
	}
	service, _ := youtube.New(client)
	// Only for testing
	service.BasePath = YoutubeAPIURLBase
	return service
}
