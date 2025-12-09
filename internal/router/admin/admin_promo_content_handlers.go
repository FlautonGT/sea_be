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
// ADMIN PROMO MANAGEMENT
// ============================================

// handleAdminGetPromosImpl returns all promos with filters
func HandleAdminGetPromosImpl(deps *Dependencies) http.HandlerFunc {
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
				id, code, title, description,
				promo_percentage, promo_flat, max_promo_amount,
				min_amount, max_usage, total_usage,
				is_active, start_at, expired_at, created_at
			FROM promos
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		promos := []map[string]interface{}{}
		for rows.Next() {
			var id, code, title, description string
			var promoPercentage, promoFlat int
			var maxPromoAmount, minAmount int64
			var maxUsage, totalUsage int
			var isActive bool
			var startAt, expiredAt, createdAt time.Time

			rows.Scan(
				&id, &code, &title, &description,
				&promoPercentage, &promoFlat, &maxPromoAmount,
				&minAmount, &maxUsage, &totalUsage,
				&isActive, &startAt, &expiredAt, &createdAt,
			)

			promos = append(promos, map[string]interface{}{
				"id":              id,
				"code":            code,
				"title":           title,
				"description":     description,
				"promoPercentage": promoPercentage,
				"promoFlat":       promoFlat,
				"maxPromoAmount":  maxPromoAmount,
				"minAmount":       minAmount,
				"maxUsage":        maxUsage,
				"totalUsage":      totalUsage,
				"isActive":        isActive,
				"startAt":         startAt.Format(time.RFC3339),
				"expiredAt":       expiredAt.Format(time.RFC3339),
				"createdAt":       createdAt.Format(time.RFC3339),
			})
		}

		var totalRows int
		deps.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM promos").Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"promos": promos,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

// handleAdminGetPromoImpl returns detailed promo info
func HandleAdminGetPromoImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promoID := chi.URLParam(r, "promoId")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var id, code, title, description, note string
		var promoPercentage, promoFlat int
		var maxPromoAmount, minAmount int64
		var maxUsage, maxDailyUsage, maxUsagePerID, maxUsagePerDevice, maxUsagePerIP, totalUsage int
		var isActive bool
		var startAt, expiredAt, createdAt time.Time
		var daysAvailable []string

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, code, title, description, note,
				promo_percentage, promo_flat, max_promo_amount, min_amount,
				max_usage, max_daily_usage, max_usage_per_id, max_usage_per_device, max_usage_per_ip,
				total_usage, days_available, is_active, start_at, expired_at, created_at
			FROM promos
			WHERE id = $1
		`, promoID).Scan(
			&id, &code, &title, &description, &note,
			&promoPercentage, &promoFlat, &maxPromoAmount, &minAmount,
			&maxUsage, &maxDailyUsage, &maxUsagePerID, &maxUsagePerDevice, &maxUsagePerIP,
			&totalUsage, &daysAvailable, &isActive, &startAt, &expiredAt, &createdAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PROMO_NOT_FOUND",
					"Promo not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get products, payments, regions
		var products, payments, regions []string
		deps.DB.Pool.QueryRow(ctx, "SELECT ARRAY_AGG(product_code) FROM promo_products WHERE promo_id = $1", promoID).Scan(&products)
		deps.DB.Pool.QueryRow(ctx, "SELECT ARRAY_AGG(payment_code) FROM promo_payment_channels WHERE promo_id = $1", promoID).Scan(&payments)
		deps.DB.Pool.QueryRow(ctx, "SELECT ARRAY_AGG(region_code) FROM promo_regions WHERE promo_id = $1", promoID).Scan(&regions)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"id":                 id,
			"code":               code,
			"title":              title,
			"description":        description,
			"note":               note,
			"products":           products,
			"paymentChannels":    payments,
			"regions":            regions,
			"daysAvailable":      daysAvailable,
			"promoPercentage":    promoPercentage,
			"promoFlat":          promoFlat,
			"maxPromoAmount":     maxPromoAmount,
			"minAmount":          minAmount,
			"maxUsage":           maxUsage,
			"maxDailyUsage":      maxDailyUsage,
			"maxUsagePerID":      maxUsagePerID,
			"maxUsagePerDevice":  maxUsagePerDevice,
			"maxUsagePerIP":      maxUsagePerIP,
			"totalUsage":         totalUsage,
			"isActive":           isActive,
			"startAt":            startAt.Format(time.RFC3339),
			"expiredAt":          expiredAt.Format(time.RFC3339),
			"createdAt":          createdAt.Format(time.RFC3339),
		})
	}
}

// CreatePromoRequest represents the request to create a promo
type CreatePromoRequest struct {
	Code              string   `json:"code"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Products          []string `json:"products"`
	PaymentChannels   []string `json:"paymentChannels"`
	Regions           []string `json:"regions"`
	DaysAvailable     []string `json:"daysAvailable"`
	MaxDailyUsage     int      `json:"maxDailyUsage"`
	MaxUsage          int      `json:"maxUsage"`
	MaxUsagePerID     int      `json:"maxUsagePerId"`
	MaxUsagePerDevice int      `json:"maxUsagePerDevice"`
	MaxUsagePerIP     int      `json:"maxUsagePerIp"`
	StartAt           string   `json:"startAt"`
	ExpiredAt         string   `json:"expiredAt"`
	MinAmount         int64    `json:"minAmount"`
	MaxPromoAmount    int64    `json:"maxPromoAmount"`
	PromoFlat         int      `json:"promoFlat"`
	PromoPercentage   int      `json:"promoPercentage"`
	IsActive          bool     `json:"isActive"`
	Note              string   `json:"note"`
}

// handleCreatePromoImpl creates a new promo
func HandleCreatePromoImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreatePromoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fmt.Printf("Error decoding request: %v\n", err)
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Parse dates
		var startAt, expiredAt *time.Time
		if req.StartAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.StartAt)
			if err != nil {
				fmt.Printf("Error parsing startAt: %v\n", err)
				utils.WriteBadRequestError(w, "Invalid startAt format. Use RFC3339 format (e.g., 2025-12-11T17:00:00Z)")
				return
			}
			startAt = &parsed
		}
		if req.ExpiredAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.ExpiredAt)
			if err != nil {
				fmt.Printf("Error parsing expiredAt: %v\n", err)
				utils.WriteBadRequestError(w, "Invalid expiredAt format. Use RFC3339 format (e.g., 2025-12-12T16:59:00Z)")
				return
			}
			expiredAt = &parsed
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			fmt.Printf("Error beginning transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Insert promo
		var promoID string
		err = tx.QueryRow(ctx, `
			INSERT INTO promos (
				code, title, description, note,
				promo_percentage, promo_flat, max_promo_amount, min_amount,
				max_usage, max_daily_usage, max_usage_per_id, max_usage_per_device, max_usage_per_ip,
				days_available, is_active, start_at, expired_at, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())
			RETURNING id
		`, req.Code, req.Title, req.Description, req.Note,
			req.PromoPercentage, req.PromoFlat, req.MaxPromoAmount, req.MinAmount,
			req.MaxUsage, req.MaxDailyUsage, req.MaxUsagePerID, req.MaxUsagePerDevice, req.MaxUsagePerIP,
			req.DaysAvailable, req.IsActive, startAt, expiredAt,
		).Scan(&promoID)

		if err != nil {
			fmt.Printf("Error inserting promo: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Insert products (convert product_code to product_id)
		for _, productCode := range req.Products {
			var productID string
			err := tx.QueryRow(ctx, "SELECT id FROM products WHERE code = $1", productCode).Scan(&productID)
			if err != nil {
				if err == pgx.ErrNoRows {
					fmt.Printf("Product with code %s not found\n", productCode)
					continue
				}
				fmt.Printf("Error finding product %s: %v\n", productCode, err)
				continue
			}
			_, err = tx.Exec(ctx, "INSERT INTO promo_products (promo_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", promoID, productID)
			if err != nil {
				fmt.Printf("Error inserting promo product: %v\n", err)
			}
		}

		// Insert payments (convert payment_code to channel_id)
		for _, paymentCode := range req.PaymentChannels {
			var channelID string
			err := tx.QueryRow(ctx, "SELECT id FROM payment_channels WHERE code = $1", paymentCode).Scan(&channelID)
			if err != nil {
				if err == pgx.ErrNoRows {
					fmt.Printf("Payment channel with code %s not found\n", paymentCode)
					continue
				}
				fmt.Printf("Error finding payment channel %s: %v\n", paymentCode, err)
				continue
			}
			_, err = tx.Exec(ctx, "INSERT INTO promo_payment_channels (promo_id, channel_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", promoID, channelID)
			if err != nil {
				fmt.Printf("Error inserting promo payment channel: %v\n", err)
			}
		}

		// Insert regions
		for _, region := range req.Regions {
			_, err = tx.Exec(ctx, "INSERT INTO promo_regions (promo_id, region_code) VALUES ($1, $2) ON CONFLICT DO NOTHING", promoID, region)
			if err != nil {
				fmt.Printf("Error inserting promo region: %v\n", err)
			}
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'PROMO', $2, $3, NOW())
		`, adminID, promoID, "Created promo "+req.Code)

		if err := tx.Commit(ctx); err != nil {
			fmt.Printf("Error committing transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      promoID,
			"code":    req.Code,
			"message": "Promo created successfully",
		})
	}
}

// handleUpdatePromoImpl updates an existing promo
func HandleUpdatePromoImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promoID := chi.URLParam(r, "promoId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CreatePromoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Update promo
		_, err = tx.Exec(ctx, `
			UPDATE promos SET
				title = $1, description = $2, note = $3,
				promo_percentage = $4, promo_flat = $5, max_promo_amount = $6, min_amount = $7,
				max_usage = $8, max_daily_usage = $9, max_usage_per_id = $10, max_usage_per_device = $11, max_usage_per_ip = $12,
				days_available = $13, is_active = $14, start_at = $15, expired_at = $16
			WHERE id = $17
		`, req.Title, req.Description, req.Note,
			req.PromoPercentage, req.PromoFlat, req.MaxPromoAmount, req.MinAmount,
			req.MaxUsage, req.MaxDailyUsage, req.MaxUsagePerID, req.MaxUsagePerDevice, req.MaxUsagePerIP,
			req.DaysAvailable, req.IsActive, req.StartAt, req.ExpiredAt, promoID,
		)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Delete and re-insert associations
		tx.Exec(ctx, "DELETE FROM promo_products WHERE promo_id = $1", promoID)
		tx.Exec(ctx, "DELETE FROM promo_payment_channels WHERE promo_id = $1", promoID)
		tx.Exec(ctx, "DELETE FROM promo_regions WHERE promo_id = $1", promoID)

		for _, product := range req.Products {
			tx.Exec(ctx, "INSERT INTO promo_products (promo_id, product_code) VALUES ($1, $2)", promoID, product)
		}

		for _, payment := range req.PaymentChannels {
			tx.Exec(ctx, "INSERT INTO promo_payment_channels (promo_id, payment_code) VALUES ($1, $2)", promoID, payment)
		}

		for _, region := range req.Regions {
			tx.Exec(ctx, "INSERT INTO promo_regions (promo_id, region_code) VALUES ($1, $2)", promoID, region)
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'PROMO', $2, $3, NOW())
		`, adminID, promoID, "Updated promo "+req.Code)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Promo updated successfully",
		})
	}
}

// handleDeletePromoImpl deletes a promo
func HandleDeletePromoImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promoID := chi.URLParam(r, "promoId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Delete associations
		tx.Exec(ctx, "DELETE FROM promo_products WHERE promo_id = $1", promoID)
		tx.Exec(ctx, "DELETE FROM promo_payment_channels WHERE promo_id = $1", promoID)
		tx.Exec(ctx, "DELETE FROM promo_regions WHERE promo_id = $1", promoID)

		// Delete promo
		_, err = tx.Exec(ctx, "DELETE FROM promos WHERE id = $1", promoID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'PROMO', $2, 'Deleted promo', NOW())
		`, adminID, promoID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Promo deleted successfully",
		})
	}
}

// handleGetPromoStatsImpl returns promo usage statistics
func HandleGetPromoStatsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promoID := chi.URLParam(r, "promoId")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get promo code
		var promoCode string
		deps.DB.Pool.QueryRow(ctx, "SELECT code FROM promos WHERE id = $1", promoID).Scan(&promoCode)

		// Get stats (mock data for now - need promo_usages table)
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"promoCode":     promoCode,
			"totalUsage":    0,
			"totalDiscount": 0,
			"todayUsage":    0,
			"todayDiscount": 0,
			"usageByProduct": []map[string]interface{}{},
			"usageByPayment": []map[string]interface{}{},
			"usageByRegion":  []map[string]interface{}{},
		})
	}
}

// ============================================
// ADMIN CONTENT MANAGEMENT
// ============================================

// handleAdminGetBannersImpl returns all banners
func HandleAdminGetBannersImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT b.id, b.title, b.description, b.href, b.image, b.sort_order, b.is_active, b.start_at, b.expired_at, b.created_at,
			       COALESCE(array_agg(br.region_code) FILTER (WHERE br.region_code IS NOT NULL), '{}') as regions
			FROM banners b
			LEFT JOIN banner_regions br ON b.id = br.banner_id
			GROUP BY b.id
			ORDER BY b.sort_order ASC, b.created_at DESC
		`)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		banners := []map[string]interface{}{}
		for rows.Next() {
			var id, title, description, href, image string
			var regions []string
			var sortOrder int
			var isActive bool
			var startAt, expiredAt *time.Time
			var createdAt time.Time

			rows.Scan(&id, &title, &description, &href, &image, &sortOrder, &isActive, &startAt, &expiredAt, &createdAt, &regions)

			banner := map[string]interface{}{
				"id":          id,
				"title":       title,
				"description": description,
				"href":        href,
				"image":       image,
				"regions":     regions,
				"order":       sortOrder,
				"isActive":    isActive,
				"createdAt":   createdAt.Format(time.RFC3339),
			}

			if startAt != nil {
				banner["startAt"] = startAt.Format(time.RFC3339)
			}
			if expiredAt != nil {
				banner["expiredAt"] = expiredAt.Format(time.RFC3339)
			}

			banners = append(banners, banner)
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"banners": banners,
		})
	}
}

// CreateBannerRequest represents the request to create a banner
type CreateBannerRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Href        string   `json:"href"`
	Image       string   `json:"image"`
	Regions     []string `json:"regions"`
	Order       int      `json:"order"`
	IsActive    bool     `json:"isActive"`
	StartAt     string   `json:"startAt"`
	ExpiredAt   string   `json:"expiredAt"`
}

// handleCreateBannerImpl creates a new banner
func HandleCreateBannerImpl(deps *Dependencies) http.HandlerFunc {
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
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		href := strings.TrimSpace(r.FormValue("href"))
		orderStr := strings.TrimSpace(r.FormValue("order"))
		isActiveStr := r.FormValue("isActive")
		startAtStr := strings.TrimSpace(r.FormValue("startAt"))
		expiredAtStr := strings.TrimSpace(r.FormValue("expiredAt"))

		// Get regions array from form
		var regions []string
		if r.MultipartForm != nil {
			regions = r.MultipartForm.Value["regions[]"]
		}

		// Validation
		if title == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"title": "Title is required"})
			return
		}

		order := 1
		if orderStr != "" {
			if parsedOrder, err := strconv.Atoi(orderStr); err == nil {
				order = parsedOrder
			}
		}
		isActive := isActiveStr == "true"

		// Parse dates
		var startAt, expiredAt *time.Time
		if startAtStr != "" {
			if t, err := time.Parse(time.RFC3339, startAtStr); err == nil {
				startAt = &t
			}
		}
		if expiredAtStr != "" {
			if t, err := time.Parse(time.RFC3339, expiredAtStr); err == nil {
				expiredAt = &t
			}
		}

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
				result, err := deps.S3.Upload(ctx, storage.FolderBanner, file, header)
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
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"image": "Image file is required"})
			return
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		var bannerID string
		err = tx.QueryRow(ctx, `
			INSERT INTO banners (title, description, href, image, sort_order, is_active, start_at, expired_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
			RETURNING id
		`, title, description, href, imageURL, order, isActive, startAt, expiredAt).Scan(&bannerID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Insert regions into banner_regions table
		for _, regionCode := range regions {
			regionCode = strings.ToUpper(strings.TrimSpace(regionCode))
			if regionCode != "" {
				_, err = tx.Exec(ctx, `
					INSERT INTO banner_regions (banner_id, region_code)
					VALUES ($1, $2)
					ON CONFLICT (banner_id, region_code) DO NOTHING
				`, bannerID, regionCode)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
			}
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'BANNER', $2, 'Created banner', NOW())
		`, adminID, bannerID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      bannerID,
			"message": "Banner created successfully",
		})
	}
}

// handleUpdateBannerImpl updates an existing banner
func HandleUpdateBannerImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bannerID := chi.URLParam(r, "bannerId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.WriteBadRequestError(w, fmt.Sprintf("Invalid form data: %v", err))
			return
		}

		// Get form values
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		href := strings.TrimSpace(r.FormValue("href"))
		orderStr := strings.TrimSpace(r.FormValue("order"))
		isActiveStr := r.FormValue("isActive")
		startAtStr := strings.TrimSpace(r.FormValue("startAt"))
		expiredAtStr := strings.TrimSpace(r.FormValue("expiredAt"))

		// Get regions array from form
		var regions []string
		if r.MultipartForm != nil {
			regions = r.MultipartForm.Value["regions[]"]
		}

		// Validation
		if title == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"title": "Title is required"})
			return
		}

		order := 1
		if orderStr != "" {
			if parsedOrder, err := strconv.Atoi(orderStr); err == nil {
				order = parsedOrder
			}
		}
		isActive := isActiveStr == "true"

		// Parse dates
		var startAt, expiredAt *time.Time
		if startAtStr != "" {
			if t, err := time.Parse(time.RFC3339, startAtStr); err == nil {
				startAt = &t
			}
		}
		if expiredAtStr != "" {
			if t, err := time.Parse(time.RFC3339, expiredAtStr); err == nil {
				expiredAt = &t
			}
		}

		// Get existing banner image URL (if not uploading new one)
		var existingImageURL sql.NullString
		err := deps.DB.Pool.QueryRow(ctx, "SELECT image FROM banners WHERE id = $1", bannerID).Scan(&existingImageURL)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "BANNER_NOT_FOUND", "Banner not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Handle image upload
		var imageURL string
		if deps.S3 != nil {
			file, header, err := r.FormFile("image")
			if err == nil {
				// New image uploaded
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				result, err := deps.S3.Upload(ctx, storage.FolderBanner, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			} else if errors.Is(err, http.ErrMissingFile) {
				// No new image, keep existing
				if existingImageURL.Valid {
					imageURL = existingImageURL.String
				} else {
					// Check if image URL provided as form value
					imageURLValue := strings.TrimSpace(r.FormValue("image"))
					if imageURLValue != "" {
						imageURL = imageURLValue
					}
				}
			} else {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			}
		} else {
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			} else {
				if existingImageURL.Valid {
					imageURL = existingImageURL.String
				}
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

		_, err = tx.Exec(ctx, `
			UPDATE banners SET
				title = $1, description = $2, href = $3, image = $4,
				sort_order = $5, is_active = $6, start_at = $7, expired_at = $8, updated_at = NOW()
			WHERE id = $9
		`, title, description, href, imageURL, order, isActive, startAt, expiredAt, bannerID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Delete existing regions
		_, err = tx.Exec(ctx, `DELETE FROM banner_regions WHERE banner_id = $1`, bannerID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Insert new regions
		for _, regionCode := range regions {
			regionCode = strings.ToUpper(strings.TrimSpace(regionCode))
			if regionCode != "" {
				_, err = tx.Exec(ctx, `
					INSERT INTO banner_regions (banner_id, region_code)
					VALUES ($1, $2)
				`, bannerID, regionCode)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
			}
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'BANNER', $2, 'Updated banner', NOW())
		`, adminID, bannerID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Banner updated successfully",
		})
	}
}

// handleDeleteBannerImpl deletes a banner
func HandleDeleteBannerImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bannerID := chi.URLParam(r, "bannerId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "DELETE FROM banners WHERE id = $1", bannerID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'DELETE', 'BANNER', $2, 'Deleted banner', NOW())
		`, adminID, bannerID)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Banner deleted successfully",
		})
	}
}

// handleAdminGetPopupsImpl returns all popups
func HandleAdminGetPopupsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT id, region_code, title, content, image, href, is_active, created_at
			FROM popups
			ORDER BY region_code ASC
		`)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		popups := []map[string]interface{}{}
		for rows.Next() {
			var id, regionCode, title, content, image, href string
			var isActive bool
			var createdAt time.Time

			rows.Scan(&id, &regionCode, &title, &content, &image, &href, &isActive, &createdAt)

			popups = append(popups, map[string]interface{}{
				"id":        id,
				"region":    regionCode,
				"title":     title,
				"content":   content,
				"image":     image,
				"href":      href,
				"isActive":  isActive,
				"createdAt": createdAt.Format(time.RFC3339),
			})
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"popups": popups,
		})
	}
}

// handleCreatePopupImpl creates a new popup for a region
func HandleCreatePopupImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID := middleware.GetAdminIDFromContext(r.Context())

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.WriteBadRequestError(w, "Invalid form data")
			return
		}

		// Get form values
		regionCode := strings.ToUpper(strings.TrimSpace(r.FormValue("region")))
		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		href := strings.TrimSpace(r.FormValue("href"))
		isActiveStr := r.FormValue("isActive")
		isActive := isActiveStr == "true"

		// Validation
		if regionCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"region": "Region is required"})
			return
		}
		if title == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"title": "Title is required"})
			return
		}

		// Check if popup already exists for this region
		var existingID string
		err := deps.DB.Pool.QueryRow(ctx, "SELECT id FROM popups WHERE region_code = $1", regionCode).Scan(&existingID)
		if err == nil {
			// Popup already exists
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"region": fmt.Sprintf("Popup for region %s already exists. Please use update endpoint or edit existing popup.", regionCode),
			})
			return
		}
		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

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
				result, err := deps.S3.Upload(ctx, storage.FolderPopup, file, header)
				if err != nil {
					utils.WriteInternalServerError(w)
					return
				}
				imageURL = result.URL
			}
		}

		if imageURL == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"image": "Image file is required"})
			return
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Insert popup
		var popupID string
		err = tx.QueryRow(ctx, `
			INSERT INTO popups (region_code, title, content, image, href, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			RETURNING id
		`, regionCode, title, content, imageURL, href, isActive).Scan(&popupID)

		if err != nil {
			// Check if unique constraint violation
			if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"region": fmt.Sprintf("Popup for region %s already exists", regionCode),
				})
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CREATE', 'POPUP', $2, $3, NOW())
		`, adminID, popupID, fmt.Sprintf("Created popup for region %s", regionCode))

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      popupID,
			"message": "Popup created successfully",
		})
	}
}

// UpdatePopupRequest represents the request to update a popup
type UpdatePopupRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Image    string `json:"image"`
	Href     string `json:"href"`
	IsActive bool   `json:"isActive"`
}

// handleUpdatePopupImpl updates a popup for a region
func HandleUpdatePopupImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		region := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "region")))
		adminID := middleware.GetAdminIDFromContext(r.Context())

		fmt.Printf("[UPDATE POPUP] Starting update for region: %s\n", region)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			fmt.Printf("[UPDATE POPUP] Error parsing multipart form: %v\n", err)
			utils.WriteBadRequestError(w, fmt.Sprintf("Invalid form data: %v", err))
			return
		}
		fmt.Printf("[UPDATE POPUP] Multipart form parsed successfully\n")

		// Get form values
		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		href := strings.TrimSpace(r.FormValue("href"))
		isActiveStr := r.FormValue("isActive")
		isActive := isActiveStr == "true"

		// Validation
		if title == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"title": "Title is required"})
			return
		}

		// Get existing popup image URL (if not uploading new one)
		var existingImageURL sql.NullString
		err := deps.DB.Pool.QueryRow(ctx, "SELECT image FROM popups WHERE region_code = $1", region).Scan(&existingImageURL)
		if err != nil && err != pgx.ErrNoRows {
			fmt.Printf("Error querying existing popup for region %s: %v\n", region, err)
			utils.WriteInternalServerError(w)
			return
		}

		// Handle image upload
		var imageURL string
		if deps.S3 != nil {
			fmt.Printf("[UPDATE POPUP] S3 configured, checking for image file\n")
			file, header, err := r.FormFile("image")
			if err == nil {
				fmt.Printf("[UPDATE POPUP] Image file found, size: %d bytes\n", header.Size)
				// New image uploaded
				defer file.Close()
				if err := storage.ValidateImageFile(header); err != nil {
					fmt.Printf("[UPDATE POPUP] Image validation failed: %v\n", err)
					utils.WriteBadRequestError(w, err.Error())
					return
				}
				fmt.Printf("[UPDATE POPUP] Starting S3 upload...\n")
				result, err := deps.S3.Upload(ctx, storage.FolderPopup, file, header)
				if err != nil {
					fmt.Printf("[UPDATE POPUP] S3 upload failed: %v\n", err)
					utils.WriteInternalServerError(w)
					return
				}
				fmt.Printf("[UPDATE POPUP] S3 upload successful, URL: %s\n", result.URL)
				imageURL = result.URL
			} else if errors.Is(err, http.ErrMissingFile) {
				fmt.Printf("[UPDATE POPUP] No new image file, keeping existing\n")
				// No new image, keep existing
				if existingImageURL.Valid {
					imageURL = existingImageURL.String
					fmt.Printf("[UPDATE POPUP] Using existing image: %s\n", imageURL)
				} else {
					fmt.Printf("[UPDATE POPUP] No existing image found\n")
				}
			} else {
				fmt.Printf("[UPDATE POPUP] Error getting form file: %v\n", err)
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			}
		} else {
			fmt.Printf("[UPDATE POPUP] S3 not configured\n")
			// S3 not configured, check if image URL provided in form
			imageURLValue := strings.TrimSpace(r.FormValue("image"))
			if imageURLValue != "" {
				imageURL = imageURLValue
			} else {
				if existingImageURL.Valid {
					imageURL = existingImageURL.String
				}
			}
		}

		if imageURL == "" {
			fmt.Printf("[UPDATE POPUP] Image URL is empty after processing\n")
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"image": "Image is required"})
			return
		}
		fmt.Printf("[UPDATE POPUP] Image URL determined: %s\n", imageURL)

		// Begin transaction
		fmt.Printf("[UPDATE POPUP] Starting database transaction\n")
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			fmt.Printf("[UPDATE POPUP] Error beginning transaction: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Upsert popup
		fmt.Printf("[UPDATE POPUP] Executing upsert query\n")
		_, err = tx.Exec(ctx, `
			INSERT INTO popups (region_code, title, content, image, href, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			ON CONFLICT (region_code) DO UPDATE SET
				title = EXCLUDED.title, content = EXCLUDED.content, image = EXCLUDED.image, href = EXCLUDED.href, is_active = EXCLUDED.is_active, updated_at = NOW()
		`, region, title, content, imageURL, href, isActive)

		if err != nil {
			fmt.Printf("[UPDATE POPUP] Error upserting popup for region %s: %v\n", region, err)
			utils.WriteInternalServerError(w)
			return
		}
		fmt.Printf("[UPDATE POPUP] Upsert successful\n")

		// Get popup ID for audit log
		var popupID string
		err = tx.QueryRow(ctx, "SELECT id FROM popups WHERE region_code = $1", region).Scan(&popupID)
		if err != nil {
			fmt.Printf("Error getting popup ID for region %s: %v\n", region, err)
			// Continue even if we can't get the ID for audit log
		}

		// Create audit log (non-blocking)
		if popupID != "" {
			tx.Exec(ctx, `
				INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
				VALUES ($1, 'UPDATE', 'POPUP', $2, $3, NOW())
			`, adminID, popupID, fmt.Sprintf("Updated popup for region %s", region))
		}

		fmt.Printf("[UPDATE POPUP] Committing transaction\n")
		if err := tx.Commit(ctx); err != nil {
			fmt.Printf("[UPDATE POPUP] Error committing transaction for popup update (region %s): %v\n", region, err)
			utils.WriteInternalServerError(w)
			return
		}
		fmt.Printf("[UPDATE POPUP] Transaction committed successfully\n")

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Popup updated successfully",
		})
	}
}
