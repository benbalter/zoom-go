package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"time"

	humanize "github.com/dustin/go-humanize"
	open "github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

func inConfigDir(file string) string {
	usr, err := user.Current()

	if err != nil {
		log.Fatalf("Can't find home directory: %v", err)
	}

	return path.Join(usr.HomeDir, ".config", "google", file)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {

	tokFile := inConfigDir("token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println("Your browser is about to open. When it does, please authorize the application when prompted and paste the token it gives you below")
	fmt.Println("Authorization token: ")
	time.Sleep(5 * time.Second)
	open.Run(authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func zoomURL(meeting *calendar.Event) (*url.URL, error) {
	fullURL, meetingID, matched := meetingURLParts(meeting)

	if !matched {
		return url.Parse("")
	} else if _, err := strconv.Atoi(meetingID); err == nil {
		return url.Parse("zoommtg://zoom.us/join?confno=" + meetingID)
	} else {
		return url.Parse(fullURL)
	}
}

func inPast(meeting *calendar.Event) bool {
	return meeting.Start.DateTime < time.Now().Format(time.RFC3339)
}

func startTime(meeting *calendar.Event) (time.Time, error) {
	return time.Parse(time.RFC3339, meeting.Start.DateTime)
}

func humanizedStartTime(meeting *calendar.Event) string {
	time, _ := startTime(meeting)
	return humanize.Time(time)
}

func moreThanFiveMinutesFromNow(meeting *calendar.Event) bool {
	start, _ := startTime(meeting)
	return time.Until(start).Minutes() > 5
}

func service() *calendar.Service {
	file := inConfigDir("client_secrets.json")
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := calendar.New(getClient(config))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	return srv
}

func events() *calendar.Events {
	t := time.Now().Format(time.RFC3339)
	events, err := service().Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(1).OrderBy("startTime").Do()

	if err != nil {
		log.Fatalf("Unable to retrieve next meeting: %v", err)
	}

	return events
}

func meetingURLParts(meeting *calendar.Event) (string, string, bool) {
	r, _ := regexp.Compile(`https://.*?\.zoom\.us/(?:j/(\d+)|my/(\S+))`)
	matches := r.FindAllStringSubmatch(meeting.Location+" "+meeting.Description, -1)

	if len(matches) > 0 {
		return matches[0][0], matches[0][1], true
	}

	return "", "", false
}

func main() {
	events := events()

	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
		return
	}

	meeting := events.Items[0]
	url, err := zoomURL(meeting)

	if err != nil || url.String() == "" {
		fmt.Println("Your next meeting isn't a Zoom meeting")
		return
	}

	fmt.Printf("Your next Zoom meeting is %s\n", meeting.Summary)

	isWas := ""
	if inPast(meeting) {
		isWas = "was"
	} else {
		isWas = "is"
	}

	fmt.Printf("It %s scheduled to start %s\n", isWas, humanizedStartTime(meeting))

	if moreThanFiveMinutesFromNow(meeting) {
		fmt.Printf("Here's the URL %s\n", url)
	} else {
		fmt.Printf("Opening %s...\n", url)
		open.Run(url.String())
	}

	fmt.Printf("Oh, and here's the URL in case you need it: %s\n", meeting.HtmlLink)
}
