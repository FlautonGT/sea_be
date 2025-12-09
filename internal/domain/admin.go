package domain

import (
	"time"

	"github.com/google/uuid"
)

type AdminStatus string

const (
	AdminStatusActive    AdminStatus = "ACTIVE"
	AdminStatusInactive  AdminStatus = "INACTIVE"
	AdminStatusSuspended AdminStatus = "SUSPENDED"
)

type RoleCode string

const (
	RoleSuperAdmin RoleCode = "SUPERADMIN"
	RoleAdmin      RoleCode = "ADMIN"
	RoleFinance    RoleCode = "FINANCE"
	RoleCSLead     RoleCode = "CS_LEAD"
	RoleCS         RoleCode = "CS"
)

type Admin struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Email       string      `json:"email" db:"email"`
	PhoneNumber *string     `json:"phoneNumber" db:"phone_number"`
	Password    string      `json:"-" db:"password"`
	RoleID      uuid.UUID   `json:"roleId" db:"role_id"`
	Status      AdminStatus `json:"status" db:"status"`
	MFAEnabled  bool        `json:"mfaEnabled" db:"mfa_enabled"`
	MFASecret   *string     `json:"-" db:"mfa_secret"`
	CreatedBy   *uuid.UUID  `json:"createdBy" db:"created_by"`
	LastLoginAt *time.Time  `json:"lastLoginAt" db:"last_login_at"`
	CreatedAt   time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time   `json:"updatedAt" db:"updated_at"`
}

type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        RoleCode  `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Level       int       `json:"level" db:"level"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"roleId" db:"role_id"`
	PermissionID uuid.UUID `json:"permissionId" db:"permission_id"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type AdminSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	AdminID      uuid.UUID  `json:"adminId" db:"admin_id"`
	RefreshToken string     `json:"-" db:"refresh_token"`
	UserAgent    string     `json:"userAgent" db:"user_agent"`
	IPAddress    string     `json:"ipAddress" db:"ip_address"`
	ExpiresAt    time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	RevokedAt    *time.Time `json:"revokedAt" db:"revoked_at"`
}

type AdminBackupCode struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	AdminID   uuid.UUID  `json:"adminId" db:"admin_id"`
	Code      string     `json:"-" db:"code"`
	UsedAt    *time.Time `json:"usedAt" db:"used_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

// Admin Response DTOs
type AdminResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	PhoneNumber *string     `json:"phoneNumber,omitempty"`
	Role        RoleInfo    `json:"role"`
	Status      AdminStatus `json:"status"`
	MFAEnabled  bool        `json:"mfaEnabled"`
	CreatedBy   *CreatedBy  `json:"createdBy,omitempty"`
	LastLoginAt *time.Time  `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type RoleInfo struct {
	Code        RoleCode `json:"code"`
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions,omitempty"`
}

type CreatedBy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoleResponse struct {
	Code        RoleCode `json:"code"`
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	AdminCount  int      `json:"adminCount"`
}

// Admin Request DTOs
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AdminVerifyMFARequest struct {
	MFAToken string `json:"mfaToken" validate:"required"`
	Code     string `json:"code" validate:"required,len=6"`
}

type CreateAdminRequest struct {
	Name        string      `json:"name" validate:"required,min=1,max=100"`
	Email       string      `json:"email" validate:"required,email"`
	PhoneNumber *string     `json:"phoneNumber" validate:"omitempty,min=10,max=15"`
	Password    string      `json:"password" validate:"required,min=8,max=100"`
	RoleCode    RoleCode    `json:"roleCode" validate:"required"`
	Status      AdminStatus `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
}

type UpdateAdminRequest struct {
	Name        *string      `json:"name" validate:"omitempty,min=1,max=100"`
	PhoneNumber *string      `json:"phoneNumber" validate:"omitempty,min=10,max=15"`
	RoleCode    *RoleCode    `json:"roleCode"`
	Status      *AdminStatus `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED"`
}

type UpdateRolePermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"required"`
}

// Audit Log
type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	AuditActionLogin  AuditAction = "LOGIN"
	AuditActionLogout AuditAction = "LOGOUT"
)

type AuditLog struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	AdminID     uuid.UUID              `json:"adminId" db:"admin_id"`
	Action      AuditAction            `json:"action" db:"action"`
	Resource    string                 `json:"resource" db:"resource"`
	ResourceID  *string                `json:"resourceId" db:"resource_id"`
	Description string                 `json:"description" db:"description"`
	OldValue    map[string]interface{} `json:"oldValue" db:"old_value"`
	NewValue    map[string]interface{} `json:"newValue" db:"new_value"`
	IPAddress   string                 `json:"ipAddress" db:"ip_address"`
	UserAgent   string                 `json:"userAgent" db:"user_agent"`
	CreatedAt   time.Time              `json:"createdAt" db:"created_at"`
}

type AuditLogResponse struct {
	ID          string        `json:"id"`
	Admin       CreatedBy     `json:"admin"`
	Action      AuditAction   `json:"action"`
	Resource    string        `json:"resource"`
	ResourceID  *string       `json:"resourceId"`
	Description string        `json:"description"`
	Changes     *AuditChanges `json:"changes,omitempty"`
	IPAddress   string        `json:"ipAddress"`
	UserAgent   string        `json:"userAgent"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type AuditChanges struct {
	Before map[string]interface{} `json:"before"`
	After  map[string]interface{} `json:"after"`
}
