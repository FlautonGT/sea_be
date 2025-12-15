package user

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"seaply/internal/domain"
	"seaply/internal/middleware"
	"seaply/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// ============================================
// USER AUTHENTICATION HANDLERS
// ============================================

// handleRegisterImpl implements user registration
func HandleRegisterImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Get region from query parameter (default: ID)
		regionParam := r.URL.Query().Get("region")
		if regionParam == "" {
			regionParam = "ID"
		}
		regionParam = strings.ToUpper(strings.TrimSpace(regionParam))

		// Sanitize inputs
		req.Email = utils.SanitizeEmail(req.Email)
		req.FirstName = utils.SanitizeString(req.FirstName)
		req.LastName = utils.SanitizeString(req.LastName)
		req.PhoneNumber = utils.SanitizeString(req.PhoneNumber)
		req.PrimaryRegion = strings.ToUpper(utils.SanitizeString(req.PrimaryRegion))

		// Use region from query parameter if primaryRegion not provided
		if req.PrimaryRegion == "" {
			req.PrimaryRegion = regionParam
		}

		// Validate required fields
		validationErrors := make(map[string]string)
		if req.FirstName == "" {
			validationErrors["firstName"] = "First name is required"
		}
		if req.Email == "" {
			validationErrors["email"] = "Email is required"
		} else if !utils.ValidateEmail(req.Email) {
			validationErrors["email"] = "Invalid email format"
		}
		if req.Password == "" {
			validationErrors["password"] = "Password is required"
		} else if !validatePasswordStrength(req.Password) {
			validationErrors["password"] = "Password must be at least 8 characters and contain uppercase, lowercase, and number"
		}
		if req.ConfirmPassword == "" {
			validationErrors["confirmPassword"] = "Confirm password is required"
		} else if req.Password != req.ConfirmPassword {
			validationErrors["confirmPassword"] = "Passwords do not match"
		}
		if req.PrimaryRegion == "" {
			req.PrimaryRegion = "ID" // Default to Indonesia
		} else if !utils.ValidateRegion(req.PrimaryRegion) {
			validationErrors["primaryRegion"] = "Invalid region code"
		}

		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Check if email already exists
		var existingID string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id FROM users WHERE LOWER(email) = LOWER($1)
		`, req.Email).Scan(&existingID)

		if err == nil {
			utils.WriteErrorJSON(w, http.StatusConflict, "EMAIL_EXISTS",
				"Email sudah terdaftar. Silakan login atau gunakan email lain.", "")
			return
		} else if err != pgx.ErrNoRows {
			fmt.Printf("Error checking email existence: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Check if phone number already exists (if provided)
		if req.PhoneNumber != "" {
			var existingPhoneID string
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT id FROM users WHERE phone_number = $1
			`, req.PhoneNumber).Scan(&existingPhoneID)

			if err == nil {
				utils.WriteErrorJSON(w, http.StatusConflict, "PHONE_EXISTS",
					"Nomor telepon sudah terdaftar. Silakan gunakan nomor telepon lain.", "")
				return
			} else if err != pgx.ErrNoRows {
				fmt.Printf("Error checking phone number existence: %v\n", err)
				utils.WriteInternalServerError(w)
				return
			}
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create user with INACTIVE status
		// Set current_region same as primary_region at registration
		var userID string
		var createdAt time.Time
		err = deps.DB.Pool.QueryRow(ctx, `
			INSERT INTO users (
				first_name, last_name, email, password_hash,
				phone_number, status, primary_region, current_region,
				membership_level, mfa_status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at
		`, req.FirstName, nullString(req.LastName), req.Email, hashedPassword,
			nullString(req.PhoneNumber), "INACTIVE", req.PrimaryRegion, req.PrimaryRegion,
			"CLASSIC", "INACTIVE").Scan(&userID, &createdAt)

		if err != nil {
			fmt.Printf("Error inserting user: %v\n", err)
			fmt.Printf("User data: FirstName=%s, LastName=%s, Email=%s, PhoneNumber=%s, PrimaryRegion=%s\n",
				req.FirstName, req.LastName, req.Email, req.PhoneNumber, req.PrimaryRegion)

			// Check if unique constraint violation
			errStr := err.Error()
			if strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key") {
				if strings.Contains(errStr, "email") || strings.Contains(errStr, "users_email") {
					utils.WriteErrorJSON(w, http.StatusConflict, "EMAIL_EXISTS",
						"Email sudah terdaftar. Silakan login atau gunakan email lain.", "")
					return
				}
				if strings.Contains(errStr, "phone") || strings.Contains(errStr, "users_phone") {
					utils.WriteErrorJSON(w, http.StatusConflict, "PHONE_EXISTS",
						"Nomor telepon sudah terdaftar. Silakan gunakan nomor telepon lain.", "")
					return
				}
			}

			utils.WriteInternalServerError(w)
			return
		}

		// Generate email verification token
		verificationToken, err := deps.JWTService.GenerateValidationToken(map[string]interface{}{
			"userId": userID,
			"email":  req.Email,
			"type":   "email_verification",
		})
		if err != nil {
			fmt.Printf("Error generating verification token: %v\n", err)
			utils.WriteInternalServerError(w)
			return
		}

		// Store verification token in Redis (30 minutes expiry)
		tokenKey := deps.Redis.ValidationTokenKey(userID)
		err = deps.Redis.Set(ctx, tokenKey, verificationToken, 30*time.Minute)
		if err != nil {
			// Log error but continue - user can resend verification
		}

		// Send verification email
		go func() {
			emailService := deps.EmailService
			if emailService != nil {
				err := emailService.SendVerificationEmail(req.Email, req.FirstName, verificationToken)
				if err != nil {
					// Log error but don't fail registration
					// Error sending email is non-critical, user can resend verification
					fmt.Printf("Failed to send verification email to %s: %v\n", req.Email, err)
				}
			}
		}()

		// Get currency from region
		currency := "IDR"
		var regionCurrency string
		err = deps.DB.Pool.QueryRow(ctx, `SELECT currency FROM regions WHERE code = $1`, req.PrimaryRegion).Scan(&regionCurrency)
		if err == nil {
			currency = regionCurrency
		}

		// Build user response object
		userResponse := map[string]interface{}{
			"id":             userID,
			"firstName":      req.FirstName,
			"lastName":       req.LastName,
			"email":          req.Email,
			"phoneNumber":    req.PhoneNumber,
			"profilePicture": nil,
			"status":         "INACTIVE",
			"primaryRegion":  req.PrimaryRegion,
			"currentRegion":  req.PrimaryRegion, // Same as primaryRegion at registration
			"currency":       currency,
			"balance": map[string]interface{}{
				"IDR": int64(0),
				"MYR": int64(0),
				"PHP": int64(0),
				"SGD": int64(0),
				"THB": int64(0),
			},
			"membership": map[string]interface{}{
				"level": "CLASSIC",
				"name":  "Classic",
			},
			"mfaStatus": "INACTIVE",
			"createdAt": createdAt.Format(time.RFC3339),
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"step": "EMAIL_VERIFICATION",
			"user": userResponse,
		})
	}
}

// handleRegisterGoogleImpl implements Google OAuth registration
func HandleRegisterGoogleImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.GoogleAuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.IDToken == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"idToken": "ID token is required",
			})
			return
		}

		// Get region from query parameter (default: ID)
		regionParam := r.URL.Query().Get("region")
		if regionParam == "" {
			regionParam = "ID"
		}
		regionParam = strings.ToUpper(strings.TrimSpace(regionParam))
		if !utils.ValidateRegion(regionParam) {
			regionParam = "ID"
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Decode Google ID token (without verification for now)
		// WARNING: In production, MUST verify the token with Google's public keys
		// TODO: Use google.golang.org/api/oauth2/v2 or similar library for proper verification
		tokenParts := strings.Split(req.IDToken, ".")
		if len(tokenParts) != 3 {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Invalid Google ID token format", "")
			return
		}

		// Decode payload (second part of JWT)
		payload, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
		if err != nil {
			fmt.Printf("Error decoding Google token payload: %v\n", err)
			fmt.Printf("Token parts count: %d, Payload part: %s\n", len(tokenParts), tokenParts[1])
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Failed to decode token payload", "")
			return
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			fmt.Printf("Error parsing Google token claims: %v\n", err)
			fmt.Printf("Payload: %s\n", string(payload))
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Failed to parse token claims", "")
			return
		}

		// Extract user info from Google token claims
		googleEmail, _ := claims["email"].(string)
		googleName, _ := claims["name"].(string)
		googlePicture, _ := claims["picture"].(string)
		googleID, _ := claims["sub"].(string) // Google user ID

		if googleEmail == "" || googleID == "" {
			fmt.Printf("Missing required fields in Google token. Email: %s, GoogleID: %s\n", googleEmail, googleID)
			fmt.Printf("Claims: %+v\n", claims)
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Missing required fields in Google token", "")
			return
		}

		// Set default name if not provided
		if googleName == "" {
			googleName = googleEmail
		}

		// Check if user already exists by email or google_id
		var existingUserID string
		var existingGoogleID sql.NullString
		var existingStatus string
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT id, google_id, status
			FROM users
			WHERE LOWER(email) = LOWER($1) OR google_id = $2
			LIMIT 1
		`, googleEmail, googleID).Scan(&existingUserID, &existingGoogleID, &existingStatus)

		if err == nil {
			// User exists
			if existingGoogleID.Valid && existingGoogleID.String == googleID {
				// User already registered with Google - login instead
				// TODO: Generate tokens and return login response
				utils.WriteErrorJSON(w, http.StatusConflict, "USER_EXISTS",
					"User already registered. Please login instead.", "")
				return
			} else {
				// Email exists but not with Google - conflict
				utils.WriteErrorJSON(w, http.StatusConflict, "EMAIL_EXISTS",
					"Email already registered. Please login with password or use different email.", "")
				return
			}
		} else if err != pgx.ErrNoRows {
			fmt.Printf("Error checking existing user for Google registration: %v\n", err)
			fmt.Printf("Google email: %s, Google ID: %s\n", googleEmail, googleID)
			utils.WriteInternalServerError(w)
			return
		}

		// Split name into first and last name
		nameParts := strings.Fields(googleName)
		firstName := googleName
		lastName := ""
		if len(nameParts) > 1 {
			firstName = strings.Join(nameParts[:len(nameParts)-1], " ")
			lastName = nameParts[len(nameParts)-1]
		}

		// Get currency from region
		currency := "IDR"
		var regionCurrency string
		err = deps.DB.Pool.QueryRow(ctx, `SELECT currency FROM regions WHERE code = $1`, regionParam).Scan(&regionCurrency)
		if err == nil {
			currency = regionCurrency
		}

		// Create user with ACTIVE status (Google verified)
		var userID string
		var createdAt time.Time
		err = deps.DB.Pool.QueryRow(ctx, `
			INSERT INTO users (
				first_name, last_name, email, password_hash,
				phone_number, status, primary_region, current_region,
				membership_level, mfa_status, google_id, profile_picture,
				email_verified_at
			) VALUES ($1, $2, $3, NULL, NULL, $4, $5, $6, $7, $8, $9, $10, NOW())
			RETURNING id, created_at
		`, firstName, nullString(lastName), googleEmail, "ACTIVE", regionParam, regionParam,
			"CLASSIC", "INACTIVE", googleID, nullString(googlePicture)).Scan(&userID, &createdAt)

		if err != nil {
			fmt.Printf("Error inserting Google user: %v\n", err)
			fmt.Printf("User data: FirstName=%s, LastName=%s, Email=%s, GoogleID=%s, PrimaryRegion=%s\n",
				firstName, lastName, googleEmail, googleID, regionParam)

			// Check if unique constraint violation
			errStr := err.Error()
			if strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "duplicate key") {
				if strings.Contains(errStr, "email") || strings.Contains(errStr, "users_email") {
					utils.WriteErrorJSON(w, http.StatusConflict, "EMAIL_EXISTS",
						"Email sudah terdaftar. Silakan login atau gunakan email lain.", "")
					return
				}
				if strings.Contains(errStr, "google") || strings.Contains(errStr, "users_google") {
					utils.WriteErrorJSON(w, http.StatusConflict, "USER_EXISTS",
						"User already registered with Google. Please login instead.", "")
					return
				}
			}

			utils.WriteInternalServerError(w)
			return
		}

		// Generate auth tokens
		user := UserRow{
			ID:              userID,
			FirstName:       firstName,
			LastName:        &lastName,
			Email:           googleEmail,
			Status:          "ACTIVE",
			ProfilePicture:  &googlePicture,
			PrimaryRegion:   regionParam,
			MFAStatus:       "INACTIVE",
			MembershipLevel: "CLASSIC",
		}

		accessToken, refreshToken, err := generateUserTokens(deps, user)
		if err != nil {
			fmt.Printf("Error generating user tokens for Google registration: %v\n", err)
			fmt.Printf("UserID: %s, Email: %s\n", userID, googleEmail)
			utils.WriteInternalServerError(w)
			return
		}

		// Store refresh token in Redis
		refreshKey := fmt.Sprintf("refresh_token:user:%s", userID)
		err = deps.Redis.Set(ctx, refreshKey, refreshToken, deps.Config.JWT.RefreshTokenExpiry)
		if err != nil {
			// Log error but continue
		}

		// Build user response object
		userResponse := map[string]interface{}{
			"id":             userID,
			"firstName":      firstName,
			"lastName":       lastName,
			"email":          googleEmail,
			"phoneNumber":    nil,
			"profilePicture": googlePicture,
			"status":         "ACTIVE",
			"primaryRegion":  regionParam,
			"currentRegion":  regionParam,
			"currency":       currency,
			"balance": map[string]interface{}{
				"IDR": int64(0),
				"MYR": int64(0),
				"PHP": int64(0),
				"SGD": int64(0),
				"THB": int64(0),
			},
			"membership": map[string]interface{}{
				"level": "CLASSIC",
				"name":  "Classic",
			},
			"mfaStatus": "INACTIVE",
			"createdAt": createdAt.Format(time.RFC3339),
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"step": "SUCCESS",
			"token": map[string]interface{}{
				"accessToken":  accessToken,
				"refreshToken": refreshToken,
				"expiresIn":    int64(deps.Config.JWT.AccessTokenExpiry.Seconds()),
				"tokenType":    "Bearer",
			},
			"user": userResponse,
		})
	}
}

// handleVerifyEmailImpl implements email verification
func HandleVerifyEmailImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.Token == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"token": "Verification token is required",
			})
			return
		}

		// Validate verification token
		tokenData, err := deps.JWTService.ValidateValidationToken(req.Token)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Token verifikasi tidak valid atau sudah kadaluarsa", "")
			return
		}

		userID, ok := tokenData["userId"].(string)
		if !ok {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Token verifikasi tidak valid", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user
		var user UserRow
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, first_name, last_name, email, status,
				profile_picture, primary_region, membership_level, mfa_status
			FROM users
			WHERE id = $1
		`, userID).Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Status,
			&user.ProfilePicture, &user.PrimaryRegion, &user.MembershipLevel, &user.MFAStatus,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND",
					"Pengguna tidak ditemukan", "")
			} else {
				utils.WriteInternalServerError(w)
			}
			return
		}

		// Check if already verified
		if user.Status == "ACTIVE" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "ALREADY_VERIFIED",
				"Email sudah diverifikasi. Silakan login.", "")
			return
		}

		// Update user status to ACTIVE and set email_verified_at
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE users
			SET status = $1, email_verified_at = NOW()
			WHERE id = $2
		`, "ACTIVE", userID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Delete verification token from Redis
		tokenKey := deps.Redis.ValidationTokenKey(userID)
		_ = deps.Redis.Delete(ctx, tokenKey)

		// Return success message only (no token)
		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Email berhasil diverifikasi. Silakan login untuk melanjutkan.",
		})
	}
}

// handleLogoutImpl implements user logout
func HandleLogoutImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Delete refresh token from Redis
		refreshKey := fmt.Sprintf("refresh_token:user:%s", userID)
		_ = deps.Redis.Delete(ctx, refreshKey)

		// Delete user cache
		_ = deps.Redis.InvalidateUserCache(ctx, userID)

		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Berhasil logout",
		})
	}
}

// handleRefreshTokenImpl implements token refresh
func HandleRefreshTokenImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.RefreshToken == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"refreshToken": "Refresh token is required",
			})
			return
		}

		// Validate refresh token
		claims, err := deps.JWTService.ValidateRefreshToken(req.RefreshToken)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN",
				"Refresh token tidak valid atau sudah kadaluarsa", "")
			return
		}

		userID, err := claims.GetSubject()
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "INVALID_TOKEN",
				"Token tidak valid", "")
			return
		}

		// Check token type (user or admin)
		tokenType := "user"
		if len(claims.Audience) > 0 {
			if strings.Contains(claims.Audience[0], "admin") {
				tokenType = "admin"
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Check if refresh token exists in Redis
		refreshKey := fmt.Sprintf("refresh_token:%s:%s", tokenType, userID)
		exists, err := deps.Redis.Exists(ctx, refreshKey)
		if err != nil || !exists {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "TOKEN_REVOKED",
				"Refresh token sudah tidak valid. Silakan login kembali.", "")
			return
		}

		// Get user/admin data and generate new tokens
		if tokenType == "user" {
			var user UserRow
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT
					id, first_name, last_name, email, status,
					profile_picture, primary_region, membership_level, mfa_status
				FROM users
				WHERE id = $1
			`, userID).Scan(
				&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Status,
				&user.ProfilePicture, &user.PrimaryRegion, &user.MembershipLevel, &user.MFAStatus,
			)

			if err != nil {
				utils.WriteErrorJSON(w, http.StatusUnauthorized, "USER_NOT_FOUND",
					"Pengguna tidak ditemukan", "")
				return
			}

			if user.Status != "ACTIVE" {
				utils.WriteErrorJSON(w, http.StatusForbidden, "ACCOUNT_INACTIVE",
					"Akun tidak aktif", "")
				return
			}

			// Generate new tokens
			accessToken, newRefreshToken, err := generateUserTokens(deps, user)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Update refresh token in Redis
			_ = deps.Redis.Set(ctx, refreshKey, newRefreshToken, deps.Config.JWT.RefreshTokenExpiry)

			utils.WriteSuccessJSON(w, TokenResponse{
				AccessToken:  accessToken,
				RefreshToken: newRefreshToken,
				ExpiresIn:    int64(deps.Config.JWT.AccessTokenExpiry.Seconds()),
				TokenType:    "Bearer",
			})
		} else {
			// Admin refresh token (not implemented in this task, but structure is here)
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN_TYPE",
				"Invalid token type", "")
		}
	}
}

// handleGetProfileImpl implements get user profile
func HandleGetProfileImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get full user profile
		var user UserRow
		var currentRegion string
		var createdAt, lastLoginAt, updatedAt *time.Time
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, first_name, last_name, email, phone_number,
				profile_picture, status, primary_region, current_region, 
				mfa_status, membership_level,
				balance_idr, balance_myr, balance_php, balance_sgd, balance_thb,
				total_spent_idr, email_verified_at, created_at, last_login_at, updated_at
			FROM users
			WHERE id = $1
		`, userID).Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber,
			&user.ProfilePicture, &user.Status, &user.PrimaryRegion, &currentRegion,
			&user.MFAStatus, &user.MembershipLevel,
			&user.BalanceIDR, &user.BalanceMYR, &user.BalancePHP, &user.BalanceSGD, &user.BalanceTHB,
			&user.TotalSpentIDR, &user.EmailVerifiedAt, &createdAt, &lastLoginAt, &updatedAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND",
					"Pengguna tidak ditemukan", "")
			} else {
				utils.WriteInternalServerError(w)
			}
			return
		}

		// Get currency based on region
		currency := getCurrencyByRegion(currentRegion)

		// Calculate membership progress
		membershipProgress := getMembershipProgress(user.MembershipLevel, user.TotalSpentIDR)

		// Build profile response (matching docs format)
		profile := map[string]interface{}{
			"id":             user.ID,
			"firstName":      user.FirstName,
			"lastName":       stringOrEmpty(user.LastName),
			"email":          user.Email,
			"phoneNumber":    stringOrEmpty(user.PhoneNumber),
			"profilePicture": stringOrEmpty(user.ProfilePicture),
			"status":         user.Status,
			"primaryRegion":  user.PrimaryRegion,
			"currentRegion":  currentRegion,
			"currency":       currency,
			"balance": map[string]interface{}{
				"IDR": user.BalanceIDR,
				"MYR": user.BalanceMYR,
				"PHP": user.BalancePHP,
				"SGD": user.BalanceSGD,
				"THB": user.BalanceTHB,
			},
			"membership": map[string]interface{}{
				"level":    user.MembershipLevel,
				"name":     getMembershipName(user.MembershipLevel),
				"benefits": getMembershipBenefits(user.MembershipLevel),
				"progress": membershipProgress,
			},
			"mfaStatus": user.MFAStatus,
		}

		// Add timestamps
		if user.EmailVerifiedAt != nil {
			profile["emailVerifiedAt"] = user.EmailVerifiedAt.Format(time.RFC3339)
		}
		if createdAt != nil {
			profile["createdAt"] = createdAt.Format(time.RFC3339)
		}
		if lastLoginAt != nil {
			profile["lastLoginAt"] = lastLoginAt.Format(time.RFC3339)
		}
		if updatedAt != nil {
			profile["updatedAt"] = updatedAt.Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, profile)
	}
}

// getCurrencyByRegion returns the currency for a region
func getCurrencyByRegion(region string) string {
	switch region {
	case "ID":
		return "IDR"
	case "MY":
		return "MYR"
	case "PH":
		return "PHP"
	case "SG":
		return "SGD"
	case "TH":
		return "THB"
	default:
		return "IDR"
	}
}

// getMembershipBenefits returns benefits for a membership level
func getMembershipBenefits(level string) []string {
	switch level {
	case "CLASSIC":
		return []string{
			"Akses ke semua produk",
			"Bonus poin 1%",
		}
	case "PRESTIGE":
		return []string{
			"Diskon eksklusif hingga 5%",
			"Priority customer support",
			"Bonus poin 3%",
			"Akses promo premium",
		}
	case "ROYAL":
		return []string{
			"Diskon eksklusif hingga 10%",
			"Dedicated account manager",
			"Bonus poin 5%",
			"Akses promo VIP",
			"Cashback mingguan",
		}
	default:
		return []string{"Akses ke semua produk"}
	}
}

// getMembershipProgress calculates progress to next level
func getMembershipProgress(level string, totalSpent int64) map[string]interface{} {
	var current, target int64
	var nextLevel string

	switch level {
	case "CLASSIC":
		current = totalSpent
		target = 5000000 // 5 juta untuk Prestige
		nextLevel = "PRESTIGE"
	case "PRESTIGE":
		current = totalSpent
		target = 10000000 // 10 juta untuk Royal
		nextLevel = "ROYAL"
	case "ROYAL":
		current = totalSpent
		target = totalSpent // Already max level
		nextLevel = ""
	default:
		current = totalSpent
		target = 5000000
		nextLevel = "PRESTIGE"
	}

	percentage := float64(0)
	if target > 0 {
		percentage = float64(current) / float64(target) * 100
		if percentage > 100 {
			percentage = 100
		}
	}

	return map[string]interface{}{
		"current":    current,
		"target":     target,
		"percentage": percentage,
		"nextLevel":  nextLevel,
		"currency":   "IDR",
	}
}

// handleUpdateProfileImpl implements update user profile
// Partial update - only updates fields that are provided, keeps existing values for others
func HandleUpdateProfileImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		contentType := r.Header.Get("Content-Type")

		// Parse multipart form data (5MB max)
		if err := r.ParseMultipartForm(5 << 20); err != nil {
			log.Warn().
				Err(err).
				Str("endpoint", "/v2/user/profile").
				Str("user_id", userID).
				Str("content_type", contentType).
				Msg("Failed to parse form data")
			utils.WriteBadRequestError(w, "Invalid form data or file too large (max 5MB)")
			return
		}

		// Parse form values
		firstName := utils.SanitizeString(r.FormValue("firstName"))
		lastName := utils.SanitizeString(r.FormValue("lastName"))
		phoneNumber := utils.SanitizeString(r.FormValue("phoneNumber"))

		// Check for profile picture file
		var profilePictureFile multipart.File
		var profilePictureHeader *multipart.FileHeader
		hasProfilePicture := false

		file, header, err := r.FormFile("profilePicture")
		if err == nil && file != nil {
			defer file.Close()
			profilePictureFile = file
			profilePictureHeader = header
			hasProfilePicture = true

			// Validate file type (JPG/PNG only)
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_FILE_TYPE",
					"Only JPG and PNG files are allowed", "")
				return
			}

			// Validate file size (max 5MB)
			if header.Size > 5<<20 {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "FILE_TOO_LARGE",
					"Profile picture must be less than 5MB", "")
				return
			}
		}

		// Validate phone number format if provided
		if phoneNumber != "" && !utils.ValidatePhone(phoneNumber) {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"phoneNumber": "Invalid phone number format",
			})
			return
		}

		// Check if at least one field is provided
		if firstName == "" && lastName == "" && phoneNumber == "" && !hasProfilePicture {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST",
				"Please provide at least one field to update", "")
			return
		}

		// Get current profile data (for partial update)
		var currentFirstName, currentLastName, currentPhoneNumber string
		var currentProfilePicture *string
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT first_name, COALESCE(last_name, ''), COALESCE(phone_number, ''), profile_picture
			FROM users WHERE id = $1
		`, userID).Scan(&currentFirstName, &currentLastName, &currentPhoneNumber, &currentProfilePicture)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to get current profile")
			utils.WriteInternalServerError(w)
			return
		}

		// Handle profile picture upload
		var newProfilePictureURL *string
		if hasProfilePicture && deps.S3 != nil {
			// Delete old profile picture if exists
			if currentProfilePicture != nil && *currentProfilePicture != "" {
				if err := deps.S3.DeleteByURL(ctx, *currentProfilePicture); err != nil {
					log.Warn().Err(err).Str("url", *currentProfilePicture).Msg("Failed to delete old profile picture")
				}
			}

			// Upload new profile picture with custom filename: /profiles/{userId}.{ext}
			ext := strings.ToLower(filepath.Ext(profilePictureHeader.Filename))
			filename := fmt.Sprintf("%s%s", userID, ext)

			result, err := deps.S3.UploadFromReader(ctx, "profiles", profilePictureFile, filename, profilePictureHeader.Header.Get("Content-Type"))
			if err != nil {
				log.Error().Err(err).Str("user_id", userID).Msg("Failed to upload profile picture to S3")
				utils.WriteErrorJSON(w, http.StatusInternalServerError, "UPLOAD_FAILED",
					"Failed to upload profile picture", "")
				return
			}

			newProfilePictureURL = &result.URL
			log.Info().
				Str("user_id", userID).
				Str("url", result.URL).
				Msg("Profile picture uploaded successfully")
		}

		// Prepare update values (partial update - keep existing if not provided)
		updateFirstName := currentFirstName
		updateLastName := currentLastName
		updatePhoneNumber := currentPhoneNumber
		updateProfilePicture := currentProfilePicture

		if firstName != "" {
			updateFirstName = firstName
		}
		if lastName != "" {
			updateLastName = lastName
		}
		if phoneNumber != "" {
			updatePhoneNumber = phoneNumber
		}
		if newProfilePictureURL != nil {
			updateProfilePicture = newProfilePictureURL
		}

		// Update user profile
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE users
			SET
				first_name = $1,
				last_name = NULLIF($2, ''),
				phone_number = NULLIF($3, ''),
				profile_picture = $4,
				updated_at = NOW()
			WHERE id = $5
		`, updateFirstName, updateLastName, updatePhoneNumber, updateProfilePicture, userID)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/user/profile").
				Str("user_id", userID).
				Msg("Failed to update user profile")
			utils.WriteInternalServerError(w)
			return
		}

		// Invalidate user cache
		_ = deps.Redis.InvalidateUserCache(ctx, userID)

		// Get updated profile
		var user UserRow
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, first_name, last_name, email, phone_number,
				profile_picture, status, primary_region, membership_level,
				balance_idr, balance_myr, balance_php, balance_sgd, balance_thb,
				total_spent_idr
			FROM users
			WHERE id = $1
		`, userID).Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber,
			&user.ProfilePicture, &user.Status, &user.PrimaryRegion, &user.MembershipLevel,
			&user.BalanceIDR, &user.BalanceMYR, &user.BalancePHP, &user.BalanceSGD, &user.BalanceTHB,
			&user.TotalSpentIDR,
		)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/user/profile").
				Str("user_id", userID).
				Msg("Failed to fetch updated user profile")
			utils.WriteInternalServerError(w)
			return
		}

		// Get total transactions count
		var totalTransactions int64
		_ = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM transactions WHERE user_id = $1
		`, userID).Scan(&totalTransactions)

		// Build profile response
		profile := map[string]interface{}{
			"id":             user.ID,
			"firstName":      user.FirstName,
			"lastName":       stringOrEmpty(user.LastName),
			"email":          user.Email,
			"phoneNumber":    stringOrEmpty(user.PhoneNumber),
			"profilePicture": stringOrEmpty(user.ProfilePicture),
			"status":         user.Status,
			"primaryRegion":  user.PrimaryRegion,
			"membership": map[string]interface{}{
				"level": user.MembershipLevel,
				"name":  getMembershipName(user.MembershipLevel),
			},
			"balance": map[string]interface{}{
				"idr": user.BalanceIDR,
				"myr": user.BalanceMYR,
				"php": user.BalancePHP,
				"sgd": user.BalanceSGD,
				"thb": user.BalanceTHB,
			},
			"stats": map[string]interface{}{
				"totalTransactions": totalTransactions,
				"totalSpent":        user.TotalSpentIDR,
			},
		}

		utils.WriteSuccessJSON(w, profile)
	}
}

// handleChangePasswordImpl implements change password
func HandleChangePasswordImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate inputs
		validationErrors := make(map[string]string)
		if req.CurrentPassword == "" {
			validationErrors["currentPassword"] = "Current password is required"
		}
		if req.NewPassword == "" {
			validationErrors["newPassword"] = "New password is required"
		} else if !validatePasswordStrength(req.NewPassword) {
			validationErrors["newPassword"] = "Password must be at least 8 characters and contain uppercase, lowercase, and number"
		}
		if req.CurrentPassword == req.NewPassword {
			validationErrors["newPassword"] = "New password must be different from current password"
		}

		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get current password hash
		var currentHash string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT password_hash FROM users WHERE id = $1
		`, userID).Scan(&currentHash)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Verify current password
		if !utils.CheckPassword(req.CurrentPassword, currentHash) {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_PASSWORD",
				"Password saat ini salah", "")
			return
		}

		// Hash new password
		newHash, err := utils.HashPassword(req.NewPassword)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Update password
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE users SET password_hash = $1 WHERE id = $2
		`, newHash, userID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Invalidate all refresh tokens for this user
		refreshKey := fmt.Sprintf("refresh_token:user:%s", userID)
		_ = deps.Redis.Delete(ctx, refreshKey)

		// Invalidate user cache
		_ = deps.Redis.InvalidateUserCache(ctx, userID)

		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Password berhasil diubah. Silakan login kembali dengan password baru.",
		})
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// validatePasswordStrength validates password strength
func validatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasUpper && hasLower && hasNumber
}

// stringOrEmpty returns string value or empty string
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ============================================
// RESEND VERIFICATION, FORGOT PASSWORD, RESET PASSWORD, LOGIN GOOGLE
// ============================================

// handleResendVerificationImpl implements resend verification email
func HandleResendVerificationImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResendVerificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.Email == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"email": "Email is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user by email
		var userID, firstName string
		var status string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id, first_name, status
			FROM users
			WHERE LOWER(email) = LOWER($1)
		`, req.Email).Scan(&userID, &firstName, &status)

		if err != nil {
			if err == pgx.ErrNoRows {
				// Don't reveal if email exists for security
				utils.WriteSuccessJSON(w, map[string]string{
					"message": "Jika email terdaftar, link verifikasi telah dikirim.",
				})
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// If already verified, don't resend
		if status == "ACTIVE" {
			utils.WriteSuccessJSON(w, map[string]string{
				"message": "Email sudah diverifikasi. Silakan login.",
			})
			return
		}

		// Check if token exists and is still valid
		tokenKey := deps.Redis.ValidationTokenKey(userID)
		var existingToken string
		getErr := deps.Redis.Get(ctx, tokenKey, &existingToken)
		tokenExpired := true

		if getErr == nil && existingToken != "" {
			// Check if token is still valid
			_, validateErr := deps.JWTService.ValidateValidationToken(existingToken)
			if validateErr == nil {
				tokenExpired = false
			}
		}

		// Only generate new token if expired
		var verificationToken string
		if tokenExpired {
			// Generate new email verification token
			verificationToken, err = deps.JWTService.GenerateValidationToken(map[string]interface{}{
				"userId": userID,
				"email":  req.Email,
				"type":   "email_verification",
			})
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Store verification token in Redis (30 minutes expiry)
			err = deps.Redis.Set(ctx, tokenKey, verificationToken, 30*time.Minute)
			if err != nil {
				// Log error but continue
			}
		} else {
			verificationToken = existingToken
		}

		// Send verification email
		go func() {
			emailService := deps.EmailService
			if emailService != nil {
				err := emailService.SendVerificationEmail(req.Email, firstName, verificationToken)
				if err != nil {
					fmt.Printf("Failed to send verification email to %s: %v\n", req.Email, err)
				}
			}
		}()

		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Link verifikasi telah dikirim ke email Anda.",
		})
	}
}

// handleForgotPasswordImpl implements forgot password
func HandleForgotPasswordImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ForgotPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.Email == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"email": "Email is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user by email (allow unverified users)
		var userID, firstName string
		var status string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id, first_name, status
			FROM users
			WHERE LOWER(email) = LOWER($1)
		`, req.Email).Scan(&userID, &firstName, &status)

		if err != nil {
			if err == pgx.ErrNoRows {
				// Don't reveal if email exists for security
				utils.WriteSuccessJSON(w, map[string]string{
					"message": "Jika email terdaftar, link reset password telah dikirim.",
				})
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Check if user has password (not OAuth-only user)
		var passwordHash sql.NullString
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT password_hash FROM users WHERE id = $1
		`, userID).Scan(&passwordHash)

		if err != nil || !passwordHash.Valid || passwordHash.String == "" {
			// User doesn't have password, can't reset
			utils.WriteErrorJSON(w, http.StatusBadRequest, "OAUTH_ACCOUNT",
				"Akun ini terdaftar menggunakan Google. Silakan login dengan Google.", "")
			return
		}

		// Generate password reset token
		resetToken, err := deps.JWTService.GenerateValidationToken(map[string]interface{}{
			"userId": userID,
			"email":  req.Email,
			"type":   "password_reset",
		})
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Store reset token in Redis (1 hour expiry)
		resetTokenKey := fmt.Sprintf("password_reset:%s", userID)
		err = deps.Redis.Set(ctx, resetTokenKey, resetToken, 1*time.Hour)
		if err != nil {
			// Log error but continue
		}

		// Send password reset email
		go func() {
			emailService := deps.EmailService
			if emailService != nil {
				err := emailService.SendPasswordResetEmail(req.Email, firstName, resetToken)
				if err != nil {
					fmt.Printf("Failed to send password reset email to %s: %v\n", req.Email, err)
				}
			}
		}()

		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Link reset password telah dikirim ke email Anda.",
		})
	}
}

// handleResetPasswordImpl implements reset password
func HandleResetPasswordImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate request
		validationErrors := make(map[string]string)
		if req.Token == "" {
			validationErrors["token"] = "Reset token is required"
		}
		if req.NewPassword == "" {
			validationErrors["newPassword"] = "New password is required"
		} else if !validatePasswordStrength(req.NewPassword) {
			validationErrors["newPassword"] = "Password must be at least 8 characters and contain uppercase, lowercase, and number"
		}
		if req.ConfirmPassword == "" {
			validationErrors["confirmPassword"] = "Confirm password is required"
		} else if req.NewPassword != req.ConfirmPassword {
			validationErrors["confirmPassword"] = "Passwords do not match"
		}

		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		// Validate reset token
		tokenData, err := deps.JWTService.ValidateValidationToken(req.Token)
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Token reset password tidak valid atau sudah kadaluarsa", "")
			return
		}

		userID, ok := tokenData["userId"].(string)
		if !ok {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Token reset password tidak valid", "")
			return
		}

		tokenType, _ := tokenData["type"].(string)
		if tokenType != "password_reset" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Token tidak valid untuk reset password", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Check if reset token exists in Redis
		resetTokenKey := fmt.Sprintf("password_reset:%s", userID)
		exists, err := deps.Redis.Exists(ctx, resetTokenKey)
		if err != nil || !exists {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "TOKEN_EXPIRED",
				"Token reset password sudah tidak valid. Silakan request reset password baru.", "")
			return
		}

		// Get user
		var user UserRow
		var status string
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, first_name, last_name, email, status,
				profile_picture, primary_region, membership_level, mfa_status
			FROM users
			WHERE id = $1
		`, userID).Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &status,
			&user.ProfilePicture, &user.PrimaryRegion, &user.MembershipLevel, &user.MFAStatus,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND",
					"Pengguna tidak ditemukan", "")
			} else {
				utils.WriteInternalServerError(w)
			}
			return
		}

		// Hash new password
		newHash, err := utils.HashPassword(req.NewPassword)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Update password and auto-verify email if not verified
		var updateQuery string
		var updateArgs []interface{}

		if status == "INACTIVE" {
			// Auto-verify email on password reset
			updateQuery = `
				UPDATE users
				SET password_hash = $1, status = $2, email_verified_at = NOW()
				WHERE id = $3
			`
			updateArgs = []interface{}{newHash, "ACTIVE", userID}
			user.Status = "ACTIVE"
		} else {
			updateQuery = `
				UPDATE users
				SET password_hash = $1
				WHERE id = $2
			`
			updateArgs = []interface{}{newHash, userID}
		}

		_, err = deps.DB.Pool.Exec(ctx, updateQuery, updateArgs...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Delete reset token from Redis
		_ = deps.Redis.Delete(ctx, resetTokenKey)

		// Delete verification token if exists
		tokenKey := deps.Redis.ValidationTokenKey(userID)
		_ = deps.Redis.Delete(ctx, tokenKey)

		// Invalidate all refresh tokens for this user
		refreshKey := fmt.Sprintf("refresh_token:user:%s", userID)
		_ = deps.Redis.Delete(ctx, refreshKey)

		// Invalidate user cache
		_ = deps.Redis.InvalidateUserCache(ctx, userID)

		utils.WriteSuccessJSON(w, map[string]string{
			"message": "Password berhasil direset. Silakan login dengan password baru.",
		})
	}
}

// handleLoginGoogleImpl implements login with Google
func HandleLoginGoogleImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.GoogleAuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.IDToken == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"idToken": "ID token is required",
			})
			return
		}

		// Get region from query parameter (optional)
		regionParam := r.URL.Query().Get("region")
		if regionParam != "" {
			regionParam = strings.ToUpper(strings.TrimSpace(regionParam))
			if !utils.ValidateRegion(regionParam) {
				regionParam = ""
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Decode Google ID token (without verification for now)
		// WARNING: In production, MUST verify the token with Google's public keys
		tokenParts := strings.Split(req.IDToken, ".")
		if len(tokenParts) != 3 {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Invalid Google ID token format", "")
			return
		}

		// Decode payload (second part of JWT)
		payload, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
		if err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Failed to decode token payload", "")
			return
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Failed to parse token claims", "")
			return
		}

		// Extract user info from Google token claims
		googleID, _ := claims["sub"].(string) // Google user ID
		googleEmail, _ := claims["email"].(string)

		if googleID == "" || googleEmail == "" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Missing required fields in Google token", "")
			return
		}

		// Find user by Google ID or email
		var user UserRow
		var currentRegion string
		var createdAt, lastLoginAt *time.Time
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, first_name, last_name, email, phone_number, status,
				profile_picture, primary_region, current_region, membership_level, mfa_status,
				balance_idr, balance_myr, balance_php, balance_sgd, balance_thb,
				created_at, last_login_at
			FROM users
			WHERE google_id = $1 OR (LOWER(email) = LOWER($2) AND google_id IS NOT NULL)
			LIMIT 1
		`, googleID, googleEmail).Scan(
			&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.Status,
			&user.ProfilePicture, &user.PrimaryRegion, &currentRegion, &user.MembershipLevel, &user.MFAStatus,
			&user.BalanceIDR, &user.BalanceMYR, &user.BalancePHP, &user.BalanceSGD, &user.BalanceTHB,
			&createdAt, &lastLoginAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND",
					"Akun tidak ditemukan. Silakan registrasi terlebih dahulu.", "")
			} else {
				utils.WriteInternalServerError(w)
			}
			return
		}

		// Check if account is suspended
		if user.Status == "SUSPENDED" {
			utils.WriteErrorJSON(w, http.StatusForbidden, "ACCOUNT_SUSPENDED",
				"Akun Anda telah dinonaktifkan", "")
			return
		}

		// Update current region if provided
		if regionParam != "" && regionParam != user.PrimaryRegion {
			_, _ = deps.DB.Pool.Exec(ctx, `
				UPDATE users SET current_region = $1 WHERE id = $2
			`, regionParam, user.ID)
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

		// Build user response matching docs format
		lastName := ""
		if user.LastName != nil {
			lastName = *user.LastName
		}
		profilePic := ""
		if user.ProfilePicture != nil {
			profilePic = *user.ProfilePicture
		}
		phoneNumber := ""
		if user.PhoneNumber != nil {
			phoneNumber = *user.PhoneNumber
		}

		// Get currency from region
		currency := getCurrencyByRegion(currentRegion)

		userResponse := map[string]interface{}{
			"id":             user.ID,
			"firstName":      user.FirstName,
			"lastName":       lastName,
			"email":          user.Email,
			"phoneNumber":    phoneNumber,
			"profilePicture": profilePic,
			"status":         user.Status,
			"primaryRegion":  user.PrimaryRegion,
			"currentRegion":  currentRegion,
			"currency":       currency,
			"balance": map[string]interface{}{
				"IDR": user.BalanceIDR,
				"MYR": user.BalanceMYR,
				"PHP": user.BalancePHP,
				"SGD": user.BalanceSGD,
				"THB": user.BalanceTHB,
			},
			"membership": map[string]interface{}{
				"level": user.MembershipLevel,
				"name":  getMembershipName(user.MembershipLevel),
			},
			"mfaStatus": user.MFAStatus,
		}

		// Add timestamps
		if createdAt != nil {
			userResponse["createdAt"] = createdAt.Format(time.RFC3339)
		}
		if lastLoginAt != nil {
			userResponse["lastLoginAt"] = lastLoginAt.Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, UserLoginSuccessResponse{
			Step: "SUCCESS",
			Token: TokenResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				ExpiresIn:    int64(deps.Config.JWT.AccessTokenExpiry.Seconds()),
				TokenType:    "Bearer",
			},
			User: userResponse,
		})
	}
}

// HandleEnableMFAImpl implements enable MFA
func HandleEnableMFAImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get user email and current MFA status
		var email, mfaStatus string
		var mfaSecret *string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT email, mfa_status, mfa_secret FROM users WHERE id = $1
		`, userID).Scan(&email, &mfaStatus, &mfaSecret)

		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for MFA enable")
			utils.WriteInternalServerError(w)
			return
		}

		// Check if MFA is already active
		if mfaStatus == "ACTIVE" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "MFA_ALREADY_ENABLED",
				"MFA is already enabled for this account", "")
			return
		}

		// Generate new MFA secret with issuer "Seaply"
		secret, otpURL, err := utils.GenerateMFASecret(email, "Seaply")
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to generate MFA secret")
			utils.WriteInternalServerError(w)
			return
		}

		// Generate backup codes
		backupCodes, err := utils.GenerateBackupCodes(5)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to generate backup codes")
			utils.WriteInternalServerError(w)
			return
		}

		// Store secret and backup codes (MFA still INACTIVE until verified)
		// We track pending state by having mfa_secret set but mfa_status = INACTIVE
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE users 
			SET mfa_secret = $1, mfa_backup_codes = $2, updated_at = NOW()
			WHERE id = $3
		`, secret, backupCodes, userID)

		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to store MFA secret")
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"step":        "SETUP",
			"qrCode":      otpURL,
			"secret":      secret,
			"backupCodes": backupCodes,
		})
	}
}

// HandleVerifyMFASetupImpl implements verify MFA setup
func HandleVerifyMFASetupImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.Code == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"code": "Verification code is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get user MFA secret and status
		var mfaStatus string
		var mfaSecret *string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT mfa_status, mfa_secret FROM users WHERE id = $1
		`, userID).Scan(&mfaStatus, &mfaSecret)

		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for MFA verify")
			utils.WriteInternalServerError(w)
			return
		}

		// Check if MFA is already active
		if mfaStatus == "ACTIVE" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "MFA_ALREADY_ACTIVE",
				"MFA is already active for this account", "")
			return
		}

		// Check if MFA setup is pending (secret exists but status is INACTIVE)
		if mfaSecret == nil || *mfaSecret == "" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "MFA_NOT_SETUP",
				"MFA secret not found. Please enable MFA first.", "")
			return
		}

		// Validate the code
		if !utils.ValidateMFACode(*mfaSecret, req.Code) {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_CODE",
				"Invalid verification code", "")
			return
		}

		// Activate MFA
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE users SET mfa_status = 'ACTIVE', updated_at = NOW() WHERE id = $1
		`, userID)

		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to activate MFA")
			utils.WriteInternalServerError(w)
			return
		}

		log.Info().Str("user_id", userID).Msg("MFA enabled successfully")

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"step":      "SUCCESS",
			"message":   "MFA has been enabled successfully",
			"mfaStatus": "ACTIVE",
		})
	}
}

// HandleDisableMFAImpl implements disable MFA
func HandleDisableMFAImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED",
				"Authentication required", "")
			return
		}

		var req struct {
			Code     string `json:"code"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate required fields
		validationErrors := make(map[string]string)
		if req.Code == "" {
			validationErrors["code"] = "MFA code is required"
		}
		if req.Password == "" {
			validationErrors["password"] = "Password is required"
		}
		if len(validationErrors) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", validationErrors)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get user data
		var mfaStatus string
		var mfaSecret, passwordHash *string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT mfa_status, mfa_secret, password_hash FROM users WHERE id = $1
		`, userID).Scan(&mfaStatus, &mfaSecret, &passwordHash)

		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user for MFA disable")
			utils.WriteInternalServerError(w)
			return
		}

		// Check if MFA is active
		if mfaStatus != "ACTIVE" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "MFA_NOT_ACTIVE",
				"MFA is not active for this account", "")
			return
		}

		// Verify password (if user has password - not OAuth only)
		if passwordHash != nil && *passwordHash != "" {
			if !utils.CheckPassword(req.Password, *passwordHash) {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_PASSWORD",
					"Invalid password", "")
				return
			}
		}

		// Verify MFA code
		if mfaSecret == nil || !utils.ValidateMFACode(*mfaSecret, req.Code) {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_CODE",
				"Invalid MFA code", "")
			return
		}

		// Disable MFA
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE users 
			SET mfa_status = 'INACTIVE', mfa_secret = NULL, mfa_backup_codes = NULL, updated_at = NOW()
			WHERE id = $1
		`, userID)

		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to disable MFA")
			utils.WriteInternalServerError(w)
			return
		}

		log.Info().Str("user_id", userID).Msg("MFA disabled successfully")

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":   "MFA has been disabled",
			"mfaStatus": "INACTIVE",
		})
	}
}
