package public

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"gate-v2/internal/middleware"
	"gate-v2/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// ============================================
// REGION HANDLERS
// ============================================

func handleGetRegionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT code, country, currency, image, is_default
			FROM regions
			WHERE is_active = true
			ORDER BY sort_order ASC
		`)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var regions []map[string]interface{}
		for rows.Next() {
			var code, country, currency string
			var image *string
			var isDefault bool

			if err := rows.Scan(&code, &country, &currency, &image, &isDefault); err != nil {
				continue
			}

			region := map[string]interface{}{
				"country":   country,
				"code":      code,
				"currency":  currency,
				"isDefault": isDefault,
			}
			if image != nil {
				region["image"] = *image
			}
			regions = append(regions, region)
		}

		if regions == nil {
			regions = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, regions)
	}
}

// ============================================
// LANGUAGE HANDLERS
// ============================================

func handleGetLanguagesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT code, name, country, image, is_default
			FROM languages
			WHERE is_active = true
			ORDER BY sort_order ASC
		`)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var languages []map[string]interface{}
		for rows.Next() {
			var code, name, country string
			var image *string
			var isDefault bool

			if err := rows.Scan(&code, &name, &country, &image, &isDefault); err != nil {
				continue
			}

			lang := map[string]interface{}{
				"code":      code,
				"name":      name,
				"country":   country,
				"isDefault": isDefault,
			}
			if image != nil {
				lang["image"] = *image
			}
			languages = append(languages, lang)
		}

		if languages == nil {
			languages = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, languages)
	}
}

// ============================================
// CATEGORY HANDLERS
// ============================================

func handleGetCategoriesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT c.code, c.title, c.description, c.icon, c.sort_order
			FROM categories c
			JOIN category_regions cr ON c.id = cr.category_id
			WHERE c.is_active = true AND cr.region_code = $1
			ORDER BY c.sort_order ASC
		`, region)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var categories []map[string]interface{}
		for rows.Next() {
			var code, title string
			var description, icon *string
			var sortOrder int

			if err := rows.Scan(&code, &title, &description, &icon, &sortOrder); err != nil {
				continue
			}

			cat := map[string]interface{}{
				"title": title,
				"code":  code,
				"order": sortOrder,
			}
			if description != nil {
				cat["description"] = *description
			}
			if icon != nil {
				cat["icon"] = *icon
			}
			categories = append(categories, cat)
		}

		if categories == nil {
			categories = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, categories)
	}
}

// ============================================
// PRODUCT HANDLERS
// ============================================

func handleGetProductsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		productCode := r.URL.Query().Get("productCode")
		categoryCode := r.URL.Query().Get("categoryCode")

		// If productCode is provided, return single product
		if productCode != "" {
			var code, slug, title, catCode, catTitle string
			var subtitle, description, publisher, thumbnail, banner *string
			var isPopular, isActive bool

			err := deps.DB.Pool.QueryRow(ctx, `
				SELECT p.code, p.slug, p.title, p.subtitle, p.description, p.publisher,
				       p.thumbnail, p.banner, p.is_popular, p.is_active,
				       c.code as category_code, c.title as category_title
				FROM products p
				JOIN categories c ON p.category_id = c.id
				JOIN product_regions pr ON p.id = pr.product_id
				WHERE p.code = $1 AND p.is_active = true AND pr.region_code = $2
			`, productCode, region).Scan(&code, &slug, &title, &subtitle, &description, &publisher,
				&thumbnail, &banner, &isPopular, &isActive, &catCode, &catTitle)

			if err != nil {
				utils.WriteNotFoundError(w, "Product")
				return
			}

			product := map[string]interface{}{
				"code":        code,
				"slug":        slug,
				"title":       title,
				"isPopular":   isPopular,
				"isAvailable": isActive,
				"tags":        []string{}, // TODO: implement tags from database
				"category": map[string]interface{}{
					"title": catTitle,
					"code":  catCode,
				},
			}
			if subtitle != nil {
				product["subtitle"] = *subtitle
			}
			if description != nil {
				product["description"] = *description
			}
			if publisher != nil {
				product["publisher"] = *publisher
			}
			if thumbnail != nil {
				product["thumbnail"] = *thumbnail
			}
			if banner != nil {
				product["banner"] = *banner
			}

			utils.WriteSuccessJSON(w, product)
			return
		}

		// Otherwise, return list of products
		var rows interface{ Close() }
		var err error

		query := `
			SELECT p.code, p.slug, p.title, p.subtitle, p.description, p.publisher,
			       p.thumbnail, p.banner, p.is_popular, p.is_active,
			       c.code as category_code, c.title as category_title
			FROM products p
			JOIN categories c ON p.category_id = c.id
			JOIN product_regions pr ON p.id = pr.product_id
			WHERE p.is_active = true AND pr.region_code = $1
		`

		if categoryCode != "" {
			query += ` AND c.code = $2`
			query += ` ORDER BY p.is_popular DESC, p.title ASC`
			rows, err = deps.DB.Pool.Query(ctx, query, region, categoryCode)
		} else {
			query += ` ORDER BY p.is_popular DESC, p.title ASC`
			rows, err = deps.DB.Pool.Query(ctx, query, region)
		}

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var products []map[string]interface{}
		rowsTyped := rows.(interface {
			Next() bool
			Scan(dest ...interface{}) error
		})

		for rowsTyped.Next() {
			var code, slug, title, catCode, catTitle string
			var subtitle, description, publisher, thumbnail, banner *string
			var isPopular, isActive bool

			if err := rowsTyped.Scan(&code, &slug, &title, &subtitle, &description, &publisher,
				&thumbnail, &banner, &isPopular, &isActive,
				&catCode, &catTitle); err != nil {
				continue
			}

			product := map[string]interface{}{
				"code":        code,
				"slug":        slug,
				"title":       title,
				"isPopular":   isPopular,
				"isAvailable": isActive,
				"tags":        []string{}, // TODO: implement tags from database
				"category": map[string]interface{}{
					"title": catTitle,
					"code":  catCode,
				},
			}
			if subtitle != nil {
				product["subtitle"] = *subtitle
			}
			if description != nil {
				product["description"] = *description
			}
			if publisher != nil {
				product["publisher"] = *publisher
			}
			if thumbnail != nil {
				product["thumbnail"] = *thumbnail
			}
			if banner != nil {
				product["banner"] = *banner
			}
			products = append(products, product)
		}

		if products == nil {
			products = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, products)
	}
}

func handleGetPopularProductsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT p.code, p.slug, p.title, p.subtitle, p.description, p.publisher,
			       p.thumbnail, p.banner, p.is_popular, p.is_active,
			       c.code as category_code, c.title as category_title
			FROM products p
			JOIN categories c ON p.category_id = c.id
			JOIN product_regions pr ON p.id = pr.product_id
			WHERE p.is_active = true AND p.is_popular = true AND pr.region_code = $1
			ORDER BY p.title ASC
			LIMIT 12
		`, region)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var products []map[string]interface{}
		for rows.Next() {
			var code, slug, title, categoryCode, categoryTitle string
			var subtitle, description, publisher, thumbnail, banner *string
			var isPopular, isActive bool

			if err := rows.Scan(&code, &slug, &title, &subtitle, &description, &publisher,
				&thumbnail, &banner, &isPopular, &isActive,
				&categoryCode, &categoryTitle); err != nil {
				continue
			}

			product := map[string]interface{}{
				"code":        code,
				"slug":        slug,
				"title":       title,
				"isPopular":   isPopular,
				"isAvailable": isActive,
				"tags":        []string{},
				"category": map[string]interface{}{
					"title": categoryTitle,
					"code":  categoryCode,
				},
			}
			if subtitle != nil {
				product["subtitle"] = *subtitle
			}
			if description != nil {
				product["description"] = *description
			}
			if publisher != nil {
				product["publisher"] = *publisher
			}
			if thumbnail != nil {
				product["thumbnail"] = *thumbnail
			}
			if banner != nil {
				product["banner"] = *banner
			}
			products = append(products, product)
		}

		if products == nil {
			products = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, products)
	}
}

func handleGetProductDetailImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		slug := chi.URLParam(r, "slug")
		if slug == "" {
			utils.WriteNotFoundError(w, "Product")
			return
		}

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		var id, code, title, categoryCode, categoryTitle string
		var subtitle, description, publisher, thumbnail, banner *string
		var features, howToOrder []byte

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT p.id, p.code, p.title, p.subtitle, p.description, p.publisher,
			       p.thumbnail, p.banner, p.features, p.how_to_order,
			       c.code as category_code, c.title as category_title
			FROM products p
			JOIN categories c ON p.category_id = c.id
			JOIN product_regions pr ON p.id = pr.product_id
			WHERE p.slug = $1 AND p.is_active = true AND pr.region_code = $2
		`, slug, region).Scan(&id, &code, &title, &subtitle, &description, &publisher,
			&thumbnail, &banner, &features, &howToOrder, &categoryCode, &categoryTitle)

		if err != nil {
			utils.WriteNotFoundError(w, "Product")
			return
		}

		product := map[string]interface{}{
			"id":    id,
			"code":  code,
			"slug":  slug,
			"title": title,
			"category": map[string]interface{}{
				"code":  categoryCode,
				"title": categoryTitle,
			},
		}

		if subtitle != nil {
			product["subtitle"] = *subtitle
		}
		if description != nil {
			product["description"] = *description
		}
		if publisher != nil {
			product["publisher"] = *publisher
		}
		if thumbnail != nil {
			product["thumbnail"] = *thumbnail
		}
		if banner != nil {
			product["banner"] = *banner
		}

		utils.WriteSuccessJSON(w, product)
	}
}

// ============================================
// SKU HANDLERS
// ============================================

func handleGetSKUsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Support both path param and query param
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			slug = r.URL.Query().Get("slug")
		}
		// Also support productCode query param
		if slug == "" {
			slug = r.URL.Query().Get("productCode")
		}
		if slug == "" {
			utils.WriteBadRequestError(w, "Product slug is required")
			return
		}

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT s.code, s.name, s.description, sp.currency, sp.sell_price, sp.original_price,
			       s.image, s.info, s.process_time, s.is_active, s.is_featured,
			       s.badge_text, s.badge_color,
			       sec.code as section_code, sec.title as section_title
			FROM skus s
			JOIN products p ON s.product_id = p.id
			JOIN sku_pricing sp ON s.id = sp.sku_id
			LEFT JOIN sections sec ON s.section_id = sec.id
			WHERE (p.slug = $1 OR p.code = $1) AND s.is_active = true AND sp.region_code = $2 AND sp.is_active = true
			ORDER BY sec.sort_order ASC, sp.sell_price ASC
		`, slug, region)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var skus []map[string]interface{}
		for rows.Next() {
			var code, name, currency string
			var description, image, info, badgeText, badgeColor, sectionCode, sectionTitle *string
			var processTime int
			var isActive, isFeatured bool
			var sellPrice, originalPrice int64

			if err := rows.Scan(&code, &name, &description, &currency, &sellPrice, &originalPrice,
				&image, &info, &processTime, &isActive, &isFeatured,
				&badgeText, &badgeColor,
				&sectionCode, &sectionTitle); err != nil {
				continue
			}

			// Calculate discount percentage
			var discount float64
			if originalPrice > 0 && originalPrice > sellPrice {
				discount = float64(originalPrice-sellPrice) / float64(originalPrice) * 100
			}

			sku := map[string]interface{}{
				"code":          code,
				"name":          name,
				"currency":      currency,
				"price":         sellPrice,
				"originalPrice": originalPrice,
				"discount":      discount,
				"processTime":   processTime,
				"isAvailable":   isActive,
				"isFeatured":    isFeatured,
			}

			if description != nil {
				sku["description"] = *description
			}
			if image != nil {
				sku["image"] = *image
			}
			if info != nil {
				sku["info"] = *info
			}
			if badgeText != nil && *badgeText != "" {
				sku["badge"] = map[string]interface{}{
					"text":  *badgeText,
					"color": *badgeColor,
				}
			}
			if sectionCode != nil {
				sku["section"] = map[string]interface{}{
					"title": *sectionTitle,
					"code":  *sectionCode,
				}
			}

			skus = append(skus, sku)
		}

		if skus == nil {
			skus = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, skus)
	}
}

// handleGetPromoSKUsImpl returns SKUs that have discounts (original_price > sell_price)
func handleGetPromoSKUsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT s.code, s.name, s.description, sp.currency, sp.sell_price, sp.original_price,
			       s.image, s.info, s.process_time, s.is_active, s.is_featured,
			       s.badge_text, s.badge_color,
			       sec.code as section_code, sec.title as section_title,
			       p.code as product_code, p.title as product_title, p.slug as product_slug
			FROM skus s
			JOIN products p ON s.product_id = p.id
			JOIN sku_pricing sp ON s.id = sp.sku_id
			LEFT JOIN sections sec ON s.section_id = sec.id
			WHERE s.is_active = true 
			  AND sp.region_code = $1 
			  AND sp.is_active = true
			  AND sp.original_price > sp.sell_price
			ORDER BY (sp.original_price - sp.sell_price) DESC
			LIMIT 20
		`, region)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var skus []map[string]interface{}
		for rows.Next() {
			var code, name, currency, productCode, productTitle, productSlug string
			var description, image, info, badgeText, badgeColor, sectionCode, sectionTitle *string
			var processTime int
			var isActive, isFeatured bool
			var sellPrice, originalPrice int64

			if err := rows.Scan(&code, &name, &description, &currency, &sellPrice, &originalPrice,
				&image, &info, &processTime, &isActive, &isFeatured,
				&badgeText, &badgeColor,
				&sectionCode, &sectionTitle,
				&productCode, &productTitle, &productSlug); err != nil {
				continue
			}

			// Calculate discount percentage
			var discount float64
			if originalPrice > 0 && originalPrice > sellPrice {
				discount = float64(originalPrice-sellPrice) / float64(originalPrice) * 100
			}

			sku := map[string]interface{}{
				"code":          code,
				"name":          name,
				"currency":      currency,
				"price":         sellPrice,
				"originalPrice": originalPrice,
				"discount":      discount,
				"processTime":   processTime,
				"isAvailable":   isActive,
				"isFeatured":    isFeatured,
				"product": map[string]interface{}{
					"code":  productCode,
					"title": productTitle,
					"slug":  productSlug,
				},
			}

			if description != nil {
				sku["description"] = *description
			}
			if image != nil {
				sku["image"] = *image
			}
			if info != nil {
				sku["info"] = *info
			}
			if badgeText != nil && *badgeText != "" {
				sku["badge"] = map[string]interface{}{
					"text":  *badgeText,
					"color": *badgeColor,
				}
			}
			if sectionCode != nil {
				sku["section"] = map[string]interface{}{
					"title": *sectionTitle,
					"code":  *sectionCode,
				}
			}

			skus = append(skus, sku)
		}

		if skus == nil {
			skus = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, skus)
	}
}

// ============================================
// BANNER HANDLERS
// ============================================

func handleGetBannersImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT b.title, b.description, b.href, b.image, b.sort_order
			FROM banners b
			JOIN banner_regions br ON b.id = br.banner_id
			WHERE b.is_active = true 
			  AND br.region_code = $1
			  AND (b.start_at IS NULL OR b.start_at <= NOW())
			  AND (b.expired_at IS NULL OR b.expired_at > NOW())
			ORDER BY b.sort_order ASC
		`, region)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var banners []map[string]interface{}
		for rows.Next() {
			var image string
			var title, description, href *string
			var sortOrder int

			if err := rows.Scan(&title, &description, &href, &image, &sortOrder); err != nil {
				continue
			}

			banner := map[string]interface{}{
				"image": image,
				"order": sortOrder,
			}
			if title != nil {
				banner["title"] = *title
			}
			if description != nil {
				banner["description"] = *description
			}
			if href != nil {
				banner["href"] = *href
			}
			banners = append(banners, banner)
		}

		if banners == nil {
			banners = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, banners)
	}
}

// ============================================
// POPUP HANDLERS
// ============================================

func handleGetPopupsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		var title, content, image, href *string
		var isActive bool

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT title, content, image, href, is_active
			FROM popups
			WHERE region_code = $1 AND is_active = true
		`, region).Scan(&title, &content, &image, &href, &isActive)

		if err != nil {
			utils.WriteSuccessJSON(w, nil)
			return
		}

		popup := map[string]interface{}{
			"isActive": isActive,
		}
		if title != nil {
			popup["title"] = *title
		}
		if content != nil {
			popup["content"] = *content
		}
		if image != nil {
			popup["image"] = *image
		}
		if href != nil {
			popup["href"] = *href
		}

		utils.WriteSuccessJSON(w, popup)
	}
}

// ============================================
// CONTACT HANDLERS
// ============================================

func handleGetContactsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var email, phone *string
		var whatsapp, instagram, facebook, x, youtube, telegram, discord *string

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT email, phone, whatsapp, instagram, facebook, x, youtube, telegram, discord
			FROM contacts
			LIMIT 1
		`).Scan(&email, &phone, &whatsapp, &instagram, &facebook, &x, &youtube, &telegram, &discord)

		if err != nil {
			utils.WriteSuccessJSON(w, map[string]interface{}{})
			return
		}

		contacts := map[string]interface{}{}
		if email != nil {
			contacts["email"] = *email
		}
		if phone != nil {
			contacts["phone"] = *phone
		}
		if whatsapp != nil {
			contacts["whatsapp"] = *whatsapp
		}
		if instagram != nil {
			contacts["instagram"] = *instagram
		}
		if facebook != nil {
			contacts["facebook"] = *facebook
		}
		if x != nil {
			contacts["x"] = *x
		}
		if youtube != nil {
			contacts["youtube"] = *youtube
		}
		if telegram != nil {
			contacts["telegram"] = *telegram
		}
		if discord != nil {
			contacts["discord"] = *discord
		}

		utils.WriteSuccessJSON(w, contacts)
	}
}

// ============================================
// PROMO HANDLERS
// ============================================

func handleGetPromosImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT p.code, p.title, p.description, p.note,
			       p.max_daily_usage, p.max_usage, p.max_usage_per_id, p.max_usage_per_device, p.max_usage_per_ip,
			       p.expired_at, p.min_amount, p.max_promo_amount, p.promo_flat, p.promo_percentage,
			       p.is_active, p.total_usage, p.days_available
			FROM promos p
			JOIN promo_regions pr ON p.id = pr.promo_id
			WHERE p.is_active = true 
			  AND pr.region_code = $1
			  AND (p.start_at IS NULL OR p.start_at <= NOW())
			  AND (p.expired_at IS NULL OR p.expired_at > NOW())
			ORDER BY p.created_at DESC
		`, region)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var promos []map[string]interface{}
		for rows.Next() {
			var code, title string
			var description, note *string
			var maxDailyUsage, maxUsage, maxUsagePerId, maxUsagePerDevice, maxUsagePerIp int
			var minAmount, maxPromoAmount, promoFlat, totalUsage int64
			var promoPercentage float64
			var isActive bool
			var expiredAt *time.Time
			var daysAvailable []string

			if err := rows.Scan(&code, &title, &description, &note,
				&maxDailyUsage, &maxUsage, &maxUsagePerId, &maxUsagePerDevice, &maxUsagePerIp,
				&expiredAt, &minAmount, &maxPromoAmount, &promoFlat, &promoPercentage,
				&isActive, &totalUsage, &daysAvailable); err != nil {
				continue
			}

			promo := map[string]interface{}{
				"code":              code,
				"title":             title,
				"products":          []interface{}{}, // TODO: fetch from promo_products table
				"paymentChannels":   []interface{}{}, // TODO: fetch from promo_payment_channels table
				"daysAvailable":     daysAvailable,
				"maxDailyUsage":     maxDailyUsage,
				"maxUsage":          maxUsage,
				"maxUsagePerId":     maxUsagePerId,
				"maxUsagePerDevice": maxUsagePerDevice,
				"maxUsagePerIp":     maxUsagePerIp,
				"minAmount":         minAmount,
				"maxPromoAmount":    maxPromoAmount,
				"promoFlat":         promoFlat,
				"promoPercentage":   promoPercentage,
				"isAvailable":       isActive,
				"totalUsage":        totalUsage,
				"totalDailyUsage":   0, // TODO: calculate from today's usage
			}
			if description != nil {
				promo["description"] = *description
			}
			if note != nil {
				promo["note"] = *note
			}
			if expiredAt != nil {
				promo["expiredAt"] = expiredAt.Format(time.RFC3339)
			}
			promos = append(promos, promo)
		}

		if promos == nil {
			promos = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, promos)
	}
}

// ============================================
// PAYMENT HANDLERS
// ============================================

func handleGetPaymentCategoriesImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT code, title, icon, sort_order
			FROM payment_channel_categories
			WHERE is_active = true
			ORDER BY sort_order ASC
		`)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var categories []map[string]interface{}
		for rows.Next() {
			var code, title string
			var icon *string
			var sortOrder int

			if err := rows.Scan(&code, &title, &icon, &sortOrder); err != nil {
				continue
			}

			cat := map[string]interface{}{
				"title": title,
				"code":  code,
				"order": sortOrder,
			}
			if icon != nil {
				cat["icon"] = *icon
			}
			categories = append(categories, cat)
		}

		if categories == nil {
			categories = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, categories)
	}
}

func handleGetPaymentChannelsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID"
		}

		paymentType := r.URL.Query().Get("paymentType")
		if paymentType == "" {
			paymentType = "purchase"
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT pc.code, pc.name, pc.description, pc.image, r.currency,
			       pc.fee_amount, pc.fee_percentage,
			       pc.min_amount, pc.max_amount, pc.instruction, pc.is_featured,
			       pcc.code as category_code, pcc.title as category_title
			FROM payment_channels pc
			LEFT JOIN payment_channel_categories pcc ON pc.category_id = pcc.id
			JOIN payment_channel_regions pcr ON pc.id = pcr.channel_id
			JOIN regions r ON pcr.region_code = r.code
			WHERE pc.is_active = true 
			  AND pcr.region_code = $1
			  AND $2 = ANY(pc.supported_types)
			ORDER BY 
			  CASE WHEN pcc.id IS NULL THEN 0 ELSE 1 END ASC,
			  pc.is_featured DESC, 
			  COALESCE(pcc.sort_order, 999) ASC, 
			  pc.sort_order ASC
		`, region, paymentType)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var channels []map[string]interface{}
		for rows.Next() {
			var code, name, currency string
			var description, image, instruction *string
			var feeAmount, minAmount, maxAmount int64
			var feePercentage float64
			var isFeatured bool
			var categoryCode, categoryTitle sql.NullString

			if err := rows.Scan(&code, &name, &description, &image, &currency,
				&feeAmount, &feePercentage,
				&minAmount, &maxAmount, &instruction, &isFeatured,
				&categoryCode, &categoryTitle); err != nil {
				continue
			}

			channel := map[string]interface{}{
				"code":          code,
				"name":          name,
				"currency":      currency,
				"feeAmount":     feeAmount,
				"feePercentage": feePercentage,
				"minAmount":     minAmount,
				"maxAmount":     maxAmount,
				"featured":      isFeatured,
			}

			// Only include category if it exists
			if categoryCode.Valid && categoryTitle.Valid {
				channel["category"] = map[string]interface{}{
					"title": categoryTitle.String,
					"code":  categoryCode.String,
				}
			}

			if description != nil {
				channel["description"] = *description
			}
			if image != nil {
				channel["image"] = *image
			}
			if instruction != nil {
				channel["instruction"] = *instruction
			}
			channels = append(channels, channel)
		}

		if channels == nil {
			channels = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, channels)
	}
}

// ============================================
// SECTION HANDLERS
// ============================================

func handleGetSectionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Support both path param and query param
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			slug = r.URL.Query().Get("slug")
		}
		// Also support productCode query param
		if slug == "" {
			slug = r.URL.Query().Get("productCode")
		}
		if slug == "" {
			utils.WriteBadRequestError(w, "Product slug is required")
			return
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT DISTINCT s.code, s.title, s.icon, s.sort_order
			FROM sections s
			JOIN skus sk ON sk.section_id = s.id
			JOIN products p ON sk.product_id = p.id
			WHERE (p.slug = $1 OR p.code = $1) AND s.is_active = true AND sk.is_active = true
			ORDER BY s.sort_order ASC
		`, slug)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var sections []map[string]interface{}
		for rows.Next() {
			var code, title string
			var icon *string
			var sortOrder int

			if err := rows.Scan(&code, &title, &icon, &sortOrder); err != nil {
				continue
			}

			section := map[string]interface{}{
				"title": title,
				"code":  code,
				"order": sortOrder,
			}
			if icon != nil {
				section["icon"] = *icon
			}
			sections = append(sections, section)
		}

		if sections == nil {
			sections = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, sections)
	}
}

// ============================================
// PRODUCT FIELDS HANDLERS
// ============================================

func handleGetFieldsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Support both path param and query param
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			slug = r.URL.Query().Get("slug")
		}
		// Also support productCode query param
		if slug == "" {
			slug = r.URL.Query().Get("productCode")
		}
		if slug == "" {
			utils.WriteBadRequestError(w, "Product slug is required")
			return
		}

		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT pf.name, pf.key, pf.field_type, pf.label, 
			       pf.is_required, pf.min_length, pf.max_length,
			       pf.placeholder, pf.pattern, pf.hint
			FROM product_fields pf
			JOIN products p ON pf.product_id = p.id
			WHERE p.slug = $1 OR p.code = $1
			ORDER BY pf.sort_order ASC
		`, slug)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		var fields []map[string]interface{}
		for rows.Next() {
			var name, key, fieldType, label string
			var placeholder, hint, pattern *string
			var isRequired bool
			var minLength, maxLength *int

			if err := rows.Scan(&name, &key, &fieldType, &label,
				&isRequired, &minLength, &maxLength,
				&placeholder, &pattern, &hint); err != nil {
				continue
			}

			field := map[string]interface{}{
				"name":     name,
				"key":      key,
				"type":     fieldType,
				"label":    label,
				"required": isRequired,
			}
			// Always include minLength and maxLength, even if null
			if minLength != nil {
				field["minLength"] = *minLength
			} else {
				field["minLength"] = nil
			}
			if maxLength != nil {
				field["maxLength"] = *maxLength
			} else {
				field["maxLength"] = nil
			}
			if placeholder != nil && *placeholder != "" {
				field["placeholder"] = *placeholder
			}
			if pattern != nil && *pattern != "" {
				field["pattern"] = *pattern
			}
			if hint != nil && *hint != "" {
				field["hint"] = *hint
			}
			fields = append(fields, field)
		}

		if fields == nil {
			fields = []map[string]interface{}{}
		}

		utils.WriteSuccessJSON(w, fields)
	}
}

// ============================================
// INVOICE HANDLERS
// ============================================

func handleGetInvoiceImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get invoice number from query parameter
		invoiceNumber := r.URL.Query().Get("invoiceNumber")
		if invoiceNumber == "" {
			log.Warn().
				Str("endpoint", "/v2/invoices").
				Str("error_type", "VALIDATION_ERROR").
				Msg("Invoice number parameter is missing")
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"invoiceNumber": "Invoice number parameter is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		log.Info().
			Str("endpoint", "/v2/invoices").
			Str("invoice_number", invoiceNumber).
			Msg("Fetching invoice details")

		// Query transaction with related data using joins
		// Note: Prices in database are stored in rupiah (not cents)
		var (
			id               string
			status           string
			paymentStatus    string
			productCode      string
			productName      string
			skuCode          string
			skuName          string
			quantity         int
			buyPrice         int64
			sellPrice        int64
			discount         int64
			paymentFee       int64
			total            int64
			currency         string
			paymentCode      string
			paymentName      string
			paymentCategoryCode *string
			providerResponse *string
			accountNickname  *string
			accountInputs    string
			email            *string
			phoneNumber      *string
			serialNumber     *string
			paidAt           *time.Time
			processedAt      *time.Time
			completedAt      *time.Time
			expiredAt        time.Time
			createdAt        time.Time
		)

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				t.id,
				t.status,
				t.payment_status,
				p.code,
				p.title,
				s.code,
				s.name,
				t.quantity,
				t.buy_price,
				t.sell_price,
				t.discount_amount,
				t.payment_fee,
				t.total_amount,
				t.currency,
				pc.code,
				pc.name,
				pcc.code,
				t.provider_response,
				t.account_nickname,
				t.account_inputs,
				t.contact_email,
				t.contact_phone,
				t.provider_serial_number,
				t.paid_at,
				t.processed_at,
				t.completed_at,
				t.expired_at,
				t.created_at
			FROM transactions t
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			LEFT JOIN payment_channel_categories pcc ON pc.category_id = pcc.id
			WHERE t.invoice_number = $1
		`, invoiceNumber).Scan(
			&id, &status, &paymentStatus, &productCode, &productName,
			&skuCode, &skuName, &quantity, &buyPrice, &sellPrice,
			&discount, &paymentFee, &total, &currency, &paymentCode, &paymentName,
			&paymentCategoryCode, &providerResponse,
			&accountNickname, &accountInputs, &email, &phoneNumber,
			&serialNumber, &paidAt, &processedAt, &completedAt, &expiredAt, &createdAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				log.Warn().
					Str("endpoint", "/v2/invoices").
					Str("error_type", "INVOICE_NOT_FOUND").
					Str("invoice_number", invoiceNumber).
					Msg("Invoice not found in database")
				utils.WriteErrorJSON(w, http.StatusNotFound, "INVOICE_NOT_FOUND",
					"Invoice not found", "The invoice number does not exist")
				return
			}
			log.Error().
				Err(err).
				Str("endpoint", "/v2/invoices").
				Str("error_type", "DB_QUERY_ERROR").
				Str("invoice_number", invoiceNumber).
				Msg("Failed to query invoice from database")
			utils.WriteInternalServerError(w)
			return
		}

		log.Info().
			Str("endpoint", "/v2/invoices").
			Str("invoice_number", invoiceNumber).
			Str("transaction_id", id).
			Str("status", status).
			Str("payment_status", paymentStatus).
			Msg("Invoice found successfully")

		// Fetch timeline entries
		timelineRows, err := deps.DB.Pool.Query(ctx, `
			SELECT status, message, created_at
			FROM transaction_logs
			WHERE transaction_id = $1
			ORDER BY created_at ASC
		`, id)
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/invoices").
				Str("error_type", "TIMELINE_QUERY_ERROR").
				Str("transaction_id", id).
				Msg("Failed to query transaction timeline")
			// Non-fatal, continue without timeline
		}
		defer timelineRows.Close()

		var timeline []map[string]interface{}
		for timelineRows.Next() {
			var tlStatus, tlMessage string
			var tlTimestamp time.Time
			if err := timelineRows.Scan(&tlStatus, &tlMessage, &tlTimestamp); err != nil {
				continue
			}
			timeline = append(timeline, map[string]interface{}{
				"status":    tlStatus,
				"message":   tlMessage,
				"timestamp": tlTimestamp.Format(time.RFC3339),
			})
		}

		if timeline == nil {
			timeline = []map[string]interface{}{}
		}

		// Build account object from account_inputs JSON
		accountData := map[string]interface{}{}
		if accountInputs != "" {
			var inputs map[string]interface{}
			if err := json.Unmarshal([]byte(accountInputs), &inputs); err == nil {
				if userId, ok := inputs["userId"].(string); ok && userId != "" {
					accountData["userId"] = userId
				}
				if zoneId, ok := inputs["zoneId"].(string); ok && zoneId != "" {
					accountData["zoneId"] = zoneId
				}
			}
		}
		if accountNickname != nil {
			accountData["nickname"] = *accountNickname
		}

		// Build contact object
		var contact map[string]interface{}
		if email != nil || phoneNumber != nil {
			contact = make(map[string]interface{})
			if email != nil {
				contact["email"] = *email
			}
			if phoneNumber != nil {
				contact["phoneNumber"] = *phoneNumber
			}
		}

		// Build pricing object
		// Prices in database are stored in rupiah
		pricing := map[string]interface{}{
			"subtotal":   float64(sellPrice * int64(quantity)),
			"discount":   float64(discount),
			"paymentFee": float64(paymentFee),
			"total":      float64(total),
			"currency":   currency,
		}

		// Build payment object
		payment := map[string]interface{}{
			"code": paymentCode,
			"name": paymentName,
		}
		
		// Add payment category code if available
		if paymentCategoryCode != nil && *paymentCategoryCode != "" {
			payment["categoryCode"] = *paymentCategoryCode
		}
		
		// Parse and add payment data from provider_response
		if providerResponse != nil && *providerResponse != "" {
			var paymentData map[string]interface{}
			if err := json.Unmarshal([]byte(*providerResponse), &paymentData); err == nil {
				// Unified paymentCode (QRIS string, VA number, redirect URL, or retail code)
				if pCode, ok := paymentData["paymentCode"].(string); ok && pCode != "" {
					payment["paymentCode"] = pCode
				}
				if pType, ok := paymentData["paymentType"].(string); ok && pType != "" {
					payment["paymentType"] = pType
				}
				
				// Additional data for specific payment types
				if bankCode, ok := paymentData["bankCode"].(string); ok && bankCode != "" {
					payment["bankCode"] = bankCode
				}
				if accountName, ok := paymentData["accountName"].(string); ok && accountName != "" {
					payment["accountName"] = accountName
				}
				if deeplink, ok := paymentData["deeplink"].(string); ok && deeplink != "" {
					payment["deeplink"] = deeplink
				}
				
				// Instructions
				if instructions, ok := paymentData["instructions"].([]interface{}); ok && len(instructions) > 0 {
					payment["instructions"] = instructions
				}
				
				// Expiry from payment data
				if paymentExpiredAt, ok := paymentData["expiredAt"].(string); ok && paymentExpiredAt != "" {
					payment["expiredAt"] = paymentExpiredAt
				}
			}
		}
		
		// Add expiredAt if not already set from payment data
		if _, hasExpiry := payment["expiredAt"]; !hasExpiry {
			payment["expiredAt"] = expiredAt.Format(time.RFC3339)
		}
		
		if paidAt != nil {
			payment["paidAt"] = paidAt.Format(time.RFC3339)
		}

		// Build response
		response := map[string]interface{}{
			"invoiceNumber": invoiceNumber,
			"status":        status,
			"paymentStatus": paymentStatus,
			"productCode":   productCode,
			"productName":   productName,
			"skuCode":       skuCode,
			"skuName":       skuName,
			"quantity":      quantity,
			"account":       accountData,
			"pricing":       pricing,
			"payment":       payment,
			"contact":       contact,
			"timeline":      timeline,
			"createdAt":     createdAt.Format(time.RFC3339),
			"expiredAt":     expiredAt.Format(time.RFC3339),
		}

		// Add optional fields
		if serialNumber != nil {
			response["serialNumber"] = *serialNumber
		}
		if completedAt != nil {
			response["completedAt"] = completedAt.Format(time.RFC3339)
		}
		if paidAt != nil {
			response["paidAt"] = paidAt.Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, response)
	}
}

