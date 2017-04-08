package command

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Video is all the video data needed to create an rss entry
type Video struct {
	ID          string
	Title       string
	Description string
	Published   time.Time
	Filename    string
}

// FeedInfo represents youtube channel info
type FeedInfo struct {
	Title       string
	Description string
}

type rss struct {
	Version string  `xml:"version,attr"`
	Channel channel `xml:"channel"`
}

type channel struct {
	Title         string `xml:"title"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []item
}

type item struct {
	XMLName     struct{}  `xml:"item"`
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	GUID        string    `xml:"guid"`
	PubDate     string    `xml:"pubDate"`
	Enclosure   enclosure `xml:"enclosure"`
}

type enclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

// buildRssFile builds the rss file from a given list of videos
func buildRssFile(videos []*Video, baseURL string, feedInfo FeedInfo) []byte {
	rssData := rss{Version: "2.0"}
	channelData := channel{Title: feedInfo.Title, Description: feedInfo.Description, LastBuildDate: time.Now().Format(time.RFC1123Z)}
	channelData.Items = []item{}

	for _, video := range videos {
		channelData.Items = append(channelData.Items, *buildItem(video, baseURL))
	}

	rssData.Channel = channelData
	bytes, _ := xml.MarshalIndent(rssData, "", "  ")
	return bytes
}

func buildItem(video *Video, baseURL string) *item {
	thisItem := &item{Title: video.Title, Description: video.Description, GUID: video.ID, PubDate: video.Published.Format(time.RFC1123Z)}
	enc := enclosure{URL: fmt.Sprintf("%s/%s.mp3", baseURL, video.Filename), Type: "audio/mpeg"}
	thisItem.Enclosure = enc
	return thisItem
}
