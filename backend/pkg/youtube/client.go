package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	apiKey string
	client *http.Client
}

type VideoInfo struct {
	ID          string
	Title       string
	Description string
	ChannelID   string
	ChannelName string
	Duration    int
	ViewCount   int64
	LikeCount   int64
	PublishedAt time.Time
	ThumbnailURL string
	Tags        []string
	Category    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ExtractVideoID(videoURL string) (string, error) {
	patterns := []string{
		`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})`,
		`youtube\.com\/watch\?.*v=([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(videoURL)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("invalid YouTube URL")
}

func (c *Client) GetVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error) {
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?id=%s&key=%s&part=snippet,statistics,contentDetails",
		videoID, c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s", string(body))
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string    `json:"title"`
				Description string    `json:"description"`
				ChannelID   string    `json:"channelId"`
				ChannelName string    `json:"channelTitle"`
				PublishedAt time.Time `json:"publishedAt"`
				Thumbnails  struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
				Tags     []string `json:"tags"`
				Category string   `json:"categoryId"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
				LikeCount string `json:"likeCount"`
			} `json:"statistics"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	item := result.Items[0]

	// Parse duration (ISO 8601 format: PT1H2M10S)
	duration := parseDuration(item.ContentDetails.Duration)

	// Parse counts
	var viewCount, likeCount int64
	fmt.Sscanf(item.Statistics.ViewCount, "%d", &viewCount)
	fmt.Sscanf(item.Statistics.LikeCount, "%d", &likeCount)

	thumbnailURL := item.Snippet.Thumbnails.High.URL
	if thumbnailURL == "" {
		thumbnailURL = item.Snippet.Thumbnails.Default.URL
	}

	// Get channel name
	channelName := item.Snippet.ChannelName

	return &VideoInfo{
		ID:           item.ID,
		Title:        item.Snippet.Title,
		Description:  item.Snippet.Description,
		ChannelID:    item.Snippet.ChannelID,
		ChannelName:  channelName,
		Duration:     duration,
		ViewCount:    viewCount,
		LikeCount:    likeCount,
		PublishedAt:  item.Snippet.PublishedAt,
		ThumbnailURL: thumbnailURL,
		Tags:         item.Snippet.Tags,
		Category:     item.Snippet.Category,
	}, nil
}

func parseDuration(durationStr string) int {
	// Parse ISO 8601 duration (PT1H2M10S)
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(durationStr)

	var hours, minutes, seconds int
	if len(matches) > 1 && matches[1] != "" {
		fmt.Sscanf(matches[1], "%d", &hours)
	}
	if len(matches) > 2 && matches[2] != "" {
		fmt.Sscanf(matches[2], "%d", &minutes)
	}
	if len(matches) > 3 && matches[3] != "" {
		fmt.Sscanf(matches[3], "%d", &seconds)
	}

	return hours*3600 + minutes*60 + seconds
}

func (c *Client) GetCaptions(ctx context.Context, videoID string) ([]CaptionTrack, error) {
	// This would require scraping or using YouTube Data API v3 captions endpoint
	// For now, return empty - will be implemented with actual caption fetching
	return nil, fmt.Errorf("caption fetching not yet implemented")
}

type CaptionTrack struct {
	Language string
	URL      string
}

// SearchRelatedVideos searches for videos related to the given video ID
// First gets the video info, then searches for similar videos based on title, description, and tags
func (c *Client) SearchRelatedVideos(ctx context.Context, videoID string, maxResults int) ([]VideoInfo, error) {
	if maxResults > 50 {
		maxResults = 50 // YouTube API limit
	}
	if maxResults <= 0 {
		maxResults = 10
	}

	// Get video info first to get channel ID and other metadata
	videoInfo, err := c.GetVideoInfo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// Try multiple strategies to find related videos
	// Strategy 1: Search by channel (most reliable)
	var apiURL string
	if videoInfo.ChannelID != "" {
		// Search for videos from the same channel, excluding current video
		apiURL = fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/search?channelId=%s&type=video&part=snippet&maxResults=%d&key=%s&order=date&publishedAfter=2020-01-01T00:00:00Z",
			videoInfo.ChannelID, maxResults+1, c.apiKey,
		)
	} else {
		// Fallback: Search by title and tags
		searchQuery := videoInfo.Title
		if len(videoInfo.Tags) > 0 {
			tagCount := 3
			if len(videoInfo.Tags) < tagCount {
				tagCount = len(videoInfo.Tags)
			}
			for i := 0; i < tagCount; i++ {
				searchQuery += " " + videoInfo.Tags[i]
			}
		}
		encodedQuery := url.QueryEscape(strings.TrimSpace(searchQuery))
		apiURL = fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/search?q=%s&type=video&part=snippet&maxResults=%d&key=%s&order=relevance",
			encodedQuery, maxResults+1, c.apiKey,
		)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s", string(body))
	}

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title       string    `json:"title"`
				Description string    `json:"description"`
				ChannelID   string    `json:"channelId"`
				ChannelName string    `json:"channelTitle"`
				PublishedAt time.Time `json:"publishedAt"`
				Thumbnails  struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to VideoInfo and fetch full details for each video
	var videos []VideoInfo
	for _, item := range result.Items {
		// Skip the original video
		if item.ID.VideoID == videoID {
			continue
		}

		// Get full video details (for statistics, duration, etc.)
		videoInfo, err := c.GetVideoInfo(ctx, item.ID.VideoID)
		if err != nil {
			// If we can't get full details, create minimal info from search result
			thumbnailURL := item.Snippet.Thumbnails.High.URL
			if thumbnailURL == "" {
				thumbnailURL = item.Snippet.Thumbnails.Default.URL
			}
			videoInfo = &VideoInfo{
				ID:           item.ID.VideoID,
				Title:        item.Snippet.Title,
				Description:  item.Snippet.Description,
				ChannelID:    item.Snippet.ChannelID,
				ChannelName:  item.Snippet.ChannelName,
				PublishedAt:  item.Snippet.PublishedAt,
				ThumbnailURL: thumbnailURL,
			}
		}
		videos = append(videos, *videoInfo)
		
		// Stop if we have enough results
		if len(videos) >= maxResults {
			break
		}
	}

	return videos, nil
}

