package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gate-v2/internal/middleware"
	"gate-v2/internal/storage"
	"gate-v2/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// ============================================
// ADMIN SETTINGS MANAGEMENT
// ============================================

// handleGetSettingsImpl returns all settings
func HandleGetSettingsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock settings data
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"general": map[string]interface{}{
				"siteName":           "Gate.co.id",
				"siteDescription":    "Top Up Game & Voucher Digital Terpercaya",
				"maintenanceMode":    false,
				"maintenanceMessage": nil,
			},
			"transaction": map[string]interface{}{
				"orderExpiry":      3600,
				"autoRefundOnFail": true,
				"maxRetryAttempts": 3,
			},
			"notification": map[string]interface{}{
				"emailEnabled":    true,
				"whatsappEnabled": true,
				"telegramEnabled": false,
			},
			"security": map[string]interface{}{
				"maxLoginAttempts": 5,
				"lockoutDuration":  900,
				"sessionTimeout":   3600,
				"mfaRequired":      true,
			},
		})
	}
}

// UpdateSettingsRequest represents the request to update settings
type UpdateSettingsRequest struct {
	Settings map[string]interface{} `json:"settings"`
}

// handleUpdateSettingsImpl updates settings for a category
func HandleUpdateSettingsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		category := chi.URLParam(r, "category")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req UpdateSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Create audit log
		deps.DB.Pool.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'SETTINGS', $2, 'Updated settings', NOW())
		`, adminID, category)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Settings updated successfully",
		})
	}
}

// handleGetContactSettingsImpl returns contact settings
func HandleGetContactSettingsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock contact settings
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"email":     "support@gate.co.id",
			"phone":     "+6281234567890",
			"whatsapp":  "https://wa.me/6281234567890",
			"instagram": "https://instagram.com/gate.official",
			"facebook":  "https://facebook.com/gate.official",
			"x":         "https://x.com/gate_official",
			"youtube":   "https://youtube.com/@gateofficial",
			"telegram":  "https://t.me/gate_official",
			"discord":   "https://discord.gg/gate",
		})
	}
}

// UpdateContactSettingsRequest represents the request to update contact settings
type UpdateContactSettingsRequest struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	WhatsApp  string `json:"whatsapp"`
	Instagram string `json:"instagram"`
	Facebook  string `json:"facebook"`
	X         string `json:"x"`
	YouTube   string `json:"youtube"`
	Telegram  string `json:"telegram"`
	Discord   string `json:"discord"`
}

// handleUpdateContactSettingsImpl updates contact settings
func HandleUpdateContactSettingsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req UpdateContactSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Create audit log
		deps.DB.Pool.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'SETTINGS', 'contacts', 'Updated contact settings', NOW())
		`, adminID)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Contact settings updated successfully",
		})
	}
}

// ============================================
// REGION MANAGEMENT
// ============================================

// handleAdminGetRegionsImpl returns all regions
func HandleAdminGetRegionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT id, code, country, currency, currency_symbol, image, is_default, is_active, sort_order, created_at, updated_at
			FROM regions
			ORDER BY sort_order ASC
		`)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		regions := []map[string]interface{}{}
		for rows.Next() {
			var id, code, country, currency, currencySymbol string
			var image *string
			var isDefault, isActive bool
			var sortOrder int
			var createdAt time.Time
			var updatedAt *time.Time

			if err := rows.Scan(&id, &code, &country, &currency, &currencySymbol, &image, &isDefault, &isActive, &sortOrder, &createdAt, &updatedAt); err != nil {
				fmt.Printf("Error scanning region row: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}

			region := map[string]interface{}{
				"id":             id,
				"code":           code,
				"country":        country,
				"currency":       currency,
				"currencySymbol": currencySymbol,
				"isDefault":      isDefault,
				"isActive":       isActive,
				"sortOrder":      sortOrder,
				"createdAt":      createdAt.Format(time.RFC3339),
			}
			if image != nil {
				region["image"] = *image
			}
			if updatedAt != nil {
				region["updatedAt"] = updatedAt.Format(time.RFC3339)
			}
			regions = append(regions, region)
		}

		if err := rows.Err(); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, regions)
	}
}

// CreateRegionRequest represents the request to create a region
type CreateRegionRequest struct {
	Code           string `json:"code"`
	Country        string `json:"country"`
	Currency       string `json:"currency"`
	CurrencySymbol string `json:"currencySymbol"`
	Image          string `json:"image"`
	IsDefault      bool   `json:"isDefault"`
	IsActive       bool   `json:"isActive"`
	Order          int    `json:"order"`
}

// handleCreateRegionImpl creates a new region
func HandleCreateRegionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.WriteBadRequestError(w, fmt.Sprintf("Invalid form data: %v", err))
			return
		}

		// Get form values
		code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
		country := strings.TrimSpace(r.FormValue("country"))
		currency := strings.ToUpper(strings.TrimSpace(r.FormValue("currency")))
		currencySymbol := strings.TrimSpace(r.FormValue("currencySymbol"))
		orderStr := strings.TrimSpace(r.FormValue("order"))
		isDefaultStr := r.FormValue("isDefault")
		isActiveStr := r.FormValue("isActive")

		// Validation
		if code == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Code is required"})
			return
		}
		if country == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"country": "Country is required"})
			return
		}
		if currency == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"currency": "Currency is required"})
			return
		}

		order := 1
		if orderStr != "" {
			if parsedOrder, err := strconv.Atoi(orderStr); err == nil {
				order = parsedOrder
			}
		}
		isDefault := isDefaultStr == "true"
		isActive := isActiveStr != "false" // Default to true

		// Upload image to S3
		var imageURL string
		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err != nil && !errors.Is(err, http.ErrMissingFile) {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			}
			if err == nil {
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderFlag, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			}
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// If setting as default, unset other defaults first
		if isDefault {
			_, err = tx.Exec(ctx, `UPDATE regions SET is_default = false WHERE is_default = true`)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		var regionID string
		err = tx.QueryRow(ctx, `
			INSERT INTO regions (code, country, currency, currency_symbol, image, is_default, is_active, sort_order, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			RETURNING id
		`, code, country, currency, currencySymbol, imageURL, isDefault, isActive, order).Scan(&regionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'REGION', $2, 'Created region', NOW())
		`, adminID, regionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      regionID,
			"message": "Region created successfully",
		})
	}
}

// handleUpdateRegionImpl updates an existing region
func HandleUpdateRegionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		regionID := chi.URLParam(r, "regionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.WriteBadRequestError(w, fmt.Sprintf("Invalid form data: %v", err))
			return
		}

		// Get existing region first to preserve unchanged fields
		var existingCountry, existingCurrency, existingCurrencySymbol, existingCode string
		var existingImage sql.NullString
		var existingIsDefault, existingIsActive bool
		var existingSortOrder int

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT code, country, currency, currency_symbol, image, is_default, is_active, sort_order
			FROM regions
			WHERE id = $1
		`, regionID).Scan(&existingCode, &existingCountry, &existingCurrency, &existingCurrencySymbol, &existingImage, &existingIsDefault, &existingIsActive, &existingSortOrder)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "REGION_NOT_FOUND", "Region not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get form values (use existing values if not provided)
		country := strings.TrimSpace(r.FormValue("country"))
		if country == "" {
			country = existingCountry
		}
		currency := strings.ToUpper(strings.TrimSpace(r.FormValue("currency")))
		if currency == "" {
			currency = existingCurrency
		}
		currencySymbol := strings.TrimSpace(r.FormValue("currencySymbol"))
		if currencySymbol == "" {
			currencySymbol = existingCurrencySymbol
		}
		orderStr := strings.TrimSpace(r.FormValue("order"))
		isDefaultStr := r.FormValue("isDefault")
		isActiveStr := r.FormValue("isActive")

		sortOrder := existingSortOrder
		if orderStr != "" {
			if parsedOrder, err := strconv.Atoi(orderStr); err == nil {
				sortOrder = parsedOrder
			}
		}

		isDefault := existingIsDefault
		if isDefaultStr != "" {
			isDefault = isDefaultStr == "true"
		}

		isActive := existingIsActive
		if isActiveStr != "" {
			isActive = isActiveStr == "true"
		}

		// Handle image upload
		imageURL := ""
		if existingImage.Valid {
			imageURL = existingImage.String
		}

		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err == nil {
				// New image uploaded
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderFlag, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			} else if !errors.Is(err, http.ErrMissingFile) {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			} else {
				// Check if image URL provided as form value
				imageURLValue := strings.TrimSpace(r.FormValue("image"))
				if imageURLValue != "" {
					imageURL = imageURLValue
				}
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			}
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// If setting as default, unset other defaults first
		if isDefault && !existingIsDefault {
			_, err = tx.Exec(ctx, `UPDATE regions SET is_default = false WHERE is_default = true AND id != $1`, regionID)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		// Build update query dynamically
		updateFields := []string{"country = $1", "currency = $2", "currency_symbol = $3", "is_active = $4", "sort_order = $5", "updated_at = NOW()"}
		args := []interface{}{country, currency, currencySymbol, isActive, sortOrder}
		argPos := 6

		if imageURL != "" {
			updateFields = append(updateFields, fmt.Sprintf("image = $%d", argPos))
			args = append(args, imageURL)
			argPos++
		}

		if isDefault != existingIsDefault {
			updateFields = append(updateFields, fmt.Sprintf("is_default = $%d", argPos))
			args = append(args, isDefault)
			argPos++
		}

		args = append(args, regionID)

		query := fmt.Sprintf(`
			UPDATE regions SET
				%s
			WHERE id = $%d
		`, strings.Join(updateFields, ", "), argPos)

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'REGION', $2, 'Updated region', NOW())
		`, adminID, regionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Region updated successfully",
		})
	}
}

// handleDeleteRegionImpl deletes a region
func HandleDeleteRegionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		regionID := chi.URLParam(r, "regionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get region code first
		var regionCode string
		err := deps.DB.Pool.QueryRow(ctx, "SELECT code FROM regions WHERE id = $1", regionID).Scan(&regionCode)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "REGION_NOT_FOUND", "Region not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Check if there are active users with this region
		var userCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM users WHERE region = $1
		`, regionCode).Scan(&userCount)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if userCount > 0 {
			utils.WriteValidationErrorJSON(w, "Cannot delete region", map[string]string{
				"region": fmt.Sprintf("Region cannot be deleted because there are %d active users associated with it", userCount),
			})
			return
		}

		// Check if there are transactions with this region
		var transactionCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM transactions WHERE region = $1
		`, regionCode).Scan(&transactionCount)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if transactionCount > 0 {
			utils.WriteValidationErrorJSON(w, "Cannot delete region", map[string]string{
				"region": fmt.Sprintf("Region cannot be deleted because there are %d transactions associated with it", transactionCount),
			})
			return
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "DELETE FROM regions WHERE id = $1", regionID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'REGION', $2, 'Deleted region', NOW())
		`, adminID, regionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Region deleted successfully",
		})
	}
}

// ============================================
// LANGUAGE MANAGEMENT
// ============================================

// handleAdminGetLanguagesImpl returns all languages
func HandleAdminGetLanguagesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT id, code, name, country, image, is_default, is_active, sort_order, created_at
			FROM languages
			ORDER BY sort_order ASC
		`)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		languages := []map[string]interface{}{}
		for rows.Next() {
			var id, code, name, country string
			var image *string
			var isDefault, isActive bool
			var sortOrder int
			var createdAt time.Time

			if err := rows.Scan(&id, &code, &name, &country, &image, &isDefault, &isActive, &sortOrder, &createdAt); err != nil {
				fmt.Printf("Error scanning language row: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}

			lang := map[string]interface{}{
				"id":        id,
				"code":      code,
				"name":      name,
				"country":   country,
				"isDefault": isDefault,
				"isActive":  isActive,
				"sortOrder": sortOrder,
				"createdAt": createdAt.Format(time.RFC3339),
			}
			if image != nil {
				lang["image"] = *image
			}
			languages = append(languages, lang)
		}

		if err := rows.Err(); err != nil {
			fmt.Printf("Error iterating language rows: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, languages)
	}
}

// CreateLanguageRequest represents the request to create a language
type CreateLanguageRequest struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	Image     string `json:"image"`
	IsDefault bool   `json:"isDefault"`
	IsActive  bool   `json:"isActive"`
	Order     int    `json:"order"`
}

// handleCreateLanguageImpl creates a new language
func HandleCreateLanguageImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.WriteBadRequestError(w, fmt.Sprintf("Invalid form data: %v", err))
			return
		}

		// Get form values
		code := strings.ToLower(strings.TrimSpace(r.FormValue("code")))
		name := strings.TrimSpace(r.FormValue("name"))
		country := strings.TrimSpace(r.FormValue("country"))
		orderStr := strings.TrimSpace(r.FormValue("order"))
		isDefaultStr := r.FormValue("isDefault")
		isActiveStr := r.FormValue("isActive")

		// Validation
		if code == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Code is required"})
			return
		}
		if name == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"name": "Name is required"})
			return
		}
		if country == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"country": "Country is required"})
			return
		}

		order := 1
		if orderStr != "" {
			if parsedOrder, err := strconv.Atoi(orderStr); err == nil {
				order = parsedOrder
			}
		}
		isDefault := isDefaultStr == "true"
		isActive := isActiveStr != "false" // Default to true

		// Upload image to S3
		var imageURL string
		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err != nil && !errors.Is(err, http.ErrMissingFile) {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			}
			if err == nil {
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderFlag, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			}
		}

		if imageURL == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"image": "Image is required"})
			return
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// If setting as default, unset other defaults first
		if isDefault {
			_, err = tx.Exec(ctx, `UPDATE languages SET is_default = false WHERE is_default = true`)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		var languageID string
		err = tx.QueryRow(ctx, `
			INSERT INTO languages (code, name, country, image, is_default, is_active, sort_order, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			RETURNING id
		`, code, name, country, imageURL, isDefault, isActive, order).Scan(&languageID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'LANGUAGE', $2, 'Created language', NOW())
		`, adminID, languageID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      languageID,
			"message": "Language created successfully",
		})
	}
}

// handleUpdateLanguageImpl updates an existing language
func HandleUpdateLanguageImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		languageID := chi.URLParam(r, "languageId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.WriteBadRequestError(w, fmt.Sprintf("Invalid form data: %v", err))
			return
		}

		// Get existing language first to preserve unchanged fields
		var existingCode, existingName, existingCountry string
		var existingImage sql.NullString
		var existingIsDefault, existingIsActive bool
		var existingSortOrder int

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT code, name, country, image, is_default, is_active, sort_order
			FROM languages
			WHERE id = $1
		`, languageID).Scan(&existingCode, &existingName, &existingCountry, &existingImage, &existingIsDefault, &existingIsActive, &existingSortOrder)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "LANGUAGE_NOT_FOUND", "Language not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get form values (use existing values if not provided)
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			name = existingName
		}
		country := strings.TrimSpace(r.FormValue("country"))
		if country == "" {
			country = existingCountry
		}
		orderStr := strings.TrimSpace(r.FormValue("order"))
		isDefaultStr := r.FormValue("isDefault")
		isActiveStr := r.FormValue("isActive")

		sortOrder := existingSortOrder
		if orderStr != "" {
			if parsedOrder, err := strconv.Atoi(orderStr); err == nil {
				sortOrder = parsedOrder
			}
		}

		isDefault := existingIsDefault
		if isDefaultStr != "" {
			isDefault = isDefaultStr == "true"
		}

		isActive := existingIsActive
		if isActiveStr != "" {
			isActive = isActiveStr == "true"
		}

		// Handle image upload
		imageURL := ""
		if existingImage.Valid {
			imageURL = existingImage.String
		}

		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err == nil {
				// New image uploaded
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderFlag, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			} else if !errors.Is(err, http.ErrMissingFile) {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			} else {
				// Check if image URL provided as form value
				imageURLValue := strings.TrimSpace(r.FormValue("image"))
				if imageURLValue != "" {
					imageURL = imageURLValue
				}
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			}
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// If setting as default, unset other defaults first
		if isDefault && !existingIsDefault {
			_, err = tx.Exec(ctx, `UPDATE languages SET is_default = false WHERE is_default = true AND id != $1`, languageID)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		// Build update query dynamically
		updateFields := []string{"name = $1", "country = $2", "is_active = $3", "sort_order = $4", "updated_at = NOW()"}
		args := []interface{}{name, country, isActive, sortOrder}
		argPos := 5

		if imageURL != "" {
			updateFields = append(updateFields, fmt.Sprintf("image = $%d", argPos))
			args = append(args, imageURL)
			argPos++
		}

		if isDefault != existingIsDefault {
			updateFields = append(updateFields, fmt.Sprintf("is_default = $%d", argPos))
			args = append(args, isDefault)
			argPos++
		}

		args = append(args, languageID)

		query := fmt.Sprintf(`
			UPDATE languages SET
				%s
			WHERE id = $%d
		`, strings.Join(updateFields, ", "), argPos)

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'LANGUAGE', $2, 'Updated language', NOW())
		`, adminID, languageID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Language updated successfully",
		})
	}
}

// handleDeleteLanguageImpl deletes a language
func HandleDeleteLanguageImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		languageID := chi.URLParam(r, "languageId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Check if language is default
		var isDefault bool
		err := deps.DB.Pool.QueryRow(ctx, "SELECT is_default FROM languages WHERE id = $1", languageID).Scan(&isDefault)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "LANGUAGE_NOT_FOUND", "Language not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		if isDefault {
			utils.WriteValidationErrorJSON(w, "Cannot delete language", map[string]string{
				"language": "Cannot delete default language",
			})
			return
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "DELETE FROM languages WHERE id = $1", languageID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'LANGUAGE', $2, 'Deleted language', NOW())
		`, adminID, languageID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Language deleted successfully",
		})
	}
}

// ============================================
// CATEGORY & SECTION MANAGEMENT
// ============================================

// handleAdminGetCategoriesImpl returns all categories
func HandleAdminGetCategoriesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT c.id, c.code, c.title, c.description, c.icon, c.is_active, c.sort_order, c.created_at,
			       COALESCE(array_agg(cr.region_code) FILTER (WHERE cr.region_code IS NOT NULL), '{}') as regions
			FROM categories c
			LEFT JOIN category_regions cr ON c.id = cr.category_id
			GROUP BY c.id
			ORDER BY c.sort_order ASC
		`)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		categories := []map[string]interface{}{}
		for rows.Next() {
			var id, code, title, description, icon string
			var isActive bool
			var sortOrder int
			var regions []string
			var createdAt time.Time

			if err := rows.Scan(&id, &code, &title, &description, &icon, &isActive, &sortOrder, &createdAt, &regions); err != nil {
				fmt.Printf("Error scanning category row: %v\n", err)
				continue
			}

			categories = append(categories, map[string]interface{}{
				"id":          id,
				"code":        code,
				"title":       title,
				"description": description,
				"icon":        icon,
				"isActive":    isActive,
				"sortOrder":   sortOrder,
				"regions":     regions,
				"createdAt":   createdAt.Format(time.RFC3339),
			})
		}

		utils.WriteSuccessJSON(w, categories)
	}
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Code        string   `json:"code"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	IsActive    bool     `json:"isActive"`
	Order       int      `json:"order"`
	Regions     []string `json:"regions"`
}

// handleCreateCategoryImpl creates a new category
func HandleCreateCategoryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreateCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		var categoryID string
		err = tx.QueryRow(ctx, `
			INSERT INTO categories (code, title, description, icon, is_active, sort_order, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			RETURNING id
		`, req.Code, req.Title, req.Description, req.Icon, req.IsActive, req.Order).Scan(&categoryID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Insert regions into category_regions table
		for _, regionCode := range req.Regions {
			_, err = tx.Exec(ctx, `
				INSERT INTO category_regions (category_id, region_code)
				VALUES ($1, $2)
				ON CONFLICT (category_id, region_code) DO NOTHING
			`, categoryID, regionCode)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'CATEGORY', $2, 'Created category', NOW())
		`, adminID, categoryID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      categoryID,
			"message": "Category created successfully",
		})
	}
}

// handleUpdateCategoryImpl updates an existing category
func HandleUpdateCategoryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := chi.URLParam(r, "categoryId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		// Parse multipart form if present (for icon upload)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			// If not multipart, try JSON
			var req CreateCategoryRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				utils.WriteBadRequestError(w, "Invalid request body")
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()

			tx, err := deps.DB.Pool.Begin(ctx)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			defer tx.Rollback(ctx)

			// Always update icon if provided in request (emoji or URL)
			_, err = tx.Exec(ctx, `
				UPDATE categories SET
					title = $1, description = $2, icon = $3, is_active = $4, sort_order = $5, updated_at = NOW()
				WHERE id = $6
			`, req.Title, req.Description, req.Icon, req.IsActive, req.Order, categoryID)

			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Delete existing regions
			_, err = tx.Exec(ctx, `DELETE FROM category_regions WHERE category_id = $1`, categoryID)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Insert new regions
			for _, regionCode := range req.Regions {
				_, err = tx.Exec(ctx, `
					INSERT INTO category_regions (category_id, region_code)
					VALUES ($1, $2)
				`, categoryID, regionCode)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
			}

			tx.Exec(ctx, `
				INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
				VALUES ($1, 'UPDATE', 'CATEGORY', $2, 'Updated category', NOW())
			`, adminID, categoryID)

			if err := tx.Commit(ctx); err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			utils.WriteSuccessJSON(w, map[string]interface{}{
				"message": "Category updated successfully",
			})
			return
		}

		// Handle multipart form (with icon file upload)
		payloadStr := r.FormValue("payload")
		if payloadStr == "" {
			utils.WriteBadRequestError(w, "Payload is required")
			return
		}

		var req CreateCategoryRequest
		if err := json.Unmarshal([]byte(payloadStr), &req); err != nil {
			utils.WriteBadRequestError(w, "Invalid payload JSON")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Handle icon upload
		var iconURL string
		if file, header, err := r.FormFile("icon"); err == nil {
			if deps.S3 != nil {
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderCategory, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				iconURL = result.URL
			}
		} else if req.Icon != "" {
			// Use icon from payload if no file uploaded
			iconURL = req.Icon
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Update icon if provided
		if iconURL != "" {
			_, err = tx.Exec(ctx, `
				UPDATE categories SET
					title = $1, description = $2, icon = $3, is_active = $4, sort_order = $5, updated_at = NOW()
				WHERE id = $6
			`, req.Title, req.Description, iconURL, req.IsActive, req.Order, categoryID)
		} else {
			_, err = tx.Exec(ctx, `
				UPDATE categories SET
					title = $1, description = $2, is_active = $3, sort_order = $4, updated_at = NOW()
				WHERE id = $5
			`, req.Title, req.Description, req.IsActive, req.Order, categoryID)
		}

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Delete existing regions
		_, err = tx.Exec(ctx, `DELETE FROM category_regions WHERE category_id = $1`, categoryID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Insert new regions
		for _, regionCode := range req.Regions {
			_, err = tx.Exec(ctx, `
				INSERT INTO category_regions (category_id, region_code)
				VALUES ($1, $2)
			`, categoryID, regionCode)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'CATEGORY', $2, 'Updated category', NOW())
		`, adminID, categoryID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Category updated successfully",
		})
	}
}

// handleDeleteCategoryImpl deletes a category
func HandleDeleteCategoryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := chi.URLParam(r, "categoryId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "DELETE FROM categories WHERE id = $1", categoryID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'CATEGORY', $2, 'Deleted category', NOW())
		`, adminID, categoryID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Category deleted successfully",
		})
	}
}

// handleAdminGetSectionsImpl returns all sections
func HandleAdminGetSectionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT s.id, s.code, s.title, s.icon, s.is_active, s.sort_order, s.created_at,
			       COALESCE(array_agg(p.code) FILTER (WHERE p.code IS NOT NULL), '{}') as products
			FROM sections s
			LEFT JOIN product_sections ps ON s.id = ps.section_id
			LEFT JOIN products p ON ps.product_id = p.id
			GROUP BY s.id
			ORDER BY s.sort_order ASC
		`)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		sections := []map[string]interface{}{}
		for rows.Next() {
			var id, code, title, icon string
			var isActive bool
			var sortOrder int
			var products []string
			var createdAt time.Time

			if err := rows.Scan(&id, &code, &title, &icon, &isActive, &sortOrder, &createdAt, &products); err != nil {
				fmt.Printf("Error scanning section row: %v\n", err)
				continue
			}

			sections = append(sections, map[string]interface{}{
				"id":        id,
				"code":      code,
				"title":     title,
				"icon":      icon,
				"isActive":  isActive,
				"sortOrder": sortOrder,
				"products":  products,
				"createdAt": createdAt.Format(time.RFC3339),
			})
		}

		utils.WriteSuccessJSON(w, sections)
	}
}

// CreateSectionRequest represents the request to create a section
type CreateSectionRequest struct {
	Code     string   `json:"code"`
	Title    string   `json:"title"`
	Icon     string   `json:"icon"`
	IsActive bool     `json:"isActive"`
	Order    int      `json:"order"`
	Products []string `json:"products"`
}

// handleCreateSectionImpl creates a new section
func HandleCreateSectionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreateSectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		var sectionID string
		err = tx.QueryRow(ctx, `
			INSERT INTO sections (code, title, icon, is_active, sort_order, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			RETURNING id
		`, req.Code, req.Title, req.Icon, req.IsActive, req.Order).Scan(&sectionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Insert products into product_sections table
		for _, productCode := range req.Products {
			var productID string
			err = tx.QueryRow(ctx, `SELECT id FROM products WHERE code = $1`, productCode).Scan(&productID)
			if err != nil {
				continue // Skip if product not found
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO product_sections (product_id, section_id)
				VALUES ($1, $2)
				ON CONFLICT (product_id, section_id) DO NOTHING
			`, productID, sectionID)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'SECTION', $2, 'Created section', NOW())
		`, adminID, sectionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      sectionID,
			"message": "Section created successfully",
		})
	}
}

// handleUpdateSectionImpl updates an existing section
func HandleUpdateSectionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sectionID := chi.URLParam(r, "sectionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreateSectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `
			UPDATE sections SET
				title = $1, icon = $2, is_active = $3, sort_order = $4, updated_at = NOW()
			WHERE id = $5
		`, req.Title, req.Icon, req.IsActive, req.Order, sectionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Delete existing product associations
		_, err = tx.Exec(ctx, `DELETE FROM product_sections WHERE section_id = $1`, sectionID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Insert new product associations
		for _, productCode := range req.Products {
			var productID string
			err = tx.QueryRow(ctx, `SELECT id FROM products WHERE code = $1`, productCode).Scan(&productID)
			if err != nil {
				continue // Skip if product not found
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO product_sections (product_id, section_id)
				VALUES ($1, $2)
			`, productID, sectionID)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'SECTION', $2, 'Updated section', NOW())
		`, adminID, sectionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Section updated successfully",
		})
	}
}

// handleDeleteSectionImpl deletes a section
func HandleDeleteSectionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sectionID := chi.URLParam(r, "sectionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "DELETE FROM sections WHERE id = $1", sectionID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'SECTION', $2, 'Deleted section', NOW())
		`, adminID, sectionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Section deleted successfully",
		})
	}
}

// AssignSectionProductsRequest represents the request to assign products to a section
type AssignSectionProductsRequest struct {
	Products []string `json:"products"`
}

// handleAssignSectionProductsImpl assigns products to a section
func HandleAssignSectionProductsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sectionID := chi.URLParam(r, "sectionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req AssignSectionProductsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "UPDATE sections SET products = $1 WHERE id = $2", req.Products, sectionID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'SECTION', $2, 'Assigned products to section', NOW())
		`, adminID, sectionID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Products assigned successfully",
		})
	}
}

// ============================================
// PAYMENT CHANNEL MANAGEMENT (ADMIN)
// ============================================

// handleAdminGetPaymentChannelsImpl returns all payment channels (admin view)
func HandleAdminGetPaymentChannelsImplAdmin(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get query parameters
		categoryCode := r.URL.Query().Get("categoryCode")
		region := r.URL.Query().Get("region")
		isActiveStr := r.URL.Query().Get("isActive")

		// Build query with filters
		query := `
			SELECT
				pc.id, pc.code, pc.name, pc.description, pc.image, pc.instruction,
				pc.fee_type, pc.fee_amount, pc.fee_percentage,
				pc.min_amount, pc.max_amount,
				pc.supported_types,
				pc.is_active, pc.is_featured, pc.sort_order,
				pc.created_at, pc.updated_at,
				pcc.code as category_code, pcc.title as category_title
			FROM payment_channels pc
			LEFT JOIN payment_channel_categories pcc ON pc.category_id = pcc.id
			WHERE 1=1
		`
		args := []interface{}{}
		argCount := 0

		if categoryCode != "" {
			argCount++
			query += fmt.Sprintf(" AND pcc.code = $%d", argCount)
			args = append(args, categoryCode)
		}

		if region != "" {
			argCount++
			query += fmt.Sprintf(`
				AND EXISTS (
					SELECT 1 FROM payment_channel_regions pcr
					WHERE pcr.channel_id = pc.id AND pcr.region_code = $%d
				)
			`, argCount)
			args = append(args, region)
		}

		if isActiveStr != "" {
			argCount++
			isActive, err := strconv.ParseBool(isActiveStr)
			if err == nil {
				query += fmt.Sprintf(" AND pc.is_active = $%d", argCount)
				args = append(args, isActive)
			}
		}

		query += " ORDER BY pc.sort_order ASC"

		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			fmt.Printf("Error querying payment channels: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		channels := []map[string]interface{}{}

		for rows.Next() {
			var id, code, name string
			var description, image, instruction sql.NullString
			var feeType string
			var feeAmount int64
			var feePercentage float64
			var minAmount, maxAmount int64
			var supportedTypes []string
			var isActive, isFeatured bool
			var sortOrder int
			var createdAt, updatedAt time.Time
			var categoryCode, categoryTitle sql.NullString

			if err := rows.Scan(
				&id, &code, &name, &description, &image, &instruction,
				&feeType, &feeAmount, &feePercentage,
				&minAmount, &maxAmount,
				&supportedTypes,
				&isActive, &isFeatured, &sortOrder,
				&createdAt, &updatedAt,
				&categoryCode, &categoryTitle,
			); err != nil {
				fmt.Printf("Error scanning payment channel row: %v\n", err)
				continue
			}

			categoryMap := map[string]interface{}{}
			if categoryCode.Valid {
				categoryMap["code"] = categoryCode.String
			}
			if categoryTitle.Valid {
				categoryMap["title"] = categoryTitle.String
			}

			descValue := ""
			if description.Valid {
				descValue = description.String
			}
			imageValue := ""
			if image.Valid {
				imageValue = image.String
			}
			instructionValue := ""
			if instruction.Valid {
				instructionValue = instruction.String
			}

			// Get regions for this channel
			regionRows, err := deps.DB.Pool.Query(ctx, `
				SELECT region_code FROM payment_channel_regions WHERE channel_id = $1
			`, id)
			regions := []string{}
			if err == nil {
				defer regionRows.Close()
				for regionRows.Next() {
					var regionCode string
					if err := regionRows.Scan(&regionCode); err == nil {
						regions = append(regions, regionCode)
					}
				}
			}

			// Get stats (today transactions and volume)
			var todayTransactions, todayVolume int64
			deps.DB.Pool.QueryRow(ctx, `
				SELECT
					COUNT(*) as today_transactions,
					COALESCE(SUM(total_amount), 0) as today_volume
				FROM orders
				WHERE payment_channel_id = $1
					AND DATE(created_at) = CURRENT_DATE
			`, id).Scan(&todayTransactions, &todayVolume)

			channels = append(channels, map[string]interface{}{
				"id":          id,
				"code":        code,
				"name":        name,
				"description": descValue,
				"image":       imageValue,
				"category":    categoryMap,
				"fee": map[string]interface{}{
					"feeType":       feeType,
					"feeAmount":     feeAmount,
					"feePercentage": feePercentage,
				},
				"limits": map[string]interface{}{
					"minAmount": minAmount,
					"maxAmount": maxAmount,
				},
				"regions":        regions,
				"supportedTypes": supportedTypes,
				"isActive":       isActive,
				"isFeatured":     isFeatured,
				"order":          sortOrder,
				"instruction":    instructionValue,
				"stats": map[string]interface{}{
					"todayTransactions": todayTransactions,
					"todayVolume":       todayVolume,
				},
				"createdAt": createdAt.Format(time.RFC3339),
				"updatedAt": updatedAt.Format(time.RFC3339),
			})
		}

		if err := rows.Err(); err != nil {
			fmt.Printf("Error iterating payment channel rows: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, channels)
	}
}

// handleGetChannelAssignmentsImpl returns payment channel gateway assignments
func HandleGetChannelAssignmentsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data - need payment_channel_gateway_assignment table
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"assignments": []map[string]interface{}{},
		})
	}
}

// handleGetPaymentChannelImpl returns detailed payment channel info
func HandleGetPaymentChannelImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelID := chi.URLParam(r, "channelId")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var id, code, name, description, image, instruction string
		var feeAmount, feePercentage int
		var minAmount, maxAmount int64
		var regions, supportedTypes []string
		var isActive, isFeatured bool
		var sortOrder int

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, code, name, description, image, instruction,
				fee_amount, fee_percentage, min_amount, max_amount,
				regions, supported_types, is_active, is_featured, sort_order
			FROM payment_channels
			WHERE id = $1
		`, channelID).Scan(
			&id, &code, &name, &description, &image, &instruction,
			&feeAmount, &feePercentage, &minAmount, &maxAmount,
			&regions, &supportedTypes, &isActive, &isFeatured, &sortOrder,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_CHANNEL_NOT_FOUND",
					"Payment channel not found", "")
				return
			}
			fmt.Printf("Error scanning payment channel: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"id":          id,
			"code":        code,
			"name":        name,
			"description": description,
			"image":       image,
			"instruction": instruction,
			"fee": map[string]interface{}{
				"feeAmount":     feeAmount,
				"feePercentage": feePercentage,
			},
			"limits": map[string]interface{}{
				"minAmount": minAmount,
				"maxAmount": maxAmount,
			},
			"regions":        regions,
			"supportedTypes": supportedTypes,
			"isActive":       isActive,
			"isFeatured":     isFeatured,
			"sortOrder":      sortOrder,
		})
	}
}

// CreatePaymentChannelRequest represents the request to create a payment channel
type CreatePaymentChannelRequest struct {
	Code           string   `json:"code"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Image          string   `json:"image"`
	CategoryCode   string   `json:"categoryCode"`
	GatewayCode    string   `json:"gatewayCode"` // Gateway-specific code (e.g., "002" for BRI)
	Gateway        string   `json:"gateway"`     // Gateway to use (e.g., DANA_DIRECT, XENDIT, MIDTRANS)
	FeeType        string   `json:"feeType"`     // FIXED, PERCENTAGE, MIXED
	FeeAmount      int64    `json:"feeAmount"`
	FeePercentage  float64  `json:"feePercentage"`
	MinAmount      int64    `json:"minAmount"`
	MaxAmount      int64    `json:"maxAmount"`
	Regions        []string `json:"regions"`
	SupportedTypes []string `json:"supportedTypes"`
	IsActive       bool     `json:"isActive"`
	IsFeatured     bool     `json:"isFeatured"`
	Order          int      `json:"order"`
	Instruction    string   `json:"instruction"`
}

// handleCreatePaymentChannelImpl creates a new payment channel
func HandleCreatePaymentChannelImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
			fmt.Printf("Error parsing multipart form: %v\n", err)
			utils.WriteBadRequestError(w, "Invalid form data")
			return
		}

		// Get form values
		code := strings.TrimSpace(r.FormValue("code"))
		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		categoryCode := strings.TrimSpace(r.FormValue("categoryCode"))
		feeType := strings.TrimSpace(r.FormValue("feeType"))
		if feeType == "" {
			feeType = "PERCENTAGE"
		}
		feeAmountStr := strings.TrimSpace(r.FormValue("feeAmount"))
		feePercentageStr := strings.TrimSpace(r.FormValue("feePercentage"))
		minAmountStr := strings.TrimSpace(r.FormValue("minAmount"))
		maxAmountStr := strings.TrimSpace(r.FormValue("maxAmount"))
		orderStr := strings.TrimSpace(r.FormValue("order"))
		instruction := strings.TrimSpace(r.FormValue("instruction"))
		isActiveStr := r.FormValue("isActive")
		isFeaturedStr := r.FormValue("isFeatured")

		// Parse arrays
		regions := []string{}
		if regionsStr := r.FormValue("regions"); regionsStr != "" {
			if err := json.Unmarshal([]byte(regionsStr), &regions); err != nil {
				regions = strings.Split(regionsStr, ",")
			}
		}
		supportedTypes := []string{}
		if supportedTypesStr := r.FormValue("supportedTypes"); supportedTypesStr != "" {
			if err := json.Unmarshal([]byte(supportedTypesStr), &supportedTypes); err != nil {
				supportedTypes = strings.Split(supportedTypesStr, ",")
			}
		}

		// Validation
		if code == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Code is required"})
			return
		}
		if name == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"name": "Name is required"})
			return
		}
		if feeType != "FIXED" && feeType != "PERCENTAGE" && feeType != "MIXED" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"feeType": "Fee type must be FIXED, PERCENTAGE, or MIXED"})
			return
		}

		// Parse numeric values
		feeAmount := int64(0)
		if feeAmountStr != "" {
			if parsed, err := strconv.ParseInt(feeAmountStr, 10, 64); err == nil {
				feeAmount = parsed
			}
		}
		feePercentage := float64(0)
		if feePercentageStr != "" {
			if parsed, err := strconv.ParseFloat(feePercentageStr, 64); err == nil {
				feePercentage = parsed
			}
		}
		minAmount := int64(0)
		if minAmountStr != "" {
			if parsed, err := strconv.ParseInt(minAmountStr, 10, 64); err == nil {
				minAmount = parsed
			}
		}
		maxAmount := int64(0)
		if maxAmountStr != "" {
			if parsed, err := strconv.ParseInt(maxAmountStr, 10, 64); err == nil {
				maxAmount = parsed
			}
		}
		order := 0
		if orderStr != "" {
			if parsed, err := strconv.Atoi(orderStr); err == nil {
				order = parsed
			}
		}
		isActive := isActiveStr == "true"
		isFeatured := isFeaturedStr == "true"

		// Check if code already exists
		var existingID string
		err := deps.DB.Pool.QueryRow(ctx, "SELECT id FROM payment_channels WHERE code = $1", code).Scan(&existingID)
		if err == nil {
			utils.WriteErrorJSON(w, http.StatusConflict, "CODE_ALREADY_EXISTS",
				"Payment channel code already exists", "")
			return
		}
		if err != nil && err != pgx.ErrNoRows {
			fmt.Printf("Error checking code: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Get category ID (optional)
		var categoryID sql.NullString
		if categoryCode != "" {
			err = deps.DB.Pool.QueryRow(ctx, "SELECT id FROM payment_channel_categories WHERE code = $1", categoryCode).Scan(&categoryID)
			if err != nil {
				if err == pgx.ErrNoRows {
					utils.WriteErrorJSON(w, http.StatusNotFound, "CATEGORY_NOT_FOUND",
						"Payment channel category not found", "")
					return
				}
				fmt.Printf("Error finding category: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}
		}

		// Handle image upload
		var imageURL string
		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err != nil && !errors.Is(err, http.ErrMissingFile) {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			}
			if err == nil {
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderPayment, file, header)
				if err != nil {
					fmt.Printf("Error uploading image to S3: %v\n", err)
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			} else {
				// Check if image URL provided as form value
				imageURLValue := strings.TrimSpace(r.FormValue("image"))
				if imageURLValue != "" {
					imageURL = imageURLValue
				}
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			}
		}

		if imageURL == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"image": "Image file or URL is required"})
			return
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			fmt.Printf("Error beginning transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		var channelID string
		if categoryID.Valid {
			err = tx.QueryRow(ctx, `
				INSERT INTO payment_channels (
					code, name, description, image, category_id,
					fee_type, fee_amount, fee_percentage, min_amount, max_amount,
					supported_types, is_active, is_featured, sort_order, instruction, created_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
				RETURNING id
			`, code, name, description, imageURL, categoryID.String,
				feeType, feeAmount, feePercentage, minAmount, maxAmount,
				supportedTypes, isActive, isFeatured, order, instruction,
			).Scan(&channelID)
		} else {
			// Insert without category_id (NULL)
			err = tx.QueryRow(ctx, `
				INSERT INTO payment_channels (
					code, name, description, image, category_id,
					fee_type, fee_amount, fee_percentage, min_amount, max_amount,
					supported_types, is_active, is_featured, sort_order, instruction, created_at
				) VALUES ($1, $2, $3, $4, NULL, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
				RETURNING id
			`, code, name, description, imageURL,
				feeType, feeAmount, feePercentage, minAmount, maxAmount,
				supportedTypes, isActive, isFeatured, order, instruction,
			).Scan(&channelID)
		}

		if err != nil {
			fmt.Printf("Error inserting payment channel: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Insert regions
		for _, region := range regions {
			_, err = tx.Exec(ctx, "INSERT INTO payment_channel_regions (channel_id, region_code) VALUES ($1, $2) ON CONFLICT DO NOTHING", channelID, region)
			if err != nil {
				fmt.Printf("Error inserting region: %v\n", err)
			}
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'PAYMENT_CHANNEL', $2, 'Created payment channel', NOW())
		`, adminID, channelID)

		if err := tx.Commit(ctx); err != nil {
			fmt.Printf("Error committing transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      channelID,
			"code":    code,
			"message": "Payment channel created successfully",
		})
	}
}

// UpdatePaymentChannelRequest represents the request to update a payment channel
type UpdatePaymentChannelRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	Image          *string  `json:"image"`
	CategoryCode   *string  `json:"categoryCode"`
	FeeType        *string  `json:"feeType"`     // FIXED, PERCENTAGE, MIXED
	FeeAmount      *int64   `json:"feeAmount"`
	FeePercentage  *float64 `json:"feePercentage"`
	MinAmount      *int64   `json:"minAmount"`
	MaxAmount      *int64   `json:"maxAmount"`
	Regions        []string `json:"regions"`
	SupportedTypes []string `json:"supportedTypes"`
	IsActive       *bool    `json:"isActive"`
	IsFeatured     *bool    `json:"isFeatured"`
	Order          *int     `json:"order"`
	Instruction    *string  `json:"instruction"`
}

// handleUpdatePaymentChannelImpl updates an existing payment channel
func HandleUpdatePaymentChannelImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelID := chi.URLParam(r, "channelId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Check if channel exists and get existing image
		var existingCode string
		var existingImage sql.NullString
		err := deps.DB.Pool.QueryRow(ctx, "SELECT code, image FROM payment_channels WHERE id = $1", channelID).Scan(&existingCode, &existingImage)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_CHANNEL_NOT_FOUND",
					"Payment channel not found", "")
				return
			}
			fmt.Printf("Error checking payment channel: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
			fmt.Printf("Error parsing multipart form: %v\n", err)
			utils.WriteBadRequestError(w, "Invalid form data")
			return
		}

		// Get form values
		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		feeType := strings.TrimSpace(r.FormValue("feeType"))
		feeAmountStr := strings.TrimSpace(r.FormValue("feeAmount"))
		feePercentageStr := strings.TrimSpace(r.FormValue("feePercentage"))
		minAmountStr := strings.TrimSpace(r.FormValue("minAmount"))
		maxAmountStr := strings.TrimSpace(r.FormValue("maxAmount"))
		orderStr := strings.TrimSpace(r.FormValue("order"))
		instruction := strings.TrimSpace(r.FormValue("instruction"))
		isActiveStr := r.FormValue("isActive")
		isFeaturedStr := r.FormValue("isFeatured")

		// Parse arrays
		regions := []string{}
		if regionsStr := r.FormValue("regions"); regionsStr != "" {
			if err := json.Unmarshal([]byte(regionsStr), &regions); err != nil {
				regions = strings.Split(regionsStr, ",")
			}
		}
		supportedTypes := []string{}
		if supportedTypesStr := r.FormValue("supportedTypes"); supportedTypesStr != "" {
			if err := json.Unmarshal([]byte(supportedTypesStr), &supportedTypes); err != nil {
				supportedTypes = strings.Split(supportedTypesStr, ",")
			}
		}

		// Handle image upload
		var imageURL string
		if existingImage.Valid {
			imageURL = existingImage.String
		}

		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err == nil {
				// New image uploaded - delete old image first
				if existingImage.Valid && existingImage.String != "" {
					deleteS3Object(ctx, deps, existingImage.String)
				}
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderPayment, file, header)
				if err != nil {
					fmt.Printf("Error uploading image to S3: %v\n", err)
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			} else if !errors.Is(err, http.ErrMissingFile) {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			} else {
				// Check if image URL provided as form value
				imageURLValue := strings.TrimSpace(r.FormValue("image"))
				if imageURLValue != "" {
					// If URL changed and old image exists, delete old image
					if existingImage.Valid && existingImage.String != "" && imageURLValue != existingImage.String {
						deleteS3Object(ctx, deps, existingImage.String)
					}
					imageURL = imageURLValue
				}
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			}
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			fmt.Printf("Error beginning transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Build update query dynamically
		updateFields := []string{}
		args := []interface{}{}
		argCount := 0

		if name != "" {
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("name = $%d", argCount))
			args = append(args, name)
		}
		if description != "" || r.FormValue("description") != "" {
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("description = $%d", argCount))
			args = append(args, description)
		}
		if imageURL != "" {
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("image = $%d", argCount))
			args = append(args, imageURL)
		}
		// Handle category update - always check categoryCode in form
		categoryCodeFormValue := strings.TrimSpace(r.FormValue("categoryCode"))
		if categoryCodeFormValue != "" {
			// Category code provided - update to this category
			var categoryID string
			err := tx.QueryRow(ctx, "SELECT id FROM payment_channel_categories WHERE code = $1", categoryCodeFormValue).Scan(&categoryID)
			if err != nil {
				if err == pgx.ErrNoRows {
					utils.WriteErrorJSON(w, http.StatusNotFound, "CATEGORY_NOT_FOUND",
						"Payment channel category not found", "")
					return
				}
				fmt.Printf("Error finding category: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("category_id = $%d", argCount))
			args = append(args, categoryID)
		} else {
			// categoryCode is empty string or not provided - set to NULL
			updateFields = append(updateFields, "category_id = NULL")
		}
		if feeType != "" {
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("fee_type = $%d", argCount))
			args = append(args, feeType)
		}
		if feeAmountStr != "" {
			if parsed, err := strconv.ParseInt(feeAmountStr, 10, 64); err == nil {
				argCount++
				updateFields = append(updateFields, fmt.Sprintf("fee_amount = $%d", argCount))
				args = append(args, parsed)
			}
		}
		if feePercentageStr != "" {
			if parsed, err := strconv.ParseFloat(feePercentageStr, 64); err == nil {
				// Validate feePercentage (must be between 0 and 100)
				if parsed < 0 || parsed > 100 {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
						"feePercentage": "Fee percentage must be between 0 and 100",
					})
					return
				}
				argCount++
				updateFields = append(updateFields, fmt.Sprintf("fee_percentage = $%d", argCount))
				args = append(args, parsed)
			}
		}
		if minAmountStr != "" {
			if parsed, err := strconv.ParseInt(minAmountStr, 10, 64); err == nil {
				argCount++
				updateFields = append(updateFields, fmt.Sprintf("min_amount = $%d", argCount))
				args = append(args, parsed)
			}
		}
		if maxAmountStr != "" {
			if parsed, err := strconv.ParseInt(maxAmountStr, 10, 64); err == nil {
				argCount++
				updateFields = append(updateFields, fmt.Sprintf("max_amount = $%d", argCount))
				args = append(args, parsed)
			}
		}
		if len(supportedTypes) > 0 {
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("supported_types = $%d", argCount))
			args = append(args, supportedTypes)
		}
		if isActiveStr != "" {
			isActive := isActiveStr == "true"
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("is_active = $%d", argCount))
			args = append(args, isActive)
		}
		if isFeaturedStr != "" {
			isFeatured := isFeaturedStr == "true"
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("is_featured = $%d", argCount))
			args = append(args, isFeatured)
		}
		if orderStr != "" {
			if parsed, err := strconv.Atoi(orderStr); err == nil {
				argCount++
				updateFields = append(updateFields, fmt.Sprintf("sort_order = $%d", argCount))
				args = append(args, parsed)
			}
		}
		if instruction != "" || r.FormValue("instruction") != "" {
			argCount++
			updateFields = append(updateFields, fmt.Sprintf("instruction = $%d", argCount))
			args = append(args, instruction)
		}

		// Always update updated_at (no parameter needed)
		updateFields = append(updateFields, "updated_at = NOW()")

		if len(updateFields) > 0 {
			argCount++
			args = append(args, channelID)
			query := fmt.Sprintf("UPDATE payment_channels SET %s WHERE id = $%d", strings.Join(updateFields, ", "), argCount)
			fmt.Printf("[UPDATE PAYMENT CHANNEL] Query: %s\n", query)
			fmt.Printf("[UPDATE PAYMENT CHANNEL] Args: %v\n", args)
			_, err = tx.Exec(ctx, query, args...)
			if err != nil {
				fmt.Printf("Error updating payment channel: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}
		} else {
			fmt.Printf("[UPDATE PAYMENT CHANNEL] No fields to update\n")
		}

		// Update regions if provided
		if len(regions) > 0 || r.FormValue("regions") != "" {
			// Delete existing regions
			_, err = tx.Exec(ctx, "DELETE FROM payment_channel_regions WHERE channel_id = $1", channelID)
			if err != nil {
				fmt.Printf("Error deleting regions: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}
			// Insert new regions
			for _, region := range regions {
				_, err = tx.Exec(ctx, "INSERT INTO payment_channel_regions (channel_id, region_code) VALUES ($1, $2) ON CONFLICT DO NOTHING", channelID, region)
				if err != nil {
					fmt.Printf("Error inserting region: %v\n", err)
				}
			}
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'PAYMENT_CHANNEL', $2, 'Updated payment channel', NOW())
		`, adminID, channelID)

		if err := tx.Commit(ctx); err != nil {
			fmt.Printf("Error committing transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Payment channel updated successfully",
		})
	}
}

// UpdateChannelAssignmentRequest represents the request to update channel gateway assignment
type UpdateChannelAssignmentRequest struct {
	PurchaseGateway string `json:"purchaseGateway"`
	DepositGateway  string `json:"depositGateway"`
}

// handleUpdateChannelAssignmentImpl updates gateway assignment for a payment channel
func HandleUpdateChannelAssignmentImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paymentCode := chi.URLParam(r, "paymentCode")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req UpdateChannelAssignmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Create audit log
		deps.DB.Pool.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'PAYMENT_CHANNEL', $2, 'Updated gateway assignment', NOW())
		`, adminID, paymentCode)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Gateway assignment updated successfully",
		})
	}
}

// handleDeletePaymentChannelImpl deletes a payment channel
func HandleDeletePaymentChannelImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelID := chi.URLParam(r, "channelId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Check if channel exists and get image URL
		var channelCode string
		var existingImage sql.NullString
		err := deps.DB.Pool.QueryRow(ctx, "SELECT code, image FROM payment_channels WHERE id = $1", channelID).Scan(&channelCode, &existingImage)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_CHANNEL_NOT_FOUND",
					"Payment channel not found", "")
				return
			}
			fmt.Printf("Error checking payment channel: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Check for pending transactions
		var pendingCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM orders 
			WHERE payment_channel_id = $1 
			AND status IN ('PENDING', 'PROCESSING')
		`, channelID).Scan(&pendingCount)
		if err != nil {
			fmt.Printf("Error checking pending transactions: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		if pendingCount > 0 {
			utils.WriteErrorJSON(w, http.StatusConflict, "CANNOT_DELETE_CHANNEL",
				fmt.Sprintf("Cannot delete payment channel with %d pending transactions", pendingCount),
				"Please wait for all pending transactions to complete or cancel them first")
			return
		}

		// Check for pending deposits
		var pendingDeposits int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM deposits 
			WHERE payment_channel_id = $1 
			AND status IN ('PENDING', 'PROCESSING')
		`, channelID).Scan(&pendingDeposits)
		if err != nil {
			fmt.Printf("Error checking pending deposits: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		if pendingDeposits > 0 {
			utils.WriteErrorJSON(w, http.StatusConflict, "CANNOT_DELETE_CHANNEL",
				fmt.Sprintf("Cannot delete payment channel with %d pending deposits", pendingDeposits),
				"Please wait for all pending deposits to complete or cancel them first")
			return
		}

		// Delete image from S3 if exists
		if existingImage.Valid && existingImage.String != "" {
			deleteS3Object(ctx, deps, existingImage.String)
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			fmt.Printf("Error beginning transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Delete will cascade to:
		// - payment_channel_regions (ON DELETE CASCADE)
		// - payment_channel_gateways (ON DELETE CASCADE)
		// - promo_payment_channels (ON DELETE CASCADE)
		_, err = tx.Exec(ctx, "DELETE FROM payment_channels WHERE id = $1", channelID)
		if err != nil {
			fmt.Printf("Error deleting payment channel: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'PAYMENT_CHANNEL', $2, 'Deleted payment channel', NOW())
		`, adminID, channelID)

		if err := tx.Commit(ctx); err != nil {
			fmt.Printf("Error committing transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Payment channel deleted successfully",
		})
	}
}

// handleGetPaymentChannelCategoriesImpl returns all payment channel categories
func HandleGetPaymentChannelCategoriesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT 
				pcc.id, pcc.code, pcc.title, pcc.icon, 
				pcc.is_active, pcc.sort_order, 
				pcc.created_at, pcc.updated_at,
				COUNT(pc.id) as channel_count
			FROM payment_channel_categories pcc
			LEFT JOIN payment_channels pc ON pc.category_id = pcc.id
			GROUP BY pcc.id, pcc.code, pcc.title, pcc.icon, pcc.is_active, pcc.sort_order, pcc.created_at, pcc.updated_at
			ORDER BY pcc.sort_order ASC
		`)

		if err != nil {
			fmt.Printf("Error querying payment channel categories: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		categories := []map[string]interface{}{}
		for rows.Next() {
			var id, code, title string
			var icon sql.NullString
			var isActive bool
			var sortOrder int
			var createdAt, updatedAt time.Time
			var channelCount int

			if err := rows.Scan(&id, &code, &title, &icon, &isActive, &sortOrder, &createdAt, &updatedAt, &channelCount); err != nil {
				fmt.Printf("Error scanning payment channel category row: %v\n", err)
				continue
			}

			iconValue := ""
			if icon.Valid {
				iconValue = icon.String
			}

			categories = append(categories, map[string]interface{}{
				"id":           id,
				"code":         code,
				"title":        title,
				"icon":         iconValue,
				"isActive":     isActive,
				"order":        sortOrder,
				"channelCount": channelCount,
				"createdAt":    createdAt.Format(time.RFC3339),
				"updatedAt":    updatedAt.Format(time.RFC3339),
			})
		}

		if err := rows.Err(); err != nil {
			fmt.Printf("Error iterating payment channel category rows: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, categories)
	}
}

// CreatePaymentChannelCategoryRequest represents the request to create a payment channel category
type CreatePaymentChannelCategoryRequest struct {
	Code     string `json:"code"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	IsActive bool   `json:"isActive"`
	Order    int    `json:"order"`
}

// handleCreatePaymentChannelCategoryImpl creates a new payment channel category
func HandleCreatePaymentChannelCategoryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreatePaymentChannelCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		var categoryID string
		err = tx.QueryRow(ctx, `
			INSERT INTO payment_channel_categories (code, title, icon, is_active, sort_order, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			RETURNING id
		`, req.Code, req.Title, req.Icon, req.IsActive, req.Order).Scan(&categoryID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'PAYMENT_CHANNEL_CATEGORY', $2, 'Created payment channel category', NOW())
		`, adminID, categoryID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      categoryID,
			"message": "Payment channel category created successfully",
		})
	}
}

// handleUpdatePaymentChannelCategoryImpl updates a payment channel category
func HandleUpdatePaymentChannelCategoryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := chi.URLParam(r, "categoryId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreatePaymentChannelCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `
			UPDATE payment_channel_categories SET
				title = $1, icon = $2, is_active = $3, sort_order = $4, updated_at = NOW()
			WHERE id = $5
		`, req.Title, req.Icon, req.IsActive, req.Order, categoryID)

		if err != nil {
			fmt.Printf("Error updating payment channel category: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'PAYMENT_CHANNEL_CATEGORY', $2, 'Updated payment channel category', NOW())
		`, adminID, categoryID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Payment channel category updated successfully",
		})
	}
}

// handleDeletePaymentChannelCategoryImpl deletes a payment channel category
func HandleDeletePaymentChannelCategoryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := chi.URLParam(r, "categoryId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "DELETE FROM payment_channel_categories WHERE id = $1", categoryID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'PAYMENT_CHANNEL_CATEGORY', $2, 'Deleted payment channel category', NOW())
		`, adminID, categoryID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Payment channel category deleted successfully",
		})
	}
}

// ============================================
// REPORTS & AUDIT
// ============================================

// handleGetRevenueReportImpl returns revenue report
func HandleGetRevenueReportImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"revenue": []map[string]interface{}{},
		})
	}
}

// handleGetTransactionReportImpl returns transaction report
func HandleGetTransactionReportImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"transactions": []map[string]interface{}{},
		})
	}
}

// handleGetProductReportImpl returns product report
func HandleGetProductReportImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"products": []map[string]interface{}{},
		})
	}
}

// handleGetProviderReportImpl returns provider report
func HandleGetProviderReportImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock data
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"providers": []map[string]interface{}{},
		})
	}
}

// ExportReportRequest represents the request to export a report
type ExportReportRequest struct {
	ReportType string                 `json:"reportType"`
	Format     string                 `json:"format"`
	Filters    map[string]interface{} `json:"filters"`
}

// handleExportReportImpl initiates report export
func HandleExportReportImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ExportReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Generate export ID
		randomStr, err := utils.GenerateRandomString(10)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		exportID := "exp_" + randomStr

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"exportId":    exportID,
			"status":      "PROCESSING",
			"downloadUrl": nil,
			"expiresAt":   nil,
		})
	}
}

// handleGetExportStatusImpl returns export status
func HandleGetExportStatusImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exportID := chi.URLParam(r, "exportId")

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"exportID":    exportID,
			"status":      "COMPLETED",
			"downloadUrl": "https://nos.jkt-1.neo.id/gate/exports/report_123.xlsx",
			"expiresAt":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		})
	}
}

// handleGetAuditLogsImpl returns audit logs
func HandleGetAuditLogsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT
				al.id, al.action, al.resource, al.resource_id, al.description,
				a.id, a.first_name || ' ' || a.last_name, a.email,
				al.created_at
			FROM audit_logs al
			JOIN admins a ON al.admin_id = a.id
			ORDER BY al.created_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		logs := []map[string]interface{}{}
		for rows.Next() {
			var id, action, resource, resourceID, description string
			var adminID, adminName, adminEmail string
			var createdAt time.Time

			rows.Scan(&id, &action, &resource, &resourceID, &description, &adminID, &adminName, &adminEmail, &createdAt)

			logs = append(logs, map[string]interface{}{
				"id":          id,
				"action":      action,
				"resource":    resource,
				"resourceId":  resourceID,
				"description": description,
				"admin": map[string]interface{}{
					"id":    adminID,
					"name":  adminName,
					"email": adminEmail,
				},
				"createdAt": createdAt.Format(time.RFC3339),
			})
		}

		var totalRows int
		deps.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"logs": logs,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

// handleAdminGetInvoicesImpl returns all invoices (alias for transactions)
func HandleAdminGetInvoicesImpl(deps *Dependencies) http.HandlerFunc {
	return HandleAdminGetTransactionsImpl(deps)
}

// handleSearchInvoiceImpl searches for invoices
func HandleSearchInvoiceImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			utils.WriteBadRequestError(w, "Search query is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT
				t.invoice_number, t.status,
				u.first_name || ' ' || u.last_name, u.email,
				p.title, s.name,
				t.total_amount, t.currency, t.created_at
			FROM transactions t
			LEFT JOIN users u ON t.user_id = u.id
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			WHERE t.invoice_number ILIKE $1
				OR u.email ILIKE $1
				OR u.phone_number ILIKE $1
			LIMIT 10
		`, "%"+query+"%")

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		results := []map[string]interface{}{}
		for rows.Next() {
			var invoiceNumber, status, userName, userEmail, product, sku, currency string
			var total int64
			var createdAt time.Time

			rows.Scan(&invoiceNumber, &status, &userName, &userEmail, &product, &sku, &total, &currency, &createdAt)

			results = append(results, map[string]interface{}{
				"invoiceNumber": invoiceNumber,
				"type":          "PURCHASE",
				"status":        status,
				"user": map[string]interface{}{
					"name":  userName,
					"email": userEmail,
				},
				"product":   product + " - " + sku,
				"total":     total,
				"currency":  currency,
				"createdAt": createdAt.Format(time.RFC3339),
			})
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"results": results,
		})
	}
}

// SendInvoiceEmailRequest represents the request to send invoice email
type SendInvoiceEmailRequest struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

// handleSendInvoiceEmailImpl sends invoice email to customer
func HandleSendInvoiceEmailImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invoiceNumber := chi.URLParam(r, "invoiceNumber")

		var req SendInvoiceEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// TODO: Implement email sending logic with invoiceNumber

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":       "Invoice email sent successfully",
			"invoiceNumber": invoiceNumber,
			"sentTo":        req.Email,
			"sentAt":        time.Now().Format(time.RFC3339),
		})
	}
}

// handleAdminVerifyMFAImpl verifies admin MFA code
func HandleAdminVerifyMFAImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock implementation
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"step": "SUCCESS",
		})
	}
}

// handleAdminRefreshTokenImpl refreshes admin token
func HandleAdminRefreshTokenImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock implementation
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"accessToken":  "new_access_token",
			"refreshToken": "new_refresh_token",
		})
	}
}

// handleAdminLogoutImpl logs out admin
func HandleAdminLogoutImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock implementation
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Logged out successfully",
		})
	}
}
