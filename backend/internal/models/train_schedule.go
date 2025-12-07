// backend/internal/models/train_schedule.go
package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// TrainSchedule represents a train schedule record in the system.
// This model contains arrival and departure information for trains,
// along with metadata about the specific locomotive and personnel involved.
//
// The fields are designed to support the Lithuanian Railways scheduling system
// with fields for both arrival and departure data, enabling tracking of locomotive
// movements between depots.
type TrainSchedule struct {
	ID                   string        `json:"id"`                   // Unique identifier for the record
	TrainNumberDeparture string        `json:"trainNumberDeparture"` // The train number for departure
	TrainNumberArrival   string        `json:"trainNumberArrival"`   // The train number for arrival
	VehicleName          string        `json:"vehicleName"`          // Name/model of the locomotive
	StartingLocation     string        `json:"startingLocation"`     // Departure location/depot
	EndLocation          string        `json:"endLocation"`          // Arrival location/depot
	DepartureDateTime    *time.Time    `json:"departureDateTime"`    // Scheduled departure date and time
	ArrivalDateTime      *time.Time    `json:"arrivalDateTime"`      // Scheduled arrival date and time
	StartingTrack        string        `json:"startingTrack"`        // Track number for departure
	TargetTrack          string        `json:"targetTrack"`          // Track number for arrival
	Employee1Departure   string        `json:"employee1Departure"`   // Primary employee for departure (usually driver)
	Employee1Arrival     string        `json:"employee1Arrival"`     // Primary employee for arrival
	DutyDeparture        string        `json:"dutyDeparture"`        // Duty/task description for departure
	DutyArrival          string        `json:"dutyArrival"`          // Duty/task description for arrival
	Notes                string        `json:"notes"`                // Additional notes about the schedule
	RawData              string        `json:"rawData"`              // Original raw data for reference
	CreatedAt            time.Time     `json:"createdAt"`            // When the record was created
	UpdatedAt            time.Time     `json:"updatedAt"`            // When the record was last updated
	UserID               sql.NullInt64 `json:"userId"`               // ID of user who created/owns this record
}

// TrainScheduleList is a collection of train schedule records.
// Used for API responses when returning multiple records.
type TrainScheduleList struct {
	Records []TrainSchedule `json:"records"` // Array of train schedule records
	Total   int             `json:"total"`   // Total count of records
}

// GetTrainSchedules retrieves all train schedule records from the database.
// Supports filtering by user ID if provided.
//
// Parameters:
//   - db: Database connection
//   - userID: Optional user ID to filter records (0 for all records)
//
// Returns:
//   - Train schedule records matching the filter
//   - Error if the database operation fails
func GetTrainSchedules(db *sql.DB, userID int) ([]TrainSchedule, error) {
	// Base SQL query
	query := `
		SELECT 
			id, train_number_departure, train_number_arrival, vehicle_name,
			starting_location, end_location, departure_date_time, arrival_date_time,
			starting_track, target_track, employee1_departure, employee1_arrival,
			duty_departure, duty_arrival, notes, raw_data, created_at, updated_at, user_id
		FROM train_schedules
	`

	// Add user filter if specified
	var params []interface{}
	if userID > 0 {
		query += " WHERE user_id = ?"
		params = append(params, userID)
	}

	// Add ordering
	query += " ORDER BY COALESCE(departure_date_time, arrival_date_time) DESC"

	// Execute query
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var schedules []TrainSchedule
	for rows.Next() {
		var schedule TrainSchedule
		var departureTime, arrivalTime sql.NullTime

		if err := rows.Scan(
			&schedule.ID, &schedule.TrainNumberDeparture, &schedule.TrainNumberArrival, &schedule.VehicleName,
			&schedule.StartingLocation, &schedule.EndLocation, &departureTime, &arrivalTime,
			&schedule.StartingTrack, &schedule.TargetTrack, &schedule.Employee1Departure, &schedule.Employee1Arrival,
			&schedule.DutyDeparture, &schedule.DutyArrival, &schedule.Notes, &schedule.RawData,
			&schedule.CreatedAt, &schedule.UpdatedAt, &schedule.UserID,
		); err != nil {
			return nil, err
		}

		// Handle nullable datetime fields
		if departureTime.Valid {
			schedule.DepartureDateTime = &departureTime.Time
		}
		if arrivalTime.Valid {
			schedule.ArrivalDateTime = &arrivalTime.Time
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// SaveTrainSchedules saves or updates a batch of train schedule records.
// This function handles both insertion of new records and updating of existing ones
// based on their ID field.
//
// Parameters:
//   - db: Database connection
//   - schedules: Array of train schedule records to save
//   - userID: ID of the user performing the operation
//
// Returns:
//   - Number of records processed
//   - Error if the database operation fails
func SaveTrainSchedules(db *sql.DB, schedules []TrainSchedule, userID int) (int, error) {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert or update each record
	processed := 0
	for _, schedule := range schedules {
		// Check if record already exists
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM train_schedules WHERE id = ?)", schedule.ID).
			Scan(&exists)
		if err != nil {
			return processed, err
		}

		// Prepare raw data as JSON if needed
		var rawData []byte
		if schedule.RawData == "" {
			// If no raw data provided, create JSON from the record itself
			rawData, _ = json.Marshal(schedule)
		} else {
			rawData = []byte(schedule.RawData)
		}

		if exists {
			// Update existing record
			_, err = tx.Exec(
				`
				UPDATE train_schedules SET
					train_number_departure = ?,
					train_number_arrival = ?,
					vehicle_name = ?,
					starting_location = ?,
					end_location = ?,
					departure_date_time = ?,
					arrival_date_time = ?,
					starting_track = ?,
					target_track = ?,
					employee1_departure = ?,
					employee1_arrival = ?,
					duty_departure = ?,
					duty_arrival = ?,
					notes = ?,
					raw_data = ?,
					updated_at = NOW(),
					user_id = ?
				WHERE id = ?
			`,
				schedule.TrainNumberDeparture,
				schedule.TrainNumberArrival,
				schedule.VehicleName,
				schedule.StartingLocation,
				schedule.EndLocation,
				schedule.DepartureDateTime,
				schedule.ArrivalDateTime,
				schedule.StartingTrack,
				schedule.TargetTrack,
				schedule.Employee1Departure,
				schedule.Employee1Arrival,
				schedule.DutyDeparture,
				schedule.DutyArrival,
				schedule.Notes,
				string(rawData),
				userID,
				schedule.ID,
			)
		} else {
			// Insert new record
			_, err = tx.Exec(`
				INSERT INTO train_schedules (
					id, train_number_departure, train_number_arrival, vehicle_name,
					starting_location, end_location, departure_date_time, arrival_date_time,
					starting_track, target_track, employee1_departure, employee1_arrival,
					duty_departure, duty_arrival, notes, raw_data, created_at, updated_at, user_id
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
			`,
				schedule.ID, schedule.TrainNumberDeparture, schedule.TrainNumberArrival, schedule.VehicleName,
				schedule.StartingLocation, schedule.EndLocation, schedule.DepartureDateTime, schedule.ArrivalDateTime,
				schedule.StartingTrack, schedule.TargetTrack, schedule.Employee1Departure, schedule.Employee1Arrival,
				schedule.DutyDeparture, schedule.DutyArrival, schedule.Notes, string(rawData),
				userID,
			)
		}

		if err != nil {
			return processed, err
		}
		processed++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return processed, err
	}

	return processed, nil
}

// UpdateTrainScheduleField updates a specific field of a train schedule record.
// This allows for partial updates of individual fields without having to send
// the entire record.
//
// Parameters:
//   - db: Database connection
//   - id: ID of the record to update
//   - field: Name of the field to update
//   - value: New value for the field
//   - userID: ID of the user performing the update
//
// Returns:
//   - Error if the database operation fails
func UpdateTrainScheduleField(db *sql.DB, id string, field string, value string, userID int) error {
	// Validate field name to prevent SQL injection
	allowedFields := map[string]string{
		"startingTrack": "starting_track",
		"targetTrack":   "target_track",
		"notes":         "notes",
	}

	dbField, allowed := allowedFields[field]
	if !allowed {
		return sql.ErrNoRows // Use standard error to avoid exposing details
	}

	// Check if record exists and belongs to user
	var exists bool
	var recordUserID sql.NullInt64
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM train_schedules WHERE id = ?), 
		(SELECT user_id FROM train_schedules WHERE id = ?)
	`, id, id).Scan(&exists, &recordUserID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	// Allow update if record belongs to user or user is admin
	// (Admin check would be done at the handler level)
	if recordUserID.Valid && int(recordUserID.Int64) != userID {
		return sql.ErrNoRows // Use standard error for security
	}

	// Build and execute the update query
	query := "UPDATE train_schedules SET " + dbField + " = ?, updated_at = NOW() WHERE id = ?"
	_, err = db.Exec(query, value, id)
	return err
}

// DeleteTrainSchedule removes a train schedule record from the database.
//
// Parameters:
//   - db: Database connection
//   - id: ID of the record to delete
//   - userID: ID of the user requesting the deletion
//
// Returns:
//   - Error if the database operation fails
func DeleteTrainSchedule(db *sql.DB, id string, userID int) error {
	// Check if record exists and belongs to user
	var exists bool
	var recordUserID sql.NullInt64
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM train_schedules WHERE id = ?), 
		(SELECT user_id FROM train_schedules WHERE id = ?)
	`, id, id).Scan(&exists, &recordUserID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	// Allow deletion if record belongs to user or user is admin
	// (Admin check would be done at the handler level)
	if recordUserID.Valid && int(recordUserID.Int64) != userID {
		return sql.ErrNoRows // Use standard error for security
	}

	// Delete the record
	_, err = db.Exec("DELETE FROM train_schedules WHERE id = ?", id)
	return err
}
