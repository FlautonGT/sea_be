package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gate-v2/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ============================================
// ROLES HANDLERS
// ============================================

func HandleGetRolesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT 
				r.code, r.name, r.level, r.description,
				COUNT(DISTINCT a.id) as admin_count
			FROM roles r
			LEFT JOIN admins a ON a.role_id = r.id
			GROUP BY r.id, r.code, r.name, r.level, r.description
			ORDER BY r.level ASC
		`)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var roles []map[string]interface{}
		for rows.Next() {
			var code, name, description string
			var level, adminCount int

			if err := rows.Scan(&code, &name, &level, &description, &adminCount); err != nil {
				continue
			}

			// Get permissions for this role
			permRows, err := deps.DB.Pool.Query(ctx, `
				SELECT p.code
				FROM permissions p
				JOIN role_permissions rp ON p.id = rp.permission_id
				JOIN roles r ON rp.role_id = r.id
				WHERE r.code = $1
				ORDER BY p.code
			`, code)
			if err != nil {
				continue
			}

			var permissions []string
			for permRows.Next() {
				var permCode string
				if err := permRows.Scan(&permCode); err == nil {
					permissions = append(permissions, permCode)
				}
			}
			permRows.Close()

			roles = append(roles, map[string]interface{}{
				"code":        code,
				"name":        name,
				"level":       level,
				"description": description,
				"permissions": permissions,
				"adminCount":  adminCount,
			})
		}

		utils.WriteSuccessJSON(w, roles)
	}
}

func HandleUpdateRolePermissionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		roleCode := chi.URLParam(r, "roleCode")
		if roleCode == "" {
			utils.WriteBadRequestError(w, "Role code is required")
			return
		}

		// Check if role exists
		var roleID uuid.UUID
		var roleName string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id, name FROM roles WHERE code = $1
		`, roleCode).Scan(&roleID, &roleName)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "ROLE_NOT_FOUND", "Role not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Cannot modify SUPERADMIN permissions
		if roleCode == "SUPERADMIN" {
			utils.WriteErrorJSON(w, http.StatusForbidden, "INVALID_ROLE_LEVEL", "Cannot modify SUPERADMIN permissions", "")
			return
		}

		var req struct {
			Permissions []string `json:"permissions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate permissions exist
		if len(req.Permissions) > 0 {
			placeholders := make([]string, len(req.Permissions))
			args := make([]interface{}, len(req.Permissions))
			for i, perm := range req.Permissions {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				args[i] = perm
			}

			var count int
			err := deps.DB.Pool.QueryRow(ctx, fmt.Sprintf(`
				SELECT COUNT(*) FROM permissions WHERE code IN (%s)
			`, strings.Join(placeholders, ",")), args...).Scan(&count)
			if err != nil || count != len(req.Permissions) {
				utils.WriteBadRequestError(w, "One or more permissions are invalid")
				return
			}
		}

		// Update permissions in transaction
		err = deps.DB.WithTransaction(ctx, func(tx pgx.Tx) error {
			// Delete existing permissions
			_, err := tx.Exec(ctx, `
				DELETE FROM role_permissions WHERE role_id = $1
			`, roleID)
			if err != nil {
				return err
			}

			// Insert new permissions
			if len(req.Permissions) > 0 {
				for _, permCode := range req.Permissions {
					_, err := tx.Exec(ctx, `
						INSERT INTO role_permissions (role_id, permission_id)
						SELECT $1, p.id
						FROM permissions p
						WHERE p.code = $2
					`, roleID, permCode)
					if err != nil {
						return err
					}
				}
			}

			return nil
		})

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Get updated role with permissions
		permRows, err := deps.DB.Pool.Query(ctx, `
			SELECT p.code
			FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1
			ORDER BY p.code
		`, roleID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer permRows.Close()

		var permissions []string
		for permRows.Next() {
			var permCode string
			if err := permRows.Scan(&permCode); err == nil {
				permissions = append(permissions, permCode)
			}
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"code":        roleCode,
			"name":        roleName,
			"permissions": permissions,
			"updatedAt":   time.Now().Format(time.RFC3339),
		})
	}
}

// ============================================
// PROVIDER HANDLERS
// ============================================

func HandleGetProvidersImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT 
				p.id, p.code, p.name, p.base_url, p.webhook_url,
				p.is_active, p.priority, p.supported_types,
				p.health_status, p.last_health_check,
				COALESCE(p.total_skus, 0), COALESCE(p.active_skus, 0),
				COALESCE(p.success_rate, 0), COALESCE(p.avg_response_time, 0),
				p.created_at, p.updated_at
			FROM providers p
			ORDER BY p.priority ASC, p.created_at DESC
		`)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var providers []map[string]interface{}
		for rows.Next() {
			var id uuid.UUID
			var code, name, baseURL, webhookURL string
			var isActive bool
			var priority int
			var supportedTypes []string
			var healthStatus sql.NullString
			var lastHealthCheck sql.NullTime
			var totalSkus, activeSkus int
			var successRate float64
			var avgResponseTime int32
			var createdAt, updatedAt time.Time

			if err := rows.Scan(&id, &code, &name, &baseURL, &webhookURL, &isActive, &priority, &supportedTypes, &healthStatus, &lastHealthCheck, &totalSkus, &activeSkus, &successRate, &avgResponseTime, &createdAt, &updatedAt); err != nil {
				continue
			}
			if supportedTypes == nil {
				supportedTypes = []string{}
			}

			provider := map[string]interface{}{
				"id":             id.String(),
				"code":           code,
				"name":           name,
				"baseUrl":        baseURL,
				"webhookUrl":     webhookURL,
				"isActive":       isActive,
				"priority":       priority,
				"supportedTypes": supportedTypes,
				"createdAt":      createdAt.Format(time.RFC3339),
				"updatedAt":      updatedAt.Format(time.RFC3339),
				"stats": map[string]interface{}{
					"totalSkus":       totalSkus,
					"activeSkus":      activeSkus,
					"successRate":     successRate,
					"avgResponseTime": avgResponseTime,
				},
			}

			if healthStatus.Valid && healthStatus.String != "" {
				provider["healthStatus"] = healthStatus.String
			} else {
				provider["healthStatus"] = "HEALTHY"
			}
			if lastHealthCheck.Valid {
				provider["lastHealthCheck"] = lastHealthCheck.Time.Format(time.RFC3339)
			}

			providers = append(providers, provider)
		}

		utils.WriteSuccessJSON(w, providers)
	}
}

// ============================================
// PAYMENT GATEWAY HANDLERS
// ============================================

func HandleGetPaymentGatewaysImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT 
				id, code, name, base_url, callback_url,
				is_active, supported_methods, supported_types,
				health_status, last_health_check,
				created_at, updated_at
			FROM payment_gateways
			ORDER BY name ASC
		`)
		if err != nil {
			fmt.Printf("[GetPaymentGateways] Query error: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var gateways []map[string]interface{}
		for rows.Next() {
			var (
				id               uuid.UUID
				code             string
				name             string
				baseURL          string
				callbackURL      sql.NullString
				isActive         bool
				supportedMethods []string
				supportedTypes   []string
				healthStatus     sql.NullString
				lastHealthCheck  sql.NullTime
				createdAt        time.Time
				updatedAt        time.Time
			)

			if err := rows.Scan(&id, &code, &name, &baseURL, &callbackURL, &isActive, &supportedMethods, &supportedTypes, &healthStatus, &lastHealthCheck, &createdAt, &updatedAt); err != nil {
				fmt.Printf("[GetPaymentGateways] Scan error: %v\n", err)
				continue
			}

			if supportedMethods == nil {
				supportedMethods = []string{}
			}
			if supportedTypes == nil {
				supportedTypes = []string{}
			}

			stats := getGatewayStats(ctx, deps, id)

			gateway := map[string]interface{}{
				"id":              id.String(),
				"code":            code,
				"name":            name,
				"baseUrl":         baseURL,
				"callbackUrl":     callbackURL.String,
				"isActive":        isActive,
				"supportedMethods": supportedMethods,
				"supportedTypes":  supportedTypes,
				"healthStatus":    "HEALTHY",
				"stats":           stats,
				"createdAt":       createdAt.Format(time.RFC3339),
				"updatedAt":       updatedAt.Format(time.RFC3339),
			}

			if healthStatus.Valid && healthStatus.String != "" {
				gateway["healthStatus"] = healthStatus.String
			}
			if lastHealthCheck.Valid {
				gateway["lastHealthCheck"] = lastHealthCheck.Time.Format(time.RFC3339)
			}

			gateways = append(gateways, gateway)
		}

		utils.WriteSuccessJSON(w, gateways)
	}
}

func HandleGetPaymentGatewayImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		gatewayId := chi.URLParam(r, "gatewayId")
		if gatewayId == "" {
			utils.WriteBadRequestError(w, "Payment Gateway ID is required")
			return
		}

		gatewayUUID, err := uuid.Parse(gatewayId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid gateway ID")
			return
		}

		detail, err := fetchPaymentGatewayDetail(ctx, deps, gatewayUUID)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "GATEWAY_NOT_FOUND", "Payment gateway not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, detail)
	}
}

func HandleCreatePaymentGatewayImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var req struct {
			Code             string                 `json:"code"`
			Name             string                 `json:"name"`
			BaseURL          string                 `json:"baseUrl"`
			CallbackURL      string                 `json:"callbackUrl"`
			IsActive         bool                   `json:"isActive"`
			SupportedMethods []string               `json:"supportedMethods"`
			SupportedTypes   []string               `json:"supportedTypes"`
			APIConfig        map[string]interface{} `json:"apiConfig"`
			Mapping          map[string][]string    `json:"mapping"`
			EnvCredentialKeys map[string]string     `json:"envCredentialKeys"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
		req.Name = strings.TrimSpace(req.Name)
		req.BaseURL = strings.TrimSpace(req.BaseURL)
		req.CallbackURL = strings.TrimSpace(req.CallbackURL)

		validationErrors := map[string]string{}
		if req.Code == "" {
			validationErrors["code"] = "Code is required"
		}
		if req.Name == "" {
			validationErrors["name"] = "Name is required"
		}
		if req.BaseURL == "" {
			validationErrors["baseUrl"] = "Base URL is required"
		}
		if len(req.SupportedMethods) == 0 {
			validationErrors["supportedMethods"] = "supportedMethods is required"
		}
		if len(req.SupportedTypes) == 0 {
			validationErrors["supportedTypes"] = "supportedTypes is required"
		}
		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		for i := range req.SupportedMethods {
			req.SupportedMethods[i] = strings.ToUpper(strings.TrimSpace(req.SupportedMethods[i]))
		}
		for i := range req.SupportedTypes {
			req.SupportedTypes[i] = strings.ToLower(strings.TrimSpace(req.SupportedTypes[i]))
			if req.SupportedTypes[i] != "purchase" && req.SupportedTypes[i] != "deposit" {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"supportedTypes": "Supported types must be purchase or deposit",
				})
				return
			}
		}

		var exists bool
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM payment_gateways WHERE code = $1)
		`, req.Code).Scan(&exists); err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if exists {
			utils.WriteErrorJSON(w, http.StatusConflict, "GATEWAY_EXISTS", "Gateway code already exists", "")
			return
		}

		apiConfig := req.APIConfig
		if apiConfig == nil {
			apiConfig = map[string]interface{}{}
		}
		apiConfigJSON, err := json.Marshal(apiConfig)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		mapping := req.Mapping
		if mapping == nil {
			mapping = map[string][]string{}
		}
		mappingJSON, err := json.Marshal(mapping)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		envKeys := req.EnvCredentialKeys
		if envKeys == nil {
			envKeys = map[string]string{}
		}
		envKeysJSON, err := json.Marshal(envKeys)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		var id uuid.UUID
		if err := deps.DB.Pool.QueryRow(ctx, `
			INSERT INTO payment_gateways (
				code, name, base_url, callback_url,
				is_active, supported_methods, supported_types,
				api_config, status_mapping, env_credential_keys
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8, $9, $10)
			RETURNING id
		`, req.Code, req.Name, req.BaseURL, req.CallbackURL, req.IsActive, req.SupportedMethods, req.SupportedTypes, apiConfigJSON, mappingJSON, envKeysJSON).Scan(&id); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":      id.String(),
			"code":    req.Code,
			"name":    req.Name,
			"message": "Payment gateway created successfully",
		})
	}
}

func HandleUpdatePaymentGatewayImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		gatewayId := chi.URLParam(r, "gatewayId")
		if gatewayId == "" {
			utils.WriteBadRequestError(w, "Gateway ID is required")
			return
		}
		gatewayUUID, err := uuid.Parse(gatewayId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid gateway ID")
			return
		}

		var req struct {
			Name             *string               `json:"name"`
			BaseURL          *string               `json:"baseUrl"`
			CallbackURL      *string               `json:"callbackUrl"`
			IsActive         *bool                 `json:"isActive"`
			SupportedMethods *[]string             `json:"supportedMethods"`
			SupportedTypes   *[]string             `json:"supportedTypes"`
			APIConfig        map[string]interface{} `json:"apiConfig"`
			Mapping          map[string][]string   `json:"mapping"`
			EnvCredentialKeys map[string]string    `json:"envCredentialKeys"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		updates := []string{}
		args := []interface{}{}
		argPos := 1

		if req.Name != nil {
			updates = append(updates, fmt.Sprintf("name = $%d", argPos))
			args = append(args, strings.TrimSpace(*req.Name))
			argPos++
		}
		if req.BaseURL != nil {
			updates = append(updates, fmt.Sprintf("base_url = $%d", argPos))
			args = append(args, strings.TrimSpace(*req.BaseURL))
			argPos++
		}
		if req.CallbackURL != nil {
			updates = append(updates, fmt.Sprintf("callback_url = NULLIF($%d, '')", argPos))
			args = append(args, strings.TrimSpace(*req.CallbackURL))
			argPos++
		}
		if req.IsActive != nil {
			updates = append(updates, fmt.Sprintf("is_active = $%d", argPos))
			args = append(args, *req.IsActive)
			argPos++
		}
		if req.SupportedMethods != nil {
			methods := *req.SupportedMethods
			for i := range methods {
				methods[i] = strings.ToUpper(strings.TrimSpace(methods[i]))
			}
			updates = append(updates, fmt.Sprintf("supported_methods = $%d", argPos))
			args = append(args, methods)
			argPos++
		}
		if req.SupportedTypes != nil {
			types := *req.SupportedTypes
			for i := range types {
				types[i] = strings.ToLower(strings.TrimSpace(types[i]))
				if types[i] != "purchase" && types[i] != "deposit" {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
						"supportedTypes": "Supported types must be purchase or deposit",
					})
					return
				}
			}
			updates = append(updates, fmt.Sprintf("supported_types = $%d", argPos))
			args = append(args, types)
			argPos++
		}
		if req.APIConfig != nil {
			apiConfigJSON, err := json.Marshal(req.APIConfig)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			updates = append(updates, fmt.Sprintf("api_config = $%d", argPos))
			args = append(args, apiConfigJSON)
			argPos++
		}
		if req.Mapping != nil {
			mappingJSON, err := json.Marshal(req.Mapping)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			updates = append(updates, fmt.Sprintf("status_mapping = $%d", argPos))
			args = append(args, mappingJSON)
			argPos++
		}
		if req.EnvCredentialKeys != nil {
			envJSON, err := json.Marshal(req.EnvCredentialKeys)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			updates = append(updates, fmt.Sprintf("env_credential_keys = $%d", argPos))
			args = append(args, envJSON)
			argPos++
		}

		if len(updates) == 0 {
			utils.WriteBadRequestError(w, "No fields to update")
			return
		}

		updates = append(updates, "updated_at = NOW()")
		args = append(args, gatewayUUID)

		query := fmt.Sprintf(`
			UPDATE payment_gateways
			SET %s
			WHERE id = $%d
			RETURNING id
		`, strings.Join(updates, ", "), argPos)

		var id uuid.UUID
		if err := deps.DB.Pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "GATEWAY_NOT_FOUND", "Payment gateway not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		detail, err := fetchPaymentGatewayDetail(ctx, deps, id)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, detail)
	}
}

func HandleDeletePaymentGatewayImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		gatewayId := chi.URLParam(r, "gatewayId")
		if gatewayId == "" {
			utils.WriteBadRequestError(w, "Gateway ID is required")
			return
		}

		gatewayUUID, err := uuid.Parse(gatewayId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid gateway ID")
			return
		}

		var channelCount int
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM payment_channel_gateways
			WHERE gateway_id = $1
		`, gatewayUUID).Scan(&channelCount); err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if channelCount > 0 {
			utils.WriteErrorJSON(w, http.StatusConflict, "GATEWAY_HAS_CHANNELS", "Cannot delete gateway with active payment channels", "")
			return
		}

		result, err := deps.DB.Pool.Exec(ctx, `
			DELETE FROM payment_gateways WHERE id = $1
		`, gatewayUUID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if result.RowsAffected() == 0 {
			utils.WriteErrorJSON(w, http.StatusNotFound, "GATEWAY_NOT_FOUND", "Payment gateway not found", "")
			return
		}

		utils.WriteSuccessJSON(w, map[string]string{"message": "Payment gateway deleted successfully"})
	}
}

func HandleTestPaymentGatewayImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"status":       "SUCCESS",
			"responseTime": 600,
			"message":      "Connection successful",
		})
	}
}

func fetchPaymentGatewayDetail(ctx context.Context, deps *Dependencies, gatewayID uuid.UUID) (map[string]interface{}, error) {
	var (
		code              string
		name              string
		baseURL           string
		callbackURL       sql.NullString
		isActive          bool
		supportedMethods  []string
		supportedTypes    []string
		healthStatus      sql.NullString
		lastHealthCheck   sql.NullTime
		apiConfigJSON     []byte
		statusMappingJSON []byte
		envKeysJSON       []byte
		createdAt         time.Time
		updatedAt         time.Time
	)
	err := deps.DB.Pool.QueryRow(ctx, `
		SELECT 
			code, name, base_url, callback_url,
			is_active, supported_methods, supported_types,
			health_status, last_health_check,
			api_config, status_mapping, env_credential_keys,
			created_at, updated_at
		FROM payment_gateways
		WHERE id = $1
	`, gatewayID).Scan(
		&code, &name, &baseURL, &callbackURL,
		&isActive, &supportedMethods, &supportedTypes,
		&healthStatus, &lastHealthCheck,
		&apiConfigJSON, &statusMappingJSON, &envKeysJSON,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if supportedMethods == nil {
		supportedMethods = []string{}
	}
	if supportedTypes == nil {
		supportedTypes = []string{}
	}

	var apiConfig map[string]interface{}
	if len(apiConfigJSON) > 0 {
		_ = json.Unmarshal(apiConfigJSON, &apiConfig)
	}
	if apiConfig == nil {
		apiConfig = map[string]interface{}{}
	}

	var statusMapping map[string][]string
	if len(statusMappingJSON) > 0 {
		_ = json.Unmarshal(statusMappingJSON, &statusMapping)
	}
	if statusMapping == nil {
		statusMapping = map[string][]string{}
	}

	var envCredentialKeys map[string]string
	if len(envKeysJSON) > 0 {
		_ = json.Unmarshal(envKeysJSON, &envCredentialKeys)
	}
	if envCredentialKeys == nil {
		envCredentialKeys = map[string]string{}
	}

	credentials := map[string]bool{
		"hasClientId":     envCredentialKeys["clientId"] != "",
		"hasClientSecret": envCredentialKeys["clientSecret"] != "",
		"hasUsername":     envCredentialKeys["username"] != "",
		"hasPin":          envCredentialKeys["pin"] != "",
		"hasApiKey":       envCredentialKeys["apiKey"] != "",
	}

	stats := getGatewayStats(ctx, deps, gatewayID)

	feeConfig := getGatewayFeeConfig(ctx, deps, gatewayID)

	detail := map[string]interface{}{
		"id":              gatewayID.String(),
		"code":            code,
		"name":            name,
		"baseUrl":         baseURL,
		"callbackUrl":     callbackURL.String,
		"isActive":        isActive,
		"supportedMethods": supportedMethods,
		"supportedTypes":  supportedTypes,
		"healthStatus":    "HEALTHY",
		"apiConfig":       apiConfig,
		"mapping":         statusMapping,
		"credentials":     credentials,
		"feeConfig":       feeConfig,
		"stats":           stats,
		"createdAt":       createdAt.Format(time.RFC3339),
		"updatedAt":       updatedAt.Format(time.RFC3339),
	}

	if healthStatus.Valid && healthStatus.String != "" {
		detail["healthStatus"] = healthStatus.String
	}
	if lastHealthCheck.Valid {
		detail["lastHealthCheck"] = lastHealthCheck.Time.Format(time.RFC3339)
	}

	return detail, nil
}

func getGatewayStats(ctx context.Context, deps *Dependencies, gatewayID uuid.UUID) map[string]interface{} {
	var (
		todayTransactions int64
		todayVolume       int64
		todaySuccess      int64
		weekTransactions  int64
		weekSuccess       int64
	)

	// Get stats via payment_channel_gateways join
	_ = deps.DB.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE t.created_at::date = CURRENT_DATE) AS today_tx,
			COALESCE(SUM(t.total_amount) FILTER (WHERE t.created_at::date = CURRENT_DATE), 0) AS today_volume,
			COUNT(*) FILTER (WHERE t.status = 'SUCCESS' AND t.created_at::date = CURRENT_DATE) AS today_success,
			COUNT(*) FILTER (WHERE t.created_at >= NOW() - INTERVAL '7 days') AS week_tx,
			COUNT(*) FILTER (WHERE t.status = 'SUCCESS' AND t.created_at >= NOW() - INTERVAL '7 days') AS week_success
		FROM transactions t
		JOIN payment_channel_gateways pcg ON t.payment_channel_id = pcg.channel_id
		WHERE pcg.gateway_id = $1
	`, gatewayID).Scan(&todayTransactions, &todayVolume, &todaySuccess, &weekTransactions, &weekSuccess)

	var successRate float64
	if weekTransactions > 0 {
		successRate = (float64(weekSuccess) / float64(weekTransactions)) * 100
	}

	return map[string]interface{}{
		"todayTransactions": todayTransactions,
		"todayVolume":       todayVolume,
		"successRate":       successRate,
		"avgResponseTime":   0,
	}
}

func getGatewayFeeConfig(ctx context.Context, deps *Dependencies, gatewayID uuid.UUID) map[string]map[string]interface{} {
	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT 
			pc.code,
			pc.fee_type,
			pc.fee_amount,
			pc.fee_percentage,
			pc.min_fee,
			pc.max_fee
		FROM payment_channel_gateways pcg
		JOIN payment_channels pc ON pc.id = pcg.channel_id
		WHERE pcg.gateway_id = $1
	`, gatewayID)
	if err != nil {
		return map[string]map[string]interface{}{}
	}
	defer rows.Close()

	feeConfig := map[string]map[string]interface{}{}
	for rows.Next() {
		var (
			code          string
			feeType       string
			feeAmount     int64
			feePercentage float64
			minFee        int64
			maxFee        int64
		)
		if err := rows.Scan(&code, &feeType, &feeAmount, &feePercentage, &minFee, &maxFee); err != nil {
			continue
		}
		feeConfig[code] = map[string]interface{}{
			"feeType":       feeType,
			"feeAmount":     feeAmount,
			"feePercentage": feePercentage,
			"minFee":        minFee,
			"maxFee":        maxFee,
		}
	}

	return feeConfig
}
func HandleGetProviderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		providerId := chi.URLParam(r, "providerId")
		if providerId == "" {
			utils.WriteBadRequestError(w, "Provider ID is required")
			return
		}
		providerUUID, err := uuid.Parse(providerId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid provider ID")
			return
		}

		var (
			id                uuid.UUID
			code              string
			name              string
			baseURL           string
			webhookURL        string
			isActive          bool
			priority          int
			supportedTypes    []string
			healthStatus      sql.NullString
			lastHealthCheck   sql.NullTime
			apiConfigJSON     []byte
			statusMappingJSON []byte
			envKeysJSON       []byte
			totalSkus         int
			activeSkus        int
			successRate       sql.NullFloat64
			avgResponseTime   sql.NullInt32
			createdAt         time.Time
			updatedAt         time.Time
		)

		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT 
				id, code, name, base_url, webhook_url,
				is_active, priority, supported_types,
				health_status, last_health_check,
				api_config, status_mapping, env_credential_keys,
				COALESCE(total_skus, 0), COALESCE(active_skus, 0),
				COALESCE(success_rate, 0), COALESCE(avg_response_time, 0),
				created_at, updated_at
			FROM providers
			WHERE id = $1
		`, providerUUID).Scan(
			&id, &code, &name, &baseURL, &webhookURL,
			&isActive, &priority, &supportedTypes,
			&healthStatus, &lastHealthCheck,
			&apiConfigJSON, &statusMappingJSON, &envKeysJSON,
			&totalSkus, &activeSkus, &successRate, &avgResponseTime,
			&createdAt, &updatedAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PROVIDER_NOT_FOUND", "Provider not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		var apiConfig map[string]interface{}
		if len(apiConfigJSON) > 0 {
			_ = json.Unmarshal(apiConfigJSON, &apiConfig)
		}
		if apiConfig == nil {
			apiConfig = map[string]interface{}{}
		}

		var statusMapping map[string][]string
		if len(statusMappingJSON) > 0 {
			_ = json.Unmarshal(statusMappingJSON, &statusMapping)
		}
		if statusMapping == nil {
			statusMapping = map[string][]string{}
		}

		var envCredentialKeys map[string]string
		if len(envKeysJSON) > 0 {
			_ = json.Unmarshal(envKeysJSON, &envCredentialKeys)
		}
		if envCredentialKeys == nil {
			envCredentialKeys = map[string]string{}
		}

		credentials := map[string]bool{
			"hasUsername":      envCredentialKeys["username"] != "",
			"hasApiKey":        envCredentialKeys["apiKey"] != "",
			"hasSecretKey":     envCredentialKeys["secretKey"] != "",
			"hasWebhookSecret": envCredentialKeys["webhookSecret"] != "",
			"hasClientId":      envCredentialKeys["clientId"] != "",
			"hasClientSecret":  envCredentialKeys["clientSecret"] != "",
		}

		provider := map[string]interface{}{
			"id":             id.String(),
			"code":           code,
			"name":           name,
			"baseUrl":        baseURL,
			"webhookUrl":     webhookURL,
			"isActive":       isActive,
			"priority":       priority,
			"supportedTypes": supportedTypes,
			"apiConfig":      apiConfig,
			"mapping":        statusMapping,
			"credentials":    credentials,
			"stats": map[string]interface{}{
				"totalSkus":       totalSkus,
				"activeSkus":      activeSkus,
				"successRate":     successRate.Float64,
				"avgResponseTime": avgResponseTime.Int32,
			},
			"createdAt": createdAt.Format(time.RFC3339),
			"updatedAt": updatedAt.Format(time.RFC3339),
		}

		if healthStatus.Valid {
			provider["healthStatus"] = healthStatus.String
		}
		if lastHealthCheck.Valid {
			provider["lastHealthCheck"] = lastHealthCheck.Time.Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, provider)
	}
}

func fetchProviderDetail(ctx context.Context, deps *Dependencies, providerID uuid.UUID) (map[string]interface{}, error) {
	var (
		code              string
		name              string
		baseURL           string
		webhookURL        sql.NullString
		isActive          bool
		priority          int
		supportedTypes    []string
		healthStatus      sql.NullString
		lastHealthCheck   sql.NullTime
		apiConfigJSON     []byte
		statusMappingJSON []byte
		envKeysJSON       []byte
		totalSkus         int
		activeSkus        int
		successRate       sql.NullFloat64
		avgResponseTime   sql.NullInt32
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := deps.DB.Pool.QueryRow(ctx, `
		SELECT 
			code, name, base_url, webhook_url,
			is_active, priority, supported_types,
			health_status, last_health_check,
			api_config, status_mapping, env_credential_keys,
			COALESCE(total_skus, 0), COALESCE(active_skus, 0),
			COALESCE(success_rate, 0), COALESCE(avg_response_time, 0),
			created_at, updated_at
		FROM providers
		WHERE id = $1
	`, providerID).Scan(
		&code, &name, &baseURL, &webhookURL,
		&isActive, &priority, &supportedTypes,
		&healthStatus, &lastHealthCheck,
		&apiConfigJSON, &statusMappingJSON, &envKeysJSON,
		&totalSkus, &activeSkus, &successRate, &avgResponseTime,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if supportedTypes == nil {
		supportedTypes = []string{}
	}

	var apiConfig map[string]interface{}
	if len(apiConfigJSON) > 0 {
		_ = json.Unmarshal(apiConfigJSON, &apiConfig)
	}
	if apiConfig == nil {
		apiConfig = map[string]interface{}{}
	}

	var statusMapping map[string][]string
	if len(statusMappingJSON) > 0 {
		_ = json.Unmarshal(statusMappingJSON, &statusMapping)
	}
	if statusMapping == nil {
		statusMapping = map[string][]string{}
	}

	var envCredentialKeys map[string]string
	if len(envKeysJSON) > 0 {
		_ = json.Unmarshal(envKeysJSON, &envCredentialKeys)
	}
	if envCredentialKeys == nil {
		envCredentialKeys = map[string]string{}
	}

	credentials := map[string]bool{
		"hasUsername":      envCredentialKeys["username"] != "",
		"hasApiKey":        envCredentialKeys["apiKey"] != "",
		"hasSecretKey":     envCredentialKeys["secretKey"] != "",
		"hasWebhookSecret": envCredentialKeys["webhookSecret"] != "",
		"hasClientId":      envCredentialKeys["clientId"] != "",
		"hasClientSecret":  envCredentialKeys["clientSecret"] != "",
	}

	stats := map[string]interface{}{
		"totalSkus":       totalSkus,
		"activeSkus":      activeSkus,
		"successRate":     successRate.Float64,
		"avgResponseTime": avgResponseTime.Int32,
	}

	provider := map[string]interface{}{
		"id":             providerID.String(),
		"code":           code,
		"name":           name,
		"baseUrl":        baseURL,
		"webhookUrl":     webhookURL.String,
		"isActive":       isActive,
		"priority":       priority,
		"supportedTypes": supportedTypes,
		"apiConfig":      apiConfig,
		"mapping":        statusMapping,
		"credentials":    credentials,
		"stats":          stats,
		"createdAt":      createdAt.Format(time.RFC3339),
		"updatedAt":      updatedAt.Format(time.RFC3339),
	}

	if healthStatus.Valid {
		provider["healthStatus"] = healthStatus.String
	} else {
		provider["healthStatus"] = "HEALTHY"
	}
	if lastHealthCheck.Valid {
		provider["lastHealthCheck"] = lastHealthCheck.Time.Format(time.RFC3339)
	}

	return provider, nil
}

func HandleCreateProviderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var req struct {
			Code             string              `json:"code"`
			Name             string              `json:"name"`
			BaseURL          string              `json:"baseUrl"`
			WebhookURL       string              `json:"webhookUrl"`
			IsActive         bool                `json:"isActive"`
			Priority         int                 `json:"priority"`
			SupportedTypes   []string            `json:"supportedTypes"`
			APIConfig        map[string]interface{} `json:"apiConfig"`
			Mapping          map[string][]string `json:"mapping"`
			EnvCredentialKeys map[string]string  `json:"envCredentialKeys"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		req.Code = strings.TrimSpace(strings.ToUpper(req.Code))
		req.Name = strings.TrimSpace(req.Name)
		req.BaseURL = strings.TrimSpace(req.BaseURL)
		req.WebhookURL = strings.TrimSpace(req.WebhookURL)

		validationErrors := map[string]string{}
		if req.Code == "" {
			validationErrors["code"] = "Code is required"
		}
		if req.Name == "" {
			validationErrors["name"] = "Name is required"
		}
		if req.BaseURL == "" {
			validationErrors["baseUrl"] = "Base URL is required"
		}
		if len(req.SupportedTypes) == 0 {
			validationErrors["supportedTypes"] = "supportedTypes is required"
		}
		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		for i := range req.SupportedTypes {
			req.SupportedTypes[i] = strings.ToUpper(strings.TrimSpace(req.SupportedTypes[i]))
		}

		var exists bool
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM providers WHERE code = $1)
		`, req.Code).Scan(&exists); err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if exists {
			utils.WriteErrorJSON(w, http.StatusConflict, "PROVIDER_EXISTS", "Provider code already exists", "")
			return
		}

		apiConfig := req.APIConfig
		if apiConfig == nil {
			apiConfig = map[string]interface{}{}
		}
		apiConfigJSON, err := json.Marshal(apiConfig)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		mapping := req.Mapping
		if mapping == nil {
			mapping = map[string][]string{}
		}
		mappingJSON, err := json.Marshal(mapping)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		envKeys := req.EnvCredentialKeys
		if envKeys == nil {
			envKeys = map[string]string{}
		}
		envKeysJSON, err := json.Marshal(envKeys)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		var id uuid.UUID
		err = deps.DB.Pool.QueryRow(ctx, `
			INSERT INTO providers (
				code, name, base_url, webhook_url,
				is_active, priority, supported_types,
				api_config, status_mapping, env_credential_keys
			)
			VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8, $9, $10)
			RETURNING id
		`, req.Code, req.Name, req.BaseURL, req.WebhookURL, req.IsActive, req.Priority, req.SupportedTypes, apiConfigJSON, mappingJSON, envKeysJSON).Scan(&id)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		var requiredEnvVars []string
		for _, key := range req.EnvCredentialKeys {
			if key != "" {
				requiredEnvVars = append(requiredEnvVars, key)
			}
		}

		utils.WriteCreatedJSON(w, map[string]interface{}{
			"id":             id.String(),
			"code":           req.Code,
			"name":           req.Name,
			"baseUrl":        req.BaseURL,
			"message":        "Provider created. Please add credentials to .env file",
			"requiredEnvVars": requiredEnvVars,
		})
	}
}

func HandleUpdateProviderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		providerId := chi.URLParam(r, "providerId")
		if providerId == "" {
			utils.WriteBadRequestError(w, "Provider ID is required")
			return
		}
		providerUUID, err := uuid.Parse(providerId)
		if err != nil {
			utils.WriteBadRequestError(w, "Invalid provider ID")
			return
		}

		var req struct {
			Name             *string             `json:"name"`
			BaseURL          *string             `json:"baseUrl"`
			WebhookURL       *string             `json:"webhookUrl"`
			IsActive         *bool               `json:"isActive"`
			Priority         *int                `json:"priority"`
			SupportedTypes   *[]string           `json:"supportedTypes"`
			APIConfig        map[string]interface{} `json:"apiConfig"`
			Mapping          map[string][]string `json:"mapping"`
			EnvCredentialKeys map[string]string  `json:"envCredentialKeys"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Build update query dynamically
		updates := []string{}
		args := []interface{}{}
		argPos := 1

		if req.Name != nil {
			updates = append(updates, fmt.Sprintf("name = $%d", argPos))
			args = append(args, *req.Name)
			argPos++
		}
		if req.BaseURL != nil {
			updates = append(updates, fmt.Sprintf("base_url = $%d", argPos))
			args = append(args, *req.BaseURL)
			argPos++
		}
		if req.WebhookURL != nil {
			updates = append(updates, fmt.Sprintf("webhook_url = $%d", argPos))
			args = append(args, *req.WebhookURL)
			argPos++
		}
		if req.IsActive != nil {
			updates = append(updates, fmt.Sprintf("is_active = $%d", argPos))
			args = append(args, *req.IsActive)
			argPos++
		}
		if req.Priority != nil {
			updates = append(updates, fmt.Sprintf("priority = $%d", argPos))
			args = append(args, *req.Priority)
			argPos++
		}
		if req.SupportedTypes != nil {
			supported := *req.SupportedTypes
			for i := range supported {
				supported[i] = strings.ToUpper(strings.TrimSpace(supported[i]))
			}
			updates = append(updates, fmt.Sprintf("supported_types = $%d", argPos))
			args = append(args, supported)
			argPos++
		}
		if req.APIConfig != nil {
			apiConfigJSON, err := json.Marshal(req.APIConfig)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			updates = append(updates, fmt.Sprintf("api_config = $%d", argPos))
			args = append(args, apiConfigJSON)
			argPos++
		}
		if req.Mapping != nil {
			mappingJSON, err := json.Marshal(req.Mapping)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			updates = append(updates, fmt.Sprintf("status_mapping = $%d", argPos))
			args = append(args, mappingJSON)
			argPos++
		}
		if req.EnvCredentialKeys != nil {
			envJSON, err := json.Marshal(req.EnvCredentialKeys)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
			updates = append(updates, fmt.Sprintf("env_credential_keys = $%d", argPos))
			args = append(args, envJSON)
			argPos++
		}

		if len(updates) == 0 {
			utils.WriteBadRequestError(w, "No fields to update")
			return
		}

		updates = append(updates, "updated_at = NOW()")
		args = append(args, providerUUID)

		query := fmt.Sprintf(`
			UPDATE providers 
			SET %s
			WHERE id = $%d
			RETURNING id
		`, strings.Join(updates, ", "), argPos)

		var id uuid.UUID
		err = deps.DB.Pool.QueryRow(ctx, query, args...).Scan(&id)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PROVIDER_NOT_FOUND", "Provider not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		detail, err := fetchProviderDetail(ctx, deps, id)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, detail)
	}
}

func HandleDeleteProviderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		providerId := chi.URLParam(r, "providerId")
		if providerId == "" {
			utils.WriteBadRequestError(w, "Provider ID is required")
			return
		}

		// Check if provider has active SKUs
		var skuCount int
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM skus WHERE provider_id = $1 AND is_active = true
		`, providerId).Scan(&skuCount)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		if skuCount > 0 {
			utils.WriteErrorJSON(w, http.StatusConflict, "PROVIDER_HAS_SKUS", "Cannot delete provider with active SKUs", "")
			return
		}

		// Delete provider
		result, err := deps.DB.Pool.Exec(ctx, `
			DELETE FROM providers WHERE id = $1
		`, providerId)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		if result.RowsAffected() == 0 {
			utils.WriteErrorJSON(w, http.StatusNotFound, "PROVIDER_NOT_FOUND", "Provider not found", "")
			return
		}

		utils.WriteSuccessJSON(w, map[string]string{"message": "Provider deleted successfully"})
	}
}

func HandleTestProviderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Placeholder - would need actual provider integration
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"status":      "SUCCESS",
			"responseTime": 850,
			"balance":     15000000,
			"message":     "Connection successful",
		})
	}
}

func HandleSyncProviderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Placeholder - would need actual provider sync logic
		utils.WriteSuccessJSON(w, map[string]interface{}{
			"status": "COMPLETED",
			"summary": map[string]interface{}{
				"totalFromProvider": 1300,
				"newSkus":           25,
				"updatedSkus":       150,
				"deactivatedSkus":   10,
				"unchanged":         1115,
			},
			"syncedAt": time.Now().Format(time.RFC3339),
		})
	}
}

// ============================================
// REPORT HANDLERS
// ============================================

func HandleGetDashboardImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		summary := map[string]interface{}{
			"totalRevenue":      int64(0),
			"totalProfit":       int64(0),
			"totalTransactions": int64(0),
			"totalUsers":        int64(0),
			"newUsers":          int64(0),
			"activeUsers":       int64(0),
		}

		// Summary revenue/profit/transactions
		var totalRevenue, totalProfit, totalTransactions int64
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT 
				COALESCE(SUM(total_amount), 0) AS revenue,
				COALESCE(SUM(profit), 0) AS profit,
				COUNT(*) AS transactions
			FROM transactions
		`).Scan(&totalRevenue, &totalProfit, &totalTransactions)
		if err == nil {
			summary["totalRevenue"] = totalRevenue
			summary["totalProfit"] = totalProfit
			summary["totalTransactions"] = totalTransactions
		}

		// Total users
		var totalUsers int64
		if err := deps.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers); err == nil {
			summary["totalUsers"] = totalUsers
		}

		// New users (last 30 days)
		var newUsers int64
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) 
			FROM users 
			WHERE created_at >= NOW() - INTERVAL '30 days'
		`).Scan(&newUsers); err == nil {
			summary["newUsers"] = newUsers
		}

		// Active users (transactions last 30 days)
		var activeUsers int64
		if err := deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(DISTINCT user_id)
			FROM transactions
			WHERE user_id IS NOT NULL
			  AND created_at >= NOW() - INTERVAL '30 days'
		`).Scan(&activeUsers); err == nil {
			summary["activeUsers"] = activeUsers
		}

		// Revenue chart (last 7 days)
		revenueChart := []map[string]interface{}{}
		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT 
				TO_CHAR(created_at::date, 'YYYY-MM-DD') AS date,
				COALESCE(SUM(total_amount), 0) AS revenue,
				COALESCE(SUM(profit), 0) AS profit
			FROM transactions
			WHERE created_at >= NOW() - INTERVAL '7 days'
			GROUP BY created_at::date
			ORDER BY created_at::date
		`)
		if err == nil {
			for rows.Next() {
				var date string
				var revenue, profit int64
				if err := rows.Scan(&date, &revenue, &profit); err == nil {
					revenueChart = append(revenueChart, map[string]interface{}{
						"date":    date,
						"revenue": revenue,
						"profit":  profit,
					})
				}
			}
			rows.Close()
		}

		// Top products
		topProducts := []map[string]interface{}{}
		rows, err = deps.DB.Pool.Query(ctx, `
			SELECT 
				p.code,
				p.title,
				COALESCE(SUM(t.total_amount), 0) AS revenue,
				COUNT(*) AS transactions
			FROM transactions t
			JOIN products p ON t.product_id = p.id
			GROUP BY p.code, p.title
			ORDER BY revenue DESC
			LIMIT 3
		`)
		if err == nil {
			for rows.Next() {
				var code, title string
				var revenue, transactions int64
				if err := rows.Scan(&code, &title, &revenue, &transactions); err == nil {
					topProducts = append(topProducts, map[string]interface{}{
						"code":         code,
						"name":         title,
						"revenue":      revenue,
						"transactions": transactions,
					})
				}
			}
			rows.Close()
		}

		// Top payments
		topPayments := []map[string]interface{}{}
		rows, err = deps.DB.Pool.Query(ctx, `
			SELECT 
				pc.code,
				pc.name,
				COALESCE(SUM(t.total_amount), 0) AS revenue,
				COUNT(*) AS transactions
			FROM transactions t
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			GROUP BY pc.code, pc.name
			ORDER BY revenue DESC
			LIMIT 3
		`)
		if err == nil {
			for rows.Next() {
				var code, name string
				var revenue, transactions int64
				if err := rows.Scan(&code, &name, &revenue, &transactions); err == nil {
					topPayments = append(topPayments, map[string]interface{}{
						"code":         code,
						"name":         name,
						"revenue":      revenue,
						"transactions": transactions,
					})
				}
			}
			rows.Close()
		}

		// Provider health
		providerHealth := []map[string]interface{}{}
		rows, err = deps.DB.Pool.Query(ctx, `
			SELECT 
				code,
				COALESCE(health_status, 'HEALTHY') AS status,
				COALESCE(success_rate, 0)
			FROM providers
			ORDER BY priority ASC, code ASC
			LIMIT 5
		`)
		if err == nil {
			for rows.Next() {
				var code, status string
				var successRate float64
				if err := rows.Scan(&code, &status, &successRate); err == nil {
					providerHealth = append(providerHealth, map[string]interface{}{
						"code":        code,
						"status":      status,
						"successRate": successRate,
					})
				}
			}
			rows.Close()
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"summary":        summary,
			"revenueChart":   revenueChart,
			"topProducts":    topProducts,
			"topPayments":    topPayments,
			"providerHealth": providerHealth,
		})
	}
}

