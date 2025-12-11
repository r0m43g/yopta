// backend/internal/models/field_mapping.go
package models

import (
	"database/sql"
	"fmt"
	"time"
)

// FieldMapping represents a mapping between external and internal field names
// This allows dynamic configuration of field name mappings without code changes
type FieldMapping struct {
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

// GetAllFieldMappings retrieves all field mappings ordered by sort_order
func GetAllFieldMappings(db *sql.DB) ([]FieldMapping, error) {
	query := `
		SELECT 
			id, external_name, internal_name, display_name, 
			field_type, is_required, sort_order, description,
			created_at, updated_at
		FROM field_mappings
		ORDER BY sort_order ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query field mappings: %w", err)
	}
	defer rows.Close()

	var mappings []FieldMapping
	for rows.Next() {
		var m FieldMapping
		err := rows.Scan(
			&m.ID, &m.ExternalName, &m.InternalName, &m.DisplayName,
			&m.FieldType, &m.IsRequired, &m.SortOrder, &m.Description,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field mapping: %w", err)
		}
		mappings = append(mappings, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating field mappings: %w", err)
	}

	return mappings, nil
}

// GetFieldMappingByID retrieves a single field mapping by ID
func GetFieldMappingByID(db *sql.DB, id int) (*FieldMapping, error) {
	query := `
		SELECT 
			id, external_name, internal_name, display_name,
			field_type, is_required, sort_order, description,
			created_at, updated_at
		FROM field_mappings
		WHERE id = ?
	`

	var m FieldMapping
	err := db.QueryRow(query, id).Scan(
		&m.ID, &m.ExternalName, &m.InternalName, &m.DisplayName,
		&m.FieldType, &m.IsRequired, &m.SortOrder, &m.Description,
		&m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("field mapping not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get field mapping: %w", err)
	}

	return &m, nil
}

// CreateFieldMapping creates a new field mapping
func CreateFieldMapping(db *sql.DB, m *FieldMapping) error {
	query := `
		INSERT INTO field_mappings 
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
		return fmt.Errorf("failed to create field mapping: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	m.ID = int(id)
	return nil
}

// UpdateFieldMapping updates an existing field mapping
func UpdateFieldMapping(db *sql.DB, id int, m *FieldMapping) error {
	query := `
		UPDATE field_mappings
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
		return fmt.Errorf("failed to update field mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field mapping not found")
	}

	return nil
}

// DeleteFieldMapping deletes a field mapping by ID
func DeleteFieldMapping(db *sql.DB, id int) error {
	query := `DELETE FROM field_mappings WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete field mapping: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field mapping not found")
	}

	return nil
}

// GetFieldMappingsAsMap returns field mappings as a map for quick lookup
// Key: external_name, Value: internal_name
func GetFieldMappingsAsMap(db *sql.DB) (map[string]string, error) {
	mappings, err := GetAllFieldMappings(db)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(mappings))
	for _, m := range mappings {
		result[m.ExternalName] = m.InternalName
	}

	return result, nil
}

// ReorderFieldMappings updates sort order for multiple field mappings
func ReorderFieldMappings(db *sql.DB, orders map[int]int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE field_mappings SET sort_order = ? WHERE id = ?")
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
