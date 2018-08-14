package zoom

import (
	"context"
	"net/http"

	"github.com/benbalter/zoom-go/config"
	calendar "google.golang.org/api/calendar/v3"
)

// NewGoogleClient creates a new client using the token from the given provider.
func NewGoogleClient(provider config.Provider) (*http.Client, error) {
	conf, err := provider.GetGoogleClientConfig()
	if err != nil {
		return nil, err
	}

	token, err := provider.GetGoogleToken()
	if err != nil {
		return nil, err
	}
	return conf.Client(context.Background(), token), nil
}

// NewGoogleCalendarService creates a new Google Calendar service with the credentials in the provider.
func NewGoogleCalendarService(provider config.Provider) (*calendar.Service, error) {
	client, err := NewGoogleClient(provider)
	if err != nil {
		return nil, err
	}
	return calendar.New(client)
}
