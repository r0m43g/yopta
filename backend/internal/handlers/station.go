// backend/internal/handlers/station.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"yopta-template/internal/models"

	"github.com/go-chi/chi/v5"
)

// GetAllStations returns a list of all stations with their tracks.
// Like a train dispatcher's big board showing all stations in the network,
// this endpoint provides a comprehensive overview of available stations and their tracks.
func GetAllStations(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all stations from the database
		stations, err := models.GetAllStations(db)
		if err != nil {
			http.Error(
				w,
				"Nepavyko gauti stočių sąrašo: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		// Return as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stations)
	}
}

// GetStationByID returns a single station with its tracks.
// This is like zooming in on one station from the big board,
// showing all the detailed information about a specific location.
func GetStationByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get station ID from URL
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "Neteisingas stoties ID", http.StatusBadRequest)
			return
		}

		// Get station from database
		station, err := models.GetStationByID(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Stotis nerasta", http.StatusNotFound)
			} else {
				http.Error(w, "Nepavyko gauti stoties: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Return as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(station)
	}
}

// CreateStation adds a new station with its tracks to the database.
// This handler is like a city planner putting a brand new train station on the map,
// complete with all its tracks and platforms.
func CreateStation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var station models.Station
		if err := json.NewDecoder(r.Body).Decode(&station); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Get user ID from context
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Nepavyko gauti vartotojo ID", http.StatusUnauthorized)
			return
		}
		station.UserID = userID

		// Validate station data
		if station.Name == "" || station.Code == "" {
			http.Error(w, "Stoties pavadinimas ir kodas yra būtini", http.StatusBadRequest)
			return
		}

		// Create station in database
		stationID, err := models.CreateStation(db, station)
		if err != nil {
			http.Error(w, "Nepavyko sukurti stoties: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the created station with its ID
		newStation, err := models.GetStationByID(db, stationID)
		if err != nil {
			http.Error(
				w,
				"Stotis sukurta, bet nepavyko jos grąžinti",
				http.StatusInternalServerError,
			)
			return
		}

		// Return created station
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newStation)
	}
}

// UpdateStation modifies an existing station and its tracks.
// This endpoint is like renovating an existing station - you can
// update its name, add tracks, remove platforms, etc.
func UpdateStation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get station ID from URL
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "Neteisingas stoties ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var station models.Station
		if err := json.NewDecoder(r.Body).Decode(&station); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Ensure ID in URL matches ID in body
		station.ID = id

		// Verify station exists
		existingStation, err := models.GetStationByID(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Stotis nerasta", http.StatusNotFound)
			} else {
				http.Error(w, "Nepavyko gauti stoties: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Check if user has permission (same user or admin)
		userID, _ := getUserIDFromContext(r)
		if existingStation.UserID != userID {
			// Check if user is admin
			isAdmin, err := isUserAdmin(db, userID)
			if err != nil || !isAdmin {
				http.Error(w, "Jūs neturite teisių redaguoti šią stotį", http.StatusForbidden)
				return
			}
		}

		// Validate station data
		if station.Name == "" || station.Code == "" {
			http.Error(w, "Stoties pavadinimas ir kodas yra būtini", http.StatusBadRequest)
			return
		}

		// Update station
		if err := models.UpdateStation(db, station); err != nil {
			http.Error(
				w,
				"Nepavyko atnaujinti stoties: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		// Get updated station
		updatedStation, err := models.GetStationByID(db, id)
		if err != nil {
			http.Error(
				w,
				"Stotis atnaujinta, bet nepavyko jos grąžinti",
				http.StatusInternalServerError,
			)
			return
		}

		// Return updated station
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedStation)
	}
}

// DeleteStation removes a station and all its tracks from the database.
// Like demolishing an old train station to make way for something new,
// this endpoint completely removes a station and all its associated data.
func DeleteStation(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get station ID from URL
		idParam := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "Neteisingas stoties ID", http.StatusBadRequest)
			return
		}

		// Verify station exists
		existingStation, err := models.GetStationByID(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Stotis nerasta", http.StatusNotFound)
			} else {
				http.Error(w, "Nepavyko gauti stoties: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Check if user has permission (same user or admin)
		userID, _ := getUserIDFromContext(r)
		if existingStation.UserID != userID {
			// Check if user is admin
			isAdmin, err := isUserAdmin(db, userID)
			if err != nil || !isAdmin {
				http.Error(w, "Jūs neturite teisių ištrinti šią stotį", http.StatusForbidden)
				return
			}
		}

		// Delete station
		if err := models.DeleteStation(db, id); err != nil {
			http.Error(w, "Nepavyko ištrinti stoties: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Return success
		w.WriteHeader(http.StatusNoContent)
	}
}

// Helper function to check if a user is an admin
func isUserAdmin(db *sql.DB, userID int) (bool, error) {
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

// AddTrack adds a new track to an existing station.
// This is like adding a new platform to an existing station,
// expanding its capacity for handling trains.
func AddTrack(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get station ID from URL
		stationIDParam := chi.URLParam(r, "stationId")
		stationID, err := strconv.Atoi(stationIDParam)
		if err != nil {
			http.Error(w, "Neteisingas stoties ID", http.StatusBadRequest)
			return
		}

		// Verify station exists
		station, err := models.GetStationByID(db, stationID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Stotis nerasta", http.StatusNotFound)
			} else {
				http.Error(w, "Nepavyko gauti stoties: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Check if user has permission (same user or admin)
		userID, _ := getUserIDFromContext(r)
		if station.UserID != userID {
			// Check if user is admin
			isAdmin, err := isUserAdmin(db, userID)
			if err != nil || !isAdmin {
				http.Error(w, "Jūs neturite teisių redaguoti šią stotį", http.StatusForbidden)
				return
			}
		}

		// Parse request body
		var track models.Track
		if err := json.NewDecoder(r.Body).Decode(&track); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Set station ID on track
		track.StationID = stationID

		// Validate track data
		if track.TrackNumber == "" {
			http.Error(w, "Kelio numeris yra būtinas", http.StatusBadRequest)
			return
		}
		if track.Positions < 1 {
			track.Positions = 1 // Default to 1 position if not specified
		}

		// Add track to database
		_, err = models.AddTrack(db, track)
		if err != nil {
			http.Error(w, "Nepavyko pridėti kelio: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get updated station with new track
		updatedStation, err := models.GetStationByID(db, stationID)
		if err != nil {
			http.Error(
				w,
				"Kelias pridėtas, bet nepavyko grąžinti stoties",
				http.StatusInternalServerError,
			)
			return
		}

		// Return updated station
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(updatedStation)
	}
}

// UpdateTrack modifies an existing track.
// This is like renovating a specific platform at a station -
// you can change its length, number of positions, etc.
func UpdateTrack(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get track ID from URL
		trackIDParam := chi.URLParam(r, "trackId")
		trackID, err := strconv.Atoi(trackIDParam)
		if err != nil {
			http.Error(w, "Neteisingas kelio ID", http.StatusBadRequest)
			return
		}

		// Parse request body
		var track models.Track
		if err := json.NewDecoder(r.Body).Decode(&track); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Set track ID
		track.ID = trackID

		// Get track to check its station
		var stationID int
		err = db.QueryRow("SELECT station_id FROM tracks WHERE id = ?", trackID).Scan(&stationID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Kelias nerastas", http.StatusNotFound)
			} else {
				http.Error(w, "Nepavyko gauti kelio: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Get the station to check permissions
		station, err := models.GetStationByID(db, stationID)
		if err != nil {
			http.Error(w, "Nepavyko gauti stoties: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if user has permission (same user or admin)
		userID, _ := getUserIDFromContext(r)
		if station.UserID != userID {
			// Check if user is admin
			isAdmin, err := isUserAdmin(db, userID)
			if err != nil || !isAdmin {
				http.Error(w, "Jūs neturite teisių redaguoti šį kelią", http.StatusForbidden)
				return
			}
		}

		// Validate track data
		if track.TrackNumber == "" {
			http.Error(w, "Kelio numeris yra būtinas", http.StatusBadRequest)
			return
		}
		if track.Positions < 1 {
			track.Positions = 1 // Default to 1 position if not specified
		}

		// Make sure track stays with its station
		track.StationID = stationID

		// Update track
		if err := models.UpdateTrack(db, track); err != nil {
			http.Error(w, "Nepavyko atnaujinti kelio: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get updated station with modified track
		updatedStation, err := models.GetStationByID(db, stationID)
		if err != nil {
			http.Error(
				w,
				"Kelias atnaujintas, bet nepavyko grąžinti stoties",
				http.StatusInternalServerError,
			)
			return
		}

		// Return updated station
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedStation)
	}
}

// DeleteTrack removes a track from a station.
// This is like removing an old platform that's no longer needed,
// freeing up space and reducing maintenance costs.
func DeleteTrack(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get track ID from URL
		trackIDParam := chi.URLParam(r, "trackId")
		trackID, err := strconv.Atoi(trackIDParam)
		if err != nil {
			http.Error(w, "Neteisingas kelio ID", http.StatusBadRequest)
			return
		}

		// Get track to check its station
		var stationID int
		err = db.QueryRow("SELECT station_id FROM tracks WHERE id = ?", trackID).Scan(&stationID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Kelias nerastas", http.StatusNotFound)
			} else {
				http.Error(w, "Nepavyko gauti kelio: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Get the station to check permissions
		station, err := models.GetStationByID(db, stationID)
		if err != nil {
			http.Error(w, "Nepavyko gauti stoties: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if user has permission (same user or admin)
		userID, _ := getUserIDFromContext(r)
		if station.UserID != userID {
			// Check if user is admin
			isAdmin, err := isUserAdmin(db, userID)
			if err != nil || !isAdmin {
				http.Error(w, "Jūs neturite teisių ištrinti šį kelią", http.StatusForbidden)
				return
			}
		}

		// Delete track
		if err := models.DeleteTrack(db, trackID); err != nil {
			http.Error(w, "Nepavyko ištrinti kelio: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get updated station without the deleted track
		updatedStation, err := models.GetStationByID(db, stationID)
		if err != nil {
			http.Error(
				w,
				"Kelias ištrintas, bet nepavyko grąžinti stoties",
				http.StatusInternalServerError,
			)
			return
		}

		// Return updated station
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedStation)
	}
}
