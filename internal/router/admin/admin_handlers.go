package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gate-v2/internal/middleware"
	"gate-v2/internal/storage"
	"gate-v2/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var allowedAdminStatuses = map[string]struct{}{
	"ACTIVE":    {},
	"INACTIVE":  {},
	"SUSPENDED": {},
}

type roleInfo struct {
	ID    uuid.UUID
	Code  string
	Name  string
	Level int
}

type productPayload struct {
	Code         string   `json:"code"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Subtitle     string   `json:"subtitle"`
	Description  string   `json:"description"`
	Publisher    string   `json:"publisher"`
	CategoryCode string   `json:"categoryCode"`
	IsActive     bool     `json:"isActive"`
	IsPopular    bool     `json:"isPopular"`
	InquirySlug  string   `json:"inquirySlug"`
	Regions      []string `json:"regions"`
	Features     []string `json:"features"`
	HowToOrder   []string `json:"howToOrder"`
	Tags         []string `json:"tags"`
}

type productFieldPayload struct {
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Required    bool     `json:"required"`
	MinLength   *int     `json:"minLength"`
	MaxLength   *int     `json:"maxLength"`
	Placeholder string   `json:"placeholder"`
	Hint        string   `json:"hint"`
	Options     []string `json:"options"`
	SortOrder   int      `json:"sortOrder"`
}

func (p productFieldPayload) SortOrderOrDefault(idx int) int {
	if p.SortOrder != 0 {
		return p.SortOrder
	}
	return idx + 1
}

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func fetchRoleByCode(ctx context.Context, deps *Dependencies, code string) (*roleInfo, error) {
	var info roleInfo
	err := deps.DB.Pool.QueryRow(ctx, `
		SELECT id, code, name, level
		FROM roles
		WHERE code = $1
	`, code).Scan(&info.ID, &info.Code, &info.Name, &info.Level)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func fetchAdminDetail(ctx context.Context, deps *Dependencies, adminID uuid.UUID) (map[string]interface{}, error) {
	var (
		id            uuid.UUID
		name          string
		email         string
		phoneNumber   sql.NullString
		status        string
		mfaEnabled    bool
		createdAt     time.Time
		updatedAt     time.Time
		lastLoginAt   sql.NullTime
		roleCode      string
		roleName      string
		roleLevel     int
		createdByID   sql.NullString
		createdByName sql.NullString
	)

	err := deps.DB.Pool.QueryRow(ctx, `
		SELECT 
			a.id, a.name, a.email, a.phone_number,
			a.status, a.mfa_enabled, a.created_at, a.updated_at, a.last_login_at,
			r.code as role_code, r.name as role_name, r.level as role_level,
			creator.id as created_by_id, creator.name as created_by_name
		FROM admins a
		JOIN roles r ON a.role_id = r.id
		LEFT JOIN admins creator ON a.created_by = creator.id
		WHERE a.id = $1
	`, adminID).Scan(
		&id, &name, &email, &phoneNumber, &status, &mfaEnabled,
		&createdAt, &updatedAt, &lastLoginAt,
		&roleCode, &roleName, &roleLevel,
		&createdByID, &createdByName,
	)
	if err != nil {
		return nil, err
	}

	permRows, err := deps.DB.Pool.Query(ctx, `
		SELECT p.code
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON rp.role_id = r.id
		JOIN admins a ON a.role_id = r.id
		WHERE a.id = $1
		ORDER BY p.code
	`, adminID)
	if err != nil {
		return nil, err
	}
	defer permRows.Close()

	var permissions []string
	for permRows.Next() {
		var permCode string
		if err := permRows.Scan(&permCode); err == nil {
			permissions = append(permissions, permCode)
		}
	}

	admin := map[string]interface{}{
		"id":         id.String(),
		"name":       name,
		"email":      email,
		"status":     status,
		"mfaEnabled": mfaEnabled,
		"role": map[string]interface{}{
			"code":        roleCode,
			"name":        roleName,
			"level":       roleLevel,
			"permissions": permissions,
		},
		"createdAt": createdAt.Format(time.RFC3339),
		"updatedAt": updatedAt.Format(time.RFC3339),
	}

	if phoneNumber.Valid {
		admin["phoneNumber"] = phoneNumber.String
	}
	if lastLoginAt.Valid {
		admin["lastLoginAt"] = lastLoginAt.Time.Format(time.RFC3339)
	}
	if createdByID.Valid {
		admin["createdBy"] = map[string]interface{}{
			"id":   createdByID.String,
			"name": createdByName.String,
		}
	}

	return admin, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugSanitizer.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return uuid.New().String()
	}
	return value
}

func normalizeStringSlice(values []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, val := range values {
		normalized := strings.TrimSpace(val)
		if normalized == "" {
			continue
		}
		normalized = strings.ToUpper(normalized)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func uploadProductAsset(ctx context.Context, deps *Dependencies, fieldName string, folder storage.FolderType, r *http.Request) (string, error) {
	if deps.S3 == nil {
		return "", errors.New("S3 storage is not configured")
	}
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := storage.ValidateImageFile(header); err != nil {
		return "", err
	}

	result, err := deps.S3.Upload(ctx, folder, file, header)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

// uploadSKUImage uploads SKU image using original filename (not UUID)
func uploadSKUImage(ctx context.Context, deps *Dependencies, fieldName string, r *http.Request) (string, error) {
	if deps.S3 == nil {
		return "", errors.New("S3 storage is not configured")
	}
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := storage.ValidateImageFile(header); err != nil {
		return "", err
	}

	result, err := deps.S3.UploadWithOriginalName(ctx, storage.FolderSKU, file, header)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func deleteS3Object(ctx context.Context, deps *Dependencies, url string) {
	if deps.S3 == nil || url == "" {
		return
	}
	_ = deps.S3.DeleteByURL(ctx, url)
}

func ensureProductCodeUnique(ctx context.Context, deps *Dependencies, code string, excludeID *uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE LOWER(code) = LOWER($1)`
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
		return errors.New("PRODUCT_CODE_EXISTS")
	}
	return nil
}

func ensureProductSlugUnique(ctx context.Context, deps *Dependencies, slug string, excludeID *uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1`
	args := []interface{}{slug}
	if excludeID != nil {
		query += ` AND id <> $2`
		args = append(args, *excludeID)
	}
	query += `)`
	if err := deps.DB.Pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errors.New("PRODUCT_SLUG_EXISTS")
	}
	return nil
}

func getCategoryID(ctx context.Context, deps *Dependencies, categoryCode string) (uuid.UUID, error) {
	var categoryID uuid.UUID
	err := deps.DB.Pool.QueryRow(ctx, `
		SELECT id FROM categories WHERE code = $1
	`, categoryCode).Scan(&categoryID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, errors.New("CATEGORY_NOT_FOUND")
		}
		return uuid.Nil, err
	}
	return categoryID, nil
}

func validateRegions(ctx context.Context, deps *Dependencies, regions []string) error {
	if len(regions) == 0 {
		return errors.New("REGION_REQUIRED")
	}
	normalized := normalizeStringSlice(regions)
	if len(normalized) == 0 {
		return errors.New("REGION_REQUIRED")
	}
	var count int
	if err := deps.DB.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM regions WHERE code = ANY($1::region_code[])
	`, normalized).Scan(&count); err != nil {
		return err
	}
	if count != len(normalized) {
		return errors.New("INVALID_REGION")
	}
	return nil
}

type productRecord struct {
	ID           uuid.UUID
	Code         string
	Slug         string
	Title        string
	Subtitle     sql.NullString
	Description  sql.NullString
	Publisher    sql.NullString
	Thumbnail    sql.NullString
	Banner       sql.NullString
	CategoryID   uuid.UUID
	CategoryCode sql.NullString
	CategoryName sql.NullString
	IsActive     bool
	IsPopular    bool
	InquirySlug  sql.NullString
	Features     []string
	HowToOrder   []string
	Tags         []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func loadProductByIdentifier(ctx context.Context, deps *Dependencies, identifier string) (*productRecord, error) {
	var record productRecord
	var featuresJSON, howToJSON []byte
	query := `
		SELECT 
			p.id, p.code, p.slug, p.title, p.subtitle, p.description, p.publisher,
			p.thumbnail, p.banner, p.category_id,
			c.code as category_code, c.title as category_name,
			p.is_active, p.is_popular, p.inquiry_slug,
			COALESCE(p.features, '[]'::jsonb),
			COALESCE(p.how_to_order, '[]'::jsonb),
			COALESCE(p.tags, ARRAY[]::text[]),
			p.created_at, p.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id::text = $1 OR LOWER(p.code) = LOWER($1) OR LOWER(p.slug) = LOWER($1)
	`
	if err := deps.DB.Pool.QueryRow(ctx, query, identifier).Scan(
		&record.ID, &record.Code, &record.Slug, &record.Title, &record.Subtitle, &record.Description, &record.Publisher,
		&record.Thumbnail, &record.Banner, &record.CategoryID,
		&record.CategoryCode, &record.CategoryName,
		&record.IsActive, &record.IsPopular, &record.InquirySlug,
		&featuresJSON, &howToJSON, &record.Tags,
		&record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(featuresJSON, &record.Features); err != nil {
		record.Features = []string{}
	}
	if err := json.Unmarshal(howToJSON, &record.HowToOrder); err != nil {
		record.HowToOrder = []string{}
	}
	return &record, nil
}

func getProductRegions(ctx context.Context, deps *Dependencies, productID uuid.UUID) ([]string, error) {
	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT region_code FROM product_regions WHERE product_id = $1 ORDER BY region_code
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			regions = append(regions, code)
		}
	}
	return regions, nil
}

func replaceProductRegions(ctx context.Context, tx pgx.Tx, productID uuid.UUID, regions []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM product_regions WHERE product_id = $1`, productID); err != nil {
		return err
	}
	if len(regions) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, region := range regions {
		batch.Queue(`INSERT INTO product_regions (product_id, region_code) VALUES ($1, $2)`, productID, region)
	}
	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return err
	}
	return nil
}

func getProductStats(ctx context.Context, deps *Dependencies, productID uuid.UUID) map[string]interface{} {
	var todayTransactions int64
	var todayRevenue int64
	_ = deps.DB.Pool.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE created_at::date = CURRENT_DATE) AS today_tx,
			COALESCE(SUM(total_amount) FILTER (WHERE created_at::date = CURRENT_DATE), 0) AS today_revenue
		FROM transactions
		WHERE product_id = $1
	`, productID).Scan(&todayTransactions, &todayRevenue)

	return map[string]interface{}{
		"todayTransactions": todayTransactions,
		"todayRevenue":      todayRevenue,
	}
}

func buildProductResponse(record *productRecord, regions []string, stats map[string]interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"id":         record.ID.String(),
		"code":       record.Code,
		"slug":       record.Slug,
		"title":      record.Title,
		"isActive":   record.IsActive,
		"isPopular":  record.IsPopular,
		"regions":    regions,
		"features":   record.Features,
		"howToOrder": record.HowToOrder,
		"tags":       record.Tags,
		"stats":      stats,
		"createdAt":  record.CreatedAt.Format(time.RFC3339),
		"updatedAt":  record.UpdatedAt.Format(time.RFC3339),
	}
	if record.Subtitle.Valid {
		response["subtitle"] = record.Subtitle.String
	}
	if record.Description.Valid {
		response["description"] = record.Description.String
	}
	if record.Publisher.Valid {
		response["publisher"] = record.Publisher.String
	}
	if record.Thumbnail.Valid {
		response["thumbnail"] = record.Thumbnail.String
	}
	if record.Banner.Valid {
		response["banner"] = record.Banner.String
	}
	if record.InquirySlug.Valid {
		response["inquirySlug"] = record.InquirySlug.String
	}
	if record.CategoryCode.Valid {
		response["category"] = map[string]interface{}{
			"code":  record.CategoryCode.String,
			"title": record.CategoryName.String,
		}
	}
	return response
}

func productHasActiveSKUs(ctx context.Context, deps *Dependencies, productID uuid.UUID) (bool, error) {
	var exists bool
	if err := deps.DB.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM skus WHERE product_id = $1 AND is_active = true)
	`, productID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func productHasTransactions(ctx context.Context, deps *Dependencies, productID uuid.UUID) (bool, error) {
	var exists bool
	if err := deps.DB.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM transactions WHERE product_id = $1)
	`, productID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func fetchProductFields(ctx context.Context, deps *Dependencies, productID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT id, name, key, field_type, label, placeholder, hint,
			is_required, min_length, max_length, options, sort_order, created_at
		FROM product_fields
		WHERE product_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []map[string]interface{}
	for rows.Next() {
		var (
			id          uuid.UUID
			name        string
			key         string
			fieldType   string
			label       string
			placeholder sql.NullString
			hint        sql.NullString
			required    bool
			minLength   sql.NullInt32
			maxLength   sql.NullInt32
			optionsJSON []byte
			sortOrder   int
			createdAt   time.Time
		)
		if err := rows.Scan(&id, &name, &key, &fieldType, &label, &placeholder, &hint,
			&required, &minLength, &maxLength, &optionsJSON, &sortOrder, &createdAt); err != nil {
			continue
		}
		var options []string
		if len(optionsJSON) > 0 {
			_ = json.Unmarshal(optionsJSON, &options)
		}
		field := map[string]interface{}{
			"id":        id.String(),
			"name":      name,
			"key":       key,
			"type":      fieldType,
			"label":     label,
			"required":  required,
			"options":   options,
			"sortOrder": sortOrder,
			"createdAt": createdAt.Format(time.RFC3339),
		}
		if placeholder.Valid {
			field["placeholder"] = placeholder.String
		}
		if hint.Valid {
			field["hint"] = hint.String
		}
		// Always include minLength and maxLength, even if null
		if minLength.Valid {
			field["minLength"] = minLength.Int32
		} else {
			field["minLength"] = nil
		}
		if maxLength.Valid {
			field["maxLength"] = maxLength.Int32
		} else {
			field["maxLength"] = nil
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func normalizeAdminStatus(status string) (string, error) {
	if status == "" {
		status = "ACTIVE"
	}
	status = strings.ToUpper(status)
	if _, ok := allowedAdminStatuses[status]; !ok {
		return "", errors.New("invalid status")
	}
	return status, nil
}

func getActorRoleInfo(ctx context.Context, deps *Dependencies) (*roleInfo, error) {
	roleCode := string(middleware.GetAdminRoleFromContext(ctx))
	if roleCode == "" {
		return nil, nil
	}
	return fetchRoleByCode(ctx, deps, roleCode)
}

func canManageRole(actor *roleInfo, target *roleInfo) bool {
	if actor == nil {
		return true
	}
	return actor.Level <= target.Level
}

// ============================================
// ADMIN MANAGEMENT HANDLERS
// ============================================

func HandleGetAdminsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Parse query parameters
		limit := 10
		page := 1
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if p := r.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		offset := (page - 1) * limit

		search := r.URL.Query().Get("search")
		roleFilter := r.URL.Query().Get("role")
		statusFilter := r.URL.Query().Get("status")

		// Build query
		query := `
			SELECT 
				a.id, a.name, a.email, a.status, a.mfa_enabled, a.created_at, a.last_login_at,
				r.code as role_code, r.name as role_name, r.level as role_level
			FROM admins a
			JOIN roles r ON a.role_id = r.id
			WHERE 1=1
		`
		args := []interface{}{}
		argPos := 1

		if search != "" {
			query += ` AND (LOWER(a.name) LIKE LOWER($` + strconv.Itoa(argPos) + `) OR LOWER(a.email) LIKE LOWER($` + strconv.Itoa(argPos) + `))`
			args = append(args, "%"+search+"%")
			argPos++
		}
		if roleFilter != "" {
			query += ` AND r.code = $` + strconv.Itoa(argPos)
			args = append(args, roleFilter)
			argPos++
		}
		if statusFilter != "" {
			query += ` AND a.status = $` + strconv.Itoa(argPos)
			args = append(args, statusFilter)
			argPos++
		}

		query += ` ORDER BY a.created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
		args = append(args, limit, offset)

		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var admins []map[string]interface{}
		for rows.Next() {
			var id, name, email, status, roleCode, roleName string
			var mfaEnabled bool
			var createdAt time.Time
			var lastLoginAt sql.NullTime
			var roleLevel int

			if err := rows.Scan(&id, &name, &email, &status, &mfaEnabled, &createdAt, &lastLoginAt, &roleCode, &roleName, &roleLevel); err != nil {
				continue
			}

			admin := map[string]interface{}{
				"id":         id,
				"name":       name,
				"email":      email,
				"status":     status,
				"mfaEnabled": mfaEnabled,
				"role": map[string]interface{}{
					"code":  roleCode,
					"name":  roleName,
					"level": roleLevel,
				},
				"createdAt": createdAt.Format(time.RFC3339),
			}
			if lastLoginAt.Valid {
				admin["lastLoginAt"] = lastLoginAt.Time.Format(time.RFC3339)
			}
			admins = append(admins, admin)
		}

		// Get total count
		countQuery := `SELECT COUNT(*) FROM admins a JOIN roles r ON a.role_id = r.id WHERE 1=1`
		countArgs := []interface{}{}
		countArgPos := 1
		if search != "" {
			countQuery += ` AND (LOWER(a.name) LIKE LOWER($` + strconv.Itoa(countArgPos) + `) OR LOWER(a.email) LIKE LOWER($` + strconv.Itoa(countArgPos) + `))`
			countArgs = append(countArgs, "%"+search+"%")
			countArgPos++
		}
		if roleFilter != "" {
			countQuery += ` AND r.code = $` + strconv.Itoa(countArgPos)
			countArgs = append(countArgs, roleFilter)
			countArgPos++
		}
		if statusFilter != "" {
			countQuery += ` AND a.status = $` + strconv.Itoa(countArgPos)
			countArgs = append(countArgs, statusFilter)
			countArgPos++
		}

		var totalRows int
		if err := deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows); err != nil {
			// Log error but continue
		}

		totalPages := (totalRows + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"admins": admins,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

func HandleGetAdminImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		adminId := chi.URLParam(r, "adminId")
		if adminId == "" {
			utils.WriteBadRequestError(w, "Admin ID is required")
			return
		}

		adminUUID, err := uuid.Parse(adminId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid admin ID")
			return
		}

		admin, err := fetchAdminDetail(ctx, deps, adminUUID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				utils.WriteErrorJSON(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, admin)
	}
}

func HandleCreateAdminImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var req struct {
			Name        string `json:"name"`
			Email       string `json:"email"`
			PhoneNumber string `json:"phoneNumber"`
			Password    string `json:"password"`
			RoleCode    string `json:"roleCode"`
			Status      string `json:"status"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)
		req.PhoneNumber = strings.TrimSpace(req.PhoneNumber)
		req.RoleCode = strings.ToUpper(strings.TrimSpace(req.RoleCode))

		validationErrors := map[string]string{}
		if req.Name == "" {
			validationErrors["name"] = "Name is required"
		}
		if req.Email == "" {
			validationErrors["email"] = "Email is required"
		}
		if req.Password == "" {
			validationErrors["password"] = "Password is required"
		}
		if req.RoleCode == "" {
			validationErrors["roleCode"] = "Role code is required"
		}
		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		status, err := normalizeAdminStatus(req.Status)
		if err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"status": "Status must be ACTIVE, INACTIVE, or SUSPENDED",
			})
			return
		}

		role, err := fetchRoleByCode(ctx, deps, req.RoleCode)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				utils.WriteErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		actorRole, err := getActorRoleInfo(ctx, deps)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if actorRole != nil && !canManageRole(actorRole, role) {
			utils.WriteErrorJSON(w, http.StatusForbidden, "INVALID_ROLE_LEVEL", "Cannot create admin with higher role level", "")
			return
		}

		var exists bool
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM admins WHERE LOWER(email) = LOWER($1)
			)
		`, req.Email).Scan(&exists); err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if exists {
			utils.WriteErrorJSON(w, http.StatusConflict, "ADMIN_EXISTS", "Email already registered", "")
			return
		}

		passwordHash, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		var createdBy uuid.NullUUID
		if actorID := middleware.GetAdminIDFromContext(r.Context()); actorID != "" {
			if parsed, err := uuid.Parse(actorID); err == nil {
				createdBy = uuid.NullUUID{UUID: parsed, Valid: true}
			}
		}

		var adminID uuid.UUID
		err = deps.DB.Pool.QueryRow(ctx, `
			INSERT INTO admins (name, email, phone_number, password_hash, role_id, status, created_by)
			VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7)
			RETURNING id
		`, req.Name, req.Email, req.PhoneNumber, passwordHash, role.ID, status, createdBy).Scan(&adminID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		adminData, err := fetchAdminDetail(ctx, deps, adminID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, adminData)
	}
}

func HandleUpdateAdminImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		adminId := chi.URLParam(r, "adminId")
		if adminId == "" {
			utils.WriteBadRequestError(w, "Admin ID is required")
			return
		}

		adminUUID, err := uuid.Parse(adminId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid admin ID")
			return
		}

		var req struct {
			Name        *string `json:"name"`
			PhoneNumber *string `json:"phoneNumber"`
			RoleCode    *string `json:"roleCode"`
			Status      *string `json:"status"`
			Password    *string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		updates := []string{}
		args := []interface{}{}
		argPos := 1

		var newRole *roleInfo
		if req.RoleCode != nil {
			code := strings.ToUpper(strings.TrimSpace(*req.RoleCode))
			if code == "" {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"roleCode": "Role code cannot be empty",
				})
				return
			}
			roleInfo, err := fetchRoleByCode(ctx, deps, code)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					utils.WriteErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found", "")
					return
				}
				utils.WriteInternalServerError(w)
				return
			}
			newRole = roleInfo
		}

		var newStatus string
		if req.Status != nil {
			status, err := normalizeAdminStatus(*req.Status)
			if err != nil {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"status": "Status must be ACTIVE, INACTIVE, or SUSPENDED",
				})
				return
			}
			newStatus = status
		}

		var newPasswordHash string
		if req.Password != nil {
			if len(strings.TrimSpace(*req.Password)) < 8 {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"password": "Password must be at least 8 characters",
				})
				return
			}
			hash, err := utils.HashPassword(*req.Password)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			newPasswordHash = hash
		}

		// Fetch current admin info for level comparison
		var currentRoleID uuid.UUID
		var currentRoleCode string
		var currentRoleLevel int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT r.id, r.code, r.level
			FROM admins a
			JOIN roles r ON a.role_id = r.id
			WHERE a.id = $1
		`, adminUUID).Scan(&currentRoleID, &currentRoleCode, &currentRoleLevel)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				utils.WriteErrorJSON(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		targetRole := &roleInfo{
			ID:    currentRoleID,
			Code:  currentRoleCode,
			Level: currentRoleLevel,
		}
		if newRole != nil {
			targetRole = newRole
		}

		actorRole, err := getActorRoleInfo(ctx, deps)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if actorRole != nil && !canManageRole(actorRole, targetRole) {
			utils.WriteErrorJSON(w, http.StatusForbidden, "INVALID_ROLE_LEVEL", "Cannot update admin with higher role level", "")
			return
		}

		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"name": "Name cannot be empty",
				})
				return
			}
			updates = append(updates, `name = $`+strconv.Itoa(argPos))
			args = append(args, name)
			argPos++
		}

		if req.PhoneNumber != nil {
			updates = append(updates, `phone_number = NULLIF($`+strconv.Itoa(argPos)+`, '')`)
			args = append(args, strings.TrimSpace(*req.PhoneNumber))
			argPos++
		}

		if newRole != nil {
			updates = append(updates, `role_id = $`+strconv.Itoa(argPos))
			args = append(args, newRole.ID)
			argPos++
		}

		if newStatus != "" {
			updates = append(updates, `status = $`+strconv.Itoa(argPos))
			args = append(args, newStatus)
			argPos++
		}

		if newPasswordHash != "" {
			updates = append(updates, `password_hash = $`+strconv.Itoa(argPos))
			args = append(args, newPasswordHash)
			argPos++
		}

		if len(updates) == 0 {
			utils.WriteBadRequestError(w, "No fields to update")
			return
		}

		updates = append(updates, "updated_at = NOW()")
		args = append(args, adminUUID)

		query := `
			UPDATE admins
			SET ` + strings.Join(updates, ", ") + `
			WHERE id = $` + strconv.Itoa(argPos)

		if _, err := deps.DB.Pool.Exec(ctx, query, args...); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		adminData, err := fetchAdminDetail(ctx, deps, adminUUID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, adminData)
	}
}

func HandleDeleteAdminImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		adminId := chi.URLParam(r, "adminId")
		if adminId == "" {
			utils.WriteBadRequestError(w, "Admin ID is required")
			return
		}

		adminUUID, err := uuid.Parse(adminId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid admin ID")
			return
		}

		if actorID := middleware.GetAdminIDFromContext(r.Context()); actorID != "" && actorID == adminUUID.String() {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_OPERATION", "Cannot delete your own account", "")
			return
		}

		var targetRole roleInfo
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT r.id, r.code, r.name, r.level
			FROM admins a
			JOIN roles r ON a.role_id = r.id
			WHERE a.id = $1
		`, adminUUID).Scan(&targetRole.ID, &targetRole.Code, &targetRole.Name, &targetRole.Level)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				utils.WriteErrorJSON(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		actorRole, err := getActorRoleInfo(ctx, deps)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if actorRole != nil && !canManageRole(actorRole, &targetRole) {
			utils.WriteErrorJSON(w, http.StatusForbidden, "INVALID_ROLE_LEVEL", "Cannot delete admin with higher role level", "")
			return
		}

		result, err := deps.DB.Pool.Exec(ctx, `DELETE FROM admins WHERE id = $1`, adminUUID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if result.RowsAffected() == 0 {
			utils.WriteErrorJSON(w, http.StatusNotFound, "ADMIN_NOT_FOUND", "Admin not found", "")
			return
		}

		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Admin deleted successfully",
		})
	}
}

// ============================================
// PRODUCT MANAGEMENT HANDLERS
// ============================================

func HandleAdminGetProductsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Parse query parameters
		limit := 10
		page := 1
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if p := r.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		offset := (page - 1) * limit

		search := r.URL.Query().Get("search")
		categoryCode := r.URL.Query().Get("categoryCode")
		region := r.URL.Query().Get("region")
		isActiveStr := r.URL.Query().Get("isActive")

		// Build query
		query := `
			SELECT DISTINCT
				p.id, p.code, p.slug, p.title, p.subtitle, p.publisher, p.thumbnail, p.banner,
				c.code as category_code, c.title as category_title,
				p.is_active, p.is_popular, p.inquiry_slug, p.created_at, p.updated_at
			FROM products p
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE 1=1
		`
		args := []interface{}{}
		argPos := 1

		if search != "" {
			query += ` AND (LOWER(p.title) LIKE LOWER($` + strconv.Itoa(argPos) + `) OR LOWER(p.code) LIKE LOWER($` + strconv.Itoa(argPos) + `))`
			args = append(args, "%"+search+"%")
			argPos++
		}
		if categoryCode != "" {
			query += ` AND c.code = $` + strconv.Itoa(argPos)
			args = append(args, categoryCode)
			argPos++
		}
		if region != "" {
			query += ` AND EXISTS (
				SELECT 1 FROM product_regions pr 
				WHERE pr.product_id = p.id AND pr.region_code = $` + strconv.Itoa(argPos) + `
			)`
			args = append(args, region)
			argPos++
		}
		if isActiveStr != "" {
			isActive := strings.ToLower(isActiveStr) == "true"
			query += ` AND p.is_active = $` + strconv.Itoa(argPos)
			args = append(args, isActive)
			argPos++
		}

		query += ` ORDER BY p.created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
		args = append(args, limit, offset)

		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var products []map[string]interface{}
		for rows.Next() {
			var (
				id            uuid.UUID
				code          string
				slug          string
				title         string
				subtitle      sql.NullString
				publisher     sql.NullString
				thumbnail     sql.NullString
				banner        sql.NullString
				categoryCode  sql.NullString
				categoryTitle sql.NullString
				isActive      bool
				isPopular     bool
				inquirySlug   sql.NullString
				createdAt     time.Time
				updatedAt     time.Time
			)

			if err := rows.Scan(&id, &code, &slug, &title, &subtitle, &publisher, &thumbnail, &banner, &categoryCode, &categoryTitle, &isActive, &isPopular, &inquirySlug, &createdAt, &updatedAt); err != nil {
				continue
			}

			product := map[string]interface{}{
				"id":        id.String(),
				"code":      code,
				"slug":      slug,
				"title":     title,
				"isActive":  isActive,
				"isPopular": isPopular,
				"createdAt": createdAt.Format(time.RFC3339),
				"updatedAt": updatedAt.Format(time.RFC3339),
			}

			if inquirySlug.Valid {
				product["inquirySlug"] = inquirySlug.String
			}

			if subtitle.Valid {
				product["subtitle"] = subtitle.String
			}
			if publisher.Valid {
				product["publisher"] = publisher.String
			}
			if thumbnail.Valid {
				product["thumbnail"] = thumbnail.String
			}
			if banner.Valid {
				product["banner"] = banner.String
			}
			if categoryCode.Valid {
				product["category"] = map[string]interface{}{
					"code":  categoryCode.String,
					"title": categoryTitle.String,
				}
			}

			regions, _ := getProductRegions(ctx, deps, id)
			product["regions"] = regions

			var skuCount int
			_ = deps.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM skus WHERE product_id = $1 AND is_active = true`, id).Scan(&skuCount)
			product["skuCount"] = skuCount
			product["stats"] = getProductStats(ctx, deps, id)

			products = append(products, product)
		}

		// Get total count
		countQuery := `SELECT COUNT(DISTINCT p.id) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1`
		countArgs := []interface{}{}
		countArgPos := 1
		if search != "" {
			countQuery += ` AND (LOWER(p.title) LIKE LOWER($` + strconv.Itoa(countArgPos) + `) OR LOWER(p.code) LIKE LOWER($` + strconv.Itoa(countArgPos) + `))`
			countArgs = append(countArgs, "%"+search+"%")
			countArgPos++
		}
		if categoryCode != "" {
			countQuery += ` AND c.code = $` + strconv.Itoa(countArgPos)
			countArgs = append(countArgs, categoryCode)
			countArgPos++
		}
		if region != "" {
			countQuery += ` AND EXISTS (SELECT 1 FROM product_regions pr WHERE pr.product_id = p.id AND pr.region_code = $` + strconv.Itoa(countArgPos) + `)`
			countArgs = append(countArgs, region)
			countArgPos++
		}
		if isActiveStr != "" {
			isActive := strings.ToLower(isActiveStr) == "true"
			countQuery += ` AND p.is_active = $` + strconv.Itoa(countArgPos)
			countArgs = append(countArgs, isActive)
			countArgPos++
		}

		var totalRows int
		if err := deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows); err != nil {
			// Log error but continue
		}

		totalPages := (totalRows + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"products": products,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

func HandleAdminGetProductImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		productId := chi.URLParam(r, "productId")
		if productId == "" {
			utils.WriteBadRequestError(w, "Product ID is required")
			return
		}

		record, err := loadProductByIdentifier(ctx, deps, productId)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		regions, err := getProductRegions(ctx, deps, record.ID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		data := buildProductResponse(record, regions, getProductStats(ctx, deps, record.ID))
		utils.WriteSuccessJSON(w, data)
	}
}

func HandleCreateProductImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
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

		var payload productPayload
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			utils.WriteBadRequestError(w, "Invalid payload JSON")
			return
		}

		payload.Code = strings.ToUpper(strings.TrimSpace(payload.Code))
		if payload.Code == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Code is required"})
			return
		}

		payload.Slug = strings.TrimSpace(payload.Slug)
		if payload.Slug == "" {
			payload.Slug = slugify(payload.Title)
		} else {
			payload.Slug = slugify(payload.Slug)
		}

		payload.Title = strings.TrimSpace(payload.Title)
		if payload.Title == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"title": "Title is required"})
			return
		}

		payload.CategoryCode = strings.TrimSpace(payload.CategoryCode)
		if payload.CategoryCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"categoryCode": "Category code is required"})
			return
		}

		regions := normalizeStringSlice(payload.Regions)
		if err := validateRegions(ctx, deps, regions); err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"regions": "Invalid region codes"})
			return
		}

		categoryID, err := getCategoryID(ctx, deps, payload.CategoryCode)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "CATEGORY_NOT_FOUND", "Category not found", "")
			return
		}

		if err := ensureProductCodeUnique(ctx, deps, payload.Code, nil); err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Product code already exists"})
			return
		}

		if err := ensureProductSlugUnique(ctx, deps, payload.Slug, nil); err != nil {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"slug": "Product slug already exists"})
			return
		}

		thumbnailURL, err := uploadProductAsset(ctx, deps, "thumbnail", storage.FolderProduct, r)
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			utils.WriteBadRequestError(w, fmt.Sprintf("Thumbnail upload failed: %v", err))
			return
		}
		bannerURL, err := uploadProductAsset(ctx, deps, "banner", storage.FolderProduct, r)
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			utils.WriteBadRequestError(w, fmt.Sprintf("Banner upload failed: %v", err))
			return
		}

		if thumbnailURL == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"thumbnail": "Thumbnail file is required"})
			return
		}

		if payload.Features == nil {
			payload.Features = []string{}
		}
		if payload.HowToOrder == nil {
			payload.HowToOrder = []string{}
		}
		if payload.Tags == nil {
			payload.Tags = []string{}
		}

		featuresJSON, _ := json.Marshal(payload.Features)
		howToJSON, _ := json.Marshal(payload.HowToOrder)

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		var productID uuid.UUID
		inquirySlugValue := strings.TrimSpace(payload.InquirySlug)
		if inquirySlugValue == "" {
			inquirySlugValue = ""
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO products (
				code, slug, title, subtitle, description, publisher,
				category_id, thumbnail, banner,
				is_active, is_popular, inquiry_slug, features, how_to_order, tags
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			RETURNING id
		`, payload.Code, payload.Slug, payload.Title, payload.Subtitle, payload.Description, payload.Publisher,
			categoryID, thumbnailURL, bannerURL, payload.IsActive, payload.IsPopular,
			nullString(inquirySlugValue), featuresJSON, howToJSON, payload.Tags).Scan(&productID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if err := replaceProductRegions(ctx, tx, productID, regions); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		record, err := loadProductByIdentifier(ctx, deps, productID.String())
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		response := buildProductResponse(record, regions, getProductStats(ctx, deps, record.ID))
		utils.WriteCreatedJSON(w, response)
	}
}

func HandleUpdateProductImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		productId := chi.URLParam(r, "productId")
		if productId == "" {
			utils.WriteBadRequestError(w, "Product ID is required")
			return
		}

		record, err := loadProductByIdentifier(ctx, deps, productId)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", "")
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
		if payloadStr == "" && (r.MultipartForm == nil || (r.MultipartForm.File["thumbnail"] == nil && r.MultipartForm.File["banner"] == nil)) {
			utils.WriteBadRequestError(w, "Payload is required")
			return
		}

		var payload map[string]json.RawMessage
		if payloadStr != "" {
			if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
				utils.WriteBadRequestError(w, "Invalid payload JSON")
				return
			}
		} else {
			payload = map[string]json.RawMessage{}
		}

		updates := []string{}
		args := []interface{}{}
		argPos := 1

		if raw, ok := payload["code"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil {
				value = strings.ToUpper(strings.TrimSpace(value))
				if value != "" {
					if err := ensureProductCodeUnique(ctx, deps, value, &record.ID); err != nil {
						utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"code": "Product code already exists"})
						return
					}
					updates = append(updates, fmt.Sprintf("code = $%d", argPos))
					args = append(args, value)
					argPos++
				}
			}
		}

		if raw, ok := payload["slug"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil {
				value = slugify(value)
				if err := ensureProductSlugUnique(ctx, deps, value, &record.ID); err != nil {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"slug": "Product slug already exists"})
					return
				}
				updates = append(updates, fmt.Sprintf("slug = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		if raw, ok := payload["title"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				updates = append(updates, fmt.Sprintf("title = $%d", argPos))
				args = append(args, strings.TrimSpace(value))
				argPos++
			}
		}

		for _, field := range []struct {
			jsonKey string
			column  string
		}{
			{"subtitle", "subtitle"},
			{"description", "description"},
			{"publisher", "publisher"},
			{"inquirySlug", "inquiry_slug"},
		} {
			if raw, ok := payload[field.jsonKey]; ok {
				var value string
				if err := json.Unmarshal(raw, &value); err == nil {
					updates = append(updates, fmt.Sprintf("%s = NULLIF($%d, '')", field.column, argPos))
					args = append(args, strings.TrimSpace(value))
					argPos++
				}
			}
		}

		if raw, ok := payload["categoryCode"]; ok {
			var value string
			if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
				categoryID, err := getCategoryID(ctx, deps, strings.TrimSpace(value))
				if err != nil {
					utils.WriteErrorJSON(w, http.StatusBadRequest, "CATEGORY_NOT_FOUND", "Category not found", "")
					return
				}
				updates = append(updates, fmt.Sprintf("category_id = $%d", argPos))
				args = append(args, categoryID)
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

		if raw, ok := payload["isPopular"]; ok {
			var value bool
			if err := json.Unmarshal(raw, &value); err == nil {
				updates = append(updates, fmt.Sprintf("is_popular = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		if raw, ok := payload["features"]; ok {
			var value []string
			if err := json.Unmarshal(raw, &value); err == nil {
				data, _ := json.Marshal(value)
				updates = append(updates, fmt.Sprintf("features = $%d", argPos))
				args = append(args, data)
				argPos++
			}
		}

		if raw, ok := payload["howToOrder"]; ok {
			var value []string
			if err := json.Unmarshal(raw, &value); err == nil {
				data, _ := json.Marshal(value)
				updates = append(updates, fmt.Sprintf("how_to_order = $%d", argPos))
				args = append(args, data)
				argPos++
			}
		}

		var regions []string
		if raw, ok := payload["regions"]; ok {
			var value []string
			if err := json.Unmarshal(raw, &value); err == nil {
				regions = normalizeStringSlice(value)
				if err := validateRegions(ctx, deps, regions); err != nil {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"regions": "Invalid region codes"})
					return
				}
			}
		}

		if raw, ok := payload["tags"]; ok {
			var value []string
			if err := json.Unmarshal(raw, &value); err == nil {
				updates = append(updates, fmt.Sprintf("tags = $%d", argPos))
				args = append(args, value)
				argPos++
			}
		}

		var newThumbnail, newBanner string

		if file, header, err := r.FormFile("thumbnail"); err == nil {
			if deps.S3 == nil {
				utils.WriteInternalServerError(w)
				return
			}
			defer file.Close()
			if err := storage.ValidateImageFile(header); err != nil {
				utils.WriteBadRequestError(w, err.Error())
				return
			}
			result, err := deps.S3.Upload(ctx, storage.FolderProduct, file, header)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			newThumbnail = result.URL
			updates = append(updates, fmt.Sprintf("thumbnail = $%d", argPos))
			args = append(args, newThumbnail)
			argPos++
		}

		if file, header, err := r.FormFile("banner"); err == nil {
			if deps.S3 == nil {
				utils.WriteInternalServerError(w)
				return
			}
			defer file.Close()
			if err := storage.ValidateImageFile(header); err != nil {
				utils.WriteBadRequestError(w, err.Error())
				return
			}
			result, err := deps.S3.Upload(ctx, storage.FolderProduct, file, header)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			newBanner = result.URL
			updates = append(updates, fmt.Sprintf("banner = $%d", argPos))
			args = append(args, newBanner)
			argPos++
		}

		if len(updates) == 0 && len(regions) == 0 {
			utils.WriteBadRequestError(w, "No fields to update")
			return
		}

		updates = append(updates, "updated_at = NOW()")
		args = append(args, record.ID)

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		if len(updates) > 0 {
			query := fmt.Sprintf(`
				UPDATE products
				SET %s
				WHERE id = $%d
				RETURNING id
			`, strings.Join(updates, ", "), argPos)

			if err := tx.QueryRow(ctx, query, args...).Scan(&record.ID); err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if len(regions) > 0 {
			if err := replaceProductRegions(ctx, tx, record.ID, regions); err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if newThumbnail != "" && record.Thumbnail.Valid {
			deleteS3Object(ctx, deps, record.Thumbnail.String)
		}
		if newBanner != "" && record.Banner.Valid {
			deleteS3Object(ctx, deps, record.Banner.String)
		}

		updated, err := loadProductByIdentifier(ctx, deps, record.ID.String())
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		regions, _ = getProductRegions(ctx, deps, record.ID)
		utils.WriteSuccessJSON(w, buildProductResponse(updated, regions, getProductStats(ctx, deps, record.ID)))
	}
}

func HandleDeleteProductImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		productId := chi.URLParam(r, "productId")
		if productId == "" {
			utils.WriteBadRequestError(w, "Product ID is required")
			return
		}

		record, err := loadProductByIdentifier(ctx, deps, productId)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		if hasSKUs, err := productHasActiveSKUs(ctx, deps, record.ID); err == nil && hasSKUs {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "PRODUCT_HAS_SKUS", "Cannot delete product with active SKUs", "")
			return
		}

		if hasTx, err := productHasTransactions(ctx, deps, record.ID); err == nil && hasTx {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "PRODUCT_HAS_TRANSACTIONS", "Cannot delete product with transactions", "")
			return
		}

		if _, err := deps.DB.Pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, record.ID); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if record.Thumbnail.Valid {
			deleteS3Object(ctx, deps, record.Thumbnail.String)
		}
		if record.Banner.Valid {
			deleteS3Object(ctx, deps, record.Banner.String)
		}

		utils.WriteSuccessJSON(w, map[string]string{"message": "Product deleted successfully"})
	}
}

func HandleAdminGetFieldsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		productId := chi.URLParam(r, "productId")
		if productId == "" {
			utils.WriteBadRequestError(w, "Product ID is required")
			return
		}

		record, err := loadProductByIdentifier(ctx, deps, productId)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		fields, err := fetchProductFields(ctx, deps, record.ID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{"fields": fields})
	}
}

func HandleUpdateFieldsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		productId := chi.URLParam(r, "productId")
		if productId == "" {
			utils.WriteBadRequestError(w, "Product ID is required")
			return
		}

		record, err := loadProductByIdentifier(ctx, deps, productId)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		var req struct {
			Fields []productFieldPayload `json:"fields"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, `DELETE FROM product_fields WHERE product_id = $1`, record.ID); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		for idx, field := range req.Fields {
			field.Name = strings.TrimSpace(field.Name)
			field.Key = strings.TrimSpace(field.Key)
			field.Type = strings.ToLower(strings.TrimSpace(field.Type))
			field.Label = strings.TrimSpace(field.Label)
			if field.Name == "" || field.Key == "" || field.Type == "" || field.Label == "" {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{"fields": "Field name, key, type and label are required"})
				return
			}

			optionsJSON, _ := json.Marshal(field.Options)
			if _, err := tx.Exec(ctx, `
				INSERT INTO product_fields (
					product_id, name, key, field_type, label, placeholder, hint,
					is_required, min_length, max_length, options, sort_order
				) VALUES ($1,$2,$3,$4,$5, NULLIF($6,''), NULLIF($7,''), $8,$9,$10,$11,$12)
			`, record.ID, field.Name, field.Key, field.Type, field.Label, field.Placeholder, field.Hint,
				field.Required, field.MinLength, field.MaxLength, optionsJSON, field.SortOrderOrDefault(idx)); err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		fields, err := fetchProductFields(ctx, deps, record.ID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{"fields": fields})
	}
}
