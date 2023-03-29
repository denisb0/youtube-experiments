package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	query      = flag.String("query", "Google", "Search term")
	videoID    = flag.String("video", "", "Video id")
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

const developerKey = "DEVELOPER_KEY"

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

func main() {
	flag.Parse()

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(developerKey))
	// service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	handleError(videoDetails(service, *videoID, []string{"snippet", "player", "topicDetails", "recordingDetails"}), "show video details")

}
