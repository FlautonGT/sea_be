package admin

import (
	"database/sql"
	"strings"
)

// nullString converts a string to sql.NullString
func nullString(value string) sql.NullString {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

// getMembershipName returns the display name for membership level
func getMembershipName(level string) string {
	switch level {
	case "CLASSIC":
		return "Classic"
	case "PRESTIGE":
		return "Prestige"
	case "ROYAL":
		return "Royal"
	default:
		return "Classic"
	}
}

