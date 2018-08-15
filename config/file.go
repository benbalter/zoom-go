package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const googleClientConfigFilename = "client_secrets.json"
const googleTokenFilename = "token.json"

// FileProvider is a Provider which uses files to store data.
type FileProvider struct {
	directory string

	cachedGoogleClientConfig *oauth2.Config
	cachedGoogleToken        *oauth2.Token
}

// NewFileProvider returns a new FileProvider with the default filepath.
func NewFileProvider() (*FileProvider, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &FileProvider{
		directory: filepath.Join(usr.HomeDir, ".config", "google"),
	}, nil
}

// GoogleClientConfigExists returns true if the config is readable and valid, false otherwise.
func (f *FileProvider) GoogleClientConfigExists() bool {
	conf, err := f.GoogleClientConfig()
	return conf != nil && err == nil
}

// GoogleClientConfig returns the Google client configuration from the configuration file.
func (f *FileProvider) GoogleClientConfig() (*oauth2.Config, error) {
	if f.cachedGoogleClientConfig != nil {
		return f.cachedGoogleClientConfig, nil
	}

	conf := &oauth2.Config{}

	fd, err := os.Open(filepath.Join(f.directory, googleClientConfigFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoGoogleClientConfig
		}
		return nil, errors.WithStack(err)
	}

	f.cachedGoogleClientConfig = conf

	return conf, json.NewDecoder(fd).Decode(conf)
}

// StoreGoogleClientConfig writes the Google client config to the configuration file.
func (f *FileProvider) StoreGoogleClientConfig(conf *oauth2.Config) error {
	f.cachedGoogleClientConfig = conf

	fd, err := os.Create(filepath.Join(f.directory, googleClientConfigFilename))
	if err != nil {
		return errors.WithStack(err)
	}

	return json.NewEncoder(fd).Encode(conf)
}

// GoogleTokenExists returns true if the token is readable and valid, false otherwise.
func (f *FileProvider) GoogleTokenExists() bool {
	token, err := f.GoogleToken()
	return token != nil && err == nil
}

// GoogleToken fetches the Google token from the configuration file.
func (f *FileProvider) GoogleToken() (*oauth2.Token, error) {
	if f.cachedGoogleToken != nil {
		return f.cachedGoogleToken, nil
	}

	token := &oauth2.Token{}

	fd, err := os.Open(filepath.Join(f.directory, googleTokenFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoGoogleToken
		}
		return nil, errors.WithStack(err)
	}

	f.cachedGoogleToken = token

	return token, json.NewDecoder(fd).Decode(token)
}

// StoreGoogleToken writes the Google token to the configuration file.
func (f *FileProvider) StoreGoogleToken(token *oauth2.Token) error {
	f.cachedGoogleToken = token

	fd, err := os.Create(filepath.Join(f.directory, googleTokenFilename))
	if err != nil {
		return errors.WithStack(err)
	}

	return json.NewEncoder(fd).Encode(token)
}
