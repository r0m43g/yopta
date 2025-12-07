// backend/internal/models/station.go
package models

import (
	"database/sql"
	"time"
)

// Station represents a train station or depot in the system.
// It serves as a container for tracks and is used in train schedule planning.
type Station struct {
	ID        int       `json:"id"`         // Unique identifier
	Name      string    `json:"name"`       // Station/depot name
	Code      string    `json:"code"`       // Station/depot code (unique)
	Notes     string    `json:"notes"`      // Additional notes
	CreatedAt time.Time `json:"created_at"` // When the record was created
	UpdatedAt time.Time `json:"updated_at"` // When the record was last updated
	UserID    int       `json:"user_id"`    // ID of the user who created the record
	Tracks    []Track   `json:"tracks"`     // Associated tracks
}

// Track represents a railway track within a station/depot.
// Tracks are used for scheduling train arrivals and departures.
type Track struct {
	ID          int       `json:"id"`           // Unique identifier
	StationID   int       `json:"station_id"`   // Foreign key to station
	TrackNumber string    `json:"track_number"` // Track identifier within station
	Positions   int       `json:"positions"`    // Number of positions on the track
	Length      int       `json:"length"`       // Length of track in meters (optional)
	Type        string    `json:"type"`         // Track type: 'through' or 'dead_end'
	Rule        string    `json:"rule"`         // Track rule: 'fifo' or 'filo'
	Exceptions  bool      `json:"exceptions"`   // Whether exceptions are allowed
	Notes       string    `json:"notes"`        // Additional notes
	CreatedAt   time.Time `json:"created_at"`   // When the record was created
	UpdatedAt   time.Time `json:"updated_at"`   // When the record was last updated
}

// GetAllStations retrieves all stations with their associated tracks.
func GetAllStations(db *sql.DB) ([]Station, error) {
	// Query to get all stations
	rows, err := db.Query(`
		SELECT id, name, code, notes, created_at, updated_at, user_id
		FROM stations
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []Station
	for rows.Next() {
		var station Station
		err := rows.Scan(
			&station.ID,
			&station.Name,
			&station.Code,
			&station.Notes,
			&station.CreatedAt,
			&station.UpdatedAt,
			&station.UserID,
		)
		if err != nil {
			return nil, err
		}
		stations = append(stations, station)
	}

	// For each station, get its tracks
	for i := range stations {
		tracks, err := GetTracksByStationID(db, stations[i].ID)
		if err != nil {
			return nil, err
		}
		stations[i].Tracks = tracks
	}

	return stations, nil
}

// GetStationByID retrieves a single station by its ID, including its tracks.
func GetStationByID(db *sql.DB, id int) (Station, error) {
	var station Station
	err := db.QueryRow(`
		SELECT id, name, code, notes, created_at, updated_at, user_id
		FROM stations
		WHERE id = ?
	`, id).Scan(
		&station.ID,
		&station.Name,
		&station.Code,
		&station.Notes,
		&station.CreatedAt,
		&station.UpdatedAt,
		&station.UserID,
	)
	if err != nil {
		return Station{}, err
	}

	// Get tracks for this station
	tracks, err := GetTracksByStationID(db, station.ID)
	if err != nil {
		return Station{}, err
	}
	station.Tracks = tracks

	return station, nil
}

// CreateStation adds a new station to the database.
func CreateStation(db *sql.DB, station Station) (int, error) {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert station
	result, err := tx.Exec(`
		INSERT INTO stations (name, code, notes, user_id)
		VALUES (?, ?, ?, ?)
	`, station.Name, station.Code, station.Notes, station.UserID)
	if err != nil {
		return 0, err
	}

	// Get the auto-generated ID
	stationID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// If there are tracks, insert them
	for _, track := range station.Tracks {
		_, err := tx.Exec(`
			INSERT INTO tracks (station_id, track_number, positions, length, notes)
			VALUES (?, ?, ?, ?, ?)
		`, stationID, track.TrackNumber, track.Positions, track.Length, track.Notes)
		if err != nil {
			return 0, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return int(stationID), nil
}

// UpdateStation updates an existing station and its tracks.
func UpdateStation(db *sql.DB, station Station) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update station
	_, err = tx.Exec(`
		UPDATE stations 
		SET name = ?, code = ?, notes = ?
		WHERE id = ?
	`, station.Name, station.Code, station.Notes, station.ID)
	if err != nil {
		return err
	}

	// For simplicity, we'll delete all tracks and re-insert them
	// In a real application, you might want to update existing tracks instead
	_, err = tx.Exec(`DELETE FROM tracks WHERE station_id = ?`, station.ID)
	if err != nil {
		return err
	}

	// Insert updated tracks
	for _, track := range station.Tracks {
		_, err := tx.Exec(`
			INSERT INTO tracks (station_id, track_number, positions, length, notes)
			VALUES (?, ?, ?, ?, ?)
		`, station.ID, track.TrackNumber, track.Positions, track.Length, track.Notes)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeleteStation removes a station and all its tracks from the database.
func DeleteStation(db *sql.DB, id int) error {
	// Note: With CASCADE delete on foreign key, this will also delete associated tracks
	_, err := db.Exec(`DELETE FROM stations WHERE id = ?`, id)
	return err
}

// GetTracksByStationID retrieves all tracks for a given station.
func GetTracksByStationID(db *sql.DB, stationID int) ([]Track, error) {
	rows, err := db.Query(`
        SELECT id, station_id, track_number, positions, length, type, rule, exceptions, notes, created_at, updated_at
        FROM tracks
        WHERE station_id = ?
        ORDER BY track_number ASC
    `, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []Track
	for rows.Next() {
		var track Track
		var length sql.NullInt64 // Handle NULL values for length

		err := rows.Scan(
			&track.ID,
			&track.StationID,
			&track.TrackNumber,
			&track.Positions,
			&length,
			&track.Type,
			&track.Rule,
			&track.Exceptions,
			&track.Notes,
			&track.CreatedAt,
			&track.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert NullInt64 to int
		if length.Valid {
			track.Length = int(length.Int64)
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// AddTrack adds a new track to a station.
func AddTrack(db *sql.DB, track Track) (int, error) {
	// Validate rule based on type - dead-end tracks can only be FILO
	if track.Type == "dead_end" {
		track.Rule = "filo" // Force FILO for dead-end tracks
	}

	result, err := db.Exec(`
        INSERT INTO tracks (station_id, track_number, positions, length, type, rule, exceptions, notes)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `, track.StationID, track.TrackNumber, track.Positions, track.Length, track.Type, track.Rule, track.Exceptions, track.Notes)
	if err != nil {
		return 0, err
	}

	trackID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(trackID), nil
}

// UpdateTrack updates an existing track.
func UpdateTrack(db *sql.DB, track Track) error {
	// Validate rule based on type - dead-end tracks can only be FILO
	if track.Type == "dead_end" {
		track.Rule = "filo" // Force FILO for dead-end tracks
	}

	_, err := db.Exec(`
        UPDATE tracks 
        SET track_number = ?, positions = ?, length = ?, type = ?, rule = ?, exceptions = ?, notes = ?
        WHERE id = ?
    `, track.TrackNumber, track.Positions, track.Length, track.Type, track.Rule, track.Exceptions, track.Notes, track.ID)

	return err
}

// DeleteTrack removes a track from the database.
func DeleteTrack(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM tracks WHERE id = ?`, id)
	return err
}
