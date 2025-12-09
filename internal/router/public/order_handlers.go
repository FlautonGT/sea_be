package public

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"seaply/internal/middleware"
	"seaply/internal/payment"
	"seaply/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// OrderInquiryRequest represents the request body for order inquiry
type OrderInquiryRequest struct {
	ProductCode string `json:"productCode" validate:"required"`
	SKUCode     string `json:"skuCode" validate:"required"`
	UserID      string `json:"userId,omitempty"`
	ZoneID      string `json:"zoneId,omitempty"`
	ServerID    string `json:"serverId,omitempty"`
	Quantity    int    `json:"quantity,omitempty"`
	PaymentCode string `json:"paymentCode,omitempty"`
	PromoCode   string `json:"promoCode,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// handleOrderInquiryImpl implements order inquiry with account validation
func HandleOrderInquiryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req OrderInquiryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate required fields
		if req.ProductCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"productCode": "Product code is required",
			})
			return
		}

		if req.SKUCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"skuCode": "SKU code is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Get product info
		var productCode, productName, productSlug, inquirySlug string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT code, slug, title, COALESCE(inquiry_slug, '') as inquiry_slug
			FROM products
			WHERE code = $1 AND is_active = true
		`, req.ProductCode).Scan(&productCode, &productSlug, &productName, &inquirySlug)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND",
					"Product not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get SKU info
		var skuCode, skuName string
		var skuPrice int64
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT s.code, s.name, sp.sell_price
			FROM skus s
			JOIN sku_pricing sp ON s.id = sp.sku_id
			JOIN products p ON s.product_id = p.id
			WHERE s.code = $1 AND p.code = $2 AND s.is_active = true AND sp.is_active = true
			LIMIT 1
		`, req.SKUCode, req.ProductCode).Scan(&skuCode, &skuName, &skuPrice)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "SKU_NOT_FOUND",
					"SKU not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Check if product requires zone
		var hasZoneField bool
		deps.DB.Pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM product_fields 
				WHERE product_id = (SELECT id FROM products WHERE code = $1)
				AND key IN ('zoneId', 'serverId')
			)
		`, req.ProductCode).Scan(&hasZoneField)

		// Determine zone value (zoneId or serverId - serverId can be used as zoneId)
		zoneValue := req.ZoneID
		if zoneValue == "" {
			zoneValue = req.ServerID
		}

		// Validate zone if required
		if hasZoneField && zoneValue == "" && req.UserID != "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"zoneId":   "Zone ID / Server ID is required for this product",
				"serverId": "Zone ID / Server ID is required for this product",
			})
			return
		}

		// Validate email format if provided
		if req.Email != "" && !utils.ValidateEmail(req.Email) {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"email": "Invalid email format",
			})
			return
		}

		// Validate phone number format if provided
		if req.PhoneNumber != "" {
			// Normalize phone number - ensure it starts with +
			phoneNumber := strings.TrimSpace(req.PhoneNumber)
			if !strings.HasPrefix(phoneNumber, "+") {
				// If starts with 0, replace with country code based on region
				if strings.HasPrefix(phoneNumber, "0") {
					region := middleware.GetRegionFromContext(r.Context())
					if region == "" {
						region = "ID"
					}
					switch region {
					case "ID":
						phoneNumber = "+62" + phoneNumber[1:]
					case "MY":
						phoneNumber = "+60" + phoneNumber[1:]
					case "PH":
						phoneNumber = "+63" + phoneNumber[1:]
					case "SG":
						phoneNumber = "+65" + phoneNumber[1:]
					case "TH":
						phoneNumber = "+66" + phoneNumber[1:]
					default:
						phoneNumber = "+62" + phoneNumber[1:] // Default to Indonesia
					}
				} else {
					// If doesn't start with + or 0, assume it's missing country code
					region := middleware.GetRegionFromContext(r.Context())
					if region == "" {
						region = "ID"
					}
					switch region {
					case "ID":
						phoneNumber = "+62" + phoneNumber
					case "MY":
						phoneNumber = "+60" + phoneNumber
					case "PH":
						phoneNumber = "+63" + phoneNumber
					case "SG":
						phoneNumber = "+65" + phoneNumber
					case "TH":
						phoneNumber = "+66" + phoneNumber
					default:
						phoneNumber = "+62" + phoneNumber // Default to Indonesia
					}
				}
			}
			req.PhoneNumber = phoneNumber

			// Validate phone number format
			if !utils.ValidatePhone(req.PhoneNumber) {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"phoneNumber": "Invalid phone number format. Please use international format (e.g., +628123456789)",
				})
				return
			}
		}

		// Validate user ID or phone number
		if req.UserID == "" && req.PhoneNumber == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"userId":      "User ID or phone number is required",
				"phoneNumber": "User ID or phone number is required",
			})
			return
		}

		// Perform account validation using inquiry_slug (like account/inquiries) if available
		var accountNickname string
		if req.UserID != "" && inquirySlug != "" {
			// Use inquiry_slug validation (same as account/inquiries)
			// zoneValue already determined above (can be from zoneId or serverId)
			inquiryURL := fmt.Sprintf("%s/%s", deps.Config.App.InquiryBaseURL, inquirySlug)

			queryParams := url.Values{}
			queryParams.Set("id", req.UserID)
			if zoneValue != "" {
				queryParams.Set("zone", zoneValue)
			}
			queryParams.Set("key", deps.Config.App.InquiryKey)

			fullURL := fmt.Sprintf("%s?%s", inquiryURL, queryParams.Encode())

			httpClient := &http.Client{
				Timeout: 10 * time.Second,
			}

			httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			httpReq.Header.Set("Accept", "application/json")

			resp, err := httpClient.Do(httpReq)
			if err != nil {
				utils.WriteErrorJSON(w, http.StatusInternalServerError, "INQUIRY_SERVICE_ERROR",
					"Failed to connect to inquiry service", "")
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Parse response
			var inquiryResp struct {
				Data struct {
					UserName string `json:"userName"`
					UserID   string `json:"userId"`
					ZoneID   string `json:"zoneId,omitempty"`
					Region   string `json:"region,omitempty"`
				} `json:"data"`
				Error *struct {
					Code     string      `json:"code"`
					Message  string      `json:"message"`
					Response interface{} `json:"response,omitempty"`
				} `json:"error,omitempty"`
			}

			if err := json.Unmarshal(body, &inquiryResp); err != nil {
				utils.WriteErrorJSON(w, http.StatusInternalServerError, "INQUIRY_RESPONSE_ERROR",
					"Failed to parse inquiry response", "")
				return
			}

			// Handle error response from inquiry API
			if inquiryResp.Error != nil {
				errorCode := inquiryResp.Error.Code
				errorMessage := inquiryResp.Error.Message

				var userMessage string
				var httpStatus int

				switch errorCode {
				case "NOT_FOUND":
					userMessage = "Account not found"
					httpStatus = http.StatusNotFound
				case "BAD_REQUEST":
					userMessage = "Invalid request"
					httpStatus = http.StatusBadRequest
				case "TOO_MANY_REQUESTS":
					userMessage = "Too many requests"
					httpStatus = http.StatusTooManyRequests
				case "INTERNAL_ERROR":
					userMessage = "Internal server error"
					httpStatus = http.StatusInternalServerError
				default:
					userMessage = errorMessage
					if userMessage == "" {
						userMessage = "An error occurred while checking the account."
					}
					httpStatus = http.StatusInternalServerError
				}

				utils.WriteErrorJSON(w, httpStatus, errorCode, userMessage, "")
				return
			}

			// Check if userName is null/empty (account not found)
			if inquiryResp.Data.UserName == "" {
				utils.WriteErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND",
					"Account not found",
					"The provided User ID and Zone ID combination does not exist")
				return
			}

			accountNickname = inquiryResp.Data.UserName
		} else if req.UserID != "" || req.PhoneNumber != "" {
			// Fallback to game check if inquiry_slug is not available
			// NOTE: Game check functionality is currently disabled
			// If inquiry_slug is not available and user provides UserID/PhoneNumber,
			// we skip account validation for now
			// TODO: Implement game check service or use provider-based validation
		}

		// Calculate pricing
		quantity := req.Quantity
		if quantity <= 0 {
			quantity = 1
		}

		// sell_price in database is stored in rupiah (e.g., 20000 = 20000 IDR)
		// All calculations are done in rupiah
		subtotal := skuPrice * int64(quantity)
		discount := int64(0)
		paymentFee := int64(0)

		// Validate and calculate promo discount if provided
		if req.PromoCode != "" {
			var promoID string
			var promoPercentage, promoFlat int
			var maxPromoAmount int64
			var startAt, expiredAt *time.Time
			var isActive bool
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT id, promo_percentage, promo_flat, max_promo_amount, 
				       is_active, start_at, expired_at
				FROM promos
				WHERE LOWER(code) = LOWER($1)
			`, req.PromoCode).Scan(&promoID, &promoPercentage, &promoFlat, &maxPromoAmount,
				&isActive, &startAt, &expiredAt)

			if err != nil {
				if err == pgx.ErrNoRows {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
						"promoCode": "Promo code not found",
					})
					return
				}
				utils.WriteInternalServerError(w)
				return
			}

			// Validate promo is active
			if !isActive {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"promoCode": "Promo code is not active",
				})
				return
			}

			// Validate promo date range
			now := time.Now()
			if startAt != nil && startAt.After(now) {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"promoCode": "Promo code has not started yet",
				})
				return
			}
			if expiredAt != nil && expiredAt.Before(now) {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"promoCode": "Promo code has expired",
				})
				return
			}

			// Check if promo is applicable to this product
			var productCount int
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_products pp
				JOIN products p ON pp.product_id = p.id
				WHERE pp.promo_id = $1 AND p.code = $2
			`, promoID, req.ProductCode).Scan(&productCount)

			if err != nil || productCount == 0 {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"promoCode": "Promo code is not applicable to this product",
				})
				return
			}

			// Check if promo is applicable to payment channel (if payment code provided)
			if req.PaymentCode != "" {
				var channelCount int
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT COUNT(*)
					FROM promo_payment_channels ppc
					JOIN payment_channels pc ON ppc.channel_id = pc.id
					WHERE ppc.promo_id = $1 AND pc.code = $2
				`, promoID, req.PaymentCode).Scan(&channelCount)

				if err != nil || channelCount == 0 {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
						"promoCode": "Promo code is not applicable to this payment method",
					})
					return
				}
			}

			// Calculate discount (all in rupiah)
			// promoFlat and maxPromoAmount in database are in rupiah
			if promoPercentage > 0 {
				discount = (subtotal * int64(promoPercentage)) / 100
				if maxPromoAmount > 0 && discount > maxPromoAmount {
					discount = maxPromoAmount
				}
			} else if promoFlat > 0 {
				discount = int64(promoFlat) // promoFlat is already in rupiah
			}
		}

		// Validate and calculate payment fee if payment code provided
		if req.PaymentCode != "" {
			var paymentChannelID, paymentName string
			var feeAmount, feePercentage float64
			var isActive bool
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT id, name, fee_amount, fee_percentage, is_active
				FROM payment_channels
				WHERE code = $1
			`, req.PaymentCode).Scan(&paymentChannelID, &paymentName, &feeAmount, &feePercentage, &isActive)

			if err != nil {
				if err == pgx.ErrNoRows {
					utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
						"paymentCode": "Payment method not found",
					})
					return
				}
				utils.WriteInternalServerError(w)
				return
			}

			// Validate payment channel is active
			if !isActive {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"paymentCode": "Payment method is not active",
				})
				return
			}

			// Calculate payment fee (all in rupiah)
			// feeAmount in database is in rupiah (e.g., 4000 = 4000 IDR)
			// feePercentage is in decimal format (e.g., 0.7 for 0.7%)
			paymentFee = int64(feeAmount) // feeAmount is already in rupiah
			if feePercentage > 0 {
				// Calculate percentage fee: (subtotal * feePercentage) / 100
				// feePercentage is 0.7, so: (subtotal * 0.7) / 100
				// In integer math: (subtotal * int64(feePercentage*100)) / 10000
				percentageFee := (subtotal * int64(feePercentage*100)) / 10000
				paymentFee += percentageFee
			}
		}

		total := subtotal - discount + paymentFee

		// Generate validation token using JWT
		tokenData := map[string]interface{}{
			"productCode": productCode,
			"skuCode":     skuCode,
			"paymentCode": req.PaymentCode,
			"quantity":    quantity,
			"accountData": map[string]interface{}{
				"userId":   req.UserID,
				"zoneId":   zoneValue,
				"nickname": accountNickname,
			},
			"pricing": map[string]interface{}{
				"subtotal":   subtotal,   // Store in rupiah for token
				"discount":   discount,   // Store in rupiah for token
				"paymentFee": paymentFee, // Store in rupiah for token
				"total":      total,      // Store in rupiah for token
			},
		}

		// Add promo code if exists
		if req.PromoCode != "" && discount > 0 {
			tokenData["promoCode"] = req.PromoCode
		}

		// Add contact data if exists
		if req.Email != "" || req.PhoneNumber != "" {
			tokenData["contactData"] = map[string]interface{}{
				"email":       req.Email,
				"phoneNumber": req.PhoneNumber,
			}
		}

		validationToken, err := deps.JWTService.GenerateValidationToken(tokenData)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		expiresAt := time.Now().Add(5 * time.Minute)

		// Build response
		response := map[string]interface{}{
			"validationToken": validationToken,
			"expiresAt":       expiresAt.Format(time.RFC3339),
			"order": map[string]interface{}{
				"product": map[string]interface{}{
					"code": productCode,
					"name": productName,
				},
				"sku": map[string]interface{}{
					"code":     skuCode,
					"name":     skuName,
					"quantity": quantity,
				},
				"account": map[string]interface{}{
					"nickname": accountNickname,
					"userId":   req.UserID,
					"zoneId":   zoneValue,
				},
				"payment": map[string]interface{}{
					"code":     req.PaymentCode,
					"name":     req.PaymentCode, // Should get from DB
					"currency": "IDR",           // Should get from region
				},
				"pricing": map[string]interface{}{
					"subtotal":   float64(subtotal),   // Already in rupiah
					"discount":   float64(discount),   // Already in rupiah
					"paymentFee": float64(paymentFee), // Already in rupiah
					"total":      float64(total),      // Already in rupiah
				},
			},
		}

		// Add promo info if exists
		if req.PromoCode != "" && discount > 0 {
			response["order"].(map[string]interface{})["promo"] = map[string]interface{}{
				"code":           req.PromoCode,
				"discountAmount": float64(discount), // Already in rupiah
			}
		}

		// Add contact info if provided
		if req.Email != "" || req.PhoneNumber != "" {
			response["order"].(map[string]interface{})["contact"] = map[string]interface{}{
				"email":       req.Email,
				"phoneNumber": req.PhoneNumber,
			}
		}

		utils.WriteSuccessJSON(w, response)
	}
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	ValidationToken string `json:"validationToken" validate:"required"`
}

// handleCreateOrderImpl implements order creation with payment processing
func HandleCreateOrderImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Parse request
		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.ValidationToken == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"validationToken": "Validation token is required",
			})
			return
		}

		// Validate and decode validation token
		tokenData, err := deps.JWTService.ValidateValidationToken(req.ValidationToken)
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "INVALID_TOKEN").
				Msg("Failed to validate validation token")
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Invalid or expired validation token", "Please create a new order inquiry")
			return
		}

		// Check if token was already used (via Redis)
		tokenKey := deps.Redis.ValidationTokenKey(req.ValidationToken)
		exists, err := deps.Redis.Client.Exists(ctx, tokenKey).Result()
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "REDIS_ERROR").
				Str("token_key", tokenKey).
				Msg("Failed to check token usage in Redis")
		}
		if err == nil && exists > 0 {
			log.Warn().
				Str("endpoint", "/v2/orders").
				Str("error_type", "TOKEN_ALREADY_USED").
				Str("token_key", tokenKey).
				Msg("Validation token has already been used")
			utils.WriteErrorJSON(w, http.StatusBadRequest, "TOKEN_ALREADY_USED",
				"Validation token has already been used", "Please create a new order inquiry")
			return
		}

		// Extract order data from token
		productCode, _ := tokenData["productCode"].(string)
		skuCode, _ := tokenData["skuCode"].(string)
		paymentCode, _ := tokenData["paymentCode"].(string)
		quantity := int(tokenData["quantity"].(float64))

		// Extract account data
		accountData := tokenData["accountData"].(map[string]interface{})
		userId, _ := accountData["userId"].(string)
		zoneId, _ := accountData["zoneId"].(string)
		nickname, _ := accountData["nickname"].(string)

		// Extract contact data if exists
		var contactEmail, contactPhone *string
		if contactData, ok := tokenData["contactData"].(map[string]interface{}); ok {
			if email, ok := contactData["email"].(string); ok && email != "" {
				contactEmail = &email
			}
			if phone, ok := contactData["phoneNumber"].(string); ok && phone != "" {
				contactPhone = &phone
			}
		}

		promoCode, _ := tokenData["promoCode"].(string)

		// Get client IP and user agent
		// Extract IP address (remove port if present)
		ipAddress := extractIPAddress(r)
		userAgent := r.UserAgent()

		// Start database transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "DB_TRANSACTION_ERROR").
				Msg("Failed to begin database transaction")
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Get region from context
		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = "ID" // Default to Indonesia
		}

		// Fetch product details (check region availability)
		var productID string
		var productName, productSlug string
		err = tx.QueryRow(ctx, `
			SELECT p.id, p.title, p.slug
			FROM products p
			JOIN product_regions pr ON p.id = pr.product_id
			WHERE p.code = $1 AND p.is_active = true AND pr.region_code = $2
		`, productCode, region).Scan(&productID, &productName, &productSlug)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "PRODUCT_NOT_FOUND").
				Str("product_code", productCode).
				Str("region", region).
				Bool("is_no_rows", err == pgx.ErrNoRows).
				Msg("Failed to fetch product from database (product not found or not available in region)")
			utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND",
				"Product not found or not available in your region", "")
			return
		}

		log.Info().
			Str("endpoint", "/v2/orders").
			Str("product_code", productCode).
			Str("product_id", productID).
			Str("product_name", productName).
			Str("region", region).
			Msg("Product found successfully")

		// Fetch SKU details with pricing and provider
		var skuID, skuName, providerID string
		var buyPrice, sellPrice int64
		err = tx.QueryRow(ctx, `
			SELECT s.id, s.name, s.provider_id, sp.buy_price, sp.sell_price
			FROM skus s
			JOIN sku_pricing sp ON s.id = sp.sku_id
			WHERE s.code = $1 AND s.product_id = $2 AND s.is_active = true AND sp.is_active = true
			LIMIT 1
		`, skuCode, productID).Scan(&skuID, &skuName, &providerID, &buyPrice, &sellPrice)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "SKU_NOT_FOUND").
				Str("sku_code", skuCode).
				Str("product_id", productID).
				Bool("is_no_rows", err == pgx.ErrNoRows).
				Msg("Failed to fetch SKU from database")
			utils.WriteErrorJSON(w, http.StatusNotFound, "SKU_NOT_FOUND",
				"SKU not found", "")
			return
		}

		log.Info().
			Str("endpoint", "/v2/orders").
			Str("sku_code", skuCode).
			Str("sku_id", skuID).
			Str("sku_name", skuName).
			Str("provider_id", providerID).
			Int64("buy_price", buyPrice).
			Int64("sell_price", sellPrice).
			Msg("SKU found successfully")

		// Calculate subtotal
		subtotal := sellPrice * int64(quantity)

		// Fetch payment channel details
		var paymentChannelID string
		var paymentCategoryID *string // nullable
		var paymentName, paymentInstruction string
		var feeAmount, feePercentage float64
		err = tx.QueryRow(ctx, `
			SELECT pc.id, pc.name, pc.category_id, pc.instruction, pc.fee_amount, pc.fee_percentage
			FROM payment_channels pc
			WHERE pc.code = $1 AND pc.is_active = true
			LIMIT 1
		`, paymentCode).Scan(&paymentChannelID, &paymentName, &paymentCategoryID,
			&paymentInstruction, &feeAmount, &feePercentage)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "PAYMENT_CHANNEL_NOT_FOUND").
				Str("payment_code", paymentCode).
				Bool("is_no_rows", err == pgx.ErrNoRows).
				Msg("Failed to fetch payment channel from database")
			utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_CHANNEL_NOT_FOUND",
				"Payment channel not found", "")
			return
		}

		// Determine gateway based on payment channel code (hardcoded routing)
		gatewayName := getGatewayForChannel(paymentCode)

		log.Info().
			Str("endpoint", "/v2/orders").
			Str("payment_code", paymentCode).
			Str("payment_channel_id", paymentChannelID).
			Str("payment_name", paymentName).
			Str("gateway_name", gatewayName).
			Float64("fee_amount", feeAmount).
			Float64("fee_percentage", feePercentage).
			Msg("Payment channel found successfully")

		// Calculate payment fee
		// feeAmount in database is in rupiah (e.g., 4000 = 4000 IDR)
		// feePercentage is in decimal format (e.g., 0.7 for 0.7%)
		// All values are in rupiah, not cents
		paymentFee := int64(feeAmount) // feeAmount is already in rupiah
		if feePercentage > 0 {
			// Calculate percentage fee: (subtotal * feePercentage) / 100
			percentageFee := (subtotal * int64(feePercentage*100)) / 10000
			paymentFee += percentageFee
		}

		// Calculate discount if promo code provided
		var discountAmount int64
		var promoID *string
		if promoCode != "" {
			var dbPromoID string
			var promoPercentage, promoFlat int
			var maxPromoAmount int64
			err = tx.QueryRow(ctx, `
				SELECT id, promo_percentage, promo_flat, max_promo_amount
				FROM promos
				WHERE code = $1 AND is_active = true AND expired_at > NOW()
			`, promoCode).Scan(&dbPromoID, &promoPercentage, &promoFlat, &maxPromoAmount)

			if err == nil {
				promoID = &dbPromoID
				if promoPercentage > 0 {
					discountAmount = (subtotal * int64(promoPercentage)) / 100
					if maxPromoAmount > 0 && discountAmount > maxPromoAmount {
						discountAmount = maxPromoAmount
					}
				} else if promoFlat > 0 {
					discountAmount = int64(promoFlat)
				}
			}
		}

		// Calculate total
		totalAmount := subtotal - discountAmount + paymentFee

		// Get user ID from auth context if authenticated
		var userID *string
		if authUserID := r.Context().Value("userID"); authUserID != nil {
			if uid, ok := authUserID.(string); ok {
				userID = &uid
			}
		}

		// For BALANCE payment, check user balance
		if paymentCode == "BALANCE" {
			if userID == nil {
				log.Warn().
					Str("endpoint", "/v2/orders").
					Str("error_type", "AUTHENTICATION_REQUIRED").
					Str("payment_code", paymentCode).
					Msg("Balance payment requires authentication")
				utils.WriteErrorJSON(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED",
					"Balance payment requires authentication", "")
				return
			}

			var balanceIDR int64
			err = tx.QueryRow(ctx, `
				SELECT balance_idr FROM users WHERE id = $1
			`, *userID).Scan(&balanceIDR)

			if err != nil {
				log.Error().
					Err(err).
					Str("endpoint", "/v2/orders").
					Str("error_type", "BALANCE_CHECK_ERROR").
					Str("user_id", *userID).
					Bool("is_no_rows", err == pgx.ErrNoRows).
					Msg("Failed to check user balance")
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE",
					"Insufficient balance", "Please top up your balance or use another payment method")
				return
			}

			if balanceIDR < totalAmount {
				log.Warn().
					Str("endpoint", "/v2/orders").
					Str("error_type", "INSUFFICIENT_BALANCE").
					Str("user_id", *userID).
					Int64("balance_idr", balanceIDR).
					Int64("total_amount", totalAmount).
					Msg("User has insufficient balance")
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE",
					"Insufficient balance", "Please top up your balance or use another payment method")
				return
			}
		}

		// Generate unique invoice number
		invoiceNumber := utils.GenerateInvoiceNumber()

		// Set expiry time based on payment type
		var expiredAt time.Time
		if paymentCode == "BALANCE" {
			// Balance payment expires in 5 minutes
			expiredAt = time.Now().Add(5 * time.Minute)
		} else if paymentCode == "QRIS" {
			// QRIS expires in 30 minutes
			expiredAt = time.Now().Add(30 * time.Minute)
		} else if paymentCode == "GOPAY" || paymentCode == "SHOPEEPAY" || paymentCode == "DANA" {
			// E-Wallet expires in 1 hour
			expiredAt = time.Now().Add(1 * time.Hour)
		} else {
			// Virtual Account expires in 24 hours
			expiredAt = time.Now().Add(24 * time.Hour)
		}

		// Prepare account inputs as JSONB
		accountInputs := map[string]interface{}{
			"userId": userId,
		}
		if zoneId != "" {
			accountInputs["zoneId"] = zoneId
		}
		accountInputsJSON, _ := json.Marshal(accountInputs)

		// Get currency from region (region already set above)
		currency := "IDR"
		// Map region to currency if needed
		switch region {
		case "MY":
			currency = "MYR"
		case "SG":
			currency = "SGD"
		case "PH":
			currency = "PHP"
		case "TH":
			currency = "THB"
		default:
			currency = "IDR"
		}

		// Create transaction record
		var transactionID string
		var accountNickname *string
		if nickname != "" {
			accountNickname = &nickname
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO transactions (
				invoice_number, user_id, product_id, sku_id, quantity,
				account_inputs, account_nickname,
				provider_id, payment_channel_id,
				promo_id, promo_code,
				buy_price, sell_price, discount_amount, payment_fee, total_amount,
				currency, region,
				status, payment_status,
				contact_email, contact_phone,
				ip_address, user_agent,
				expired_at, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7,
				$8, $9,
				$10, $11,
				$12, $13, $14, $15, $16,
				$17, $18,
				$19, $20,
				$21, $22,
				$23, $24,
				$25, NOW(), NOW()
			) RETURNING id
		`, invoiceNumber, userID, productID, skuID, quantity,
			accountInputsJSON, accountNickname,
			providerID, paymentChannelID,
			promoID, promoCode,
			buyPrice, sellPrice, discountAmount, paymentFee, totalAmount,
			currency, region,
			"PENDING", "UNPAID",
			contactEmail, contactPhone,
			ipAddress, userAgent,
			expiredAt).Scan(&transactionID)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "TRANSACTION_INSERT_ERROR").
				Str("invoice_number", invoiceNumber).
				Str("product_id", productID).
				Str("sku_id", skuID).
				Str("payment_channel_id", paymentChannelID).
				Msg("Failed to insert transaction into database")
			utils.WriteInternalServerError(w)
			return
		}

		log.Info().
			Str("endpoint", "/v2/orders").
			Str("transaction_id", transactionID).
			Str("invoice_number", invoiceNumber).
			Msg("Transaction created successfully")

		// Create initial timeline entry
		_, err = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, $2, $3, NOW())
		`, transactionID, "PENDING", "Order created, waiting for payment")

		if err != nil {
			log.Warn().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "TRANSACTION_LOG_INSERT_ERROR").
				Str("transaction_id", transactionID).
				Msg("Failed to insert transaction log (non-fatal)")
			// Non-fatal, continue
		}

		// Process payment based on payment channel type
		var paymentData map[string]interface{}
		var gatewayRefID *string
		paymentStatus := "UNPAID"
		transactionStatus := "PENDING"

		if paymentCode == "BALANCE" {
			// For balance payment, deduct immediately
			_, err = tx.Exec(ctx, `
				UPDATE users
				SET balance_idr = balance_idr - $1,
					total_spent_idr = total_spent_idr + $1,
					updated_at = NOW()
				WHERE id = $2
			`, totalAmount, *userID)

			if err != nil {
				log.Error().
					Err(err).
					Str("endpoint", "/v2/orders").
					Str("error_type", "BALANCE_DEDUCT_ERROR").
					Str("transaction_id", transactionID).
					Str("user_id", *userID).
					Int64("total_amount", totalAmount).
					Msg("Failed to deduct user balance")
				utils.WriteInternalServerError(w)
				return
			}

			// Update transaction to PAID
			paymentStatus = "PAID"
			transactionStatus = "PAID"
			paidAt := time.Now()

			_, err = tx.Exec(ctx, `
				UPDATE transactions
				SET payment_status = $1, status = $2, paid_at = $3, updated_at = NOW()
				WHERE id = $4
			`, paymentStatus, transactionStatus, paidAt, transactionID)

			if err != nil {
				log.Error().
					Err(err).
					Str("endpoint", "/v2/orders").
					Str("error_type", "TRANSACTION_UPDATE_ERROR").
					Str("transaction_id", transactionID).
					Msg("Failed to update transaction status to PAID")
				utils.WriteInternalServerError(w)
				return
			}

			// Add timeline entry for payment
			_, err = tx.Exec(ctx, `
				INSERT INTO transaction_logs (transaction_id, status, message, created_at)
				VALUES ($1, $2, $3, NOW())
			`, transactionID, "PAID", "Payment completed using balance")

			paymentData = map[string]interface{}{
				"method": "BALANCE",
				"paidAt": paidAt,
			}

		} else {
			// For external payment gateways, create payment request
			log.Info().
				Str("endpoint", "/v2/orders").
				Str("transaction_id", transactionID).
				Str("payment_code", paymentCode).
				Msg("Processing external payment gateway")

			// Check if payment manager is available
			if deps.PaymentManager == nil {
				log.Error().
					Str("endpoint", "/v2/orders").
					Str("payment_code", paymentCode).
					Msg("Payment manager not configured")
				tx.Rollback(ctx)
				utils.WriteErrorJSON(w, http.StatusServiceUnavailable, "PAYMENT_GATEWAY_UNAVAILABLE",
					"Payment gateway is not available", "Please try again later or use a different payment method")
				return
			}

			// Check available channels
			supportedChannels := deps.PaymentManager.GetSupportedChannels()
			log.Info().
				Str("endpoint", "/v2/orders").
				Strs("supported_channels", supportedChannels).
				Str("requested_channel", paymentCode).
				Msg("Payment manager channels check")

			// Build description for payment
			paymentDesc := fmt.Sprintf("%s - %s", productName, skuName)
			if nickname != "" {
				paymentDesc = fmt.Sprintf("%s (%s)", paymentDesc, nickname)
			}

			// Build return URL with locale (default to id-id)
			locale := "id-id"
			if region == "MY" {
				locale = "ms-my"
			} else if region == "PH" {
				locale = "en-ph"
			} else if region == "SG" {
				locale = "en-sg"
			} else if region == "TH" {
				locale = "th-th"
			}
			frontendInvoiceURL := fmt.Sprintf("%s/%s/invoice/%s", deps.Config.App.FrontendBaseURL, locale, invoiceNumber)

			// Determine callback URL based on gateway
			callbackURL := deps.Config.App.BaseURL + "/v2/webhooks/payment"
			if gatewayName == "DANA_DIRECT" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/dana"
			} else if gatewayName == "MIDTRANS" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/midtrans"
			} else if gatewayName == "XENDIT" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/xendit"
			} else if gatewayName == "BRI_DIRECT" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/bri"
			}

			// Create payment request
			paymentReq := &payment.PaymentRequest{
				RefID:          invoiceNumber,
				Amount:         float64(totalAmount),
				Currency:       currency,
				Channel:        paymentCode,
				GatewayName:    gatewayName,
				GatewayCode:    paymentCode, // Use payment code as gateway code
				Description:    paymentDesc,
				CustomerName:   nickname,
				CustomerEmail:  "",
				CustomerPhone:  "",
				ExpiryDuration: time.Until(expiredAt),
				CallbackURL:    callbackURL,
				SuccessURL:     frontendInvoiceURL,
				FailureURL:     frontendInvoiceURL,
				Metadata: map[string]string{
					"transaction_id": transactionID,
					"product_code":   productCode,
					"sku_code":       skuCode,
					"user_id":        userId,
				},
			}

			// Add contact info if available
			if contactEmail != nil {
				paymentReq.CustomerEmail = *contactEmail
			}
			if contactPhone != nil {
				paymentReq.CustomerPhone = *contactPhone
			}

			log.Info().
				Str("endpoint", "/v2/orders").
				Str("ref_id", invoiceNumber).
				Str("channel", paymentCode).
				Str("gateway_name", gatewayName).
				Float64("amount", paymentReq.Amount).
				Str("callback_url", paymentReq.CallbackURL).
				Str("success_url", paymentReq.SuccessURL).
				Msg("Calling payment gateway")

			// Check if gateway exists, if not try fallback
			if _, err := deps.PaymentManager.Get(gatewayName); err != nil {
				// Gateway not available, try fallback for VA channels
				fallbackGateway := getFallbackGateway(paymentCode, gatewayName)
				if fallbackGateway != "" {
					log.Warn().
						Str("endpoint", "/v2/orders").
						Str("original_gateway", gatewayName).
						Str("fallback_gateway", fallbackGateway).
						Str("channel", paymentCode).
						Msg("Primary gateway not available, using fallback")
					gatewayName = fallbackGateway
					paymentReq.GatewayName = fallbackGateway

					// Update callback URL for fallback gateway
					if fallbackGateway == "XENDIT" {
						paymentReq.CallbackURL = deps.Config.App.BaseURL + "/v2/webhooks/xendit"
					}
				}
			}

			// Create payment via gateway
			paymentResult, paymentErr := deps.PaymentManager.CreatePayment(ctx, paymentReq)
			if paymentErr != nil {
				log.Error().
					Err(paymentErr).
					Str("endpoint", "/v2/orders").
					Str("payment_code", paymentCode).
					Str("invoice_number", invoiceNumber).
					Msg("Payment gateway call failed")
				tx.Rollback(ctx)
				utils.WriteErrorJSON(w, http.StatusBadGateway, "PAYMENT_GATEWAY_ERROR",
					"Failed to create payment", paymentErr.Error())
				return
			}

			// Successfully created payment via gateway
			log.Info().
				Str("endpoint", "/v2/orders").
				Str("ref_id", paymentResult.RefID).
				Str("gateway_ref", paymentResult.GatewayRefID).
				Str("status", paymentResult.Status).
				Str("payment_url", paymentResult.PaymentURL).
				Str("qr_code", paymentResult.QRCode).
				Str("qr_code_url", paymentResult.QRCodeURL).
				Msg("Payment created successfully via gateway")

			// Map gateway response to payment data
			paymentData = mapGatewayResponseToPaymentData(paymentResult, expiredAt)
			gatewayRefID = &paymentResult.GatewayRefID

			// Store payment data as JSONB
			paymentDataJSON, _ := json.Marshal(paymentData)
			_, err = tx.Exec(ctx, `
				UPDATE transactions
				SET payment_gateway_ref_id = $1, provider_response = $2, updated_at = NOW()
				WHERE id = $3
			`, gatewayRefID, paymentDataJSON, transactionID)

			if err != nil {
				log.Warn().
					Err(err).
					Str("endpoint", "/v2/orders").
					Str("error_type", "PAYMENT_DATA_UPDATE_ERROR").
					Str("transaction_id", transactionID).
					Msg("Failed to update payment data (non-fatal)")
				// Non-fatal, continue
			}
		}

		// Mark validation token as used in Redis (5 minute expiry)
		err = deps.Redis.Client.SetEx(ctx, tokenKey, "used", 5*time.Minute).Err()
		if err != nil {
			log.Warn().
				Err(err).
				Str("endpoint", "/v2/orders").
				Str("error_type", "TRANSACTION_LOG_INSERT_ERROR").
				Str("transaction_id", transactionID).
				Msg("Failed to insert transaction log (non-fatal)")
			// Non-fatal, continue
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {

			utils.WriteInternalServerError(w)
			return
		}

		// Build response
		response := buildOrderResponse(
			invoiceNumber, transactionStatus, paymentStatus,
			productCode, productName, skuCode, skuName, quantity,
			userId, zoneId, nickname,
			subtotal, discountAmount, paymentFee, totalAmount, currency,
			paymentCode, paymentName, paymentInstruction,
			paymentData, promoCode,
			contactEmail, contactPhone,
			time.Now(), expiredAt,
		)

		utils.WriteSuccessJSON(w, response)
	}
}

// mapGatewayResponseToPaymentData maps payment gateway response to payment data format
// All payment codes (QRIS string, VA number, redirect URL, retail code) stored in "paymentCode"
func mapGatewayResponseToPaymentData(resp *payment.PaymentResponse, expiredAt time.Time) map[string]interface{} {
	data := make(map[string]interface{})

	// Common fields
	data["expiredAt"] = resp.ExpiresAt
	if resp.ExpiresAt.IsZero() {
		data["expiredAt"] = expiredAt
	}
	data["gatewayRef"] = resp.GatewayRefID
	data["channel"] = resp.Channel

	// Determine payment code based on channel type
	// All codes stored in "paymentCode" for simplicity
	switch {
	case resp.QRCode != "":
		// QRIS payment - paymentCode is the QRIS string
		data["paymentCode"] = resp.QRCode
		data["paymentType"] = "QRIS"

	case resp.VirtualAccount != "":
		// Virtual Account - paymentCode is the VA number
		data["paymentCode"] = resp.VirtualAccount
		data["paymentType"] = "VIRTUAL_ACCOUNT"
		if resp.BankCode != "" {
			data["bankCode"] = resp.BankCode
		}
		if resp.AccountName != "" {
			data["accountName"] = resp.AccountName
		} else {
			data["accountName"] = "GATE INDONESIA"
		}

	case resp.PaymentURL != "":
		// E-Wallet / Redirect - paymentCode is the redirect URL
		data["paymentCode"] = resp.PaymentURL
		data["paymentType"] = "E_WALLET"

	case resp.PaymentCode != "":
		// Retail or other - paymentCode as is
		data["paymentCode"] = resp.PaymentCode
		data["paymentType"] = "RETAIL"
	}

	// Add instructions if available
	if len(resp.Instructions) > 0 {
		data["instructions"] = resp.Instructions
	}

	return data
}

// extractIPAddress extracts IP address from request, removing port if present
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			// Remove port if present
			if host, _, err := net.SplitHostPort(ip); err == nil {
				return host
			}
			return ip
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		// Remove port if present
		if host, _, err := net.SplitHostPort(xri); err == nil {
			return host
		}
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in format "IP:port" or "[IPv6]:port"
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	// If SplitHostPort fails, return as is (shouldn't happen normally)
	return r.RemoteAddr
}

// buildOrderResponse builds the order response object
func buildOrderResponse(
	invoiceNumber, status, paymentStatus string,
	productCode, productName, skuCode, skuName string, quantity int,
	userId, zoneId, nickname string,
	subtotal, discount, paymentFee, total int64, currency string,
	paymentCode, paymentName, paymentInstruction string,
	paymentData map[string]interface{}, promoCode string,
	contactEmail, contactPhone *string,
	createdAt, expiredAt time.Time,
) map[string]interface{} {

	// Build account object
	account := map[string]interface{}{
		"userId": userId,
	}
	if zoneId != "" {
		account["zoneId"] = zoneId
	}
	if nickname != "" {
		account["nickname"] = nickname
	}

	// Build pricing object
	// All values are already in rupiah (not cents), so no division needed
	pricing := map[string]interface{}{
		"subtotal":   float64(subtotal),
		"discount":   float64(discount),
		"paymentFee": float64(paymentFee),
		"total":      float64(total),
		"currency":   currency,
	}

	// Build payment object
	payment := map[string]interface{}{
		"code":        paymentCode,
		"name":        paymentName,
		"instruction": paymentInstruction,
		"expiredAt":   expiredAt.Format(time.RFC3339),
	}

	// Add unified paymentCode (QRIS string, VA number, redirect URL, or retail code)
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
	if paidAt, ok := paymentData["paidAt"].(time.Time); ok {
		payment["paidAt"] = paidAt.Format(time.RFC3339)
	}
	if instructions, ok := paymentData["instructions"].([]string); ok && len(instructions) > 0 {
		payment["instructions"] = instructions
	}

	// Build order object
	order := map[string]interface{}{
		"invoiceNumber": invoiceNumber,
		"status":        status,
		"productCode":   productCode,
		"productName":   productName,
		"skuCode":       skuCode,
		"skuName":       skuName,
		"quantity":      quantity,
		"account":       account,
		"pricing":       pricing,
		"payment":       payment,
		"createdAt":     createdAt.Format(time.RFC3339),
		"expiredAt":     expiredAt.Format(time.RFC3339),
	}

	// Add promo if exists
	if promoCode != "" && discount > 0 {
		order["promo"] = map[string]interface{}{
			"code":           promoCode,
			"discountAmount": float64(discount) / 100,
		}
	}

	// Add contact if exists
	if contactEmail != nil || contactPhone != nil {
		contact := make(map[string]interface{})
		if contactEmail != nil {
			contact["email"] = *contactEmail
		}
		if contactPhone != nil {
			contact["phoneNumber"] = *contactPhone
		}
		order["contact"] = contact
	}

	return map[string]interface{}{
		"step":  "SUCCESS",
		"order": order,
	}
}

// getProviderForProduct maps product code to game check provider
func getProviderForProduct(productCode string) string {
	providerMap := map[string]string{
		"mlbb":     "codashop",
		"ff":       "codashop",
		"pubgm":    "codashop",
		"genshin":  "codashop",
		"valorant": "codashop",
		"hsr":      "codashop",
		"dana":     "duniagames",
		"gopay":    "duniagames",
	}

	provider, exists := providerMap[productCode]
	if !exists {
		return "codashop"
	}
	return provider
}

// getFallbackGateway returns a fallback gateway if the primary gateway is unavailable
func getFallbackGateway(channelCode, primaryGateway string) string {
	// Define fallback mappings
	fallbacks := map[string]string{
		// If BRI_DIRECT not available, fall back to XENDIT for VA_BRI
		"BRI_DIRECT": "XENDIT",
		// If BCA_DIRECT not available, fall back to XENDIT
		"BCA_DIRECT": "XENDIT",
	}

	if fallback, ok := fallbacks[primaryGateway]; ok {
		return fallback
	}
	return ""
}

// getGatewayForChannel maps payment channel code to gateway name
// Hardcoded routing:
// - QRIS, DANA -> DANA_DIRECT (DANA Gapura)
// - GOPAY, SHOPEEPAY -> MIDTRANS
// - ALFAMART, INDOMARET -> XENDIT
// - BRI_VA -> BRI_DIRECT (BRI SNAP API)
// - Other Virtual Accounts -> XENDIT
func getGatewayForChannel(channelCode string) string {
	gatewayMap := map[string]string{
		// QRIS & E-Wallet via DANA Gapura
		"QRIS": "DANA_DIRECT",
		"DANA": "DANA_DIRECT",

		// GoPay & ShopeePay via Midtrans
		"GOPAY":     "MIDTRANS",
		"SHOPEEPAY": "MIDTRANS",

		// Retail via Xendit
		"ALFAMART":  "XENDIT",
		"INDOMARET": "XENDIT",

		// Virtual Accounts - BRI direct via SNAP API
		"BRI_VA": "BRI_DIRECT",
		"VA_BRI": "BRI_DIRECT", // alias

		// Virtual Accounts via Xendit
		"BCA_VA":     "XENDIT",
		"VA_BCA":     "XENDIT",
		"BNI_VA":     "XENDIT",
		"VA_BNI":     "XENDIT",
		"MANDIRI_VA": "XENDIT",
		"VA_MANDIRI": "XENDIT",
		"PERMATA_VA": "XENDIT",
		"VA_PERMATA": "XENDIT",
		"CIMB_VA":    "XENDIT",
		"VA_CIMB":    "XENDIT",
	}

	gateway, exists := gatewayMap[channelCode]
	if !exists {
		// Default to XENDIT for unknown channels
		return "XENDIT"
	}
	return gateway
}
