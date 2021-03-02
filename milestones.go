package main

import (
	"fmt"
)

func RecordMilestone(env *Env) {
	rows, err := env.db.Query(`SELECT DISTINCT activity_type FROM strava_activities ORDER BY activity_type ASC;`)
	if err != nil {
		die(err)
	}
	defer rows.Close()

	fmt.Println("== record milestone == ")
	fmt.Printf("\nWhich activity?\n")
	for rows.Next() {
		var activity_type string
		if err := rows.Scan(&activity_type); err != nil {
			die(err)
		}

		fmt.Printf("- %s\n", activity_type)
	}

	var (
		activity_type string
		name          string
		description   string
	)
	fmt.Scanln(&activity_type)

	fmt.Printf("\nTitle?\n")
	fmt.Scanln(&name)

	fmt.Printf("\nDescribe what you did!\n")
	fmt.Scanln(&description)

	fmt.Printf("Got: %s and %s", activity_type, description)

	statement, err := env.db.Prepare("INSERT INTO milestones (name, activity_type, description, started_on, created_at) VALUES (?, ?, ?, datetime('now'), datetime('now'));")
	if err != nil {
		die(err)
	}
	defer statement.Close()

	_, err = statement.Exec(name, activity_type, description)
	if err != nil {
		die(err)
	}
}
