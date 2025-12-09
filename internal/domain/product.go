package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Slug        string    `json:"slug" db:"slug"`
	Title       string    `json:"title" db:"title"`
	Subtitle    *string   `json:"subtitle" db:"subtitle"`
	Description *string   `json:"description" db:"description"`
	Publisher   *string   `json:"publisher" db:"publisher"`
	Thumbnail   *string   `json:"thumbnail" db:"thumbnail"`
	Banner      *string   `json:"banner" db:"banner"`
	CategoryID  uuid.UUID `json:"categoryId" db:"category_id"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	IsPopular   bool      `json:"isPopular" db:"is_popular"`
	InquirySlug *string   `json:"inquirySlug" db:"inquiry_slug"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type ProductRegion struct {
	ProductID  uuid.UUID `json:"productId" db:"product_id"`
	RegionCode string    `json:"regionCode" db:"region_code"`
	IsActive   bool      `json:"isActive" db:"is_active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type ProductFeature struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProductID uuid.UUID `json:"productId" db:"product_id"`
	Feature   string    `json:"feature" db:"feature"`
	Order     int       `json:"order" db:"order"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type ProductHowToOrder struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProductID uuid.UUID `json:"productId" db:"product_id"`
	Step      string    `json:"step" db:"step"`
	Order     int       `json:"order" db:"order"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type ProductTag struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProductID uuid.UUID `json:"productId" db:"product_id"`
	Tag       string    `json:"tag" db:"tag"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type ProductField struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ProductID   uuid.UUID  `json:"productId" db:"product_id"`
	Name        string     `json:"name" db:"name"`
	Key         string     `json:"key" db:"key"`
	Type        string     `json:"type" db:"type"` // number, text, email, select
	Label       string     `json:"label" db:"label"`
	Required    bool       `json:"required" db:"required"`
	MinLength   *int       `json:"minLength" db:"min_length"`
	MaxLength   *int       `json:"maxLength" db:"max_length"`
	Placeholder *string    `json:"placeholder" db:"placeholder"`
	Pattern     *string    `json:"pattern" db:"pattern"`
	Hint        *string    `json:"hint" db:"hint"`
	Order       int        `json:"order" db:"order"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

type Category struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	Icon        *string   `json:"icon" db:"icon"`
	Order       int       `json:"order" db:"order"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type CategoryRegion struct {
	CategoryID uuid.UUID `json:"categoryId" db:"category_id"`
	RegionCode string    `json:"regionCode" db:"region_code"`
	IsActive   bool      `json:"isActive" db:"is_active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type Section struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Title     string    `json:"title" db:"title"`
	Icon      *string   `json:"icon" db:"icon"`
	Order     int       `json:"order" db:"order"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type ProductSection struct {
	ProductID uuid.UUID `json:"productId" db:"product_id"`
	SectionID uuid.UUID `json:"sectionId" db:"section_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Response DTOs
type ProductResponse struct {
	Code        string         `json:"code"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Subtitle    *string        `json:"subtitle,omitempty"`
	Description *string        `json:"description,omitempty"`
	Publisher   *string        `json:"publisher,omitempty"`
	Thumbnail   *string        `json:"thumbnail,omitempty"`
	Banner      *string        `json:"banner,omitempty"`
	IsPopular   bool           `json:"isPopular"`
	IsAvailable bool           `json:"isAvailable"`
	Tags        []string       `json:"tags,omitempty"`
	Category    CategoryInfo   `json:"category"`
	Features    []string       `json:"features,omitempty"`
	HowToOrder  []string       `json:"howToOrder,omitempty"`
}

type CategoryInfo struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

type CategoryResponse struct {
	Title       string  `json:"title"`
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Order       int     `json:"order"`
}

type SectionResponse struct {
	Title string  `json:"title"`
	Code  string  `json:"code"`
	Icon  *string `json:"icon,omitempty"`
	Order int     `json:"order"`
}

type FieldResponse struct {
	Name        string  `json:"name"`
	Key         string  `json:"key"`
	Type        string  `json:"type"`
	Label       string  `json:"label"`
	Required    bool    `json:"required"`
	MinLength   *int    `json:"minLength,omitempty"`
	MaxLength   *int    `json:"maxLength,omitempty"`
	Placeholder *string `json:"placeholder,omitempty"`
	Pattern     *string `json:"pattern,omitempty"`
	Hint        *string `json:"hint,omitempty"`
}

// Admin Response DTOs
type AdminProductResponse struct {
	ID          string            `json:"id"`
	Code        string            `json:"code"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Subtitle    *string           `json:"subtitle,omitempty"`
	Publisher   *string           `json:"publisher,omitempty"`
	Thumbnail   *string           `json:"thumbnail,omitempty"`
	Category    CategoryInfo      `json:"category"`
	IsActive    bool              `json:"isActive"`
	IsPopular   bool              `json:"isPopular"`
	Regions     []string          `json:"regions"`
	SKUCount    int               `json:"skuCount"`
	Stats       *ProductStats     `json:"stats,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type ProductStats struct {
	TodayTransactions int     `json:"todayTransactions"`
	TodayRevenue      float64 `json:"todayRevenue"`
}

type AdminCategoryResponse struct {
	ID           string         `json:"id"`
	Code         string         `json:"code"`
	Title        string         `json:"title"`
	Description  *string        `json:"description,omitempty"`
	Icon         *string        `json:"icon,omitempty"`
	IsActive     bool           `json:"isActive"`
	Order        int            `json:"order"`
	Regions      []string       `json:"regions"`
	ProductCount int            `json:"productCount"`
	Stats        *CategoryStats `json:"stats,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

type CategoryStats struct {
	TotalTransactions int     `json:"totalTransactions"`
	TotalRevenue      float64 `json:"totalRevenue"`
}

type AdminSectionResponse struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Title     string    `json:"title"`
	Icon      *string   `json:"icon,omitempty"`
	IsActive  bool      `json:"isActive"`
	Order     int       `json:"order"`
	Products  []string  `json:"products"`
	SKUCount  int       `json:"skuCount"`
	CreatedAt time.Time `json:"createdAt"`
}

// Request DTOs
type CreateProductRequest struct {
	Code        string   `json:"code" validate:"required,min=1,max=50"`
	Slug        string   `json:"slug" validate:"required,min=1,max=100"`
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Subtitle    *string  `json:"subtitle" validate:"omitempty,max=200"`
	Description *string  `json:"description" validate:"omitempty,max=5000"`
	Publisher   *string  `json:"publisher" validate:"omitempty,max=100"`
	CategoryCode string  `json:"categoryCode" validate:"required"`
	IsActive    bool     `json:"isActive"`
	IsPopular   bool     `json:"isPopular"`
	Regions     []string `json:"regions" validate:"required,min=1"`
	Features    []string `json:"features"`
	HowToOrder  []string `json:"howToOrder"`
	Tags        []string `json:"tags"`
}

type UpdateProductRequest struct {
	Title       *string  `json:"title" validate:"omitempty,min=1,max=200"`
	Subtitle    *string  `json:"subtitle" validate:"omitempty,max=200"`
	Description *string  `json:"description" validate:"omitempty,max=5000"`
	Publisher   *string  `json:"publisher" validate:"omitempty,max=100"`
	CategoryCode *string `json:"categoryCode"`
	IsActive    *bool    `json:"isActive"`
	IsPopular   *bool    `json:"isPopular"`
	Regions     []string `json:"regions"`
	Features    []string `json:"features"`
	HowToOrder  []string `json:"howToOrder"`
	Tags        []string `json:"tags"`
}

type UpdateFieldsRequest struct {
	Fields []FieldRequest `json:"fields" validate:"required,dive"`
}

type FieldRequest struct {
	Name        string  `json:"name" validate:"required"`
	Key         string  `json:"key" validate:"required"`
	Type        string  `json:"type" validate:"required,oneof=number text email select"`
	Label       string  `json:"label" validate:"required"`
	Required    bool    `json:"required"`
	MinLength   *int    `json:"minLength"`
	MaxLength   *int    `json:"maxLength"`
	Placeholder *string `json:"placeholder"`
	Pattern     *string `json:"pattern"`
	Hint        *string `json:"hint"`
}

type CreateCategoryRequest struct {
	Code        string   `json:"code" validate:"required,min=1,max=50"`
	Title       string   `json:"title" validate:"required,min=1,max=100"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	IsActive    bool     `json:"isActive"`
	Order       int      `json:"order"`
	Regions     []string `json:"regions" validate:"required,min=1"`
}

type UpdateCategoryRequest struct {
	Title       *string  `json:"title" validate:"omitempty,min=1,max=100"`
	Description *string  `json:"description" validate:"omitempty,max=500"`
	IsActive    *bool    `json:"isActive"`
	Order       *int     `json:"order"`
	Regions     []string `json:"regions"`
}

type CreateSectionRequest struct {
	Code     string   `json:"code" validate:"required,min=1,max=50"`
	Title    string   `json:"title" validate:"required,min=1,max=100"`
	Icon     *string  `json:"icon"`
	IsActive bool     `json:"isActive"`
	Order    int      `json:"order"`
	Products []string `json:"products"`
}

type UpdateSectionRequest struct {
	Title    *string  `json:"title" validate:"omitempty,min=1,max=100"`
	Icon     *string  `json:"icon"`
	IsActive *bool    `json:"isActive"`
	Order    *int     `json:"order"`
}

type AssignSectionProductsRequest struct {
	Products []string `json:"products" validate:"required"`
}

