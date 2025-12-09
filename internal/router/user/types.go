package user

import (
	"database/sql"
	"strings"
	"time"

	"seaply/internal/utils"
)

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

// UserRow represents user data from database
type UserRow struct {
	ID                string
	FirstName         string
	LastName          *string
	Email             string
	PasswordHash      *string
	PhoneNumber       *string
	Status            string
	ProfilePicture    *string
	PrimaryRegion     string
	MFAStatus         string
	MembershipLevel   string
	BalanceIDR        int64
	BalanceMYR        int64
	BalancePHP        int64
	BalanceSGD        int64
	BalanceTHB        int64
	TotalTransactions int64
	TotalSpentIDR     int64
	EmailVerifiedAt   *time.Time
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

// nullString converts a string to sql.NullString
func nullString(value string) sql.NullString {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}
