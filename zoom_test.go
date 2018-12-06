package zoom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	calendar "google.golang.org/api/calendar/v3"
)

var testEventResponse = `{
	"items": [
		{
			"kind": "calendar#event",
			"htmlLink": "lalala",
			"created": "2018-10-09T17:00:00-07:00",
			"summary": "I am an in-person meeting",
			"description": "I am a description for said in-person meeting",
			"location": "In a real place!",
			"creator": {
				"email": "parkr@jithub.com",
				"displayName": "Parker Moore"
			},
			"organizer": {
				"email": "kevin@jithub.com",
				"displayName": "Kevin Jithub"
			},
			"start": {
				"dateTime": "2018-10-10T17:00:00-07:00",
				"timeZone": "America/New_York"
			}
		},
		{
			"kind": "calendar#event",
			"htmlLink": "lalala",
			"created": "2018-10-09T17:00:00-07:00",
			"summary": "I am a video call",
			"description": "I am a description for the video call",
			"location": "https://jithub.zoom.us/j/12345",
			"creator": {
				"email": "parkr@jithub.com",
				"displayName": "Parker Moore"
			},
			"organizer": {
				"email": "kevin@jithub.com",
				"displayName": "Kevin Jithub"
			},
			"start": {
				"dateTime": "2018-10-10T17:30:00-07:00",
				"timeZone": "America/New_York"
			}
		}
	]
}`

func TestNextEvent(t *testing.T) {
	mux := http.NewServeMux()

	service, shutdown := newFakeGoogleCalendarService(t, mux)
	defer shutdown()

	actualRequests := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		actualRequests++

		if r.URL.Path == "/calendars/primary/events" {
			query := r.URL.Query()
			assert.Equal(t, query.Get("alt"), "json")
			assert.Equal(t, query.Get("maxResults"), "10")
			assert.Equal(t, query.Get("orderBy"), "startTime")
			assert.Equal(t, query.Get("showDeleted"), "false")
			assert.Equal(t, query.Get("singleEvents"), "true")
			assert.Equal(t, query.Get("timeMin"), time.Now().Format(time.RFC3339))
			fmt.Fprintf(w, testEventResponse)
		} else {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	event, err := NextEvent(service)
	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, 1, actualRequests)

	assert.Equal(t, &calendar.Event{
		Created: "2018-10-09T17:00:00-07:00",
		Creator: &calendar.EventCreator{
			DisplayName: "Parker Moore",
			Email:       "parkr@jithub.com",
		},
		Description: "I am a description for the video call",
		HtmlLink:    "lalala",
		Kind:        "calendar#event",
		Location:    "https://jithub.zoom.us/j/12345",
		Organizer: &calendar.EventOrganizer{
			DisplayName: "Kevin Jithub",
			Email:       "kevin@jithub.com",
		},
		Start: &calendar.EventDateTime{
			DateTime: "2018-10-10T17:30:00-07:00",
			TimeZone: "America/New_York",
		},
		Summary: "I am a video call",
	}, event)

}

func TestNextEvent_NoUpcomingEvents(t *testing.T) {
	mux := http.NewServeMux()

	service, shutdown := newFakeGoogleCalendarService(t, mux)
	defer shutdown()

	actualRequests := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		actualRequests++

		if r.URL.Path == "/calendars/primary/events" {
			fmt.Fprintf(w, `{"items":[]}`)
		} else {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL)
		}
	})

	event, err := NextEvent(service)
	require.NoError(t, err)
	assert.Equal(t, 1, actualRequests)
	assert.Nil(t, event)
}

func newFakeGoogleCalendarService(t *testing.T, mux http.Handler) (*calendar.Service, func()) {
	service, err := calendar.New(&http.Client{})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	server := httptest.NewServer(mux)
	service.BasePath = server.URL

	return service, server.Close
}

func TestMeetingSummary(t *testing.T) {
	testCases := []struct {
		input    *calendar.Event
		expected string
	}{
		{nil, ""},
		{&calendar.Event{}, "You have a meeting coming up."},
		{&calendar.Event{Summary: "Make plans for Q4"}, `Your next meeting is "Make plans for Q4".`},
		{&calendar.Event{
			Creator: &calendar.EventCreator{DisplayName: "Mona Lisa"},
		}, `You have a meeting coming up, created by Mona Lisa.`},
		{&calendar.Event{
			Organizer: &calendar.EventOrganizer{DisplayName: "Johnny Appleseed"},
		}, `You have a meeting coming up, organized by Johnny Appleseed.`},
		{&calendar.Event{
			Summary:   "Make plans for Q4",
			Creator:   &calendar.EventCreator{DisplayName: "Mona Lisa"},
			Organizer: &calendar.EventOrganizer{DisplayName: "Johnny Appleseed"},
		}, `Your next meeting is "Make plans for Q4", organized by Johnny Appleseed.`},
	}
	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, MeetingSummary(testCase.input), "input: %+v", testCase.input)
	}
}

func TestIsMeetingSoon(t *testing.T) {
	testCases := []struct {
		input    *calendar.Event
		expected bool
	}{
		{nil, false},
		{&calendar.Event{}, false},
		{&calendar.Event{Start: &calendar.EventDateTime{}}, false},
		{&calendar.Event{Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(-5 * time.Minute).Format(googleCalendarDateTimeFormat),
		}}, false},
		{&calendar.Event{Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(-2 * time.Minute).Format(googleCalendarDateTimeFormat),
		}}, true},
		{&calendar.Event{Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(5 * time.Minute).Format(googleCalendarDateTimeFormat),
		}}, true},
		{&calendar.Event{Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(12 * time.Minute).Format(googleCalendarDateTimeFormat),
		}}, false},
	}
	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, IsMeetingSoon(testCase.input))
	}
}

func TestHumanizedStartTime(t *testing.T) {
	testCases := []struct {
		input    *calendar.Event
		expected string
	}{
		{nil, "event does not have a start datetime"},
		{&calendar.Event{}, "event does not have a start datetime"},
		{&calendar.Event{Start: &calendar.EventDateTime{}}, "event does not have a start datetime"},
		{&calendar.Event{Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(-12 * time.Minute).Format(googleCalendarDateTimeFormat),
		}}, "12 minutes ago"},
		{&calendar.Event{Start: &calendar.EventDateTime{
			DateTime: time.Now().Add(12 * time.Minute).Format(googleCalendarDateTimeFormat),
		}}, "11 minutes from now"},
	}
	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, HumanizedStartTime(testCase.input))
	}
}
