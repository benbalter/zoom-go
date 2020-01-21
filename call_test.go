package zoom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractZoomCallData(t *testing.T) {
	examples := []struct {
		input    string
		expected call
	}{
		{"https://github.com", call{}},
		{
			"\n\n\nConference: https://jithub.zoom.us/j/12345\n\n",
			call{id: "12345", originalURL: "https://jithub.zoom.us/j/12345"},
		},
		{
			"\nhttps://github.com\nConf:https://jithub.zoom.us/j/42124?pwd=ZXN2S0k1AzU1ZitEKUhTR0NMZ2NwZz09\n",
			call{id: "42124", password: "ZXN2S0k1AzU1ZitEKUhTR0NMZ2NwZz09", originalURL: "https://jithub.zoom.us/j/42124?pwd=ZXN2S0k1AzU1ZitEKUhTR0NMZ2NwZz09"},
		},
		{
			"\n\nConf:https://jithub.zoom.us/my/foobar\n",
			call{originalURL: "https://jithub.zoom.us/my/foobar"},
		},
		{
			"\n\nConf:https://jithub.zoom.us/my/foobar?pwd=ZXN2S0k1AzU1ZitEKUhTR0NMZ2NwZz09\n",
			call{password: "ZXN2S0k1AzU1ZitEKUhTR0NMZ2NwZz09", originalURL: "https://jithub.zoom.us/my/foobar?pwd=ZXN2S0k1AzU1ZitEKUhTR0NMZ2NwZz09"},
		},
	}

	for _, example := range examples {
		actual, _ := extractZoomCallData(example.input)
		assert.Equal(t, example.expected, actual)
	}
}

func TestCallGetAppURL(t *testing.T) {
	examples := []struct {
		input    call
		expected string
	}{
		{call{}, ""},
		{call{originalURL: "foo"}, "foo"},
		{call{originalURL: "foo", id: "1234"}, "zoommtg://zoom.us/join?confno=1234"},
		{call{originalURL: "foo", id: "1234", password: "baz"}, "zoommtg://zoom.us/join?confno=1234&pwd=baz"},
	}

	for _, example := range examples {
		actual := example.input.GetAppURL()
		assert.Equal(t, example.expected, actual)
	}
}
