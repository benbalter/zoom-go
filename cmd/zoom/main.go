package main

import (
	"fmt"
	"os"

	"github.com/skratchdot/open-golang/open"

	zoom "github.com/benbalter/zoom-go"
	"github.com/benbalter/zoom-go/config"
)

func main() {
	provider, err := config.NewFileProvider()
	if err != nil {
		fmt.Printf("unable to create file configuration provider: %+v\n", err)
		os.Exit(1)
	}

	calendar, err := zoom.NewGoogleCalendarService(provider)
	if err != nil {
		fmt.Printf("error creating google calendar client: %+v\n", err)
		os.Exit(1)
	}

	meeting, err := zoom.NextEvent(calendar)
	if err != nil {
		fmt.Printf("error fetching next meeting: %+v\n", err)
		os.Exit(1)
	}

	if meeting == nil {
		fmt.Println("No upcoming events found.")
		return
	}

	fmt.Printf("Your next meeting is %q.\n", meeting.Summary)

	url, ok := zoom.MeetingURLFromEvent(meeting)
	if !ok {
		fmt.Println("Your next meeting is not a Zoom meeting.")
		os.Exit(1)
	}

	fmt.Printf("It starts %s", zoom.HumanizedStartTime(meeting))

	if zoom.IsMeetingSoon(meeting) {
		fmt.Printf("Opening %s...\n", url)
		open.Run(url.String())
	} else {
		fmt.Printf("Here's the URL %s\n", url)
	}

	fmt.Printf("Oh, and here's the URL in case you need it: %s\n", meeting.HtmlLink)
}
