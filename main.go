package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "data/database.db"
const migrationDirectory = "data/migrations"

var migrationRegexp = regexp.MustCompile("([0-9]+)-([a-zA-Z0-9_-]+).sql")

func die(err error) {
	log.Fatal(err.Error())
}

// TODO: how does Go like us passing database handles around?
// where do we open and close our connection?
func createDatabase() {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Printf("creating database %s", dbFile)
		file, err := os.Create(dbFile)
		if err != nil {
			die(err)
		}
		file.Close()
	} else {
		log.Printf("database %s already exists", dbFile)
	}

	// TODO: investigate connection modes https://github.com/mattn/go-sqlite3#connection-string
	sqliteDatabase, _ := sql.Open("sqlite3", "./"+dbFile)
	defer sqliteDatabase.Close()
	if hasExistingTables(sqliteDatabase) == false {
		err := bootstrapDatabase(sqliteDatabase)
		if err != nil {
			die(err)
		}
	}

	err := migrateDatabase(sqliteDatabase)
	if err != nil {
		die(err)
	}
}

func bootstrapDatabase(db *sql.DB) error {
	sql := `CREATE TABLE schema_migrations
				(migration_id INT PRIMARY KEY NOT NULL UNIQUE)
				WITHOUT ROWID;
		 	`

	statement, err := db.Prepare(sql)
	if err != nil {
		return err
	}

	statement.Exec()
	log.Printf("created migrations table")
	return nil
}

func hasExistingTables(db *sql.DB) bool {
	var table_name string

	query := `SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%';`
	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil {
		die(err)
	}

	for rows.Next() {
		err := rows.Scan(&table_name)
		if err != nil {
			die(err)
		}

		return true
	}

	return false
}

func getExistingMigrations(db *sql.DB) ([]string, error) {
	var migration_id string
	var migration_ids []string

	query := `SELECT * FROM schema_migrations ORDER BY migration_id ASC;`
	rows, err := db.Query(query)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&migration_id)
		if err != nil {
			return nil, err
		}

		migration_ids = append(migration_ids, migration_id)
	}

	return migration_ids, nil
}

// TODO: should we eliminate records when we find them?
// what algorithm should this be?
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func migrateDatabase(db *sql.DB) error {
	files, err := ioutil.ReadDir(migrationDirectory)
	if err != nil {
		return err
	}

	existingMigrations, err := getExistingMigrations(db)
	if err != nil {
		return err
	}

	for _, file := range files {
		matchString := migrationRegexp.FindStringSubmatch(file.Name())
		if len(matchString) > 0 {
			sqlPath := fmt.Sprintf("%s/%s", migrationDirectory, file.Name())
			if err != nil {
				return err
			}

			if contains(existingMigrations, matchString[1]) {
				log.Printf("- skipping  %s      %s", matchString[1], matchString[2])
			} else {
				log.Printf("- migrating %s      %s", matchString[1], matchString[2])
				rawSQL, err := ioutil.ReadFile(sqlPath)
				statement, err := db.Prepare(string(rawSQL))
				if err != nil {
					return err
				}
				_, err = statement.Exec()
				if err != nil {
					die(err)
				}
				statement.Close()
				migration_id, err := strconv.Atoi(matchString[1])
				if err != nil {
					die(err)
				}
				recordMigration(db, migration_id)
			}
		}
	}

	return nil
}

func recordMigration(db *sql.DB, migration_id int) {
	statement, err := db.Prepare(fmt.Sprintf("INSERT INTO schema_migrations VALUES (%d);", migration_id))
	if err != nil {
		die(err)
	}

	_, err = statement.Exec()
	if err != nil {
		die(err)
	}

	statement.Close()
}

func main() {
	log.Println("= setup database =")
	createDatabase()
	// StartStravaSync()
}
