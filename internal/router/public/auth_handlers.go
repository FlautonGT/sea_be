package public

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gate-v2/internal/utils"
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
	PhoneNumber     string `json:"phoneNumber"`
	PrimaryRegion   string `json:"primaryRegion"`
}

// VerifyEmailRequest represents the email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	PhoneNumber    string `json:"phoneNumber"`
	ProfilePicture string `json:"profilePicture"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ResendVerificationRequest represents the resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordRequest represents the forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Token           string `json:"token"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
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

// UserLoginSuccessResponse represents user login success response
type UserLoginSuccessResponse struct {
	Step  string        `json:"step"`
	Token TokenResponse `json:"token"`
	User  interface{}   `json:"user"`
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

// UserRow represents user data from database
type UserRow struct {
	ID              string
	FirstName       string
	LastName        *string
	Email           string
	PasswordHash    *string
	PhoneNumber     *string
	Status          string
	ProfilePicture  *string
	PrimaryRegion   string
	MFAStatus       string
	MembershipLevel string
	BalanceIDR      int64
	BalanceMYR      int64
	BalancePHP      int64
	BalanceSGD      int64
	BalanceTHB      int64
	TotalTransactions int64
	TotalSpentIDR   int64
	EmailVerifiedAt *time.Time
}

// handleAdminLoginImpl implements the admin login logic
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
			// Admin not found or database error
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
				"Email atau password salah", "")
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

// handleUserLoginImpl implements the user login logic
func HandleUserLoginImpl(deps *Dependencies) http.HandlerFunc {
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

		// Query user from database
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var user UserRow
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT 
				id, 
				first_name, 
				last_name, 
				email, 
				password_hash, 
				status,
				profile_picture,
				primary_region,
				mfa_status,
				membership_level
			FROM users
			WHERE LOWER(email) = LOWER($1)
		`, req.Email).Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.PasswordHash,
			&user.Status,
			&user.ProfilePicture,
			&user.PrimaryRegion,
			&user.MFAStatus,
			&user.MembershipLevel,
		)

		if err != nil {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
				"Email atau password salah", "")
			return
		}

		// Check if user has password (not OAuth user)
		if user.PasswordHash == nil || *user.PasswordHash == "" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "OAUTH_ACCOUNT",
				"Akun ini terdaftar menggunakan Google. Silakan login dengan Google.", "")
			return
		}

		// Check if user is active
		if user.Status != "ACTIVE" {
			if user.Status == "INACTIVE" {
				// Check if verification token exists and is valid
				tokenKey := deps.Redis.ValidationTokenKey(user.ID)
				var existingToken string
				err := deps.Redis.Get(r.Context(), tokenKey, &existingToken)
				tokenExpired := true

				if err == nil && existingToken != "" {
					// Check if token is still valid
					_, err := deps.JWTService.ValidateValidationToken(existingToken)
					if err == nil {
						tokenExpired = false
					}
				}

				// If token expired, resend verification email
				if tokenExpired {
					// Generate new email verification token
					verificationToken, err := deps.JWTService.GenerateValidationToken(map[string]interface{}{
						"userId": user.ID,
						"email":  user.Email,
						"type":   "email_verification",
					})
					if err == nil {
						// Store verification token in Redis (30 minutes expiry)
						_ = deps.Redis.Set(r.Context(), tokenKey, verificationToken, 30*time.Minute)

						// Send verification email in background
						go func() {
							emailService := deps.EmailService
							if emailService != nil {
								_ = emailService.SendVerificationEmail(user.Email, user.FirstName, verificationToken)
							}
						}()
					}
				}

				utils.WriteErrorJSON(w, http.StatusForbidden, "EMAIL_NOT_VERIFIED",
					"Email belum diverifikasi. Link verifikasi telah dikirim ke email Anda.", "")
				return
			}
			utils.WriteErrorJSON(w, http.StatusForbidden, "ACCOUNT_SUSPENDED",
				"Akun Anda telah dinonaktifkan", "")
			return
		}

		// Check password
		if !utils.CheckPassword(req.Password, *user.PasswordHash) {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS",
				"Email atau password salah", "")
			return
		}

		// Check if MFA is enabled
		if user.MFAStatus == "ACTIVE" {
			// Generate MFA token
			mfaToken, err := deps.JWTService.GenerateMFAToken(user.ID, "user")
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
		accessToken, refreshToken, err := generateUserTokens(deps, user)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Update last login
		_, _ = deps.DB.Pool.Exec(ctx, `
			UPDATE users SET last_login_at = NOW() WHERE id = $1
		`, user.ID)

		// Build user response
		lastName := ""
		if user.LastName != nil {
			lastName = *user.LastName
		}
		profilePic := ""
		if user.ProfilePicture != nil {
			profilePic = *user.ProfilePicture
		}

		utils.WriteSuccessJSON(w, UserLoginSuccessResponse{
			Step: "SUCCESS",
			Token: TokenResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				ExpiresIn:    int64(deps.Config.JWT.AccessTokenExpiry.Seconds()),
				TokenType:    "Bearer",
			},
			User: map[string]interface{}{
				"id":             user.ID,
				"firstName":     user.FirstName,
				"lastName":      lastName,
				"email":         user.Email,
				"profilePicture": profilePic,
				"status":        user.Status,
				"primaryRegion": user.PrimaryRegion,
				"membership": map[string]interface{}{
					"level": user.MembershipLevel,
					"name":  getMembershipName(user.MembershipLevel),
				},
				"mfaStatus": user.MFAStatus,
			},
		})
	}
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

// getAdminPermissions retrieves admin permissions from database
func getAdminPermissions(ctx context.Context, deps *Dependencies, adminID string) ([]string, error) {
	rows, err := deps.DB.Pool.Query(ctx, `
		SELECT p.code
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN admins a ON a.role_id = rp.role_id
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
			return nil, err
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

// generateUserTokens generates access and refresh tokens for user
func generateUserTokens(deps *Dependencies, user UserRow) (string, string, error) {
	claims := utils.TokenClaims{
		UserID: user.ID,
		Type:   "user",
		Email:  user.Email,
	}

	accessToken, err := deps.JWTService.GenerateAccessToken(claims)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := deps.JWTService.GenerateRefreshToken(user.ID, "user")
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// extractBearerToken extracts the token from Authorization header
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

