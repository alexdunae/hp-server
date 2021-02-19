package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

func saveCredentials(env *Env, name string, raw_json string, expires_at time.Time) {
	// TODO look into transactions
	// tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	statement, err := env.db.Prepare(`
		INSERT INTO credentials (name, data, expires_at, updated_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(name) DO UPDATE SET
		data = ?,
			expires_at = ?,
			updated_at = datetime('now')
	`)
	defer statement.Close()

	if err != nil {
		log.Fatalln(err)
	}

	if _, err := statement.Exec(name, raw_json, expires_at, raw_json, expires_at); err != nil {
		log.Fatal(err)
	}

	log.Printf("saved %s credentials %v", name, string(raw_json))
}

func retrieveCredentials(env *Env, name string) ([]byte, error) {
	log.Printf("retrieveCredentials name=%s", name)
	var raw_json []byte
	var expires_at time.Time
	row := env.db.QueryRow(`SELECT data, expires_at FROM credentials WHERE name = ? LIMIT 1;`, name)

	err := row.Scan(&raw_json, &expires_at)
	if err != nil {
		return nil, err
	}

	if len(raw_json) < 1 {
		return nil, errors.New("no credentials")
	}

	if expires_at.Before(time.Now()) {
		return nil, errors.New(fmt.Sprintf("existing credentials expired on %s", expires_at))
	}

	return raw_json, nil
}
