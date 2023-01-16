package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/TheOtherDavid/gmail-retrieve-message/gmailretrieve"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func main() {
	sender := os.Getenv("TARGET_SENDER")

	message := gmailretrieve.RetrieveUnreadMessageFromSender(sender)
	fmt.Println(message)

	artists := extractArtists(message)
	if len(artists) > 0 {
		fmt.Println(artists)
	} else {
		fmt.Println("No artists found!")
	}

	//Now we create an SNS event

	now := time.Now()
	// Prepare the payload for the SNS message
	payload := map[string]interface{}{
		"PlaylistName": "WGT Announcement " + strconv.Itoa(now.Year()) + string(now.Month()) + strconv.Itoa(now.Day()),
		"ArtistList":   artists,
	}
	payloadBytes, _ := json.Marshal(payload)

	topicArn := os.Getenv("AWS_TOPIC_ARN")
	// Prepare the SNS message
	snsMessage := &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(string(payloadBytes)),
		Subject:  aws.String("New band list"),
	}

	svc := sns.New(session.New())
	// Send the SNS message
	result, err := svc.Publish(snsMessage)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("SNS message sent, message ID: " + *result.MessageId)
	}
}

func extractArtists(s string) []string {
	// Split the string into lines
	lines := strings.Split(s, "\n")

	// Initialize a slice to store the artist names
	artists := make([]string, 0)

	// Iterate over the lines
	for _, line := range lines {
		// Check if the line starts with an uppercase letter
		if len(line) > 0 && line[0] >= 'A' && line[0] <= 'Z' {
			// If it does, split the line into sections, potential artist names, separated by commas
			potentialArtists := strings.Split(line, ",")

			// Iterate over the potential artists
			for _, potentialArtist := range potentialArtists {
				potentialArtist = strings.TrimSpace(potentialArtist)
				// Check if the potential artist ends with ")"
				if strings.HasSuffix(potentialArtist, ")") {
					// If it does, remove the parenthesis and add the artist to the slice
					trimmedArtist := trimCountry(potentialArtist)
					artists = append(artists, trimmedArtist)
				}
			}
		}
	}
	return artists
}

func trimCountry(s string) string {
	// Find the first index of " ("
	i := strings.Index(s, " (")
	// Return a substring that starts at the beginning of the string and ends at the index of " ("
	return s[:i]
}
