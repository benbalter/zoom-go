// Package config provides methods for fetching and storing configuration.
package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

var (
	// ErrNoGoogleClientConfig indicates that the client configuration is missing.
	ErrNoGoogleClientConfig = errors.New("missing google client config")
	// ErrNoGoogleToken indicatges that the token is missing.
	ErrNoGoogleToken = errors.New("missing google token")
)

// Provider is a token provider.
type Provider interface {
	// GoogleClientConfig returns the Google client config.
	GoogleClientConfig() (*oauth2.Config, error)

	// StoreGoogleClientConfig writes the Google client config.
	StoreGoogleClientConfig(*oauth2.Config) error

	// GoogleClientConfigExists returns true if the client config is readable, false otherwise.
	GoogleClientConfigExists() bool

	// GoogleToken returns the Google token.
	GoogleToken() (*oauth2.Token, error)

	// StoreGoogleToken writes the Google token.
	StoreGoogleToken(*oauth2.Token) error

	// GoogleTokenExists returns true if the token is readable, false otherwise.
	GoogleTokenExists() bool
}

// ReadGoogleClientConfigFromFile reads the content of a file and parses it as an *oauth2.Config
func ReadGoogleClientConfigFromFile(filepath string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	conf, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return conf, nil
}
