// backend/internal/handlers/settings.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	ID           int    `json:"id,omitempty"`
	SettingKey   string `json:"setting_key"`
	SettingValue string `json:"setting_value"`
	Description  string `json:"description,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// BlacklistedIP represents an IP address that is blocked from accessing the system
type BlacklistedIP struct {
	ID        int    `json:"id,omitempty"`
	IPAddress string `json:"ip_address"`
	Reason    string `json:"reason,omitempty"`
	AddedBy   int    `json:"added_by,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Username  string `json:"username,omitempty"` // For displaying who added it
}

// GetSystemSettings returns all system settings
func GetSystemSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query all settings from the database
		rows, err := db.Query(
			"SELECT id, setting_key, setting_value, description, updated_at FROM system_settings",
		)
		if err != nil {
			http.Error(w, "Klaida gaunant nustatymus", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Parse settings into a slice
		var settings []SystemSetting
		for rows.Next() {
			var setting SystemSetting
			err := rows.Scan(
				&setting.ID,
				&setting.SettingKey,
				&setting.SettingValue,
				&setting.Description,
				&setting.UpdatedAt,
			)
			if err != nil {
				http.Error(w, "Klaida apdorojant nustatymus", http.StatusInternalServerError)
				return
			}
			settings = append(settings, setting)
		}

		// Return settings as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)
	}
}

// UpdateSystemSetting updates a single system setting
func UpdateSystemSetting(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var setting SystemSetting
		if err := json.NewDecoder(r.Body).Decode(&setting); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Update the setting in the database
		_, err := db.Exec(
			"UPDATE system_settings SET setting_value = ? WHERE setting_key = ?",
			setting.SettingValue, setting.SettingKey,
		)
		if err != nil {
			http.Error(w, "Klaida atnaujinant nustatymą", http.StatusInternalServerError)
			return
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Nustatymas sėkmingai atnaujintas", // Setting successfully updated
		})
	}
}

// GetBlacklistedIPs returns all blacklisted IP addresses
func GetBlacklistedIPs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query blacklisted IPs with added_by username
		rows, err := db.Query(`
			SELECT b.id, b.ip_address, b.reason, b.added_by, u.username, b.created_at
			FROM blacklisted_ips b
			LEFT JOIN users u ON b.added_by = u.id
			ORDER BY b.created_at DESC
		`)
		if err != nil {
			http.Error(
				w,
				"Klaida gaunant juodame sąraše esančius IP adresus",
				http.StatusInternalServerError,
			)
			return
		}
		defer rows.Close()

		// Parse IPs into a slice
		var ips []BlacklistedIP
		for rows.Next() {
			var ip BlacklistedIP
			var addedBy sql.NullInt64
			var username sql.NullString
			err := rows.Scan(&ip.ID, &ip.IPAddress, &ip.Reason, &addedBy, &username, &ip.CreatedAt)
			if err != nil {
				http.Error(w, "Klaida apdorojant IP adresus", http.StatusInternalServerError)
				return
			}

			if addedBy.Valid {
				ip.AddedBy = int(addedBy.Int64)
			}

			if username.Valid {
				ip.Username = username.String
			}

			ips = append(ips, ip)
		}

		// Return IPs as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ips)
	}
}

// AddBlacklistedIP adds a new IP to the blacklist
func AddBlacklistedIP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ip BlacklistedIP
		if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Get user ID from context
		userID, err := getUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Nepavyko gauti vartotojo ID", http.StatusUnauthorized)
			return
		}

		// Insert new blacklisted IP
		_, err = db.Exec(
			"INSERT INTO blacklisted_ips (ip_address, reason, added_by) VALUES (?, ?, ?)",
			ip.IPAddress, ip.Reason, userID,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida pridedant IP adresą į juodąjį sąrašą",
				http.StatusInternalServerError,
			)
			return
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "IP adresas sėkmingai pridėtas į juodąjį sąrašą",
		})
	}
}

// RemoveBlacklistedIP removes an IP from the blacklist
func RemoveBlacklistedIP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ip BlacklistedIP
		if err := json.NewDecoder(r.Body).Decode(&ip); err != nil {
			http.Error(w, "Neteisingas užklausos formatas", http.StatusBadRequest)
			return
		}

		// Delete IP from blacklist
		_, err := db.Exec(
			"DELETE FROM blacklisted_ips WHERE id = ?",
			ip.ID,
		)
		if err != nil {
			http.Error(
				w,
				"Klaida šalinant IP adresą iš juodojo sąrašo",
				http.StatusInternalServerError,
			)
			return
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "IP adresas sėkmingai pašalintas iš juodojo sąrašo",
		})
	}
}
