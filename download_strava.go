package main

import (
	"log"
	"time"
)

// determine the most recent activity timestamp by listing matching
// files in `dataDirectory`
func latestActivityTimestamp(env *Env) time.Time {
	var latestTime time.Time

	// MAX(started_on) was causing errors, e.g.
	//    `Scan error on column index 0, name "started_on": unsupported Scan, storing driver.Value type string into type *time.Time`
	// so we do this temporarily
	sql := `SELECT started_on FROM strava_activities ORDER BY started_on DESC LIMIT 1;`

	row := env.db.QueryRow(sql)
	err := row.Scan(&latestTime)
	if err != nil {
		log.Println(err)
	}

	return latestTime
}

// func maybeRequestAthlete(env *Env, httpClient *http.Client) {
// 	athleteFile := path.Join(dataDirectory, "athlete.json")
// 	if _, err := os.Stat(athleteFile); err == nil {
// 		log.Printf("%s already exists", athleteFile)
// 		return
// 	} else {
// 		log.Printf("%s does not exist", athleteFile)
// 	}

// 	endpoint := "https://www.strava.com/api/v3/athlete"
// 	body, err := fetchFromStrava(httpClient, endpoint)

// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	log.Printf("saving %s", athleteFile)
// 	err = ioutil.WriteFile(athleteFile, body, 0644)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// }

func SyncWithStrava(env *Env) {
	timestamp := latestActivityTimestamp(env)
	log.Printf("latest activity is %s (%d)", timestamp.String(), timestamp.Unix())
	// os.Exit(0)
	strava := &Strava{}
	err := strava.Connect(env)

	if err != nil {
		die(err)
	}

	err = strava.DownloadActivities(env, timestamp)
	if err != nil {
		die(err)
	}
}
