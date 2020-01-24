// Package zoom provides a way to fetch the next Zoom meeting in your Google calendar.
package zoom

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	calendar "google.golang.org/api/calendar/v3"
)

const googleCalendarDateTimeFormat = time.RFC3339

// NextEvents returns the next N calendar events in your primary calendar.
// It only returns events which contain Zoom video chats.
func NextEvents(service *calendar.Service, count int) ([]*calendar.Event, error) {
	t := time.Now().Add(-5 * time.Minute).Format(time.RFC3339)

	events, err := service.Events.
		List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(int64(count * 10)).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(events.Items) == 0 {
		return nil, nil
	}

	zoomEvents := []*calendar.Event{}
	for _, event := range events.Items {
		if _, ok := MeetingURLFromEvent(event); !ok {
			continue
		}

		zoomEvents = append(zoomEvents, event)

		if len(zoomEvents) == count {
			break
		}
	}

	if len(zoomEvents) == 0 {
		return nil, errors.Errorf("no zoom events upcoming")
	}

	return zoomEvents, nil
}

// NextEventByStartTime takes an array of events and finds the one whose start time is closest to now.
func NextEventByStartTime(events []*calendar.Event) *calendar.Event {
	if len(events) == 0 {
		return nil
	}
	if len(events) == 1 {
		return events[0]
	}

	// Sort based on how far away the start time is from now. For example,
	// a start time 5 minutes in the future (300s away) will be chosen instead
	// of one 30 minutes (1800s away) in the past.
	//
	// This is helpful when you have two events that overlap or are back-to-back.
	var closestEvent *calendar.Event
	var closestEventStartTimeDistance time.Duration = 1<<63 - 1 // start with the max duration so we always get an event
	now := time.Now()
	for _, event := range events {
		t, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		distanceFromNow := now.Sub(t)
		if distanceFromNow < 0 {
			distanceFromNow = distanceFromNow * -1 // absolute value
		}
		if distanceFromNow < closestEventStartTimeDistance {
			closestEvent = event
			closestEventStartTimeDistance = distanceFromNow
		}
	}
	return closestEvent
}

// NextEvent returns the next calendar event in your primary calendar.
// It will list at most 5 events, and select the first one with a Zoom URL if one exists.
func NextEvent(service *calendar.Service) (*calendar.Event, error) {
	events, err := NextEvents(service, 1)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	log.Println(events)
	return events[0], nil
}

// MeetingURLFromEvent returns a URL if the event is a Zoom meeting.
func MeetingURLFromEvent(event *calendar.Event) (*url.URL, bool) {
	input := event.Location + " " + event.Description
	if videoEntryPointURL, ok := conferenceVideoEntryPointURL(event); ok {
		input = videoEntryPointURL + " " + input
	}

	data, ok := extractZoomCallData(input)
	if !ok {
		return nil, ok
	}

	parsedURL, err := url.Parse(data.GetAppURL())
	if err != nil {
		return nil, false
	}
	return parsedURL, true
}

// conferenceVideoEntryPointURL returns the URL for the video entrypoint if one exists.
func conferenceVideoEntryPointURL(event *calendar.Event) (string, bool) {
	if event.ConferenceData == nil {
		return "", false
	}

	for _, entryPoint := range event.ConferenceData.EntryPoints {
		if entryPoint.EntryPointType == "video" && strings.Contains(entryPoint.Uri, "zoom") {
			return entryPoint.Uri, true
		}
	}

	return "", false
}

// IsMeetingSoon returns true if the meeting is less than 5 minutes from now.
func IsMeetingSoon(event *calendar.Event) bool {
	startTime, err := MeetingStartTime(event)
	if err != nil {
		return false
	}
	minutesUntilStart := time.Until(startTime).Minutes()
	return -5 < minutesUntilStart && minutesUntilStart < 5
}

// HumanizedStartTime converts the event's start time to a human-friendly statement.
func HumanizedStartTime(event *calendar.Event) string {
	startTime, err := MeetingStartTime(event)
	if err != nil {
		return err.Error()
	}
	return humanize.Time(startTime)
}

// MeetingStartTime returns the calendar event's start time.
func MeetingStartTime(event *calendar.Event) (time.Time, error) {
	if event == nil || event.Start == nil || event.Start.DateTime == "" {
		return time.Time{}, errors.New("event does not have a start datetime")
	}
	return time.Parse(googleCalendarDateTimeFormat, event.Start.DateTime)
}

// MeetingSummary generates a one-line summary of the meeting as a string.
func MeetingSummary(event *calendar.Event) string {
	if event == nil {
		return ""
	}

	var output bytes.Buffer

	if event.Summary != "" {
		fmt.Fprintf(&output, "Your next meeting is %q", event.Summary)
	} else {
		fmt.Fprint(&output, "You have a meeting coming up")
	}

	if event.Organizer != nil && event.Organizer.DisplayName != "" {
		fmt.Fprintf(&output, ", organized by %s.", event.Organizer.DisplayName)
	} else if event.Creator != nil && event.Creator.DisplayName != "" {
		fmt.Fprintf(&output, ", created by %s.", event.Creator.DisplayName)
	} else {
		fmt.Fprintf(&output, ".")
	}

	return output.String()
}
