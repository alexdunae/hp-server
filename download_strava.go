package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// generated via https://mholt.github.io/json-to-go/
type StravaActivities []struct {
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

var accessToken = ""
var refreshToken = ""

// determine the most recent activity timestamp by listing matching
// files in `dataDirectory`
func latestActivityTimestamp() time.Time {
	r, err := regexp.Compile("activity-([0-9]+).json")
	if err != nil {
		log.Fatalln(err)
	}

	latestTimestamp := int64(0)
	files, err := ioutil.ReadDir(dataDirectory)
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		timestampString := r.FindStringSubmatch(file.Name())
		if len(timestampString) < 1 {
			continue
		}

		timestamp, _ := strconv.ParseInt(timestampString[1], 10, 64)
		// log.Printf("strconv %d %s", timestamp, timestampString[1])
		if timestamp > latestTimestamp {
			latestTimestamp = timestamp
		}

	}
	latestTime := time.Unix(latestTimestamp, 0)
	return latestTime
}

func fetchFromStrava(endpoint string) ([]byte, error) {
	log.Printf("fetching %s", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	authorization := "Bearer " + accessToken
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == http.StatusUnauthorized {
		log.Fatalln("Strava response=unauthorized - check your STRAVA_ACCESS_TOKEN")
	} else if statusCode != http.StatusOK {
		log.Fatalf("Strava response=%d", statusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func downloadMissingActivities(latestTimestamp time.Time) {
	endpoint := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?after=%d&per_page=50", latestTimestamp.Unix())

	body, err := fetchFromStrava(endpoint)
	if err != nil {
		log.Fatalln(err)
	}

	data := StravaActivities{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("received %d activities", len(data))

	for _, activity := range data {
		timestamp := activity.StartDate.Unix()
		outfile := fmt.Sprintf("%s/activity-%d.json", dataDirectory, timestamp)
		// log.Printf("saving activity %d to %s", i, outfile)
		json, err := json.MarshalIndent(activity, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(outfile, json, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}

	timestamp := latestActivityTimestamp()
	log.Printf("new latest activity is %s (%d)", timestamp.String(), timestamp.Unix())
}

func maybeRequestAthlete() {
	athleteFile := path.Join(dataDirectory, "athlete.json")
	if _, err := os.Stat(athleteFile); err == nil {
		log.Printf("%s already exists", athleteFile)
		return
	} else {
		log.Printf("%s does not exist", athleteFile)
	}

	requestAthlete(athleteFile)
}

func requestAthlete(athleteFile string) {
	endpoint := "https://www.strava.com/api/v3/athlete"
	body, err := fetchFromStrava(endpoint)

	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("saving %s", athleteFile)
	err = ioutil.WriteFile(athleteFile, body, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func StartStravaSync() {
	godotenv.Load()

	accessToken = os.Getenv("STRAVA_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("Error loading STRAVA_ACCESS_TOKEN - you can define it in a .env file")
	}

	log.Printf("STRAVA_ACCESS_TOKEN=%s", accessToken)
	maybeRequestAthlete()

	timestamp := latestActivityTimestamp()
	log.Printf("latest activity is %s (%d)", timestamp.String(), timestamp.Unix())
	downloadMissingActivities(timestamp)
}
