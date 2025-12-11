// backend/internal/handlers/field_mappings.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"yopta-template/internal/models"

	"github.com/go-chi/chi/v5"
)

// GetFieldMappings returns all field mappings
// This endpoint provides the complete list of field name mappings
// ordered by their display order
func GetFieldMappings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mappings, err := models.GetAllFieldMappings(db)
		if err != nil {
			http.Error(w, "Nepavyko gauti laukų atvaizdavimų: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mappings)
	}
}

// GetFieldMappingsMap returns field mappings as a key-value map
// Key: external_name, Value: internal_name
// This format is optimized for the frontend import logic
func GetFieldMappingsMap(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mappingsMap, err := models.GetFieldMappingsAsMap(db)
		if err != nil {
			http.Error(w, "Nepavyko gauti laukų atvaizdavimų: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mappingsMap)
	}
}

// GetFieldMapping returns a single field mapping by ID
func GetFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Netinkamas ID", http.StatusBadRequest)
			return
		}

		mapping, err := models.GetFieldMappingByID(db, id)
		if err != nil {
			http.Error(w, "Nepavyko gauti lauko atvaizdavimo: "+err.Error(),
				http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapping)
	}
}

// CreateFieldMapping creates a new field mapping
func CreateFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mapping models.FieldMapping
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

		if err := models.CreateFieldMapping(db, &mapping); err != nil {
			http.Error(w, "Nepavyko sukurti lauko atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mapping)
	}
}

// UpdateFieldMapping updates an existing field mapping
func UpdateFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Netinkamas ID", http.StatusBadRequest)
			return
		}

		var mapping models.FieldMapping
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

		if err := models.UpdateFieldMapping(db, id, &mapping); err != nil {
			http.Error(w, "Nepavyko atnaujinti lauko atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		// Fetch updated mapping
		updated, err := models.GetFieldMappingByID(db, id)
		if err != nil {
			http.Error(w, "Nepavyko gauti atnaujinto atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}

// DeleteFieldMapping deletes a field mapping
func DeleteFieldMapping(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Netinkamas ID", http.StatusBadRequest)
			return
		}

		if err := models.DeleteFieldMapping(db, id); err != nil {
			http.Error(w, "Nepavyko ištrinti lauko atvaizdavimo: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ReorderFieldMappings updates the sort order of field mappings
func ReorderFieldMappings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var orders map[string]int
		if err := json.NewDecoder(r.Body).Decode(&orders); err != nil {
			http.Error(w, "Netinkami duomenys: "+err.Error(),
				http.StatusBadRequest)
			return
		}

		// Convert string keys to int
		intOrders := make(map[int]int)
		for k, v := range orders {
			id, err := strconv.Atoi(k)
			if err != nil {
				http.Error(w, "Netinkamas ID: "+k, http.StatusBadRequest)
				return
			}
			intOrders[id] = v
		}

		if err := models.ReorderFieldMappings(db, intOrders); err != nil {
			http.Error(w, "Nepavyko pertvarkyti laukų: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
