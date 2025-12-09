package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"seaply/internal/utils"

	"github.com/jackc/pgx/v5"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse represents the token object in login response
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

// AdminLoginSuccessResponse represents admin login success response
type AdminLoginSuccessResponse struct {
	Token TokenResponse `json:"token"`
	Admin interface{}   `json:"admin"`
}

// MFARequiredResponse represents MFA required response
type MFARequiredResponse struct {
	Step      string `json:"step"`
	MFAToken  string `json:"mfaToken"`
	ExpiresAt string `json:"expiresAt"`
}

// AdminRow represents admin data from database
type AdminRow struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	RoleCode     string
	RoleName     string
	Status       string
	MFAEnabled   bool
	LastLoginAt  *time.Time
}

// HandleAdminLoginImpl implements the admin login logic
func HandleAdminLoginImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate request
		if req.Email == "" || req.Password == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"email":    "Email is required",
				"password": "Password is required",
			})
			return
		}

		// Query admin from database
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var admin AdminRow
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT 
				a.id, 
				a.name, 
				a.email, 
				a.password_hash, 
				r.code as role_code,
				r.name as role_name,
				a.status,
				a.mfa_enabled,
				a.last_login_at
			FROM admins a
			JOIN roles r ON a.role_id = r.id
			WHERE LOWER(a.email) = LOWER($1)
		`, req.Email).Scan(
			&admin.ID,
			&admin.Name,
			&admin.Email,
			&admin.PasswordHash,
			&admin.RoleCode,
			&admin.RoleName,
			&admin.Status,
			&admin.MFAEnabled,
			&admin.LastLoginAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
					"Email atau password salah", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Check if admin is active
		if admin.Status != "ACTIVE" {
			utils.WriteErrorJSON(w, http.StatusForbidden, "ACCOUNT_SUSPENDED",
				"Akun Anda telah dinonaktifkan", "")
			return
		}

		// Check password
		if !utils.CheckPassword(req.Password, admin.PasswordHash) {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
				"Email atau password salah", "")
			return
		}

		// Get admin permissions
		permissions, err := getAdminPermissions(ctx, deps, admin.ID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Check if MFA is enabled
		if admin.MFAEnabled {
			// Generate MFA token
			mfaToken, err := deps.JWTService.GenerateMFAToken(admin.ID, "admin")
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			expiresAt := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
			utils.WriteSuccessJSON(w, MFARequiredResponse{
				Step:      "MFA_VERIFICATION",
				MFAToken:  mfaToken,
				ExpiresAt: expiresAt,
			})
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := generateAdminTokens(deps, admin, permissions)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Update last login
		_, _ = deps.DB.Pool.Exec(ctx, `
			UPDATE admins SET last_login_at = NOW() WHERE id = $1
		`, admin.ID)

		// Format lastLoginAt
		var lastLoginAt string
		if admin.LastLoginAt != nil {
			lastLoginAt = admin.LastLoginAt.Format(time.RFC3339)
		} else {
			lastLoginAt = time.Now().Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, AdminLoginSuccessResponse{
			Token: TokenResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				ExpiresIn:    int64(deps.Config.JWT.AccessTokenExpiry.Seconds()),
				TokenType:    "Bearer",
			},
			Admin: map[string]interface{}{
				"id":    admin.ID,
				"name":  admin.Name,
				"email": admin.Email,
				"role": map[string]interface{}{
					"code": admin.RoleCode,
					"name": admin.RoleName,
				},
				"status":      admin.Status,
				"lastLoginAt": lastLoginAt,
			},
		})
	}
}

// getAdminPermissions retrieves admin permissions from database
func getAdminPermissions(ctx context.Context, deps *Dependencies, adminID string) ([]string, error) {
	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT p.code
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		JOIN admins a ON rp.role_id = a.role_id
		WHERE a.id = $1
	`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			continue
		}
		permissions = append(permissions, code)
	}

	return permissions, nil
}

// generateAdminTokens generates access and refresh tokens for admin
func generateAdminTokens(deps *Dependencies, admin AdminRow, permissions []string) (string, string, error) {
	claims := utils.TokenClaims{
		UserID:      admin.ID,
		Type:        "admin",
		Email:       admin.Email,
		Role:        admin.RoleCode,
		Permissions: permissions,
	}

	accessToken, err := deps.JWTService.GenerateAccessToken(claims)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := deps.JWTService.GenerateRefreshToken(admin.ID, "admin")
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
