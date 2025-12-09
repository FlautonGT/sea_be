package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusInactive  UserStatus = "INACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

type MembershipLevel string

const (
	MembershipClassic  MembershipLevel = "CLASSIC"
	MembershipPrestige MembershipLevel = "PRESTIGE"
	MembershipRoyal    MembershipLevel = "ROYAL"
)

type MFAStatus string

const (
	MFAStatusActive   MFAStatus = "ACTIVE"
	MFAStatusInactive MFAStatus = "INACTIVE"
)

type User struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	FirstName       string          `json:"firstName" db:"first_name"`
	LastName        string          `json:"lastName" db:"last_name"`
	Email           string          `json:"email" db:"email"`
	PhoneNumber     *string         `json:"phoneNumber" db:"phone_number"`
	Password        string          `json:"-" db:"password"`
	ProfilePicture  *string         `json:"profilePicture" db:"profile_picture"`
	Status          UserStatus      `json:"status" db:"status"`
	PrimaryRegion   string          `json:"primaryRegion" db:"primary_region"`
	CurrentRegion   string          `json:"currentRegion" db:"current_region"`
	MembershipLevel MembershipLevel `json:"membershipLevel" db:"membership_level"`
	MFAStatus       MFAStatus       `json:"mfaStatus" db:"mfa_status"`
	MFASecret       *string         `json:"-" db:"mfa_secret"`
	GoogleID        *string         `json:"googleId,omitempty" db:"google_id"`
	EmailVerifiedAt *time.Time      `json:"emailVerifiedAt" db:"email_verified_at"`
	LastLoginAt     *time.Time      `json:"lastLoginAt" db:"last_login_at"`
	CreatedAt       time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time       `json:"updatedAt" db:"updated_at"`
}

type UserBalance struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Currency  string    `json:"currency" db:"currency"`
	Balance   float64   `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type UserBalanceMap map[string]float64

type Membership struct {
	Level    MembershipLevel    `json:"level"`
	Name     string             `json:"name"`
	Benefits []string           `json:"benefits,omitempty"`
	Progress *MembershipProgress `json:"progress,omitempty"`
}

type MembershipProgress struct {
	Current    float64 `json:"current"`
	Target     float64 `json:"target"`
	Percentage float64 `json:"percentage"`
	NextLevel  string  `json:"nextLevel"`
	Currency   string  `json:"currency"`
}

type UserSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	RefreshToken string     `json:"-" db:"refresh_token"`
	UserAgent    string     `json:"userAgent" db:"user_agent"`
	IPAddress    string     `json:"ipAddress" db:"ip_address"`
	ExpiresAt    time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	RevokedAt    *time.Time `json:"revokedAt" db:"revoked_at"`
}

type UserBackupCode struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"userId" db:"user_id"`
	Code      string     `json:"-" db:"code"`
	UsedAt    *time.Time `json:"usedAt" db:"used_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

type VerificationToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	Type      string    `json:"type" db:"type"` // email_verification, password_reset
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	UsedAt    *time.Time `json:"usedAt" db:"used_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

// Response DTOs
type UserResponse struct {
	ID              string         `json:"id"`
	FirstName       string         `json:"firstName"`
	LastName        string         `json:"lastName"`
	Email           string         `json:"email"`
	PhoneNumber     *string        `json:"phoneNumber"`
	ProfilePicture  *string        `json:"profilePicture"`
	Status          UserStatus     `json:"status"`
	PrimaryRegion   string         `json:"primaryRegion"`
	CurrentRegion   string         `json:"currentRegion"`
	Currency        string         `json:"currency"`
	Balance         UserBalanceMap `json:"balance"`
	Membership      Membership     `json:"membership"`
	MFAStatus       MFAStatus      `json:"mfaStatus"`
	GoogleID        *string        `json:"googleId,omitempty"`
	EmailVerifiedAt *time.Time     `json:"emailVerifiedAt,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
	LastLoginAt     *time.Time     `json:"lastLoginAt,omitempty"`
	UpdatedAt       *time.Time     `json:"updatedAt,omitempty"`
}

// Request DTOs
type RegisterRequest struct {
	FirstName       string `json:"firstName" validate:"required,min=1,max=50"`
	LastName        string `json:"lastName" validate:"required,min=1,max=50"`
	Email           string `json:"email" validate:"required,email"`
	PhoneNumber     string `json:"phoneNumber" validate:"required,min=10,max=15"`
	Password        string `json:"password" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}

type GoogleAuthRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Token string `json:"token" validate:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyMFARequest struct {
	MFAToken string `json:"mfaToken" validate:"required"`
	Code     string `json:"code" validate:"required,len=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}

type EnableMFARequest struct{}

type VerifyMFASetupRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type DisableMFARequest struct {
	Code     string `json:"code" validate:"required,len=6"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type UpdateProfileRequest struct {
	FirstName   *string `json:"firstName" validate:"omitempty,min=1,max=50"`
	LastName    *string `json:"lastName" validate:"omitempty,min=1,max=50"`
	PhoneNumber *string `json:"phoneNumber" validate:"omitempty,min=10,max=15"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}

