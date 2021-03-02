package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hitecherik/Tabulatron/pkg/tabbycat"
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
		CREATE TABLE IF NOT EXISTS reglog (
			id INTEGER NOT NULL PRIMARY KEY,
			time TEXT DEFAULT (DATETIME()),
			type TEXT NOT NULL,
			participant INTEGER NOT NULL,
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
		DELETE FROM reglog;
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

	query = `
		INSERT INTO reglog (type, participant)
		VALUES ("arrival", ?)
	`
	_, _ = d.db.Exec(query, id)

	return id, name, category == "speaker", nil
}

func (d *Database) ClearParticipantFromBarcode(barcode string) (string, error) {
	query := `
		INSERT INTO reglog (type, participant)
		SELECT "departure", id
		FROM participants
		WHERE barcode=?
		LIMIT 1
	`

	_, _ = d.db.Exec(query, barcode)

	query = fmt.Sprintf(`
		SELECT discord
		FROM participants
		WHERE barcode='%v'
		LIMIT 1
	`, barcode)

	discords, err := d.stringsQuery(query)
	if err != nil {
		return "", err
	}
	if len(discords) == 0 {
		return "", fmt.Errorf("no users with barcode %v found", barcode)
	}

	query = `
		UPDATE participants
		SET discord = NULL
		WHERE barcode = ?
	`

	if _, err := d.db.Exec(query, barcode); err != nil {
		return "", err
	}

	return discords[0], nil
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

func (d *Database) ClearParticipantFromDiscord(discord string) error {
	query := `
		INSERT INTO reglog (type, participant)
		SELECT "departure", id
		FROM participants
		WHERE discord = ?
	`

	_, _ = d.db.Exec(query, discord)

	query = `
		UPDATE participants
		SET discord = NULL
		WHERE discord = ?
	`

	_, err := d.db.Exec(query, discord)
	return err
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

func (d *Database) ParticipantsFromTeamId(teamId string) ([]string, []string, error) {
	query := fmt.Sprintf(`
		SELECT discord, urlkey
		FROM participants p
		JOIN teams t ON (t.participant=p.id)
		WHERE t.id = %v AND discord IS NOT NULL
	`, teamId)

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	snowflakes := make([]string, 0)
	urlKeys := make([]string, 0)

	for rows.Next() {
		var (
			snowflake string
			urlKey    string
		)
		if err := rows.Scan(&snowflake, &urlKey); err != nil {
			return nil, nil, err
		}

		snowflakes = append(snowflakes, snowflake)
		urlKeys = append(urlKeys, urlKey)
	}

	return snowflakes, urlKeys, nil
}

func (d *Database) DiscordFromParticipantIds(participantIds []string) ([]string, []string, error) {
	ordering := make([]string, 0, len(participantIds))
	for i, id := range participantIds {
		ordering = append(ordering, fmt.Sprintf("WHEN %v THEN %v", id, i))
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(discord, ""), urlkey
		FROM participants
		WHERE id IN (%v)
		ORDER BY CASE id %v END
	`, strings.Join(participantIds, ","), strings.Join(ordering, " "))

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	snowflakes := make([]string, 0)
	urlKeys := make([]string, 0)

	for rows.Next() {
		var (
			snowflake string
			urlKey    string
		)
		if err := rows.Scan(&snowflake, &urlKey); err != nil {
			return nil, nil, err
		}

		snowflakes = append(snowflakes, snowflake)
		urlKeys = append(urlKeys, urlKey)
	}

	return snowflakes, urlKeys, nil
}

func (d *Database) AllDiscords() ([]string, error) {
	query := `
		SELECT discord
		FROM participants
		WHERE discord IS NOT NULL
	`

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
