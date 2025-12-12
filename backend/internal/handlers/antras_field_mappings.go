// backend/internal/handlers/antras_field_mappings.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"yopta-template/internal/models"

	"github.com/go-chi/chi/v5"
)

// GetAntrasFieldMappings returns all Antras field mappings
// This endpoint provides the complete list of field name mappings for Antras page
// ordered by their display order
func GetAntrasFieldMappings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mappings, err := models.GetAllAntrasFieldMappings(db)
		if err != nil {
			http.Error(w, "Nepavyko gauti Antras laukų atvaizdavimų: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mappings)
	}
}

// GetAntrasFieldMappingsMap returns Antras field mappings as a key-value map
// Key: external_name, Value: internal_name
// This format is optimized for the frontend import logic
func GetAntrasFieldMappingsMap(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mappingsMap, err := models.GetAntrasFieldMappingsAsMap(db)
		if err != nil {
			http.Error(w, "Nepavyko gauti Antras laukų atvaizdavimų: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mappingsMap)
	}
}

// GetAntrasFieldMapping returns a single Antras field mapping by ID
func GetAntrasFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Netinkamas ID", http.StatusBadRequest)
			return
		}

		mapping, err := models.GetAntrasFieldMappingByID(db, id)
		if err != nil {
			http.Error(w, "Nepavyko gauti Antras lauko atvaizdavimo: "+err.Error(),
				http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapping)
	}
}

// CreateAntrasFieldMapping creates a new Antras field mapping
func CreateAntrasFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mapping models.AntrasFieldMapping
		if err := json.NewDecoder(r.Body).Decode(&mapping); err != nil {
			http.Error(w, "Netinkami duomenys: "+err.Error(),
				http.StatusBadRequest)
			return
		}

		// Validation
		if mapping.ExternalName == "" {
			http.Error(w, "Išorinis pavadinimas yra privalomas",
				http.StatusBadRequest)
			return
		}
		if mapping.InternalName == "" {
			http.Error(w, "Vidinis pavadinimas yra privalomas",
				http.StatusBadRequest)
			return
		}
		if mapping.DisplayName == "" {
			http.Error(w, "Rodomas pavadinimas yra privalomas",
				http.StatusBadRequest)
			return
		}

		if err := models.CreateAntrasFieldMapping(db, &mapping); err != nil {
			http.Error(w, "Nepavyko sukurti Antras lauko atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mapping)
	}
}

// UpdateAntrasFieldMapping updates an existing Antras field mapping
func UpdateAntrasFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Netinkamas ID", http.StatusBadRequest)
			return
		}

		var mapping models.AntrasFieldMapping
		if err := json.NewDecoder(r.Body).Decode(&mapping); err != nil {
			http.Error(w, "Netinkami duomenys: "+err.Error(),
				http.StatusBadRequest)
			return
		}

		// Validation
		if mapping.ExternalName == "" {
			http.Error(w, "Išorinis pavadinimas yra privalomas",
				http.StatusBadRequest)
			return
		}
		if mapping.InternalName == "" {
			http.Error(w, "Vidinis pavadinimas yra privalomas",
				http.StatusBadRequest)
			return
		}
		if mapping.DisplayName == "" {
			http.Error(w, "Rodomas pavadinimas yra privalomas",
				http.StatusBadRequest)
			return
		}

		if err := models.UpdateAntrasFieldMapping(db, id, &mapping); err != nil {
			http.Error(w, "Nepavyko atnaujinti Antras lauko atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapping)
	}
}

// DeleteAntrasFieldMapping deletes an Antras field mapping by ID
func DeleteAntrasFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Netinkamas ID", http.StatusBadRequest)
			return
		}

		if err := models.DeleteAntrasFieldMapping(db, id); err != nil {
			http.Error(w, "Nepavyko ištrinti Antras lauko atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ReorderAntrasFieldMappings updates sort order for multiple Antras field mappings
func ReorderAntrasFieldMappings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orders map[int]int
		if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
			http.Error(w, "Netinkami duomenys: "+err.Error(),
				http.StatusBadRequest)
			return
		}

		if err := models.ReorderAntrasFieldMappings(db, orders); err != nil {
			http.Error(w, "Nepavyko atnaujinti Antras laukų tvarkos: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
