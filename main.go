package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("GH_TOKEN")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	for notification := range notifications(ctx, client) {
		if notification.Unread != nil && *notification.Unread {
			popUp(notification)
		}
	}
}

func popUp(notification *github.Notification) {

	text := "UNKNOWN"
	if notification.Subject != nil && notification.Subject.Title != nil {
		text = *notification.Subject.Title
	}
	subtitle := "UNKNOWN"
	if notification.Subject != nil && notification.Subject.Type != nil {
		subtitle = *notification.Subject.Type
	}
	if notification.Reason != nil {
		subtitle += " / " + *notification.Reason
	}
	title := "Github"
	sound := "Glass"

	if notification.Subject != nil && notification.Subject.LatestCommentURL != nil {
		fmt.Printf("\nURL: %s\n", *notification.Subject.LatestCommentURL)
	}

	script := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\" sound name \"%s\"", text, title, subtitle, sound)
	output, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not send notification: %v\n", err)
	}
	if len(output) > 0 {
		fmt.Printf("Notification Log: %s\n%s\n", script, output)
	}
}

func notifications(ctx context.Context, client *github.Client) chan *github.Notification {
	channel := make(chan *github.Notification)

	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		since := time.Now().AddDate(0, 0, -7) // week ago

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				objects, _, err := client.Activity.ListNotifications(ctx, &github.NotificationListOptions{
					Since: since,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting gh notifications %v\n", err)
				}
				for _, object := range objects {
					if object.UpdatedAt != nil {
						if since.Before(*object.UpdatedAt) {
							since = *object.UpdatedAt
						} else if *object.UpdatedAt == since {
							continue
						}
					}

					channel <- object
				}
			}
		}
	}()
	return channel
}
