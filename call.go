package zoom

import (
	"net/url"
	"path"
	"strings"

	"github.com/mvdan/xurls"
)

var urlRegexp = xurls.Strict()

type call struct {
	id          string
	password    string
	originalURL string
}

func (c call) GetAppURL() string {
	if c.id == "" {
		return c.originalURL
	}

	url := "zoommtg://zoom.us/join?confno=" + c.id
	if c.password != "" {
		url = url + "&pwd=" + c.password
	}
	return url
}

func extractZoomCallURL(input string) (*url.URL, bool) {
	urls := urlRegexp.FindAllString(input, -1)
	if len(urls) == 0 {
		return nil, false
	}
	for _, inputURL := range urls {
		u, err := url.Parse(inputURL)
		if err != nil {
			continue
		}
		if strings.HasSuffix(u.Hostname(), ".zoom.us") {
			return u, true
		}
	}

	return nil, false
}

func extractZoomCallData(input string) (call, bool) {
	zoomURL, ok := extractZoomCallURL(input)
	if !ok {
		return call{}, false
	}

	// By default, match the whole URL.
	data := &call{originalURL: zoomURL.String()}

	// If we have a meeting ID in the URL, then we have a URL.
	if strings.HasPrefix(zoomURL.Path, "/j/") {
		_, data.id = path.Split(zoomURL.Path)
	}

	if password := zoomURL.Query().Get("pwd"); password != "" {
		data.password = password
	}

	return *data, true
}
