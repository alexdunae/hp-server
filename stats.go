package main

import (
	"fmt"
	"time"
)

// left aligned label, right aligned count
const countTemplate = "%-20s %6d\n"

func RenderMilestones(env *Env) {
	total_count := 0
	fmt.Println("\n== Milestones ==")
	rows, err := env.db.Query("SELECT started_on, name, activity_type FROM milestones ORDER BY started_on ASC")
	if err != nil {
		die(err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			started_on    time.Time
			name          string
			activity_type string
		)

		if err := rows.Scan(&started_on, &name, &activity_type); err != nil {
			die(err)
		}

		total_count += 1

		fmt.Printf("%-20s %-20s %s\n", started_on.Format("Jan 2, 2006"), activity_type, name)
	}

	err = rows.Err()
	if err != nil {
		die(err)
	}

	if total_count <= 0 {
		fmt.Println("No milestones recorded yet")
		return
	}

	fmt.Printf("err? %v", err)

}

func RenderActivities(env *Env) {
	fmt.Println("\n== Activities ==")

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

		fmt.Printf(countTemplate, activity_type, activity_count)
	}

	fmt.Println("---------------------------")
	fmt.Printf(countTemplate, "ALL ACTIVITIES", total_count)
}

func RenderStats(env *Env) {
	RenderActivities(env)
	RenderMilestones(env)
}
