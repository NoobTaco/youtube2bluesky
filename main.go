package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Fetch API keys from environment variables
var (
	YouTubeAPIKey    = os.Getenv("YOUTUBE_API_KEY")
	YouTubeChannelID = os.Getenv("YOUTUBE_CHANNEL_ID")
	BlueSkyUsername  = os.Getenv("BLUESKY_USERNAME")
	BlueSkyAppPass   = os.Getenv("BLUESKY_APP_PASS")
	BlueSkyTemplate  = os.Getenv("BLUESKY_MESSAGE_TEMPLATE") // Expandable message template
)

// YouTube API Response Structure
type YouTubeResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title      string `json:"title"`
			Thumbnails struct {
				High struct {
					URL string `json:"url"`
				} `json:"high"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

// BlueSky Auth Response
type BlueSkyAuthResponse struct {
	AccessJwt string `json:"accessJwt"`
}

func GetLatestYouTubeVideo() (string, string, string, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?key=%s&channelId=%s&part=snippet,id&order=date&maxResults=1", YouTubeAPIKey, YouTubeChannelID)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ytResp YouTubeResponse
	json.Unmarshal(body, &ytResp)

	if len(ytResp.Items) > 0 {
		videoID := ytResp.Items[0].ID.VideoID
		title := ytResp.Items[0].Snippet.Title
		thumbnailURL := ytResp.Items[0].Snippet.Thumbnails.High.URL
		return title, "https://www.youtube.com/watch?v=" + videoID, thumbnailURL, nil
	}
	return "", "", "", fmt.Errorf("no videos found")
}

// PostToBlueSky posts a formatted message to BlueSky
func PostToBlueSky(title, videoURL, thumbnailURL string) error {
	// Authenticate with BlueSky
	authData := map[string]string{"identifier": BlueSkyUsername, "password": BlueSkyAppPass}
	authBody, _ := json.Marshal(authData)

	authResp, err := http.Post("https://bsky.social/xrpc/com.atproto.server.createSession", "application/json", bytes.NewBuffer(authBody))
	if err != nil {
		return err
	}
	defer authResp.Body.Close()

	authResponseBody, _ := io.ReadAll(authResp.Body)
	var authResponse BlueSkyAuthResponse
	json.Unmarshal(authResponseBody, &authResponse)

	if authResponse.AccessJwt == "" {
		return fmt.Errorf("failed to authenticate with BlueSky")
	}

	// Apply custom message formatting
	if BlueSkyTemplate == "" {
		BlueSkyTemplate = "🎥 New Video: %s\n📺 Watch here: %s"
	}
	postMessage := fmt.Sprintf(BlueSkyTemplate, title, videoURL)

	// Format the post data with a YouTube Embed and Thumbnail
	postData := map[string]interface{}{
		"collection": "app.bsky.feed.post",
		"repo":       BlueSkyUsername,
		"record": map[string]interface{}{
			"text":      postMessage,
			"createdAt": time.Now().Format(time.RFC3339),
			"embed": map[string]interface{}{
				"$type": "app.bsky.embed.external",
				"external": map[string]interface{}{
					"uri":         videoURL,
					"title":       title,
					"description": "Watch now on YouTube",
					"thumbnail": map[string]interface{}{
						"uri": thumbnailURL,
					},
				},
			},
		},
	}
	postBody, _ := json.Marshal(postData)

	// Send post to BlueSky
	req, _ := http.NewRequest("POST", "https://bsky.social/xrpc/com.atproto.repo.createRecord", bytes.NewBuffer(postBody))
	req.Header.Set("Authorization", "Bearer "+authResponse.AccessJwt)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to post: %s", body)
	}

	fmt.Println("✅ Successfully posted to BlueSky with YouTube Embed and Thumbnail!")
	return nil
}

func main() {
	title, videoURL, thumbnailURL, err := GetLatestYouTubeVideo()
	if err != nil {
		fmt.Println("❌ Error fetching YouTube video:", err)
		return
	}

	err = PostToBlueSky(title, videoURL, thumbnailURL)
	if err != nil {
		fmt.Println("❌ Error posting to BlueSky:", err)
	}
}
