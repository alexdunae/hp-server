package main

import (
	"log"
)

// left aligned label, right aligned count
const countTemplate = "%-20s %6d"

// determine the most recent activity timestamp by listing matching
// files in `dataDirectory`
func RenderStats(env *Env) {
	log.Println("== Activities ==")

	var total_count int64 = 0

	rows, err := env.db.Query(`SELECT activity_type, count(*) AS activity_count FROM strava_activities GROUP BY activity_type;`)
	if err != nil {
		die(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			activity_type  string
			activity_count int64
		)

		if err := rows.Scan(&activity_type, &activity_count); err != nil {
			die(err)

		}

		total_count += activity_count

		log.Printf(countTemplate, activity_type, activity_count)
	}

	log.Println("---------------------------")
	log.Printf(countTemplate, "ALL ACTIVITIES", total_count)
}
