// package config
package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

// Provider is a token provider.
type Provider interface {
	// GetGoogleClientConfig returns the Google client config.
	GetGoogleClientConfig() (*oauth2.Config, error)

	// GetGoogleToken returns the Google token.
	GetGoogleToken() (*oauth2.Token, error)

	// SaveGoogleToken writes the Google token.
	SaveGoogleToken(token *oauth2.Token) error
}

// FileProvider is a Provider which uses files to store data.
type FileProvider struct {
	directory string
}

// NewFileProvider returns a new FileProvider with the default filepath.
func NewFileProvider() (FileProvider, error) {
	usr, err := user.Current()
	if err != nil {
		return FileProvider{}, err
	}

	return FileProvider{
		directory: filepath.Join(usr.HomeDir, ".config", "google"),
	}, nil
}

// GetGoogleClientConfig returns the Google client configuration from the configuration file.
func (f FileProvider) GetGoogleClientConfig() (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(filepath.Join(f.directory, "client_secrets.json"))
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	return google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
}

// GetGoogleToken fetches the Google token from the configuration file.
func (f FileProvider) GetGoogleToken() (*oauth2.Token, error) {
	token := &oauth2.Token{}

	fd, err := os.Open(filepath.Join(f.directory, "token.json"))
	if err != nil {
		return token, err
	}

	return token, json.NewDecoder(fd).Decode(token)
}

// SaveGoogleToken writes the Google token to the configuration file.
func (f FileProvider) SaveGoogleToken(token *oauth2.Token) error {
	fd, err := os.Create(filepath.Join(f.directory, "token.json"))
	if err != nil {
		return err
	}

	return json.NewEncoder(fd).Encode(token)
}
