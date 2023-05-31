package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	query      = flag.String("query", "Google", "Search term")
	id         = flag.String("id", "", "Video/channel/playlist id")
	maxResults = flag.Int64("max-results", 5, "Max YouTube results")
)

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

const apiKeyEnv = "API_KEY"

// Print the ID and title of each result in a list as well as a name that
// identifies the list. For example, print the word section name "Videos"
// above a list of video search results, followed by the video ID and title
// of each matching video.
func printIDs(sectionName string, matches map[string]string) {
	fmt.Printf("%v:\n", sectionName)
	for id, title := range matches {
		fmt.Printf("[%v] %v\n", id, title)
	}
	fmt.Printf("\n\n")
}

func search(s *youtube.Service, query string, maxResults int64) error {
	// Make the API call to YouTube.
	call := s.Search.List([]string{"id", "snippet"}).
		Q(query).
		MaxResults(maxResults)
	response, err := call.Do()
	if err != nil {
		return err
	}

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
		case "youtube#channel":
			channels[item.Id.ChannelId] = item.Snippet.Title
		case "youtube#playlist":
			playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
	}

	printIDs("Videos", videos)
	printIDs("Channels", channels)
	printIDs("Playlists", playlists)
	return nil
}

func videoDetails(s *youtube.Service, id string, parts []string) error {
	// Make the API call to YouTube.
	fmt.Printf("get video %s, parts %v\n", id, parts)
	call := s.Videos.List(parts).Id(id)
	response, err := call.Do()
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		fmt.Printf("%+v\n", response)
		return errors.New("no content found")
	}

	v := response.Items[0]

	j, err := json.MarshalIndent(v, "", "   ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func playlistItems(s *youtube.Service, id string, parts []string, pageToken string) error {
	fmt.Printf("get video %s, parts %v\n", id, parts)
	call := s.PlaylistItems.List(parts).PlaylistId(id).MaxResults(3).Fields("items/snippet/title", "items/snippet/resourceId/videoId")
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	response, err := call.Do()
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		fmt.Printf("%+v\n", response)
		return errors.New("no content found")
	}

	j, err := json.MarshalIndent(response, "", "   ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func playlists(s *youtube.Service, id string, parts []string) error {
	fmt.Printf("get playlist %s, parts %v\n", id, parts)
	call := s.Playlists.List(parts).ChannelId(id).MaxResults(50).Fields("items/snippet/title", "items/id")

	response, err := call.Do()
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		fmt.Printf("%+v\n", response)
		return errors.New("no content found")
	}

	j, err := json.MarshalIndent(response, "", "   ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func channels(s *youtube.Service, id string, parts []string) error {
	fmt.Printf("get channels %s, parts %v\n", id, parts)
	call := s.Channels.List(parts).Id(id)

	response, err := call.Do()
	if err != nil {
		return err
	}
	if len(response.Items) == 0 {
		fmt.Printf("%+v\n", response)
		return errors.New("no content found")
	}

	j, err := json.MarshalIndent(response, "", "   ")
	if err != nil {
		return err
	}

	fmt.Println(string(j))

	return nil
}

func getChannelID(s *youtube.Service, channelName string) (string, error) {
	fmt.Printf("get channels %s\n", channelName)

	call := s.Search.List([]string{"id"}).
		Q(channelName).
		MaxResults(1).Type("channel")
	response, err := call.Do()
	if err != nil {
		return "", err
	}

	if len(response.Items) == 0 {
		return "", errors.New("no result found")
	}
	item := response.Items[0]
	if item.Id.Kind != "youtube#channel" {
		return "", errors.New("result not channel type")
	}

	j, err := json.MarshalIndent(response, "", "   ")
	if err != nil {
		return "", err
	}
	fmt.Println(string(j))

	return item.Id.ChannelId, nil
}

func getUploadsPlaylistID(s *youtube.Service, channelID string) (string, error) {
	call := s.Channels.List([]string{"contentDetails"}).MaxResults(1).Id(channelID).Fields("items/contentDetails/relatedPlaylists/uploads")

	response, err := call.Do()
	if err != nil {
		return "", err
	}
	return response.Items[0].ContentDetails.RelatedPlaylists.Uploads, nil
}

func main() {
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv(apiKeyEnv)

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// handleError(videoDetails(service, *videoID, []string{"snippet", "player", "topicDetails", "recordingDetails"}), "show video details")
	// handleError(playlistItems(service, *id, []string{"snippet"}, ""), "show playlist")
	// handleError(playlists(service, *id, []string{"snippet"}), "show playlists")
	// handleError(channels(service, *id, []string{"contentDetails"}), "show playlists")

	// uploads, err := getUploadsPlaylistID(service, *id)
	// handleError(err, "getUploadsPlaylistID")
	// fmt.Println(uploads)
	// handleError(playlistItems(service, uploads, []string{"snippet"}, ""), "show playlist")

	chID, err := getChannelID(service, *id)
	handleError(err, "getChannelID")
	fmt.Println(chID)
}

// my channel UC64mIIOlYMWB5ac6VoaRj8w
// dailydev UCXUjtTfQWJa0G9K_SqRm3jQ
// https://www.youtube.com/@dailydotdev/shorts
