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

	"seaply/internal/storage"
	"seaply/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type skuBadgePayload struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

type skuPricingPayload struct {
	BuyPrice      json.Number `json:"buyPrice"`
	SellPrice     json.Number `json:"sellPrice"`
	OriginalPrice json.Number `json:"originalPrice"`
}

type skuPayload struct {
	Code            string                       `json:"code"`
	ProviderSku     string                       `json:"providerSkuCode"`
	Name            string                       `json:"name"`
	Description     string                       `json:"description"`
	ProductCode     string                       `json:"productCode"`
	ProviderCode    string                       `json:"providerCode"`
	SectionCode     string                       `json:"sectionCode"`
	IsActive        *bool                        `json:"isActive"`
	IsFeatured      *bool                        `json:"isFeatured"`
	ProcessTime     *int                         `json:"processTime"`
	Info            string                       `json:"info"`
	StockStatus     string                       `json:"stockStatus"`
	Pricing         map[string]skuPricingPayload `json:"pricing"`
	Badge           *skuBadgePayload             `json:"badge"`
	ImageURL        string                       `json:"image"`
	ExistingPricing map[string]skuPricingPayload `json:"-"`
}

type skuRecord struct {
	ID           uuid.UUID
	Code         string
	ProviderSku  string
	Name         string
	Description  sql.NullString
	Image        sql.NullString
	Info         sql.NullString
	ProductID    uuid.UUID
	ProductCode  string
	ProductTitle string
	ProviderID   uuid.UUID
	ProviderCode string
	ProviderName string
	SectionID    sql.NullString
	SectionCode  sql.NullString
	SectionTitle sql.NullString
	ProcessTime  int
	IsActive     bool
	IsFeatured   bool
	StockStatus  string
	BadgeText    sql.NullString
	BadgeColor   sql.NullString
	TotalSold    int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type skuPricingRecord struct {
	RegionCode string
	Currency   string
	BuyPrice   int64
	SellPrice  int64
	Original   int64
	Margin     float64
	Discount   float64
}

type skuStatsRecord struct {
	TodaySold int64
	TotalSold int64
}

func HandleAdminGetSKUsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		limit := parseQueryInt(r, "limit", 10)
		page := parseQueryInt(r, "page", 1)
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * limit

		search := strings.TrimSpace(r.URL.Query().Get("search"))
		productCode := strings.TrimSpace(r.URL.Query().Get("productCode"))
		providerCode := strings.TrimSpace(r.URL.Query().Get("providerCode"))
		regionCode := strings.TrimSpace(r.URL.Query().Get("region"))
		isActiveStr := strings.TrimSpace(r.URL.Query().Get("isActive"))

		baseQuery := `
			SELECT
				s.id, s.code, s.provider_sku_code, s.name, s.description, s.image, s.info,
				p.id as product_id, p.code as product_code, p.title,
				pr.id as provider_id, pr.code as provider_code, pr.name,
				sc.id as section_id, sc.code as section_code, sc.title,
				s.process_time, s.is_active, s.is_featured, s.stock_status,
				s.badge_text, s.badge_color,
				s.total_sold, s.created_at, s.updated_at
			FROM skus s
			JOIN products p ON s.product_id = p.id
			JOIN providers pr ON s.provider_id = pr.id
			LEFT JOIN sections sc ON s.section_id = sc.id
			WHERE 1=1
		`

		args := []interface{}{}
		argPos := 1

		if search != "" {
			baseQuery += fmt.Sprintf(` AND (LOWER(s.name) LIKE LOWER($%d) OR LOWER(s.code) LIKE LOWER($%d))`, argPos, argPos)
			args = append(args, "%"+search+"%")
			argPos++
		}
		if productCode != "" {
			baseQuery += fmt.Sprintf(` AND p.code = $%d`, argPos)
			args = append(args, productCode)
			argPos++
		}
		if providerCode != "" {
			baseQuery += fmt.Sprintf(` AND pr.code = $%d`, argPos)
			args = append(args, providerCode)
			argPos++
		}
		if regionCode != "" {
			baseQuery += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM sku_pricing sp WHERE sp.sku_id = s.id AND sp.region_code = $%d)`, argPos)
			args = append(args, strings.ToUpper(regionCode))
			argPos++
		}
		if isActiveStr != "" {
			active := strings.ToLower(isActiveStr) == "true"
			baseQuery += fmt.Sprintf(` AND s.is_active = $%d`, argPos)
			args = append(args, active)
			argPos++
		}

		countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS sub"
		var totalRows int
		if err := deps.DB.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalRows); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		baseQuery += fmt.Sprintf(" ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
		args = append(args, limit, offset)

		rows, err := deps.DB.Pool.Query(ctx, baseQuery, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var skuIDs []uuid.UUID
		var records []skuRecord

		for rows.Next() {
			var rec skuRecord
			if err := rows.Scan(
				&rec.ID, &rec.Code, &rec.ProviderSku, &rec.Name, &rec.Description, &rec.Image, &rec.Info,
				&rec.ProductID, &rec.ProductCode, &rec.ProductTitle,
				&rec.ProviderID, &rec.ProviderCode, &rec.ProviderName,
				&rec.SectionID, &rec.SectionCode, &rec.SectionTitle,
				&rec.ProcessTime, &rec.IsActive, &rec.IsFeatured, &rec.StockStatus,
				&rec.BadgeText, &rec.BadgeColor,
				&rec.TotalSold, &rec.CreatedAt, &rec.UpdatedAt,
			); err != nil {
				continue
			}
			skuIDs = append(skuIDs, rec.ID)
			records = append(records, rec)
		}

		pricingMap := fetchSKUPricingMap(ctx, deps, skuIDs)
		statsMap := fetchSKUTodaySold(ctx, deps, skuIDs)

		var skus []map[string]interface{}
		for _, rec := range records {
			stats := skuStatsRecord{
				TodaySold: statsMap[rec.ID],
				TotalSold: rec.TotalSold,
			}
			skus = append(skus, buildSKUResponse(rec, pricingMap[rec.ID], stats))
		}

		totalPages := (totalRows + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"skus": skus,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

func HandleAdminGetSKUImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()

		skuID := chi.URLParam(r, "skuId")
		if skuID == "" {
			utils.WriteBadRequestError(w, "SKU ID is required")
			return
		}

		rec, err := loadSKUByIdentifier(ctx, deps, skuID)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "SKU_NOT_FOUND", "SKU not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		pricingMap := fetchSKUPricingMap(ctx, deps, []uuid.UUID{rec.ID})
		statsMap := fetchSKUTodaySold(ctx, deps, []uuid.UUID{rec.ID})
		stats := skuStatsRecord{
			TodaySold: statsMap[rec.ID],
			TotalSold: rec.TotalSold,
		}

		utils.WriteSuccessJSON(w, buildSKUResponse(*rec, pricingMap[rec.ID], stats))
	}
}

func HandleCreateSKUImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()

		if err := r.ParseMultipartForm(25 << 20); err != nil {
			utils.WriteBadRequestError(w, "Invalid form data")
			return
		}

		payloadStr := r.FormValue("payload")
		if payloadStr == "" {
			utils.WriteBadRequestError(w, "Payload is required")
			return
		}

		var payload skuPayload
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			utils.WriteBadRequestError(w, "Invalid payload JSON")
			return
		}

		payload.Code = strings.TrimSpace(payload.Code)
		if payload.Code == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Code is required"})
			return
		}

		payload.ProviderSku = strings.TrimSpace(payload.ProviderSku)
		if payload.ProviderSku == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"providerSkuCode": "Provider SKU code is required"})
			return
		}

		payload.Name = strings.TrimSpace(payload.Name)
		if payload.Name == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"name": "Name is required"})
			return
		}

		payload.ProductCode = strings.TrimSpace(payload.ProductCode)
		payload.ProviderCode = strings.TrimSpace(payload.ProviderCode)
		if payload.ProductCode == "" || payload.ProviderCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"productCode": "Product and provider codes are required"})
			return
		}

		if len(payload.Pricing) == 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"pricing": "Pricing per region is required"})
			return
		}

		productID, err := getProductIDByCode(ctx, deps, payload.ProductCode)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "Product not found", "")
			return
		}

		providerID, err := getProviderIDByCode(ctx, deps, payload.ProviderCode)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "PROVIDER_NOT_FOUND", "Provider not found", "")
			return
		}

		var sectionID *uuid.UUID
		if payload.SectionCode != "" {
			secID, err := getSectionIDByCode(ctx, deps, payload.SectionCode)
			if err != nil {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "SECTION_NOT_FOUND", "Section not found", "")
				return
			}
			sectionID = &secID
		}

		if err := ensureSKUCodeUnique(ctx, deps, payload.Code, nil); err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "SKU code already exists"})
			return
		}
		if err := ensureProviderSKUUnique(ctx, deps, providerID, payload.ProviderSku, nil); err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"providerSkuCode": "Provider SKU code already in use"})
			return
		}

		pricingRegions := normalizeStringSlice(mapKeys(payload.Pricing))
		if err := validateRegions(ctx, deps, pricingRegions); err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"pricing": "Invalid region codes in pricing"})
			return
		}
		regionCurrencies, err := fetchRegionCurrencies(ctx, deps, pricingRegions)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		imageURL, err := uploadSKUImage(ctx, deps, "image", r)
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) && payload.ImageURL != "" {
				imageURL = payload.ImageURL
			} else {
				utils.WriteBadRequestError(w, fmt.Sprintf("Image upload failed: %v", err))
				return
			}
		}
		if imageURL == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"image": "Image file is required"})
			return
		}

		isActive := true
		if payload.IsActive != nil {
			isActive = *payload.IsActive
		}
		isFeatured := false
		if payload.IsFeatured != nil {
			isFeatured = *payload.IsFeatured
		}
		processTime := 0
		if payload.ProcessTime != nil {
			processTime = *payload.ProcessTime
		}
		stockStatus := payload.StockStatus
		if stockStatus == "" {
			stockStatus = "AVAILABLE"
		}

		priceRecords, err := buildPricingInsertRecords(payload.Pricing, regionCurrencies)
		if err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"pricing": err.Error()})
			return
		}

		var skuID uuid.UUID
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		err = tx.QueryRow(ctx, `
			INSERT INTO skus (
				code, provider_sku_code, name, description, image, info,
				product_id, provider_id, section_id,
				process_time, is_active, is_featured, stock_status,
				badge_text, badge_color
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			RETURNING id
		`, payload.Code, payload.ProviderSku, payload.Name, nullString(payload.Description), imageURL, nullString(payload.Info),
			productID, providerID, nullableUUID(sectionID),
			processTime, isActive, isFeatured, stockStatus,
			nullStringFromPtr(payload.Badge, func(b *skuBadgePayload) string { return b.Text }),
			nullStringFromPtr(payload.Badge, func(b *skuBadgePayload) string { return b.Color }),
		).Scan(&skuID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if err := insertSKUPricing(ctx, tx, skuID, priceRecords); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		rec, err := loadSKUByIdentifier(ctx, deps, skuID.String())
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		pricingMap := fetchSKUPricingMap(ctx, deps, []uuid.UUID{rec.ID})
		stats := skuStatsRecord{TodaySold: 0, TotalSold: rec.TotalSold}

		utils.WriteCreatedJSON(w, buildSKUResponse(*rec, pricingMap[rec.ID], stats))
	}
}

func HandleUpdateSKUImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()

		skuIDParam := chi.URLParam(r, "skuId")
		if skuIDParam == "" {
			utils.WriteBadRequestError(w, "SKU ID is required")
			return
		}

		rec, err := loadSKUByIdentifier(ctx, deps, skuIDParam)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "SKU_NOT_FOUND", "SKU not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		if err := r.ParseMultipartForm(25 << 20); err != nil {
			utils.WriteBadRequestError(w, "Invalid form data")
			return
		}

		payloadStr := r.FormValue("payload")
		if payloadStr == "" && (r.MultipartForm == nil || len(r.MultipartForm.File) == 0) {
			utils.WriteBadRequestError(w, "Payload is required")
			return
		}

		payload := map[string]json.RawMessage{}
		if payloadStr != "" {
			if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
				utils.WriteBadRequestError(w, "Invalid payload JSON")
				return
			}
		}

		var updates []string
		var args []interface{}
		argPos := 1

		finalProviderID := rec.ProviderID
		var providerChanged bool
		var newProviderSku string

		if raw, ok := payload["code"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				value = strings.TrimSpace(value)
				if err := ensureSKUCodeUnique(ctx, deps, value, &rec.ID); err != nil {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "SKU code already exists"})
					return
				}
				updates = append(updates, fmt.Sprintf("code = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		if raw, ok := payload["providerSkuCode"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				newProviderSku = strings.TrimSpace(value)
			}
		}

		if raw, ok := payload["providerCode"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				providerID, err := getProviderIDByCode(ctx, deps, strings.TrimSpace(value))
				if err != nil {
					utils.WriteErrorJSON(w, http.StatusBadRequest, "PROVIDER_NOT_FOUND", "Provider not found", "")
					return
				}
				finalProviderID = providerID
				updates = append(updates, fmt.Sprintf("provider_id = $%d", argPos))
				args = append(args, providerID)
				argPos++
				providerChanged = true
			}
		}

		if newProviderSku != "" || providerChanged {
			providerSku := rec.ProviderSku
			if newProviderSku != "" {
				providerSku = newProviderSku
			}
			if err := ensureProviderSKUUnique(ctx, deps, finalProviderID, providerSku, &rec.ID); err != nil {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"providerSkuCode": "Provider SKU code already in use"})
				return
			}
			if newProviderSku != "" {
				updates = append(updates, fmt.Sprintf("provider_sku_code = $%d", argPos))
				args = append(args, newProviderSku)
				argPos++
			}
		}

		if raw, ok := payload["name"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				updates = append(updates, fmt.Sprintf("name = $%d", argPos))
				args = append(args, strings.TrimSpace(value))
				argPos++
			}
		}

		for _, field := range []struct {
			Key    string
			Column string
		}{
			{"description", "description"},
			{"info", "info"},
			{"stockStatus", "stock_status"},
		} {
			if raw, ok := payload[field.Key]; ok {
				var value string
				if err := json.Unmarshal(raw, &value); err == nil {
					updates = append(updates, fmt.Sprintf("%s = NULLIF($%d, '')", field.Column, argPos))
					args = append(args, strings.TrimSpace(value))
					argPos++
				}
			}
		}

		if raw, ok := payload["productCode"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				productID, err := getProductIDByCode(ctx, deps, strings.TrimSpace(value))
				if err != nil {
					utils.WriteErrorJSON(w, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "Product not found", "")
					return
				}
				updates = append(updates, fmt.Sprintf("product_id = $%d", argPos))
				args = append(args, productID)
				argPos++
			}
		}

		if raw, ok := payload["sectionCode"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil {
				value = strings.TrimSpace(value)
				if value == "" {
					updates = append(updates, "section_id = NULL")
				} else {
					sectionID, err := getSectionIDByCode(ctx, deps, value)
					if err != nil {
						utils.WriteErrorJSON(w, http.StatusBadRequest, "SECTION_NOT_FOUND", "Section not found", "")
						return
					}
					updates = append(updates, fmt.Sprintf("section_id = $%d", argPos))
					args = append(args, sectionID)
					argPos++
				}
			}
		}

		if raw, ok := payload["processTime"]; ok {
			var value int
			if err := json.Unmarshal(raw, &value); err == nil {
				updates = append(updates, fmt.Sprintf("process_time = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		if raw, ok := payload["isActive"]; ok {
			var value bool
			if err := json.Unmarshal(raw, &value); err == nil {
				updates = append(updates, fmt.Sprintf("is_active = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		if raw, ok := payload["isFeatured"]; ok {
			var value bool
			if err := json.Unmarshal(raw, &value); err == nil {
				updates = append(updates, fmt.Sprintf("is_featured = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		if raw, ok := payload["badge"]; ok {
			var badge skuBadgePayload
			if err := json.Unmarshal(raw, &badge); err == nil {
				updates = append(updates, fmt.Sprintf("badge_text = NULLIF($%d,'')", argPos))
				args = append(args, strings.TrimSpace(badge.Text))
				argPos++
				updates = append(updates, fmt.Sprintf("badge_color = NULLIF($%d,'')", argPos))
				args = append(args, strings.TrimSpace(badge.Color))
				argPos++
			}
		}

		var newImageURL string
		if file, header, err := r.FormFile("image"); err == nil {
			if deps.S3 == nil {
				utils.WriteInternalServerError(w)
				return
			}
			defer file.Close()
			if err := storage.ValidateImageFile(header); err != nil {
				utils.WriteBadRequestError(w, err.Error())
				return
			}
			result, err := deps.S3.UploadWithOriginalName(ctx, storage.FolderSKU, file, header)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			newImageURL = result.URL
			updates = append(updates, fmt.Sprintf("image = $%d", argPos))
			args = append(args, newImageURL)
			argPos++
		} else if raw, ok := payload["image"]; ok {
			// If no file uploaded but image URL provided in payload
			var imageURL string
			if err := json.Unmarshal(raw, &imageURL); err == nil && strings.TrimSpace(imageURL) != "" {
				newImageURL = strings.TrimSpace(imageURL)
				updates = append(updates, fmt.Sprintf("image = $%d", argPos))
				args = append(args, newImageURL)
				argPos++
			}
		}

		pricingUpdates := map[string]skuPricingPayload{}
		if raw, ok := payload["pricing"]; ok {
			if err := json.Unmarshal(raw, &pricingUpdates); err != nil {
				utils.WriteBadRequestError(w, "Invalid pricing payload")
				return
			}
			if len(pricingUpdates) == 0 {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"pricing": "Pricing must include at least one region"})
				return
			}
		}

		if len(updates) == 0 && len(pricingUpdates) == 0 && newImageURL == "" {
			utils.WriteBadRequestError(w, "No fields to update")
			return
		}

		updates = append(updates, "updated_at = NOW()")
		args = append(args, rec.ID)

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		if len(updates) > 0 {
			query := fmt.Sprintf(`UPDATE skus SET %s WHERE id = $%d`, strings.Join(updates, ", "), argPos)
			if _, err := tx.Exec(ctx, query, args...); err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if len(pricingUpdates) > 0 {
			regions := normalizeStringSlice(mapKeys(pricingUpdates))
			if err := validateRegions(ctx, deps, regions); err != nil {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"pricing": "Invalid region codes in pricing"})
				return
			}
			currencies, err := fetchRegionCurrencies(ctx, deps, regions)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			priceRecords, err := buildPricingInsertRecords(pricingUpdates, currencies)
			if err != nil {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"pricing": err.Error()})
				return
			}
			if err := upsertSKUPricing(ctx, tx, rec.ID, priceRecords); err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if newImageURL != "" && rec.Image.Valid {
			deleteS3Object(ctx, deps, rec.Image.String)
		}

		updated, err := loadSKUByIdentifier(ctx, deps, rec.ID.String())
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		pricingMap := fetchSKUPricingMap(ctx, deps, []uuid.UUID{updated.ID})
		statsMap := fetchSKUTodaySold(ctx, deps, []uuid.UUID{updated.ID})
		stats := skuStatsRecord{
			TodaySold: statsMap[updated.ID],
			TotalSold: updated.TotalSold,
		}

		utils.WriteSuccessJSON(w, buildSKUResponse(*updated, pricingMap[updated.ID], stats))
	}
}

func HandleDeleteSKUImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		skuID := chi.URLParam(r, "skuId")
		if skuID == "" {
			utils.WriteBadRequestError(w, "SKU ID is required")
			return
		}

		rec, err := loadSKUByIdentifier(ctx, deps, skuID)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "SKU_NOT_FOUND", "SKU not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		var hasTransactions bool
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM transactions WHERE sku_id = $1)
		`, rec.ID).Scan(&hasTransactions); err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if hasTransactions {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "SKU_HAS_TRANSACTIONS", "Cannot delete SKU with transactions", "")
			return
		}

		if _, err := deps.DB.Pool.Exec(ctx, `DELETE FROM skus WHERE id = $1`, rec.ID); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if rec.Image.Valid {
			deleteS3Object(ctx, deps, rec.Image.String)
		}

		utils.WriteSuccessJSON(w, map[string]string{"message": "SKU deleted successfully"})
	}
}

func HandleBulkUpdatePriceImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		var req struct {
			SKUs []struct {
				Code    string `json:"code"`
				Pricing map[string]struct {
					SellPrice     json.Number `json:"sellPrice"`
					OriginalPrice json.Number `json:"originalPrice"`
				} `json:"pricing"`
			} `json:"skus"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if len(req.SKUs) == 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"skus": "At least one SKU entry is required"})
			return
		}

		updatedCount := 0
		err := deps.DB.WithTransaction(ctx, func(tx pgx.Tx) error {
			for _, item := range req.SKUs {
				code := strings.TrimSpace(item.Code)
				if code == "" {
					return errors.New("SKU code is required in bulk update")
				}
				if len(item.Pricing) == 0 {
					return fmt.Errorf("Pricing is required for SKU %s", code)
				}

				var skuID uuid.UUID
				if err := tx.QueryRow(ctx, `SELECT id FROM skus WHERE code = $1`, code).Scan(&skuID); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						return fmt.Errorf("SKU %s not found", code)
					}
					return err
				}

				for region, price := range item.Pricing {
					normalizedRegion := strings.ToUpper(strings.TrimSpace(region))
					if normalizedRegion == "" {
						return fmt.Errorf("Region code is required for SKU %s", code)
					}
					sell, err := parsePriceNumber(price.SellPrice)
					if err != nil {
						return fmt.Errorf("Invalid sellPrice for SKU %s (%s): %v", code, normalizedRegion, err)
					}
					original, err := parsePriceNumber(price.OriginalPrice)
					if err != nil {
						return fmt.Errorf("Invalid originalPrice for SKU %s (%s): %v", code, normalizedRegion, err)
					}

					cmd, err := tx.Exec(ctx, `
						UPDATE sku_pricing
						SET sell_price = $1, original_price = $2, updated_at = NOW()
						WHERE sku_id = $3 AND region_code = $4
					`, sell, original, skuID, normalizedRegion)
					if err != nil {
						return err
					}
					if cmd.RowsAffected() == 0 {
						return fmt.Errorf("Pricing for region %s not found on SKU %s", normalizedRegion, code)
					}
				}
				updatedCount++
			}
			return nil
		})

		if err != nil {
			utils.WriteBadRequestError(w, err.Error())
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"updatedSkus": updatedCount,
			"message":     "Sku prices updated successfully",
		})
	}
}

func HandleSyncSKUsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		var req struct {
			ProviderCode string  `json:"providerCode"`
			ProductCode  string  `json:"productCode"`
			AutoActivate bool    `json:"autoActivate"`
			PriceMargin  float64 `json:"priceMargin"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if strings.TrimSpace(req.ProviderCode) == "" || strings.TrimSpace(req.ProductCode) == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"providerCode": "Provider code is required",
				"productCode":  "Product code is required",
			})
			return
		}

		if _, err := getProviderIDByCode(ctx, deps, req.ProviderCode); err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "PROVIDER_NOT_FOUND", "Provider not found", "")
			return
		}
		if _, err := getProductIDByCode(ctx, deps, req.ProductCode); err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "PRODUCT_NOT_FOUND", "Product not found", "")
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"status": "COMPLETED",
			"summary": map[string]int{
				"totalFromProvider": 0,
				"newSkus":           0,
				"updatedSkus":       0,
				"skippedSkus":       0,
			},
			"newSkus":  []interface{}{},
			"syncedAt": time.Now().Format(time.RFC3339),
		})
	}
}

type skuPricingInsertRecord struct {
	Region   string
	Currency string
	Buy      int64
	Sell     int64
	Original int64
}

func buildPricingInsertRecords(pricing map[string]skuPricingPayload, currencies map[string]string) ([]skuPricingInsertRecord, error) {
	var records []skuPricingInsertRecord
	for region, value := range pricing {
		normalized := strings.ToUpper(strings.TrimSpace(region))
		if normalized == "" {
			return nil, errors.New("Region code is required in pricing")
		}
		currency, ok := currencies[normalized]
		if !ok {
			return nil, fmt.Errorf("Currency for region %s not found", normalized)
		}

		buy, err := parsePriceNumber(value.BuyPrice)
		if err != nil {
			return nil, fmt.Errorf("Invalid buyPrice for region %s: %v", normalized, err)
		}
		sell, err := parsePriceNumber(value.SellPrice)
		if err != nil {
			return nil, fmt.Errorf("Invalid sellPrice for region %s: %v", normalized, err)
		}
		original, err := parsePriceNumber(value.OriginalPrice)
		if err != nil {
			return nil, fmt.Errorf("Invalid originalPrice for region %s: %v", normalized, err)
		}

		records = append(records, skuPricingInsertRecord{
			Region:   normalized,
			Currency: currency,
			Buy:      buy,
			Sell:     sell,
			Original: original,
		})
	}
	return records, nil
}

func insertSKUPricing(ctx context.Context, tx pgx.Tx, skuID uuid.UUID, records []skuPricingInsertRecord) error {
	for _, rec := range records {
		if _, err := tx.Exec(ctx, `
			INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (sku_id, region_code) DO UPDATE
			SET currency = EXCLUDED.currency,
				buy_price = EXCLUDED.buy_price,
				sell_price = EXCLUDED.sell_price,
				original_price = EXCLUDED.original_price,
				updated_at = NOW()
		`, skuID, rec.Region, rec.Currency, rec.Buy, rec.Sell, rec.Original); err != nil {
			return err
		}
	}
	return nil
}

func upsertSKUPricing(ctx context.Context, tx pgx.Tx, skuID uuid.UUID, records []skuPricingInsertRecord) error {
	return insertSKUPricing(ctx, tx, skuID, records)
}

func fetchSKUPricingMap(ctx context.Context, deps *Dependencies, skuIDs []uuid.UUID) map[uuid.UUID][]skuPricingRecord {
	result := make(map[uuid.UUID][]skuPricingRecord)
	if len(skuIDs) == 0 {
		return result
	}

	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT sku_id, region_code, currency, buy_price, sell_price, original_price,
			COALESCE(margin_percentage, 0), COALESCE(discount_percentage, 0)
		FROM sku_pricing
		WHERE sku_id = ANY($1) AND is_active = TRUE
	`, skuIDs)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var skuID uuid.UUID
		var rec skuPricingRecord
		if err := rows.Scan(&skuID, &rec.RegionCode, &rec.Currency, &rec.BuyPrice, &rec.SellPrice, &rec.Original, &rec.Margin, &rec.Discount); err != nil {
			continue
		}
		result[skuID] = append(result[skuID], rec)
	}
	return result
}

func fetchSKUTodaySold(ctx context.Context, deps *Dependencies, skuIDs []uuid.UUID) map[uuid.UUID]int64 {
	result := make(map[uuid.UUID]int64)
	if len(skuIDs) == 0 {
		return result
	}

	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT sku_id, COUNT(*) as today_sold
		FROM transactions
		WHERE sku_id = ANY($1) AND created_at::date = CURRENT_DATE
		GROUP BY sku_id
	`, skuIDs)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var skuID uuid.UUID
		var count int64
		if err := rows.Scan(&skuID, &count); err != nil {
			continue
		}
		result[skuID] = count
	}
	return result
}

func buildSKUResponse(rec skuRecord, pricing []skuPricingRecord, stats skuStatsRecord) map[string]interface{} {
	data := map[string]interface{}{
		"id":              rec.ID.String(),
		"code":            rec.Code,
		"providerSkuCode": rec.ProviderSku,
		"name":            rec.Name,
		"processTime":     rec.ProcessTime,
		"isActive":        rec.IsActive,
		"isFeatured":      rec.IsFeatured,
		"stockStatus":     rec.StockStatus,
		"product": map[string]interface{}{
			"code":  rec.ProductCode,
			"title": rec.ProductTitle,
		},
		"provider": map[string]interface{}{
			"code": rec.ProviderCode,
			"name": rec.ProviderName,
		},
		"stats": map[string]interface{}{
			"todaySold": stats.TodaySold,
			"totalSold": stats.TotalSold,
		},
		"createdAt": rec.CreatedAt.Format(time.RFC3339),
		"updatedAt": rec.UpdatedAt.Format(time.RFC3339),
	}

	if rec.Description.Valid {
		data["description"] = rec.Description.String
	}
	if rec.Image.Valid {
		data["image"] = rec.Image.String
	}
	if rec.Info.Valid {
		data["info"] = rec.Info.String
	}
	if rec.SectionCode.Valid {
		data["section"] = map[string]interface{}{
			"code":  rec.SectionCode.String,
			"title": rec.SectionTitle.String,
		}
	}
	if rec.BadgeText.Valid || rec.BadgeColor.Valid {
		data["badge"] = map[string]interface{}{
			"text":  rec.BadgeText.String,
			"color": rec.BadgeColor.String,
		}
	}

	pricingMap := map[string]map[string]interface{}{}
	for _, price := range pricing {
		pricingMap[price.RegionCode] = map[string]interface{}{
			"currency":      price.Currency,
			"buyPrice":      price.BuyPrice,
			"sellPrice":     price.SellPrice,
			"originalPrice": price.Original,
			"margin":        price.Margin,
			"discount":      price.Discount,
		}
	}
	data["pricing"] = pricingMap
	return data
}

func loadSKUByIdentifier(ctx context.Context, deps *Dependencies, identifier string) (*skuRecord, error) {
	var rec skuRecord
	query := `
		SELECT
			s.id, s.code, s.provider_sku_code, s.name, s.description, s.image, s.info,
			p.id as product_id, p.code as product_code, p.title,
			pr.id as provider_id, pr.code as provider_code, pr.name,
			sc.id as section_id, sc.code as section_code, sc.title,
			s.process_time, s.is_active, s.is_featured, s.stock_status,
			s.badge_text, s.badge_color,
			s.total_sold, s.created_at, s.updated_at
		FROM skus s
		JOIN products p ON s.product_id = p.id
		JOIN providers pr ON s.provider_id = pr.id
		LEFT JOIN sections sc ON s.section_id = sc.id
		WHERE s.id::text = $1 OR LOWER(s.code) = LOWER($1)
	`
	if err := deps.DB.Pool.QueryRow(ctx, query, identifier).Scan(
		&rec.ID, &rec.Code, &rec.ProviderSku, &rec.Name, &rec.Description, &rec.Image, &rec.Info,
		&rec.ProductID, &rec.ProductCode, &rec.ProductTitle,
		&rec.ProviderID, &rec.ProviderCode, &rec.ProviderName,
		&rec.SectionID, &rec.SectionCode, &rec.SectionTitle,
		&rec.ProcessTime, &rec.IsActive, &rec.IsFeatured, &rec.StockStatus,
		&rec.BadgeText, &rec.BadgeColor,
		&rec.TotalSold, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &rec, nil
}

func getProductIDByCode(ctx context.Context, deps *Dependencies, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := deps.DB.Pool.QueryRow(ctx, `SELECT id FROM products WHERE LOWER(code) = LOWER($1)`, strings.TrimSpace(code)).Scan(&id)
	return id, err
}

func getProviderIDByCode(ctx context.Context, deps *Dependencies, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := deps.DB.Pool.QueryRow(ctx, `SELECT id FROM providers WHERE LOWER(code) = LOWER($1)`, strings.TrimSpace(code)).Scan(&id)
	return id, err
}

func getSectionIDByCode(ctx context.Context, deps *Dependencies, code string) (uuid.UUID, error) {
	var id uuid.UUID
	err := deps.DB.Pool.QueryRow(ctx, `SELECT id FROM sections WHERE LOWER(code) = LOWER($1)`, strings.TrimSpace(code)).Scan(&id)
	return id, err
}

// HandleAdminGetSKUImagesImpl returns unique images from existing SKUs
func HandleAdminGetSKUImagesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		productCode := strings.TrimSpace(r.URL.Query().Get("productCode"))

		query := `
			SELECT DISTINCT s.image
			FROM skus s
			JOIN products p ON s.product_id = p.id
			WHERE s.image IS NOT NULL AND s.image != ''
		`
		args := []interface{}{}
		argPos := 1

		if productCode != "" {
			query += fmt.Sprintf(` AND LOWER(p.code) = LOWER($%d)`, argPos)
			args = append(args, productCode)
			argPos++
		}

		query += ` ORDER BY s.image ASC`

		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var images []string
		for rows.Next() {
			var image string
			if err := rows.Scan(&image); err == nil && image != "" {
				images = append(images, image)
			}
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"images": images,
		})
	}
}

func ensureSKUCodeUnique(ctx context.Context, deps *Dependencies, code string, excludeID *uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM skus WHERE LOWER(code) = LOWER($1)`
	args := []interface{}{code}
	if excludeID != nil {
		query += ` AND id <> $2`
		args = append(args, *excludeID)
	}
	query += `)`
	if err := deps.DB.Pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errors.New("SKU code already exists")
	}
	return nil
}

func ensureProviderSKUUnique(ctx context.Context, deps *Dependencies, providerID uuid.UUID, skuCode string, excludeID *uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM skus WHERE provider_id = $1 AND LOWER(provider_sku_code) = LOWER($2)`
	args := []interface{}{providerID, skuCode}
	if excludeID != nil {
		query += ` AND id <> $3`
		args = append(args, *excludeID)
	}
	query += `)`
	if err := deps.DB.Pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errors.New("Provider SKU already exists")
	}
	return nil
}

func fetchRegionCurrencies(ctx context.Context, deps *Dependencies, regions []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(regions) == 0 {
		return result, nil
	}

	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT code, currency FROM regions WHERE code = ANY($1::region_code[])
	`, regions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var code, currency string
		if err := rows.Scan(&code, &currency); err != nil {
			return nil, err
		}
		result[code] = currency
	}

	if len(result) != len(regions) {
		return nil, errors.New("One or more region codes are invalid")
	}
	return result, nil
}

func parsePriceNumber(num json.Number) (int64, error) {
	if num.String() == "" {
		return 0, errors.New("price is required")
	}
	if strings.Contains(num.String(), ".") {
		return 0, errors.New("price must be integer (smallest currency unit)")
	}
	value, err := num.Int64()
	if err != nil {
		return 0, errors.New("price must be integer (smallest currency unit)")
	}
	if value < 0 {
		return 0, errors.New("price must be positive")
	}
	return value, nil
}

func nullableUUID(id *uuid.UUID) interface{} {
	if id == nil {
		return nil
	}
	return *id
}

func nullStringFromPtr(b *skuBadgePayload, getter func(*skuBadgePayload) string) sql.NullString {
	if b == nil {
		return sql.NullString{}
	}
	return nullString(getter(b))
}

func mapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, strings.TrimSpace(k))
	}
	return keys
}

func parseQueryInt(r *http.Request, key string, defaultValue int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		return parsed
	}
	return defaultValue
}
