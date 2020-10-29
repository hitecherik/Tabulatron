package db

import (
	"database/sql"
	"fmt"
	"strings"

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
			email TEXT KEY,
			type TEXT NOT NULL,
			discord TEXT KEY,
			urlkey TEXT NOT NULL
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
		REPLACE INTO teams (id, participant, emoji)
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

	insertStmt, err := d.db.Prepare(`
		INSERT INTO participants (id, barcode, name, email, type, urlkey)
		VALUES (?, ?, ?, ?, ?, ?)
	`)

	if err != nil {
		return err
	}
	defer insertStmt.Close()

	updateStmt, err := d.db.Prepare(`
		UPDATE participants
		SET barcode=?, name=?, email=?, urlkey=?
		WHERE id=?
	`)

	if err != nil {
		return err
	}
	defer updateStmt.Close()

	query := `
		SELECT COUNT(*)
		FROM participants
		WHERE id=?
	`

	for _, participant := range participants {
		row := d.db.QueryRow(query, participant.Id)
		count := 0
		_ = row.Scan(&count)

		if count == 0 {
			_, err := insertStmt.Exec(participant.Id, participant.Barcode, participant.Name, participant.Email, category, participant.UrlKey)
			if err != nil {
				return err
			}
		} else {
			_, err := updateStmt.Exec(participant.Barcode, participant.Name, participant.Email, participant.UrlKey, participant.Id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Database) UpdateDiscordId(barcode string, discordId string) (bool, error) {
	query := `
		UPDATE participants
		SET discord = ?
		WHERE barcode = ?
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

func (d *Database) ParticipantFromBarcode(barcode string, discord string) (uint, string, bool, error) {
	query := `
		SELECT p.id, "[" || COALESCE(emoji, "J")  || "] " || name, type
		FROM participants p LEFT JOIN teams t ON (p.id=t.participant)
		WHERE barcode = ?
		AND discord IS NULL
		AND (SELECT COUNT(*) FROM participants WHERE discord = ?) = 0
	`

	row := d.db.QueryRow(query, barcode, discord)
	var (
		id       uint
		name     string
		category string
	)
	if err := row.Scan(&id, &name, &category); err != nil {
		return 0, "", false, err
	}

	query = `
		UPDATE participants
		SET discord = ?
		WHERE id = ?
	`
	if _, err := d.db.Exec(query, discord, id); err != nil {
		return 0, "", false, err
	}

	return id, name, category == "speaker", nil
}

func (d *Database) ParticipantFromDiscord(discord string) (uint, bool, error) {
	query := `
		SELECT id, type
		FROM participants
		WHERE discord = ?
	`

	row := d.db.QueryRow(query, discord)
	var (
		id       uint
		category string
	)
	if err := row.Scan(&id, &category); err != nil {
		return 0, false, err
	}

	return id, category == "speaker", nil
}

func (d *Database) ParticipantNameFromEmail(email string) (string, error) {
	query := `
		SELECT name
		FROM participants
		WHERE email = ?
		LIMIT 1
	`

	row := d.db.QueryRow(query, email)
	var name string
	if err := row.Scan(&name); err != nil {
		return "", err
	}

	return name, nil
}

func (d *Database) TeamEmails(teams []string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT email
		FROM participants p JOIN teams t ON (p.id=t.participant)
		WHERE t.id IN (%v)
	`, strings.Join(teams, ","))

	return d.stringsQuery(query)
}

func (d *Database) ParticipantEmails(participants []string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT email
		FROM participants
		WHERE id IN (%v)
	`, strings.Join(participants, ","))

	return d.stringsQuery(query)
}

func (d *Database) DiscordFromTeamId(teamId string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT discord
		FROM participants p
		JOIN teams t ON (t.participant=p.id)
		WHERE t.id = %v AND discord IS NOT NULL
	`, teamId)

	return d.stringsQuery(query)
}

func (d *Database) DiscordFromParticipantIds(participantIds []string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT discord
		FROM participants
		WHERE id IN (%v) AND discord IS NOT NULL
	`, strings.Join(participantIds, ","))

	return d.stringsQuery(query)
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

func (d *Database) stringsQuery(query string) ([]string, error) {
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	strings := make([]string, 0)

	for rows.Next() {
		var str string
		if err := rows.Scan(&str); err != nil {
			return nil, err
		}

		strings = append(strings, str)
	}

	return strings, nil
}
