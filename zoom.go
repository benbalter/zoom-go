// Package zoom provides a way to fetch the next Zoom meeting in your Google calendar.
package zoom

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	calendar "google.golang.org/api/calendar/v3"
)

const googleCalendarDateTimeFormat = time.RFC3339

var zoomURLRegexp = regexp.MustCompile(`https://.*?zoom\.us/(?:j/(\d+)|my/(\S+))`)
var zoomURLRegexpPwd = regexp.MustCompile(`https://.*?zoom\.us/j/.*pwd=(.*)`)

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
	return events[0], nil
}

// MeetingURLFromEvent returns a URL if the event is a Zoom meeting.
func MeetingURLFromEvent(event *calendar.Event) (*url.URL, bool) {
	input := event.Location + " " + event.Description
	if videoEntryPointURL, ok := conferenceVideoEntryPointURL(event); ok {
		input = videoEntryPointURL + " " + input
	}

	matches := zoomURLRegexp.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
                fmt.Println("No matches...")
		return nil, false
	}

	haspass := zoomURLRegexpPwd.FindAllStringSubmatch(event.Description, -1)


	// By default, match the whole URL.
	stringURL := matches[0][0]

	// If we have a meeting ID in the URL, then use zoommtg:// instead of the HTTPS URL.
	if len(matches[0]) >= 2 {
		if _, err := strconv.Atoi(matches[0][1]); err == nil {
			if len(haspass) >= 1 {
				if len(haspass[0]) >= 2 {
					stringURL = "zoommtg://zoom.us/join?confno=" + matches[0][1] + "&pwd=" + haspass[0][1]
				} else {
					stringURL = "zoommtg://zoom.us/join?confno=" + matches[0][1]
				}
			} else {
				stringURL = "zoommtg://zoom.us/join?confno=" + matches[0][1]
			}
		}
	}

        fmt.Println(stringURL)
	parsedURL, err := url.Parse(stringURL)
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
