package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	YouTubeAPIKey    = os.Getenv("YOUTUBE_API_KEY")
	YouTubeChannelID = os.Getenv("YOUTUBE_CHANNEL_ID")
	BlueSkyUsername  = os.Getenv("BLUESKY_USERNAME")
	BlueSkyAppPass   = os.Getenv("BLUESKY_APP_PASS")
)

type YouTubeResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

func GetLatestYouTubeVideo() (string, string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?key=%s&channelId=%s&part=snippet,id&order=date&maxResults=1", YouTubeAPIKey, YouTubeChannelID)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ytResp YouTubeResponse
	json.Unmarshal(body, &ytResp)

	if len(ytResp.Items) > 0 {
		videoID := ytResp.Items[0].ID.VideoID
		title := ytResp.Items[0].Snippet.Title
		return title, "https://www.youtube.com/watch?v=" + videoID, nil
	}
	return "", "", fmt.Errorf("no videos found")
}

func main() {
	title, videoURL, err := GetLatestYouTubeVideo()
	if err != nil {
		fmt.Println("Error fetching YouTube video:", err)
		return
	}
	fmt.Println("Latest Video:", title)
	fmt.Println("Video URL:", videoURL)
}
