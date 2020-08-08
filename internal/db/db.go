package db

import (
	"database/sql"
	"fmt"

	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db   *sql.DB
	file string
}

func New(file string) (*Database, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	query := `
		CREATE TABLE IF NOT EXISTS participants (
			id INTEGER NOT NULL PRIMARY KEY,
			barcode TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			email TEXT,
			type TEXT NOT NULL,
			discord TEXT KEY
		);
		CREATE TABLE IF NOT EXISTS teams (
			id INTEGER NOT NULL,
			participant INTEGER NOT NULL,
			emoji TEXT NOT NULL,
			PRIMARY KEY (id, participant),
			FOREIGN KEY (participant) REFERENCES participants (id)
		);
	`

	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &Database{db, file}, nil
}

func (d *Database) Reset() error {
	query := `
		DELETE FROM teams;
		DELETE FROM participants;
	`

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) AddTeams(teams []tabbycat.Team) error {
	stmt, err := d.db.Prepare(`
		INSERT INTO teams (id, participant, emoji)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, team := range teams {
		if err := d.AddParticipants(true, team.Speakers); err != nil {
			return err
		}

		for _, speaker := range team.Speakers {
			if _, err := stmt.Exec(team.Id, speaker.Id, team.Emoji); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Database) AddParticipants(speakers bool, participants []tabbycat.Participant) error {
	category := "adjudicator"
	if speakers {
		category = "speaker"
	}

	stmt, err := d.db.Prepare(`
		INSERT INTO participants (id, barcode, name, email, type)
		VALUES (?, ?, ?, ?, ?)
	`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, participant := range participants {
		_, err := stmt.Exec(participant.Id, participant.Barcode, participant.Name, participant.Email, category)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) UpdateDiscordId(barcode string, discordId string) (bool, error) {
	query := `
		UPDATE participants
		SET discord = ?
		WHERE barcode = ?
		LIMIT 1;
	`

	result, err := d.db.Exec(query, discordId, barcode)
	if err != nil {
		return true, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return true, err
	}

	if rows != 1 {
		return false, fmt.Errorf("user with barcode %v doesn't exist", barcode)
	}

	return true, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) Set(value string) error {
	db, err := New(value)
	if err != nil {
		return err
	}

	*d = *db
	return nil
}

func (d *Database) SetIfNotExists(value string) error {
	if d.db == nil {
		return d.Set(value)
	}

	return nil
}

func (d *Database) String() string {
	return d.file
}
