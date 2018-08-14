// Package zoom provides a way to fetch the next Zoom meeting in your Google calendar.
package zoom

import (
	"net/url"
	"regexp"
	"strconv"
	"time"

	humanize "github.com/dustin/go-humanize"
	calendar "google.golang.org/api/calendar/v3"
)

var zoomURLRegexp = regexp.MustCompile(`https://.*?\.zoom\.us/(?:j/(\d+)|my/(\S+))`)

// NextEvent returns the next calendar event in your primary calendar.
func NextEvent(service *calendar.Service) (*calendar.Event, error) {
	t := time.Now().Format(time.RFC3339)

	events, err := service.Events.
		List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(1).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}

	if len(events.Items) == 0 {
		return nil, nil
	}

	return events.Items[0], nil
}

// MeetingURLFromEvent returns a URL if the event is a Zoom meeting.
func MeetingURLFromEvent(event *calendar.Event) (*url.URL, bool) {
	matches := zoomURLRegexp.FindAllStringSubmatch(event.Location+" "+event.Description, -1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return nil, false
	}

	// By default, match the whole URL.
	stringURL := matches[0][0]

	// If we have a meeting ID in the URL, then use zoommtg:// instead of the HTTPS URL.
	if len(matches[0]) >= 2 {
		if _, err := strconv.Atoi(matches[0][1]); err == nil {
			stringURL = "zoommtg://zoom.us/join?confno=" + matches[0][1]
		}
	}

	parsedURL, err := url.Parse(stringURL)
	if err != nil {
		return nil, false
	}
	return parsedURL, true
}

// IsMeetingSoon returns true if the meeting is less than 5 minutes from now.
func IsMeetingSoon(event *calendar.Event) bool {
	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return false
	}
	return time.Until(startTime).Minutes() < 5
}

// HumanizedStartTime converts the event's start time to a human-friendly statement.
func HumanizedStartTime(event *calendar.Event) string {
	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return err.Error()
	}
	return humanize.Time(startTime)
}
