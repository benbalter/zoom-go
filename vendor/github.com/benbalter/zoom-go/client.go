package zoom

import (
	"context"
	"net/http"

	"github.com/benbalter/zoom-go/config"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	calendar "google.golang.org/api/calendar/v3"
)

// NewGoogleClient creates a new client using the token from the given provider.
func NewGoogleClient(provider config.Provider) (*http.Client, error) {
	conf, err := provider.GoogleClientConfig()
	if err != nil {
		return nil, err
	}

	token, err := provider.GoogleToken()
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

// GoogleCalendarAuthorizationURL returns the authorization URL for the service configured in the provider.
func GoogleCalendarAuthorizationURL(provider config.Provider) (string, error) {
	conf, err := provider.GoogleClientConfig()
	if err != nil {
		return "", err
	}

	return conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline), nil
}

// HandleGoogleCalendarAuthorization takes an auth code and generates the necessary token and stores it on the provider.
func HandleGoogleCalendarAuthorization(provider config.Provider, authCode string) error {
	conf, err := provider.GoogleClientConfig()
	if err != nil {
		return err
	}

	tok, err := conf.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		return errors.WithStack(err)
	}

	return provider.StoreGoogleToken(tok)
}
