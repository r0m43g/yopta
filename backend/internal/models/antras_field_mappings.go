// backend/internal/models/antras_field_mappings.go
package models

import (
	"database/sql"
	"fmt"
	"time"
)

// AntrasFieldMapping represents a mapping between external and internal field names for Antras page
// This allows dynamic configuration of field name mappings without code changes
// Used for train movement data import (37 fields from Excel export)
type AntrasFieldMapping struct {
	ID           int       `json:"id"`
	ExternalName string    `json:"external_name"`
	InternalName string    `json:"internal_name"`
	DisplayName  string    `json:"display_name"`
	FieldType    string    `json:"field_type"`
	IsRequired   bool      `json:"is_required"`
	SortOrder    int       `json:"sort_order"`
	Description  *string   `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetAllAntrasFieldMappings retrieves all Antras field mappings ordered by sort_order
func GetAllAntrasFieldMappings(db *sql.DB) ([]AntrasFieldMapping, error) {
	query := `
		SELECT 
			id, external_name, internal_name, display_name, 
			field_type, is_required, sort_order, description,
			created_at, updated_at
		FROM antras_field_mappings
		ORDER BY sort_order ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query antras field mappings: %w", err)
	}
	defer rows.Close()

	var mappings []AntrasFieldMapping
	for rows.Next() {
		var m AntrasFieldMapping
		err := rows.Scan(
			&m.ID, &m.ExternalName, &m.InternalName, &m.DisplayName,
			&m.FieldType, &m.IsRequired, &m.SortOrder, &m.Description,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan antras field mapping: %w", err)
		}
		mappings = append(mappings, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating antras field mappings: %w", err)
	}

	return mappings, nil
}

// GetAntrasFieldMappingByID retrieves a single Antras field mapping by ID
func GetAntrasFieldMappingByID(db *sql.DB, id int) (*AntrasFieldMapping, error) {
	query := `
		SELECT 
			id, external_name, internal_name, display_name,
			field_type, is_required, sort_order, description,
			created_at, updated_at
		FROM antras_field_mappings
		WHERE id = ?
	`

	var m AntrasFieldMapping
	err := db.QueryRow(query, id).Scan(
		&m.ID, &m.ExternalName, &m.InternalName, &m.DisplayName,
		&m.FieldType, &m.IsRequired, &m.SortOrder, &m.Description,
		&m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("antras field mapping not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get antras field mapping: %w", err)
	}

	return &m, nil
}

// CreateAntrasFieldMapping creates a new Antras field mapping
func CreateAntrasFieldMapping(db *sql.DB, m *AntrasFieldMapping) error {
	query := `
		INSERT INTO antras_field_mappings 
			(external_name, internal_name, display_name, field_type, 
			 is_required, sort_order, description)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		m.ExternalName, m.InternalName, m.DisplayName, m.FieldType,
		m.IsRequired, m.SortOrder, m.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create antras field mapping: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	m.ID = int(id)
	return nil
}

// UpdateAntrasFieldMapping updates an existing Antras field mapping
func UpdateAntrasFieldMapping(db *sql.DB, id int, m *AntrasFieldMapping) error {
	query := `
		UPDATE antras_field_mappings
		SET 
			external_name = ?,
			internal_name = ?,
			display_name = ?,
			field_type = ?,
			is_required = ?,
			sort_order = ?,
			description = ?
		WHERE id = ?
	`

	result, err := db.Exec(
		query,
		m.ExternalName, m.InternalName, m.DisplayName, m.FieldType,
		m.IsRequired, m.SortOrder, m.Description, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update antras field mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("antras field mapping not found")
	}

	return nil
}

// DeleteAntrasFieldMapping deletes an Antras field mapping by ID
func DeleteAntrasFieldMapping(db *sql.DB, id int) error {
	query := `DELETE FROM antras_field_mappings WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete antras field mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("antras field mapping not found")
	}

	return nil
}

// GetAntrasFieldMappingsAsMap returns Antras field mappings as a map for quick lookup
// Key: external_name, Value: internal_name
// This format is optimized for the frontend import logic
func GetAntrasFieldMappingsAsMap(db *sql.DB) (map[string]string, error) {
	mappings, err := GetAllAntrasFieldMappings(db)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(mappings))
	for _, m := range mappings {
		result[m.ExternalName] = m.InternalName
	}

	return result, nil
}

// ReorderAntrasFieldMappings updates sort order for multiple Antras field mappings
// Takes a map of ID -> new sort order and updates them in a transaction
func ReorderAntrasFieldMappings(db *sql.DB, orders map[int]int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE antras_field_mappings SET sort_order = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for id, order := range orders {
		_, err := stmt.Exec(order, id)
		if err != nil {
			return fmt.Errorf("failed to update sort order for id %d: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
