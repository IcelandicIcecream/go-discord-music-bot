package services

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YouTubeService struct {
	Service *youtube.Service
}

func NewYouTubeService(apiKey string) (*YouTubeService, error) {
	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &YouTubeService{Service: service}, nil
}

func (y *YouTubeService) SearchVideos(searchQuery string, maxResults int64) (*youtube.SearchListResponse, error) {
	call := y.Service.Search.
		List([]string{
			"id",
			"snippet",
		}).
		Q(searchQuery).
		MaxResults(maxResults)

	return call.Do()
}
