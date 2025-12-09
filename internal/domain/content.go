package domain

import (
	"time"

	"github.com/google/uuid"
)

type Banner struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description" db:"description"`
	Href        *string    `json:"href" db:"href"`
	Image       string     `json:"image" db:"image"`
	Order       int        `json:"order" db:"order"`
	IsActive    bool       `json:"isActive" db:"is_active"`
	StartAt     *time.Time `json:"startAt" db:"start_at"`
	ExpiredAt   *time.Time `json:"expiredAt" db:"expired_at"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

type BannerRegion struct {
	BannerID   uuid.UUID `json:"bannerId" db:"banner_id"`
	RegionCode string    `json:"regionCode" db:"region_code"`
	IsActive   bool      `json:"isActive" db:"is_active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type Popup struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	RegionCode string     `json:"regionCode" db:"region_code"`
	Title      string     `json:"title" db:"title"`
	Content    *string    `json:"content" db:"content"`
	Image      *string    `json:"image" db:"image"`
	Href       *string    `json:"href" db:"href"`
	IsActive   bool       `json:"isActive" db:"is_active"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
}

type Contact struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Response DTOs
type BannerResponse struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Href        *string `json:"href,omitempty"`
	Image       string  `json:"image"`
	Order       int     `json:"order"`
}

type PopupResponse struct {
	Title    string  `json:"title"`
	Content  *string `json:"content,omitempty"`
	Image    *string `json:"image,omitempty"`
	Href     *string `json:"href,omitempty"`
	IsActive bool    `json:"isActive"`
}

type ContactResponse struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Whatsapp  string `json:"whatsapp"`
	Instagram string `json:"instagram"`
	Facebook  string `json:"facebook"`
	X         string `json:"x"`
	Youtube   string `json:"youtube"`
	Telegram  string `json:"telegram"`
	Discord   string `json:"discord"`
}

type AdminBannerResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Href        *string    `json:"href,omitempty"`
	Image       string     `json:"image"`
	Order       int        `json:"order"`
	IsActive    bool       `json:"isActive"`
	Regions     []string   `json:"regions"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	ExpiredAt   *time.Time `json:"expiredAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type AdminPopupResponse struct {
	ID         string    `json:"id"`
	RegionCode string    `json:"regionCode"`
	Title      string    `json:"title"`
	Content    *string   `json:"content,omitempty"`
	Image      *string   `json:"image,omitempty"`
	Href       *string   `json:"href,omitempty"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Request DTOs
type CreateBannerRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=200"`
	Description *string    `json:"description"`
	Href        *string    `json:"href" validate:"omitempty,url"`
	Regions     []string   `json:"regions" validate:"required,min=1"`
	Order       int        `json:"order" validate:"min=0"`
	IsActive    bool       `json:"isActive"`
	StartAt     *time.Time `json:"startAt"`
	ExpiredAt   *time.Time `json:"expiredAt"`
}

type UpdateBannerRequest struct {
	Title       *string    `json:"title" validate:"omitempty,min=1,max=200"`
	Description *string    `json:"description"`
	Href        *string    `json:"href" validate:"omitempty,url"`
	Regions     []string   `json:"regions"`
	Order       *int       `json:"order" validate:"omitempty,min=0"`
	IsActive    *bool      `json:"isActive"`
	StartAt     *time.Time `json:"startAt"`
	ExpiredAt   *time.Time `json:"expiredAt"`
}

type UpdatePopupRequest struct {
	Title    *string `json:"title" validate:"omitempty,min=1,max=200"`
	Content  *string `json:"content"`
	Href     *string `json:"href" validate:"omitempty,url"`
	IsActive *bool   `json:"isActive"`
}

type UpdateContactsRequest struct {
	Email     *string `json:"email" validate:"omitempty,email"`
	Phone     *string `json:"phone"`
	Whatsapp  *string `json:"whatsapp" validate:"omitempty,url"`
	Instagram *string `json:"instagram" validate:"omitempty,url"`
	Facebook  *string `json:"facebook" validate:"omitempty,url"`
	X         *string `json:"x" validate:"omitempty,url"`
	Youtube   *string `json:"youtube" validate:"omitempty,url"`
	Telegram  *string `json:"telegram" validate:"omitempty,url"`
	Discord   *string `json:"discord" validate:"omitempty,url"`
}

