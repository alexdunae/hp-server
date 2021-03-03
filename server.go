package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const ServerToken = "hp/4.20.69"

type ActivityJSON struct {
	Name         string    `json:"name"`
	ActivityType string    `json:"activity_type"`
	StartedOn    time.Time `json:"started_on"`
}

func getActivities(env *Env) []*ActivityJSON {
	rows, err := env.db.Query("SELECT name, activity_type, started_on FROM strava_activities ORDER BY started_on DESC LIMIT 50;")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	activities := make([]*ActivityJSON, 0)

	for rows.Next() {
		activity := new(ActivityJSON)
		err := rows.Scan(&activity.Name, &activity.ActivityType, &activity.StartedOn)
		if err != nil {
			log.Fatal(err)
		}
		activities = append(activities, activity)

	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil
	}

	return activities
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", ServerToken)
	w.WriteHeader(200)
}

func handleActivities(env *Env, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", ServerToken)
	w.Header().Set("X-Endpoint", r.URL.Path)

	activities := getActivities(env)
	log.Printf("/activities.json rendering %d", len(activities))

	json, err := json.Marshal(activities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func StartServer(env *Env) {
	fmt.Println("== start server ==")
	fmt.Println("Listening on http://localhost:3000")
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/activities.json", func(w http.ResponseWriter, r *http.Request) {
		handleActivities(env, w, r)
	})
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalln(err)
	}

}
