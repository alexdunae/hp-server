package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type Strava struct {
	// authenticated Strava API client
	httpClient *http.Client
}

// generated via https://mholt.github.io/json-to-go/
type StravaActivity struct {
	ResourceState int `json:"resource_state"`
	Athlete       struct {
		ID            int `json:"id"`
		ResourceState int `json:"resource_state"`
	} `json:"athlete"`
	Name               string      `json:"name"`
	Distance           float64     `json:"distance"`
	MovingTime         int         `json:"moving_time"`
	ElapsedTime        int         `json:"elapsed_time"`
	TotalElevationGain float64     `json:"total_elevation_gain"`
	Type               string      `json:"type"`
	WorkoutType        interface{} `json:"workout_type"`
	ID                 int         `json:"id"`
	ExternalID         string      `json:"external_id"`
	UploadID           int         `json:"upload_id"`
	StartDate          time.Time   `json:"start_date"`
	StartDateLocal     time.Time   `json:"start_date_local"`
	Timezone           string      `json:"timezone"`
	UtcOffset          float64     `json:"utc_offset"`
	StartLatlng        []float64   `json:"start_latlng"`
	EndLatlng          []float64   `json:"end_latlng"`
	LocationCity       string      `json:"location_city"`
	LocationState      string      `json:"location_state"`
	LocationCountry    string      `json:"location_country"`
	StartLatitude      float64     `json:"start_latitude"`
	StartLongitude     float64     `json:"start_longitude"`
	AchievementCount   int         `json:"achievement_count"`
	KudosCount         int         `json:"kudos_count"`
	CommentCount       int         `json:"comment_count"`
	AthleteCount       int         `json:"athlete_count"`
	PhotoCount         int         `json:"photo_count"`
	Map                struct {
		ID              string `json:"id"`
		SummaryPolyline string `json:"summary_polyline"`
		ResourceState   int    `json:"resource_state"`
	} `json:"map"`
	Trainer                    bool        `json:"trainer"`
	Commute                    bool        `json:"commute"`
	Manual                     bool        `json:"manual"`
	Private                    bool        `json:"private"`
	Visibility                 string      `json:"visibility"`
	Flagged                    bool        `json:"flagged"`
	GearID                     string      `json:"gear_id"`
	FromAcceptedTag            bool        `json:"from_accepted_tag"`
	UploadIDStr                string      `json:"upload_id_str,omitempty"`
	AverageSpeed               float64     `json:"average_speed"`
	MaxSpeed                   float64     `json:"max_speed"`
	AverageWatts               float64     `json:"average_watts,omitempty"`
	Kilojoules                 float64     `json:"kilojoules,omitempty"`
	DeviceWatts                bool        `json:"device_watts,omitempty"`
	HasHeartrate               bool        `json:"has_heartrate"`
	HeartrateOptOut            bool        `json:"heartrate_opt_out"`
	DisplayHideHeartrateOption bool        `json:"display_hide_heartrate_option"`
	ElevHigh                   float64     `json:"elev_high,omitempty"`
	ElevLow                    float64     `json:"elev_low,omitempty"`
	PrCount                    int         `json:"pr_count"`
	TotalPhotoCount            int         `json:"total_photo_count"`
	HasKudoed                  bool        `json:"has_kudoed"`
	SufferScore                interface{} `json:"suffer_score"`
}

const dataDirectory = "data/strava"
const oauthState = "hp-oauth-secret-state"

func (s Strava) retrieveStoredCredentials(env *Env) *oauth2.Token {
	raw_json, err := retrieveCredentials(env, "strava")
	if err != nil {
		log.Println(err)
		return nil
	}
	tok := &oauth2.Token{}

	json.Unmarshal(raw_json, tok)
	remaining := tok.Expiry.Sub(time.Now())
	log.Printf("retrieved strava token %s expiring in %s", tok.AccessToken, remaining.String())
	return tok
}

func listenForCode(c chan string) {
	http.HandleFunc("/strava_oauth", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world strava_oauth %s", html.EscapeString(r.URL.Path))

		err := r.ParseForm()
		if err != nil {
			log.Printf("could not parse query: %v", err)
		}
		code := r.FormValue("code")
		state := r.FormValue("state")

		if code != "" {
			if state != oauthState {
				log.Fatalf("invalid state, got '%s' expected '%s'", state, oauthState)
			}

			log.Printf("got real code!!!!: %s", code)
			c <- code
			// TODO: how to close the server
		}
	})
	log.Printf("Listening on http://127.0.0.1:8080/strava_oauth")
	http.ListenAndServe(":8080", nil)
}

func (s *Strava) Connect(env *Env) error {
	godotenv.Load()
	ctx := context.Background()

	clientID := os.Getenv("STRAVA_CLIENT_ID")
	if clientID == "" {
		return errors.New("Error loading STRAVA_CLIENT_ID - you can define it in a .env file")
	}

	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")
	if clientSecret == "" {
		return errors.New("Error loading STRAVA_CLIENT_SECRET - you can define it in a .env file")
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,

		// nb: for some reason Strava wants it's scope as a single string, not an array
		Scopes: []string{"read_all,activity:read_all,profile:read_all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.strava.com/oauth/authorize",
			TokenURL: "https://www.strava.com/oauth/token",
		},
		RedirectURL: "http://127.0.0.1:8080/strava_oauth",
	}

	tok := s.retrieveStoredCredentials(env)

	if tok != nil {
		s.httpClient = conf.Client(ctx, tok)
		return nil
	}


	c := make(chan string)
	go listenForCode(c)

	url := conf.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	log.Printf("Visit this URL to connect to Strava:\n  %v", url)

	code := <-c
	log.Printf("Got code... %s", code)

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return err
	}
	log.Printf("got token! %v", tok)

	raw_json, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	saveCredentials(env, "strava", string(raw_json), tok.Expiry)
	s.httpClient = conf.Client(ctx, tok)
	return nil
}

func (s Strava) fetchFromAPI(endpoint string) ([]byte, error) {
	log.Printf("fetching %s", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if s.httpClient == nil {
		die(errors.New("client is not setup"))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == http.StatusUnauthorized {
		log.Fatalln("Strava response=unauthorized - check your STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET")
	} else if statusCode == http.StatusTooManyRequests {
		log.Fatalln("Strava response=429 Too Many Requests - we hit the Strava rate limit: http://developers.strava.com/docs/rate-limits/")
	} else if statusCode != http.StatusOK {
		log.Fatalf("Strava response=%d", statusCode)
	}

	printRateLimit(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func printRateLimit(resp *http.Response) {
	// Strava returns two headers, each with two comma-separated integers
	// representing the 15min and daily totals
	// see http://developers.strava.com/docs/rate-limits/
	rateLimitsTotal := resp.Header.Get("X-Ratelimit-Limit")
	rateLimitsUsed := resp.Header.Get("X-Ratelimit-Usage")

	totals := strings.SplitN(rateLimitsTotal, ",", 2)
	useds := strings.SplitN(rateLimitsUsed, ",", 2)

	if len(totals) < 1 || len(useds) < 1 {
		log.Printf("Unable to parse the rate limit header")
		return
	}

	log.Printf("Strava API rate limits    15mins: %s/%s used     daily: %s/%s used", useds[0], totals[0], useds[1], totals[1])
}

func (s Strava) SaveActivity(env *Env, activity StravaActivity) error {
	log.Printf("- saving %s - %s", activity.StartDateLocal, activity.Name)

	raw_json, err := json.Marshal(activity)

	statement, err := env.db.Prepare(`
	INSERT INTO strava_activities (remote_id, external_id, name, data, started_on, created_at, updated_at) VALUES (
		?, ?, ?, ?, ?, datetime('now'), datetime('now')
	) ON CONFLICT(remote_id) DO UPDATE SET
		data = ?,
		started_on = ?,
		updated_at = datetime('now');
	`)

	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(activity.ID, activity.ExternalID, activity.Name, raw_json, activity.StartDate, raw_json, activity.StartDate)

	if err != nil {
		return err
	}

	outfile := fmt.Sprintf("%s/activity-%d.json", dataDirectory, activity.StartDate.Unix())
	json, err := json.MarshalIndent(activity, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outfile, json, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s Strava) DownloadActivities(env *Env, latestTimestamp time.Time) error {
	var newLatestTimestamp time.Time
	if latestTimestamp.Unix() < 0 {
		latestTimestamp = time.Unix(0, 0)
	}
	log.Printf("downloading activities after=%v", latestTimestamp)
	endpoint := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?after=%d&per_page=100", latestTimestamp.Unix())

	body, err := s.fetchFromAPI(endpoint)
	if err != nil {
		return err
	}

	data := []StravaActivity{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	log.Printf("received %d activities", len(data))

	for _, activity := range data {
		if activity.StartDate.After(newLatestTimestamp) {
			newLatestTimestamp = activity.StartDate
		}

		err := s.SaveActivity(env, activity)

		if err != nil {
			return err
		}
	}

	log.Printf("latest activity is %s (%d)", newLatestTimestamp.String(), newLatestTimestamp.Unix())
	return nil
}

func (s Strava) maybeRequestAthlete(env *Env) {
	athleteFile := path.Join(dataDirectory, "athlete.json")
	if _, err := os.Stat(athleteFile); err == nil {
		log.Printf("%s already exists", athleteFile)
		return
	} else {
		log.Printf("%s does not exist", athleteFile)
	}

	endpoint := "https://www.strava.com/api/v3/athlete"
	body, err := s.fetchFromAPI(endpoint)

	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("saving %s", athleteFile)
	err = ioutil.WriteFile(athleteFile, body, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
