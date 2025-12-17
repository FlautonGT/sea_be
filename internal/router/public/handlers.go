package public

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"unsafe"

	"seaply/internal/domain"
	"seaply/internal/middleware"
	"seaply/internal/payment"
	"seaply/internal/provider"
	"seaply/internal/router/user"
	"seaply/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Helper function to handle nullable strings
func nullStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Helper function to get keys from a map
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Public Handlers - Implemented
func HandleGetRegions(deps *Dependencies) http.HandlerFunc {
	return handleGetRegionsImpl(deps)
}

func HandleGetLanguages(deps *Dependencies) http.HandlerFunc {
	return handleGetLanguagesImpl(deps)
}

func HandleGetContacts(deps *Dependencies) http.HandlerFunc {
	return handleGetContactsImpl(deps)
}

func HandleGetPopups(deps *Dependencies) http.HandlerFunc {
	return handleGetPopupsImpl(deps)
}

func HandleGetBanners(deps *Dependencies) http.HandlerFunc {
	return handleGetBannersImpl(deps)
}

func HandleGetCategories(deps *Dependencies) http.HandlerFunc {
	return handleGetCategoriesImpl(deps)
}

func HandleGetProducts(deps *Dependencies) http.HandlerFunc {
	return handleGetProductsImpl(deps)
}

func HandleGetPopularProducts(deps *Dependencies) http.HandlerFunc {
	return handleGetPopularProductsImpl(deps)
}

func HandleGetFields(deps *Dependencies) http.HandlerFunc {
	return handleGetFieldsImpl(deps)
}

func HandleGetSections(deps *Dependencies) http.HandlerFunc {
	return handleGetSectionsImpl(deps)
}

func HandleGetSKUs(deps *Dependencies) http.HandlerFunc {
	return handleGetSKUsImpl(deps)
}

func HandleGetPromoSKUs(deps *Dependencies) http.HandlerFunc {
	return handleGetPromoSKUsImpl(deps)
}

func HandleGetPaymentCategories(deps *Dependencies) http.HandlerFunc {
	return handleGetPaymentCategoriesImpl(deps)
}

func HandleGetPaymentChannels(deps *Dependencies) http.HandlerFunc {
	return handleGetPaymentChannelsImpl(deps)
}

func HandleGetPromos(deps *Dependencies) http.HandlerFunc {
	return handleGetPromosImpl(deps)
}

func HandleValidatePromo(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req domain.ValidatePromoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate required fields
		if req.PromoCode == "" || req.ProductCode == "" || req.SKUCode == "" || req.PaymentCode == "" || req.Region == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"promoCode":   "Promo code is required",
				"productCode": "Product code is required",
				"skuCode":     "SKU code is required",
				"paymentCode": "Payment code is required",
				"region":      "Region is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Validate account fields match product fields
		// Get product fields to validate account structure
		var productID string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id FROM products WHERE code = $1 AND is_active = true
		`, req.ProductCode).Scan(&productID)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND",
					"Product not found", "The product code does not exist or is inactive")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get required product fields
		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT key, is_required
			FROM product_fields
			WHERE product_id = $1
			ORDER BY sort_order ASC
		`, productID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		requiredFields := make(map[string]bool)
		allFields := make(map[string]bool)
		for rows.Next() {
			var fieldKey string
			var isRequired bool
			if err := rows.Scan(&fieldKey, &isRequired); err == nil {
				allFields[fieldKey] = true
				if isRequired {
					requiredFields[fieldKey] = true
				}
			}
		}

		// Validate account contains all required fields
		if req.Account == nil {
			req.Account = make(map[string]interface{})
		}

		missingFields := []string{}
		for fieldKey, required := range requiredFields {
			if required {
				value, exists := req.Account[fieldKey]
				if !exists || value == nil || value == "" {
					missingFields = append(missingFields, fieldKey)
				}
			}
		}

		if len(missingFields) > 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"account": fmt.Sprintf("Missing required account fields: %s", strings.Join(missingFields, ", ")),
			})
			return
		}

		// Validate account only contains valid product field keys
		for accountKey := range req.Account {
			if !allFields[accountKey] {
				utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
					"account": fmt.Sprintf("Invalid account field: %s. Valid fields are: %s", accountKey, strings.Join(getMapKeys(allFields), ", ")),
				})
				return
			}
		}

		// Validate quantity (default to 1 if not provided or invalid)
		quantity := req.Quantity
		if quantity <= 0 {
			quantity = 1
		}

		// Get SKU price to calculate original amount
		var skuPrice int64
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT sp.sell_price
			FROM skus s
			JOIN sku_pricing sp ON s.id = sp.sku_id
			JOIN products p ON s.product_id = p.id
			WHERE s.code = $1 AND p.code = $2 AND s.is_active = true AND sp.is_active = true
			LIMIT 1
		`, req.SKUCode, req.ProductCode).Scan(&skuPrice)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "SKU_NOT_FOUND",
					"SKU not found", "The SKU code does not exist or is inactive")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Calculate original amount: SKU price * quantity
		originalAmount := float64(skuPrice) * float64(quantity)

		// Get region from context or use provided region
		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = strings.ToUpper(req.Region)
		}

		// Fetch promo by code
		var promoID string
		var promoCode, promoTitle, promoDesc *string
		var isActive bool
		var startAt, expiredAt *time.Time
		var minAmount, maxPromoAmount, promoFlat float64
		var promoPercentage float64
		var totalUsage, maxUsage, maxDailyUsage, maxUsagePerID, maxUsagePerDevice, maxUsagePerIP int
		var daysAvailable []string

		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT id, code, title, description, is_active, start_at, expired_at,
				   min_amount, max_promo_amount, promo_flat, promo_percentage,
				   total_usage, max_usage, max_daily_usage, max_usage_per_id,
				   max_usage_per_device, max_usage_per_ip, days_available
			FROM promos
			WHERE LOWER(code) = LOWER($1)
		`, req.PromoCode).Scan(
			&promoID, &promoCode, &promoTitle, &promoDesc, &isActive, &startAt, &expiredAt,
			&minAmount, &maxPromoAmount, &promoFlat, &promoPercentage,
			&totalUsage, &maxUsage, &maxDailyUsage, &maxUsagePerID, &maxUsagePerDevice, &maxUsagePerIP,
			&daysAvailable,
		)

		if err != nil {
			// Promo not found
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "PROMO_NOT_FOUND",
			})
			return
		}

		// Validation 1: Check if promo is active
		if !isActive {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "PROMO_NOT_ACTIVE",
			})
			return
		}

		// Validation 2: Check if promo has started
		now := time.Now()
		if startAt != nil && startAt.After(now) {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "PROMO_NOT_STARTED",
			})
			return
		}

		// Validation 3: Check if promo is expired
		if expiredAt != nil && expiredAt.Before(now) {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "PROMO_EXPIRED",
			})
			return
		}

		// Validation 4: Check if product is in allowed products list
		// If promo_products is empty/null, promo can be used for all products
		// If promo_products has entries, product must be in the list
		var totalProductCount int
		var productCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM promo_products WHERE promo_id = $1
		`, promoID).Scan(&totalProductCount)

		if err == nil && totalProductCount > 0 {
			// Promo has specific products, check if this product is allowed
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_products pp
				JOIN products p ON pp.product_id = p.id
				WHERE pp.promo_id = $1 AND p.code = $2
			`, promoID, req.ProductCode).Scan(&productCount)

			if err != nil || productCount == 0 {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "PRODUCT_NOT_APPLICABLE",
					"Promo code is not applicable to this product", "")
				return
			}
		}
		// If totalProductCount == 0, promo can be used for all products (skip validation)

		// Validation 5: Check if payment channel is in allowed payment channels
		// If promo_payment_channels is empty/null, promo can be used for all payment methods
		// If promo_payment_channels has entries, payment must be in the list
		var totalChannelCount int
		var channelCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM promo_payment_channels WHERE promo_id = $1
		`, promoID).Scan(&totalChannelCount)

		if err == nil && totalChannelCount > 0 {
			// Promo has specific payment channels, check if this payment is allowed
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_payment_channels ppc
				JOIN payment_channels pc ON ppc.channel_id = pc.id
				WHERE ppc.promo_id = $1 AND pc.code = $2
			`, promoID, req.PaymentCode).Scan(&channelCount)

			if err != nil || channelCount == 0 {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "PAYMENT_NOT_APPLICABLE",
					"Promo code is not applicable to this payment method", "")
				return
			}
		}
		// If totalChannelCount == 0, promo can be used for all payment methods (skip validation)

		// Validation 6: Check if region is in allowed regions
		var regionCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM promo_regions
			WHERE promo_id = $1 AND region_code = $2
		`, promoID, region).Scan(&regionCount)

		if err != nil || regionCount == 0 {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "REGION_NOT_APPLICABLE",
			})
			return
		}

		// Validation 7: Check if day is in allowed days (e.g., SAT, SUN only)
		if len(daysAvailable) > 0 {
			currentDay := now.Weekday().String()[:3] // MON, TUE, WED, THU, FRI, SAT, SUN
			// Convert Go weekday format to standard format
			weekdayMap := map[string]string{
				"Mon": "MON",
				"Tue": "TUE",
				"Wed": "WED",
				"Thu": "THU",
				"Fri": "FRI",
				"Sat": "SAT",
				"Sun": "SUN",
			}
			currentDay = weekdayMap[currentDay]

			dayFound := false
			for _, day := range daysAvailable {
				if day == currentDay {
					dayFound = true
					break
				}
			}

			if !dayFound {
				utils.WriteSuccessJSON(w, map[string]interface{}{
					"valid":  false,
					"reason": "DAY_NOT_APPLICABLE",
				})
				return
			}
		}

		// Validation 8: Check min_amount requirement
		if originalAmount < minAmount {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "MIN_AMOUNT_NOT_MET",
				fmt.Sprintf("Minimum amount required: %.0f", minAmount),
				fmt.Sprintf("Current amount: %.0f", originalAmount))
			return
		}

		// Validation 9: Check max usage limits
		// 9a: Check total usage limit
		if maxUsage > 0 && totalUsage >= maxUsage {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "USAGE_LIMIT_EXCEEDED",
			})
			return
		}

		// 9b: Check daily usage limit
		if maxDailyUsage > 0 {
			var dailyUsage int
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_usages
				WHERE promo_id = $1 AND DATE(created_at) = CURRENT_DATE
			`, promoID).Scan(&dailyUsage)

			if err == nil && dailyUsage >= maxDailyUsage {
				utils.WriteSuccessJSON(w, map[string]interface{}{
					"valid":  false,
					"reason": "DAILY_USAGE_LIMIT_EXCEEDED",
				})
				return
			}
		}

		// 9c: Check per user usage limit (if user is authenticated)
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID != "" && maxUsagePerID > 0 {
			var userUsage int
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_usages
				WHERE promo_id = $1 AND user_id = $2
			`, promoID, userID).Scan(&userUsage)

			if err == nil && userUsage >= maxUsagePerID {
				utils.WriteSuccessJSON(w, map[string]interface{}{
					"valid":  false,
					"reason": "USER_USAGE_LIMIT_EXCEEDED",
				})
				return
			}
		}

		// 9d: Check per device usage limit (if device ID is provided)
		deviceID := r.Header.Get("X-Device-ID")
		if deviceID != "" && maxUsagePerDevice > 0 {
			var deviceUsage int
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_usages
				WHERE promo_id = $1 AND device_id = $2
			`, promoID, deviceID).Scan(&deviceUsage)

			if err == nil && deviceUsage >= maxUsagePerDevice {
				utils.WriteSuccessJSON(w, map[string]interface{}{
					"valid":  false,
					"reason": "DEVICE_USAGE_LIMIT_EXCEEDED",
				})
				return
			}
		}

		// 9e: Check per IP usage limit
		ipAddress := r.Header.Get("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = r.RemoteAddr
		}
		if ipAddress != "" && maxUsagePerIP > 0 {
			var ipUsage int
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT COUNT(*)
				FROM promo_usages
				WHERE promo_id = $1 AND ip_address = $2
			`, promoID, ipAddress).Scan(&ipUsage)

			if err == nil && ipUsage >= maxUsagePerIP {
				utils.WriteSuccessJSON(w, map[string]interface{}{
					"valid":  false,
					"reason": "IP_USAGE_LIMIT_EXCEEDED",
				})
				return
			}
		}

		// Calculate discount amount
		var discountAmount float64
		if promoFlat > 0 {
			discountAmount = promoFlat
		} else if promoPercentage > 0 {
			discountAmount = (originalAmount * promoPercentage) / 100
		}

		// Cap discount by max promo amount
		if maxPromoAmount > 0 && discountAmount > maxPromoAmount {
			discountAmount = maxPromoAmount
		}

		finalAmount := originalAmount - discountAmount
		if finalAmount < 0 {
			finalAmount = 0
		}

		// Build promo details
		promoDetails := map[string]interface{}{
			"title":          nullStringOrEmpty(promoTitle),
			"maxPromoAmount": maxPromoAmount,
		}
		if promoPercentage > 0 {
			promoDetails["promoPercentage"] = promoPercentage
		}
		if promoFlat > 0 {
			promoDetails["promoFlat"] = promoFlat
		}

		// Build success response according to documentation
		response := map[string]interface{}{
			"promoCode":      nullStringOrEmpty(promoCode),
			"discountAmount": math.Round(discountAmount*100) / 100,
			"originalAmount": math.Round(originalAmount*100) / 100,
			"finalAmount":    math.Round(finalAmount*100) / 100,
			"promoDetails":   promoDetails,
		}

		utils.WriteSuccessJSON(w, response)
	}
}

func HandleGetInvoice(deps *Dependencies) http.HandlerFunc {
	return handleGetInvoiceImpl(deps)
}

func HandleGetDepositInvoice(deps *Dependencies) http.HandlerFunc {
	return handleGetDepositInvoiceImpl(deps)
}

func HandleGetReviews(deps *Dependencies) http.HandlerFunc {
	return handleGetReviewsImpl(deps)
}

func HandleCreateReview(deps *Dependencies) http.HandlerFunc {
	return handleCreateReviewImpl(deps)
}

func HandleQRGenerate(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := r.URL.Query().Get("data")
		if data == "" {
			utils.WriteBadRequestError(w, "data parameter required")
			return
		}

		qrCode, err := utils.GenerateQRCode(data, 256)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(qrCode)
	}
}

// Auth Handlers
func HandleRegister(deps *Dependencies) http.HandlerFunc {
	// Convert Dependencies to user.Dependencies using unsafe pointer
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleRegisterImpl(userDeps)
}

func HandleRegisterGoogle(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleRegisterGoogleImpl(userDeps)
}

func HandleVerifyEmail(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleVerifyEmailImpl(userDeps)
}

func HandleResendVerification(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleResendVerificationImpl(userDeps)
}

func HandleLogin(deps *Dependencies) http.HandlerFunc {
	return HandleUserLoginImpl(deps)
}

func HandleLoginGoogle(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleLoginGoogleImpl(userDeps)
}

func HandleVerifyMFA(deps *Dependencies) http.HandlerFunc {
	return HandleVerifyMFAImpl(deps)
}

func HandleForgotPassword(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleForgotPasswordImpl(userDeps)
}

func HandleResetPassword(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleResetPasswordImpl(userDeps)
}

func HandleRefreshToken(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleRefreshTokenImpl(userDeps)
}

func HandleEnableMFA(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleEnableMFAImpl(userDeps)
}

func HandleVerifyMFASetup(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleVerifyMFASetupImpl(userDeps)
}

func HandleDisableMFA(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleDisableMFAImpl(userDeps)
}

func HandleLogout(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleLogoutImpl(userDeps)
}

// User Handlers
func HandleGetProfile(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleGetProfileImpl(userDeps)
}

func HandleUpdateProfile(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleUpdateProfileImpl(userDeps)
}

func HandleChangePassword(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleChangePasswordImpl(userDeps)
}

func HandleGetUserTransactions(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleGetUserTransactionsImpl(userDeps)
}

func HandleGetMutations(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleGetMutationsImpl(userDeps)
}

func HandleGetReports(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleGetReportsImpl(userDeps)
}

func HandleGetDeposits(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleGetDepositsImpl(userDeps)
}

func HandleDepositInquiry(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleDepositInquiryImpl(userDeps)
}

func HandleCreateDeposit(deps *Dependencies) http.HandlerFunc {
	userDeps := (*user.Dependencies)(unsafe.Pointer(deps))
	return user.HandleCreateDepositImpl(userDeps)
}

// Transaction Handlers
func HandleAccountInquiry(deps *Dependencies) http.HandlerFunc {
	return HandleAccountInquiryImpl(deps)
}

func HandleOrderInquiry(deps *Dependencies) http.HandlerFunc {
	return HandleOrderInquiryImpl(deps)
}

func HandleCreateOrder(deps *Dependencies) http.HandlerFunc {
	return HandleCreateOrderImpl(deps)
}

// RC codes that should trigger retry with backup SKU
var digiflazzRetryableRCCodes = map[string]bool{
	"02": true, "43": true, "53": true, "55": true, "56": true,
	"58": true, "62": true, "66": true, "67": true, "68": true,
	"69": true, "70": true, "71": true,
}

// RC 49 means ref_id already used, need to generate new ref_id
const digiflazzRCRefIDUsed = "49"

// Webhook Handlers
func HandleDigiflazzWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read Digiflazz webhook body")
			utils.WriteBadRequestError(w, "Failed to read request body")
			return
		}
		defer r.Body.Close()

		log.Info().Str("body", string(body)).Msg("Received Digiflazz webhook")

		// Get Digiflazz provider
		if deps.ProviderManager == nil {
			log.Error().Msg("Provider manager is nil")
			utils.WriteInternalServerError(w)
			return
		}

		prov, err := deps.ProviderManager.Get("digiflazz")
		if err != nil {
			log.Error().Err(err).Msg("Digiflazz provider not found")
			utils.WriteInternalServerError(w)
			return
		}

		// Type assert to DigiflazzProvider to access ValidateWebhook
		digiProv, ok := prov.(*provider.DigiflazzProvider)
		if !ok {
			log.Error().Msg("Provider is not of type DigiflazzProvider")
			utils.WriteInternalServerError(w)
			return
		}

		// Validate webhook
		signature := r.Header.Get("X-Hub-Signature")
		trx, err := digiProv.ValidateWebhook(body, signature)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate Digiflazz webhook")
			utils.WriteBadRequestError(w, "Invalid webhook signature or format")
			return
		}

		// Map status
		var newStatus string
		switch trx.Status {
		case "Sukses":
			newStatus = "SUCCESS"
		case "Gagal":
			newStatus = "FAILED"
		case "Pending":
			newStatus = "PENDING"
		default:
			newStatus = "PROCESSING"
		}

		log.Info().
			Str("ref_id", trx.RefID).
			Str("provider_status", trx.Status).
			Str("rc", trx.RC).
			Str("mapped_status", newStatus).
			Msg("Processing Digiflazz webhook transaction")

		// Find transaction
		var transactionID, currentStatus, skuID, accountInputs string
		var retryCount int
		var skuCodeBackup1, skuCodeBackup2 *string
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT t.id, t.status, t.sku_id, t.account_inputs, COALESCE(t.retry_count, 0),
			       s.provider_sku_code_backup1, s.provider_sku_code_backup2
			FROM transactions t
			JOIN skus s ON t.sku_id = s.id
			WHERE t.invoice_number = $1
			LIMIT 1
		`, trx.RefID).Scan(&transactionID, &currentStatus, &skuID, &accountInputs, &retryCount, &skuCodeBackup1, &skuCodeBackup2)

		if err != nil {
			log.Warn().Err(err).Str("ref_id", trx.RefID).Msg("Transaction not found for Digiflazz webhook")
			utils.WriteSuccessJSON(w, map[string]interface{}{"status": "processed", "note": "transaction not found"})
			return
		}

		// Check if we should retry with backup SKU
		shouldRetry := false
		var backupSKU string
		var newRefID string

		if newStatus == "FAILED" && currentStatus != "SUCCESS" && currentStatus != "FAILED" {
			// Check if RC code is retryable
			if digiflazzRetryableRCCodes[trx.RC] {
				// Determine which backup to use based on retry count
				if retryCount == 0 && skuCodeBackup1 != nil && *skuCodeBackup1 != "" {
					backupSKU = *skuCodeBackup1
					shouldRetry = true
					newRefID = trx.RefID // Use same ref_id
				} else if retryCount == 1 && skuCodeBackup2 != nil && *skuCodeBackup2 != "" {
					backupSKU = *skuCodeBackup2
					shouldRetry = true
					newRefID = trx.RefID // Use same ref_id
				}
			} else if trx.RC == digiflazzRCRefIDUsed {
				// RC 49: ref_id already used, generate new ref_id and retry
				if retryCount < 2 {
					// Use backup SKU if available, otherwise use current
					if retryCount == 0 && skuCodeBackup1 != nil && *skuCodeBackup1 != "" {
						backupSKU = *skuCodeBackup1
					} else if retryCount == 1 && skuCodeBackup2 != nil && *skuCodeBackup2 != "" {
						backupSKU = *skuCodeBackup2
					} else {
						// No backup available, just retry with new ref_id using original SKU
						var originalSKU string
						err := deps.DB.Pool.QueryRow(ctx, `
							SELECT s.provider_sku_code FROM skus s WHERE s.id = $1
						`, skuID).Scan(&originalSKU)
						if err == nil {
							backupSKU = originalSKU
						}
					}
					if backupSKU != "" {
						// Generate new ref_id
						newRefID = utils.GenerateInvoiceNumber()
						shouldRetry = true
					}
				}
			}
		}

		if shouldRetry && backupSKU != "" {
			log.Info().
				Str("ref_id", trx.RefID).
				Str("new_ref_id", newRefID).
				Str("backup_sku", backupSKU).
				Str("rc", trx.RC).
				Int("retry_count", retryCount).
				Msg("Retrying transaction with backup SKU")

			// Create provider log entry for failed attempt - store full callback data
			failedLogEntry := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"type":      "PROVIDER_CALLBACK",
				"data":      trx,
			}
			failedLogJSON, _ := json.Marshal([]interface{}{failedLogEntry})

			// Update retry count, log the attempt, and add provider log
			_, err = deps.DB.Pool.Exec(ctx, `
				UPDATE transactions 
				SET retry_count = retry_count + 1,
				    provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb,
				    updated_at = NOW()
				WHERE id = $2
			`, string(failedLogJSON), transactionID)
			if err != nil {
				log.Error().Err(err).Msg("Failed to update retry count")
			}

			// Retry logic - don't add timeline entry here as it's too technical for users
			// The retry will process in background and add final timeline entry when done

			// Perform retry in goroutine - this function handles sequential retries
			go func(txID, refID, newRef, sku, custNo string, backup1, backup2 *string, digiProvider *provider.DigiflazzProvider) {
				retryCtx, retryCancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer retryCancel()

				// Retry loop - will try backup1, then backup2 if needed
				currentSKU := sku
				currentRefID := newRef
				if currentRefID == "" {
					currentRefID = refID
				}

				for attempt := 0; attempt < 2; attempt++ {
					// Determine which SKU to use
					if attempt == 0 {
						currentSKU = sku // backup1
					} else if attempt == 1 && backup2 != nil && *backup2 != "" {
						currentSKU = *backup2 // backup2
						// Update retry count for backup2
						_, _ = deps.DB.Pool.Exec(retryCtx, `
							UPDATE transactions SET retry_count = retry_count + 1 WHERE id = $1
						`, txID)
					} else {
						break // No more backups
					}

					orderReq := &provider.OrderRequest{
						RefID:      currentRefID,
						SKU:        currentSKU,
						CustomerNo: custNo,
					}

					log.Info().
						Str("transaction_id", txID).
						Str("ref_id", currentRefID).
						Str("backup_sku", currentSKU).
						Int("attempt", attempt+1).
						Msg("Sending retry order to Digiflazz")

					// Add provider log for retry request
					retryReqLog := map[string]interface{}{
						"timestamp": time.Now().Format(time.RFC3339),
						"type":      "RETRY_REQUEST",
						"data": map[string]interface{}{
							"sku":        currentSKU,
							"refId":      currentRefID,
							"customerNo": custNo,
							"attempt":    attempt + 1,
						},
					}
					retryReqJSON, _ := json.Marshal([]interface{}{retryReqLog})
					_, _ = deps.DB.Pool.Exec(retryCtx, `
						UPDATE transactions 
						SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
						WHERE id = $2
					`, string(retryReqJSON), txID)

					orderResp, err := digiProvider.CreateOrder(retryCtx, orderReq)
					if err != nil {
						log.Error().Err(err).
							Str("transaction_id", txID).
							Str("backup_sku", currentSKU).
							Int("attempt", attempt+1).
							Msg("Failed to create retry order")

						// Create provider log for retry failure
						retryFailLog := map[string]interface{}{
							"timestamp": time.Now().Format(time.RFC3339),
							"type":      "RETRY_FAILED",
							"data": map[string]interface{}{
								"error":   err.Error(),
								"sku":     currentSKU,
								"refId":   currentRefID,
								"attempt": attempt + 1,
							},
						}
						retryFailJSON, _ := json.Marshal([]interface{}{retryFailLog})
						_, _ = deps.DB.Pool.Exec(retryCtx, `
							UPDATE transactions 
							SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
							WHERE id = $2
						`, string(retryFailJSON), txID)

						// Continue to next backup if available
						if attempt == 0 && backup2 != nil && *backup2 != "" {
							log.Info().Msg("Backup1 failed, trying backup2...")
							continue
						}

						// No more backups - mark as failed
						_, _ = deps.DB.Pool.Exec(retryCtx, `
							UPDATE transactions 
							SET status = 'FAILED', updated_at = NOW()
							WHERE id = $1
						`, txID)

						_, _ = deps.DB.Pool.Exec(retryCtx, `
							INSERT INTO transaction_logs (transaction_id, status, message, created_at)
							VALUES ($1, 'FAILED', $2, NOW())
						`, txID, "Item has been failed to sent.")
						return
					}

					// Check response status
					var finalStatus string
					switch orderResp.Status {
					case "SUCCESS":
						finalStatus = "SUCCESS"
					case "FAILED":
						finalStatus = "FAILED"
					default:
						finalStatus = "PROCESSING"
					}

					// Create provider log for retry response
					retryRespLog := map[string]interface{}{
						"timestamp": time.Now().Format(time.RFC3339),
						"type":      "RETRY_RESPONSE",
						"data":      orderResp,
					}
					retryRespJSON, _ := json.Marshal([]interface{}{retryRespLog})

					respJSON, _ := json.Marshal(orderResp)

					if finalStatus == "PROCESSING" || finalStatus == "SUCCESS" {
						// Pending or Success - update and wait for callback
						_, err = deps.DB.Pool.Exec(retryCtx, `
							UPDATE transactions 
							SET status = $1::transaction_status,
							    provider_ref_id = $2,
							    provider_serial_number = $3,
							    provider_response = $4,
							    provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $5::jsonb,
							    completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END,
							    updated_at = NOW()
							WHERE id = $6
						`, finalStatus, orderResp.ProviderRefID, orderResp.SN, respJSON, string(retryRespJSON), txID)

						// Add timeline entry: Item has been successfully sent or failed to sent
						var finalMessage string
						if finalStatus == "SUCCESS" {
							finalMessage = "Item has been successfully sent."
						} else if finalStatus == "FAILED" {
							finalMessage = "Item has been failed to sent."
						} else {
							// PROCESSING status - don't add timeline entry yet, wait for callback
							finalMessage = ""
						}
						if finalMessage != "" {
							_, _ = deps.DB.Pool.Exec(retryCtx, `
								INSERT INTO transaction_logs (transaction_id, status, message, created_at)
								VALUES ($1, $2, $3, NOW())
							`, txID, finalStatus, finalMessage)
						}

						log.Info().
							Str("transaction_id", txID).
							Str("backup_sku", currentSKU).
							Str("status", finalStatus).
							Str("sn", orderResp.SN).
							Int("attempt", attempt+1).
							Msg("Retry order completed")
						return // Exit - either success or waiting for callback
					}

					// FAILED - check if we should try next backup
					_, _ = deps.DB.Pool.Exec(retryCtx, `
						UPDATE transactions 
						SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
						WHERE id = $2
					`, string(retryRespJSON), txID)

					log.Warn().
						Str("transaction_id", txID).
						Str("backup_sku", currentSKU).
						Str("status", orderResp.Status).
						Int("attempt", attempt+1).
						Msg("Retry order failed, checking for next backup")

					// Continue to next backup if available
					if attempt == 0 && backup2 != nil && *backup2 != "" {
						continue
					}

					// No more backups - mark as failed
					_, _ = deps.DB.Pool.Exec(retryCtx, `
						UPDATE transactions 
						SET status = 'FAILED',
						    provider_response = $1,
						    updated_at = NOW()
						WHERE id = $2
					`, respJSON, txID)

					_, _ = deps.DB.Pool.Exec(retryCtx, `
						INSERT INTO transaction_logs (transaction_id, status, message, created_at)
						VALUES ($1, 'FAILED', $2, NOW())
										`, txID, "Item has been failed to sent.")
					return
				}

			}(transactionID, trx.RefID, newRefID, backupSKU, trx.CustomerNo, skuCodeBackup1, skuCodeBackup2, digiProv)

			// Return OK immediately, retry happens in background
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"status": "ok",
				"note":   "retry initiated",
			})
			return
		}

		// Normal processing (no retry needed)
		shouldUpdate := false
		if currentStatus == "SUCCESS" {
			log.Info().
				Str("ref_id", trx.RefID).
				Str("current_status", currentStatus).
				Str("new_status", newStatus).
				Msg("Transaction already SUCCESS, ignoring webhook")
		} else {
			if newStatus == "SUCCESS" {
				shouldUpdate = true
				log.Info().
					Str("ref_id", trx.RefID).
					Str("current_status", currentStatus).
					Str("new_status", newStatus).
					Msg("Will update transaction to SUCCESS")
			} else if newStatus == "FAILED" && currentStatus != "FAILED" {
				shouldUpdate = true
				log.Info().
					Str("ref_id", trx.RefID).
					Str("current_status", currentStatus).
					Str("new_status", newStatus).
					Msg("Will update transaction to FAILED")
			} else {
				log.Info().
					Str("ref_id", trx.RefID).
					Str("current_status", currentStatus).
					Str("new_status", newStatus).
					Msg("Will not update transaction - conditions not met")
			}
		}

		if shouldUpdate {
			tx, err := deps.DB.Pool.Begin(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to begin transaction")
				utils.WriteInternalServerError(w)
				return
			}
			defer tx.Rollback(ctx)

			trxJSON, _ := json.Marshal(trx)

			// Create provider log entry with full raw callback data
			providerLogEntry := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"type":      "PROVIDER_CALLBACK",
				"data":      trx,
			}
			providerLogJSON, _ := json.Marshal([]interface{}{providerLogEntry})

			_, err = tx.Exec(ctx, `
				UPDATE transactions 
				SET status = $1::transaction_status, 
					provider_serial_number = $2,
					provider_response = $3,
					provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $4::jsonb,
					completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END,
					updated_at = NOW()
				WHERE id = $5
			`, newStatus, trx.SN, trxJSON, string(providerLogJSON), transactionID)

			if err != nil {
				log.Error().Err(err).Msg("Failed to update transaction")
				utils.WriteInternalServerError(w)
				return
			}

			// Add timeline entry for success/failed - only for final status
			if newStatus == "SUCCESS" || newStatus == "FAILED" {
				var timelineMessage string
				if newStatus == "SUCCESS" {
					timelineMessage = "Item has been successfully sent."
				} else {
					timelineMessage = "Item has been failed to sent."
				}

				log.Info().
					Str("transaction_id", transactionID).
					Str("status", newStatus).
					Str("message", timelineMessage).
					Msg("Adding timeline entry for Digiflazz webhook")

				_, err = tx.Exec(ctx, `
					INSERT INTO transaction_logs (transaction_id, status, message, created_at)
					VALUES ($1, $2, $3, NOW())
				`, transactionID, newStatus, timelineMessage)
				if err != nil {
					log.Error().
						Err(err).
						Str("transaction_id", transactionID).
						Str("status", newStatus).
						Msg("Failed to insert timeline")
					// Don't return here, continue to commit transaction update
				} else {
					log.Info().
						Str("transaction_id", transactionID).
						Str("status", newStatus).
						Msg("Successfully inserted timeline entry")
				}
			} else {
				log.Info().
					Str("transaction_id", transactionID).
					Str("new_status", newStatus).
					Msg("Skipping timeline entry - not final status (SUCCESS/FAILED)")
			}

			if err := tx.Commit(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to commit transaction")
				utils.WriteInternalServerError(w)
				return
			}

			log.Info().
				Str("ref_id", trx.RefID).
				Str("old_status", currentStatus).
				Str("new_status", newStatus).
				Msg("Transaction status updated via Digiflazz webhook")
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"status": "ok",
		})
	}
}

func HandleVIPResellerWebhook(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func HandleBangJeffWebhook(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func HandleLinkQuWebhook(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func HandleBCAWebhook(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func HandleBRIWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read BRI webhook body")
			// Return success response as per BRI spec
			sendBRIWebhookResponse(w, "4003400", "Bad Request")
			return
		}
		defer r.Body.Close()

		log.Info().
			Str("body", string(body)).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Received BRI webhook")

		// Parse webhook payload
		var payload struct {
			PartnerServiceID string `json:"partnerServiceId"`
			CustomerNo       string `json:"customerNo"`
			VirtualAccountNo string `json:"virtualAccountNo"`
			PaymentRequestID string `json:"paymentRequestId"`
			TrxDateTime      string `json:"trxDateTime"`
			AdditionalInfo   *struct {
				IDApp         string `json:"idApp"`
				PassApp       string `json:"passApp"`
				PaymentAmount string `json:"paymentAmount"`
				TerminalID    string `json:"terminalId"`
				BankID        string `json:"bankId"`
			} `json:"additionalInfo,omitempty"`
		}

		if err := json.Unmarshal(body, &payload); err != nil {
			log.Error().Err(err).Str("body", string(body)).Msg("Failed to parse BRI webhook payload")
			sendBRIWebhookResponse(w, "4003401", "Invalid Field Format")
			return
		}

		vaNo := strings.TrimLeft(payload.VirtualAccountNo, " ")
		log.Info().
			Str("va_no", vaNo).
			Str("customer_no", payload.CustomerNo).
			Str("payment_request_id", payload.PaymentRequestID).
			Msg("Processing BRI payment notification")

		// Find transaction by VA number (stored in payment_gateway_ref_id)
		var transactionID, invoiceNumber, currentStatus string
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT id, invoice_number, status FROM transactions 
			WHERE payment_gateway_ref_id = $1 OR payment_gateway_ref_id = $2
			LIMIT 1
		`, payload.VirtualAccountNo, vaNo).Scan(&transactionID, &invoiceNumber, &currentStatus)

		if err != nil {
			log.Warn().
				Str("va_no", vaNo).
				Err(err).
				Msg("BRI webhook: transaction not found by VA number")
			// Still return success to BRI to avoid retries
			sendBRIWebhookResponse(w, "2003400", "Successful")
			return
		}

		// Skip if already processed
		if currentStatus == "SUCCESS" || currentStatus == "FAILED" {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Str("current_status", currentStatus).
				Msg("BRI webhook: transaction already processed")
			sendBRIWebhookResponse(w, "2003400", "Successful")
			return
		}

		// Get payment amount
		var paymentAmount float64
		if payload.AdditionalInfo != nil && payload.AdditionalInfo.PaymentAmount != "" {
			fmt.Sscanf(payload.AdditionalInfo.PaymentAmount, "%f", &paymentAmount)
		}

		// Update transaction status
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to begin transaction for BRI webhook")
			sendBRIWebhookResponse(w, "5003400", "General Error")
			return
		}
		defer tx.Rollback(ctx)

		// Update to PROCESSING/PAID
		_, err = tx.Exec(ctx, `
			UPDATE transactions 
			SET status = 'PROCESSING', 
				payment_status = 'PAID',
				paid_at = NOW(),
				updated_at = NOW()
			WHERE id = $1
		`, transactionID)

		if err != nil {
			log.Error().Err(err).Str("transaction_id", transactionID).Msg("Failed to update transaction status")
			sendBRIWebhookResponse(w, "5003400", "General Error")
			return
		}

		// Insert timeline
		tx.Exec(ctx, `
			INSERT INTO transaction_timeline (transaction_id, status, message, created_at)
			VALUES ($1, 'PROCESSING', 'Pembayaran diterima via BRI VA', NOW())
		`, transactionID)

		if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to commit BRI webhook update")
			sendBRIWebhookResponse(w, "5003400", "General Error")
			return
		}

		log.Info().
			Str("invoice_number", invoiceNumber).
			Float64("amount", paymentAmount).
			Msg("BRI payment processed successfully")

		// Return success response
		sendBRIWebhookResponse(w, "2003400", "Successful")
	}
}

// sendBRIWebhookResponse sends BRI webhook response in required format
func sendBRIWebhookResponse(w http.ResponseWriter, responseCode, responseMessage string) {
	w.Header().Set("Content-Type", "application/json")

	// Determine HTTP status based on response code
	httpStatus := http.StatusOK
	if strings.HasPrefix(responseCode, "4") {
		httpStatus = http.StatusBadRequest
	} else if strings.HasPrefix(responseCode, "5") {
		httpStatus = http.StatusInternalServerError
	}
	w.WriteHeader(httpStatus)

	response := map[string]interface{}{
		"responseCode":    responseCode,
		"responseMessage": responseMessage,
	}
	json.NewEncoder(w).Encode(response)
}

func HandleDANAWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read DANA webhook body")
			sendDANAResponse(w, "5005601", "Internal Server Error")
			return
		}
		defer r.Body.Close()

		log.Info().
			Str("body", string(body)).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Received DANA webhook")

		// Parse webhook notification
		var notification struct {
			OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
			OriginalReferenceNo        string `json:"originalReferenceNo"`
			LatestTransactionStatus    string `json:"latestTransactionStatus"`
			Amount                     struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"amount"`
			PaidTime string `json:"paidTime"`
		}

		if err := json.Unmarshal(body, &notification); err != nil {
			log.Error().Err(err).Str("body", string(body)).Msg("Failed to parse DANA webhook")
			sendDANAResponse(w, "4005600", "Unexpected response")
			return
		}

		// Get invoice number (originalPartnerReferenceNo is our invoice number)
		invoiceNumber := notification.OriginalPartnerReferenceNo
		if invoiceNumber == "" {
			log.Warn().Msg("DANA webhook: missing originalPartnerReferenceNo")
			sendDANAResponse(w, "4005600", "Unexpected response")
			return
		}

		// Check if this is a deposit (starts with "SEAD") or transaction (starts with "SEAI")
		if strings.HasPrefix(invoiceNumber, "SEAD") {
			// Handle as deposit
			handleDepositPaymentCallback(ctx, deps, invoiceNumber, notification, body, w, "DANA")
			return
		}

		// Check current transaction status and fetch necessary data
		var transactionID, status, paymentStatus, providerSKU, accountInputs string
		var providerID, skuID string
		var providerCode, paymentName, productName, skuName string

		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT t.id, t.status, t.payment_status, t.sku_id, t.account_inputs, t.provider_id,
			       COALESCE(p.code, ''), COALESCE(s.provider_sku_code, ''),
			       COALESCE(pc.name, ''), COALESCE(pr.title, ''), COALESCE(s.name, '')
			FROM transactions t
			LEFT JOIN providers p ON t.provider_id = p.id
			LEFT JOIN skus s ON t.sku_id = s.id
			LEFT JOIN payment_channels pc ON t.payment_channel_id = pc.id
			LEFT JOIN products pr ON s.product_id = pr.id
			WHERE t.invoice_number = $1
		`, invoiceNumber).Scan(&transactionID, &status, &paymentStatus, &skuID, &accountInputs, &providerID, &providerCode, &providerSKU, &paymentName, &productName, &skuName)

		if err != nil {
			log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Transaction not found for DANA webhook")
			sendDANAResponse(w, "2005600", "Successful")
			return
		}

		// Idempotency: If already paid/success, return success immediately
		if paymentStatus == "PAID" || status == "SUCCESS" {
			log.Info().
				Str("invoice", invoiceNumber).
				Str("status", status).
				Str("payment_status", paymentStatus).
				Msg("DANA webhook: Transaction already paid/success, ignoring")
			sendDANAResponse(w, "2005600", "Successful")
			return
		}

		// Parse the full notification for logging
		var fullNotification map[string]interface{}
		json.Unmarshal(body, &fullNotification)

		// Create payment callback log entry
		paymentCallbackLog := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"type":      "PAYMENT_CALLBACK",
			"data":      fullNotification,
		}
		paymentCallbackJSON, _ := json.Marshal([]interface{}{paymentCallbackLog})

		// Only process if status is PENDING or payment is UNPAID/EXPIRED
		if notification.LatestTransactionStatus == "00" {
			// Payment Successful
			log.Info().
				Str("invoice", invoiceNumber).
				Msg("DANA Payment Successful, proceeding to fulfillment")

			// Update to PAID and PROCESSING with payment log
			paidAt := time.Now()
			_, err = deps.DB.Pool.Exec(ctx, `
				UPDATE transactions
				SET payment_status = 'PAID', status = 'PROCESSING', paid_at = $1, 
				    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $2::jsonb,
				    updated_at = NOW()
				WHERE id = $3
			`, paidAt, string(paymentCallbackJSON), transactionID)

			if err != nil {
				log.Error().Err(err).Msg("Failed to update transaction status to PROCESSING")
				sendDANAResponse(w, "5005601", "Internal Server Error")
				return
			}

			// Update payment_data table
			rawCallbackJSON, _ := json.Marshal(fullNotification)
			_, _ = deps.DB.Pool.Exec(ctx, `
				UPDATE payment_data 
				SET status = 'PAID', paid_at = $1, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $2::jsonb), updated_at = NOW()
				WHERE invoice_number = $3
			`, paidAt, string(rawCallbackJSON), invoiceNumber)

			// Add log: Payment received via {payment.name}
			paymentReceivedMessage := fmt.Sprintf("Payment received via %s.", paymentName)
			if paymentName == "" {
				paymentReceivedMessage = "Payment received via DANA."
			}
			deps.DB.Pool.Exec(ctx, `
				INSERT INTO transaction_logs (transaction_id, status, message, created_at)
				VALUES ($1, 'PAYMENT', $2, NOW())
			`, transactionID, paymentReceivedMessage)

			// Get Customer No (Zone if needed) from account inputs
			var accInputs map[string]interface{}
			customerNo := ""
			if err := json.Unmarshal([]byte(accountInputs), &accInputs); err == nil {
				var userIdStr string
				if userId, ok := accInputs["userId"].(string); ok && userId != "" {
					userIdStr = userId
				} else if userIdFloat, ok := accInputs["userId"].(float64); ok {
					userIdStr = strconv.FormatFloat(userIdFloat, 'f', 0, 64)
				}

				if userIdStr != "" {
					customerNo = userIdStr
					if zoneId, ok := accInputs["zoneId"].(string); ok && zoneId != "" {
						customerNo = userIdStr + zoneId
					} else if zoneIdFloat, ok := accInputs["zoneId"].(float64); ok {
						customerNo = userIdStr + strconv.FormatFloat(zoneIdFloat, 'f', 0, 64)
					} else if serverId, ok := accInputs["serverId"].(string); ok && serverId != "" {
						customerNo = userIdStr + serverId
					}
				} else if phone, ok := accInputs["phoneNumber"].(string); ok && phone != "" {
					customerNo = phone
				}
			}

			// Call Provider
			if deps.ProviderManager != nil && providerCode != "" {
				providerKey := strings.ToLower(providerCode)
				prov, err := deps.ProviderManager.Get(providerKey)

				if err != nil {
					log.Error().Err(err).Str("provider", providerKey).Msg("Provider not found in manager")
				} else {
					// Add timeline entry: Processing order {product.name} {sku.name}
					processingMessage := fmt.Sprintf("Processing order %s %s", productName, skuName)
					_, _ = deps.DB.Pool.Exec(ctx, `
						INSERT INTO transaction_logs (transaction_id, status, message, created_at)
						VALUES ($1, 'PROCESSING', $2, NOW())
					`, transactionID, processingMessage)

					req := &provider.OrderRequest{
						RefID:      invoiceNumber,
						SKU:        providerSKU,
						CustomerNo: customerNo,
					}

					result, err := prov.CreateOrder(ctx, req)

					// Log ORDER_REQUEST
					if result != nil && len(result.RawRequest) > 0 {
						var rawReqData interface{}
						json.Unmarshal(result.RawRequest, &rawReqData)
						orderReqLog := map[string]interface{}{
							"timestamp": time.Now().Format(time.RFC3339),
							"type":      "ORDER_REQUEST",
							"data":      rawReqData,
						}
						orderReqJSON, _ := json.Marshal([]interface{}{orderReqLog})
						deps.DB.Pool.Exec(ctx, `
							UPDATE transactions
							SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb, updated_at = NOW()
							WHERE id = $2
						`, string(orderReqJSON), transactionID)
					}

					if err != nil {
						log.Error().Err(err).Msg("Provider CreateOrder failed")

						// Log ORDER_FAILED
						orderFailLog := map[string]interface{}{
							"timestamp": time.Now().Format(time.RFC3339),
							"type":      "ORDER_FAILED",
							"data": map[string]interface{}{
								"error": err.Error(),
							},
						}
						orderFailJSON, _ := json.Marshal([]interface{}{orderFailLog})
						deps.DB.Pool.Exec(ctx, `
							UPDATE transactions
							SET status = 'FAILED',
							    provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb,
							    updated_at = NOW()
							WHERE id = $2
						`, string(orderFailJSON), transactionID)

						// Add timeline entry for failed
						_, _ = deps.DB.Pool.Exec(ctx, `
							INSERT INTO transaction_logs (transaction_id, status, message, created_at)
							VALUES ($1, 'FAILED', $2, NOW())
						`, transactionID, "Item has been failed to sent.")
					} else {
						// Log ORDER_RESPONSE
						var rawRespData interface{}
						if len(result.RawResponse) > 0 {
							json.Unmarshal(result.RawResponse, &rawRespData)
						} else {
							rawRespData = map[string]interface{}{
								"ref_id":  result.RefID,
								"status":  result.Status,
								"message": result.Message,
								"sn":      result.SN,
							}
						}
						orderRespLog := map[string]interface{}{
							"timestamp": time.Now().Format(time.RFC3339),
							"type":      "ORDER_RESPONSE",
							"data":      rawRespData,
						}
						orderRespJSON, _ := json.Marshal([]interface{}{orderRespLog})

						var finalStatus string
						if result.Status == "PENDING" {
							finalStatus = "PROCESSING"
							// Don't add timeline entry for PENDING - "Processing order" was already added before CreateOrder
						} else if result.Status == "FAILED" {
							finalStatus = "FAILED"
						} else if result.Status == "SUCCESS" {
							finalStatus = "SUCCESS"
						} else {
							finalStatus = "PROCESSING"
						}

						providerRespJSON, _ := json.Marshal(map[string]interface{}{
							"ref_id":  result.RefID,
							"status":  result.Status,
							"message": result.Message,
							"sn":      result.SN,
						})

						_, _ = deps.DB.Pool.Exec(ctx, `
							UPDATE transactions
							SET status = $1::transaction_status,
							    provider_ref_id = $2,
							    provider_serial_number = $3,
							    provider_response = $4,
							    provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $5::jsonb,
							    completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END,
							    updated_at = NOW()
							WHERE id = $6
						`, finalStatus, result.ProviderRefID, result.SN, string(providerRespJSON), string(orderRespJSON), transactionID)

						// Only add timeline entry for final status (SUCCESS or FAILED), not for PROCESSING/PENDING
						if finalStatus == "SUCCESS" {
							_, _ = deps.DB.Pool.Exec(ctx, `
								INSERT INTO transaction_logs (transaction_id, status, message, created_at)
								VALUES ($1, 'SUCCESS', $2, NOW())
							`, transactionID, "Item has been successfully sent.")
						} else if finalStatus == "FAILED" {
							_, _ = deps.DB.Pool.Exec(ctx, `
								INSERT INTO transaction_logs (transaction_id, status, message, created_at)
								VALUES ($1, 'FAILED', $2, NOW())
							`, transactionID, "Item has been failed to sent.")
						}
					}
				}
			} else {
				log.Error().Str("provider_code", providerCode).Msg("ProviderManager is nil or provider code empty")
			}

			sendDANAResponse(w, "2005600", "Successful")
			return

		} else if notification.LatestTransactionStatus == "05" {
			// Payment Failed
			log.Info().Str("invoice", invoiceNumber).Msg("DANA Payment Failed")

			_, _ = deps.DB.Pool.Exec(ctx, `
				UPDATE transactions
				SET status = 'FAILED', payment_status = 'FAILED', 
				    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $1::jsonb,
				    updated_at = NOW()
				WHERE id = $2
			`, string(paymentCallbackJSON), transactionID)

			deps.DB.Pool.Exec(ctx, `
				INSERT INTO transaction_logs (transaction_id, status, message, created_at)
				VALUES ($1, 'FAILED', 'Payment failed by user', NOW())
			`, transactionID)

			sendDANAResponse(w, "2005600", "Successful")
			return
		} else {
			// Other statuses (Pending 01/02), just log the callback
			log.Info().Str("invoice", invoiceNumber).Str("status", notification.LatestTransactionStatus).Msg("DANA Payment Pending")

			// Still log the callback
			_, _ = deps.DB.Pool.Exec(ctx, `
				UPDATE transactions
				SET payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $1::jsonb, updated_at = NOW()
				WHERE id = $2
			`, string(paymentCallbackJSON), transactionID)

			sendDANAResponse(w, "2005600", "Successful")
			return
		}
	}
}

// sendDANAResponse sends DANA webhook response in required format
func sendDANAResponse(w http.ResponseWriter, responseCode, responseMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-TIMESTAMP", time.Now().Format("2006-01-02T15:04:05-07:00"))

	response := map[string]interface{}{
		"responseCode":    responseCode,
		"responseMessage": responseMessage,
	}
	json.NewEncoder(w).Encode(response)
}

func HandleXenditWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read Xendit webhook body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Info().
			Str("body", string(body)).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Received Xendit webhook")

		// Try to parse as Payment Requests API webhook (for retail payments)
		var webhookEvent struct {
			Event      string `json:"event"`
			BusinessID string `json:"business_id"`
			Created    string `json:"created"`
			Data       struct {
				PaymentID        string  `json:"payment_id"`
				Status           string  `json:"status"`
				PaymentRequestID string  `json:"payment_request_id"`
				RequestAmount    float64 `json:"request_amount"`
				ChannelCode      string  `json:"channel_code"`
				ReferenceID      string  `json:"reference_id"`
				FailureCode      string  `json:"failure_code,omitempty"`
			} `json:"data"`
		}

		// Parse the full body for logging
		var fullWebhookData map[string]interface{}
		json.Unmarshal(body, &fullWebhookData)

		// Create payment callback log entry
		paymentCallbackLog := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"type":      "PAYMENT_CALLBACK",
			"data":      fullWebhookData,
		}
		paymentCallbackJSON, _ := json.Marshal([]interface{}{paymentCallbackLog})

		if err := json.Unmarshal(body, &webhookEvent); err == nil && webhookEvent.Event != "" {
			// This is a Payment Requests API webhook
			invoiceNumber := webhookEvent.Data.ReferenceID
			if invoiceNumber == "" {
				log.Warn().Msg("Xendit webhook: missing reference_id")
				w.WriteHeader(http.StatusOK)
				return
			}

			// Check if this is a deposit (starts with "SEAD") or transaction (starts with "SEAI")
			if strings.HasPrefix(invoiceNumber, "SEAD") {
				// Handle as deposit
				handleDepositPaymentCallback(ctx, deps, invoiceNumber, webhookEvent, body, w, "XENDIT")
				return
			}

			log.Info().
				Str("event", webhookEvent.Event).
				Str("invoice_number", invoiceNumber).
				Str("status", webhookEvent.Data.Status).
				Str("channel_code", webhookEvent.Data.ChannelCode).
				Float64("amount", webhookEvent.Data.RequestAmount).
				Msg("Processing Xendit Payment Requests webhook")

			// Map status
			var newStatus, newPaymentStatus string
			var timelineMessage string
			shouldProcessProvider := false

			switch webhookEvent.Event {
			case "payment.capture":
				if webhookEvent.Data.Status == "SUCCEEDED" {
					newStatus = "PROCESSING"
					newPaymentStatus = "PAID"
					timelineMessage = "Payment received via " + webhookEvent.Data.ChannelCode + "."
					shouldProcessProvider = true
				}
			case "payment.failure":
				newStatus = "FAILED"
				newPaymentStatus = "FAILED"
				timelineMessage = "Payment failed: " + webhookEvent.Data.FailureCode + "."
			default:
				log.Info().
					Str("event", webhookEvent.Event).
					Msg("Xendit webhook: unhandled event, ignoring")
				w.WriteHeader(http.StatusOK)
				return
			}

			// Update transaction status with payment log
			paidAt := time.Now()
			var updateQuery string
			var updateArgs []interface{}
			if shouldProcessProvider {
				updateQuery = `
					UPDATE transactions
					SET status = $1, payment_status = $2, paid_at = $3, processed_at = $3,
					    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $4::jsonb, updated_at = NOW()
					WHERE invoice_number = $5 AND status NOT IN ('SUCCESS', 'FAILED')
				`
				updateArgs = []interface{}{newStatus, newPaymentStatus, paidAt, string(paymentCallbackJSON), invoiceNumber}
			} else {
				updateQuery = `
					UPDATE transactions
					SET status = $1, payment_status = $2,
					    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $3::jsonb, updated_at = NOW()
					WHERE invoice_number = $4 AND status NOT IN ('SUCCESS', 'FAILED')
				`
				updateArgs = []interface{}{newStatus, newPaymentStatus, string(paymentCallbackJSON), invoiceNumber}
			}

			result, err := deps.DB.Pool.Exec(ctx, updateQuery, updateArgs...)

			if err != nil {
				log.Error().
					Err(err).
					Str("invoice_number", invoiceNumber).
					Msg("Failed to update transaction from Xendit webhook")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if result.RowsAffected() > 0 {
				// Update payment_data table
				paymentDataStatus := "PENDING"
				switch newPaymentStatus {
				case "PAID":
					paymentDataStatus = "PAID"
				case "FAILED":
					paymentDataStatus = "FAILED"
				}
				rawCallbackJSON, _ := json.Marshal(webhookEvent)
				if shouldProcessProvider {
					_, _ = deps.DB.Pool.Exec(ctx, `
						UPDATE payment_data 
						SET status = $1, paid_at = $2, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $3::jsonb), updated_at = NOW()
						WHERE invoice_number = $4
					`, paymentDataStatus, paidAt, string(rawCallbackJSON), invoiceNumber)
				} else {
					_, _ = deps.DB.Pool.Exec(ctx, `
						UPDATE payment_data 
						SET status = $1, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $2::jsonb), updated_at = NOW()
						WHERE invoice_number = $3
					`, paymentDataStatus, string(rawCallbackJSON), invoiceNumber)
				}

				// Get transaction details for timeline and provider processing
				var transactionID, providerID, accountInputs string
				var providerCode, providerSKU, paymentName, productName, skuName string
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT t.id, t.provider_id, t.account_inputs,
					       COALESCE(p.code, ''), COALESCE(s.provider_sku_code, ''),
					       COALESCE(pc.name, ''), COALESCE(pr.title, ''), COALESCE(s.name, '')
					FROM transactions t
					LEFT JOIN providers p ON t.provider_id = p.id
					LEFT JOIN skus s ON t.sku_id = s.id
					LEFT JOIN payment_channels pc ON t.payment_channel_id = pc.id
					LEFT JOIN products pr ON s.product_id = pr.id
					WHERE t.invoice_number = $1
				`, invoiceNumber).Scan(&transactionID, &providerID, &accountInputs, &providerCode, &providerSKU, &paymentName, &productName, &skuName)

				if err == nil && transactionID != "" {
					// Add timeline entry: Payment received via {payment.name}
					paymentReceivedMessage := fmt.Sprintf("Payment received via %s.", paymentName)
					if paymentName == "" {
						paymentReceivedMessage = timelineMessage
					}
					_, _ = deps.DB.Pool.Exec(ctx, `
						INSERT INTO transaction_logs (transaction_id, status, message, created_at)
						VALUES ($1, 'PAYMENT', $2, NOW())
					`, transactionID, paymentReceivedMessage)

					// Process to provider if payment successful
					if shouldProcessProvider && deps.ProviderManager != nil && providerCode != "" {
						// Parse account inputs
						var accInputs map[string]interface{}
						customerNo := ""
						if err := json.Unmarshal([]byte(accountInputs), &accInputs); err == nil {
							var userIdStr string
							if userId, ok := accInputs["userId"].(string); ok && userId != "" {
								userIdStr = userId
							} else if userIdFloat, ok := accInputs["userId"].(float64); ok {
								userIdStr = strconv.FormatFloat(userIdFloat, 'f', 0, 64)
							}
							if userIdStr != "" {
								customerNo = userIdStr
								if zoneId, ok := accInputs["zoneId"].(string); ok && zoneId != "" {
									customerNo = userIdStr + zoneId
								} else if zoneIdFloat, ok := accInputs["zoneId"].(float64); ok {
									customerNo = userIdStr + strconv.FormatFloat(zoneIdFloat, 'f', 0, 64)
								} else if serverId, ok := accInputs["serverId"].(string); ok && serverId != "" {
									customerNo = userIdStr + serverId
								}
							} else if phone, ok := accInputs["phoneNumber"].(string); ok && phone != "" {
								customerNo = phone
							}
						}

						if customerNo != "" && providerSKU != "" {
							prov, err := deps.ProviderManager.Get(strings.ToLower(providerCode))
							if err == nil {
								go func(invNum, txID, custNo, provCode, sku, prodName, skuName string) {
									providerCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
									defer cancel()

									// Add timeline entry: Processing order {product.name} {sku.name}
									processingMessage := fmt.Sprintf("Processing order %s %s.", prodName, skuName)
									_, _ = deps.DB.Pool.Exec(providerCtx, `
										INSERT INTO transaction_logs (transaction_id, status, message, created_at)
										VALUES ($1, 'PROCESSING', $2, NOW())
									`, txID, processingMessage)

									req := &provider.OrderRequest{
										RefID:      invNum,
										SKU:        sku,
										CustomerNo: custNo,
									}

									orderResp, err := prov.CreateOrder(providerCtx, req)

									// Log ORDER_REQUEST
									if orderResp != nil && len(orderResp.RawRequest) > 0 {
										var rawReqData interface{}
										json.Unmarshal(orderResp.RawRequest, &rawReqData)
										orderReqLog := map[string]interface{}{
											"timestamp": time.Now().Format(time.RFC3339),
											"type":      "ORDER_REQUEST",
											"data":      rawReqData,
										}
										orderReqJSON, _ := json.Marshal([]interface{}{orderReqLog})
										deps.DB.Pool.Exec(providerCtx, `
											UPDATE transactions
											SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
											WHERE id = $2
										`, string(orderReqJSON), txID)
									}

									if err != nil {
										log.Error().Err(err).Str("invoice_number", invNum).Msg("Provider CreateOrder failed")
										orderFailLog := map[string]interface{}{
											"timestamp": time.Now().Format(time.RFC3339),
											"type":      "ORDER_FAILED",
											"data":      map[string]interface{}{"error": err.Error()},
										}
										orderFailJSON, _ := json.Marshal([]interface{}{orderFailLog})
										deps.DB.Pool.Exec(providerCtx, `
											UPDATE transactions
											SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
											WHERE id = $2
										`, string(orderFailJSON), txID)
										return
									}

									// Log ORDER_RESPONSE
									var rawRespData interface{}
									if len(orderResp.RawResponse) > 0 {
										json.Unmarshal(orderResp.RawResponse, &rawRespData)
									} else {
										rawRespData = map[string]interface{}{
											"ref_id": orderResp.RefID, "status": orderResp.Status,
											"message": orderResp.Message, "sn": orderResp.SN,
										}
									}
									orderRespLog := map[string]interface{}{
										"timestamp": time.Now().Format(time.RFC3339),
										"type":      "ORDER_RESPONSE",
										"data":      rawRespData,
									}
									orderRespJSON, _ := json.Marshal([]interface{}{orderRespLog})

									updateStatus := "PROCESSING"
									if orderResp.Status == "SUCCESS" {
										updateStatus = "SUCCESS"
									} else if orderResp.Status == "FAILED" {
										updateStatus = "FAILED"
									}

									providerRespJSON, _ := json.Marshal(map[string]interface{}{
										"ref_id": orderResp.RefID, "status": orderResp.Status,
										"message": orderResp.Message, "sn": orderResp.SN,
									})

									deps.DB.Pool.Exec(providerCtx, `
										UPDATE transactions
										SET status = $1::transaction_status,
										    provider_ref_id = $2,
										    provider_serial_number = $3,
										    provider_response = $4,
										    provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $5::jsonb,
										    completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END,
										    updated_at = NOW()
										WHERE id = $6
									`, updateStatus, orderResp.ProviderRefID, orderResp.SN, string(providerRespJSON), string(orderRespJSON), txID)

									// Add timeline entry: Item has been successfully sent or failed to sent (only for final status)
									if updateStatus == "SUCCESS" || updateStatus == "FAILED" {
										var finalMessage string
										if updateStatus == "SUCCESS" {
											finalMessage = "Item has been successfully sent."
										} else {
											finalMessage = "Item has been failed to sent."
										}
										deps.DB.Pool.Exec(providerCtx, `
											INSERT INTO transaction_logs (transaction_id, status, message, created_at)
											VALUES ($1, $2, $3, NOW())
										`, txID, updateStatus, finalMessage)
									}
								}(invoiceNumber, transactionID, customerNo, providerCode, providerSKU, productName, skuName)
							}
						}
					}
				}

				log.Info().
					Str("invoice_number", invoiceNumber).
					Str("new_status", newStatus).
					Str("payment_status", newPaymentStatus).
					Msg("Transaction updated from Xendit webhook")
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		// Try to parse as VA callback (legacy)
		var vaCallback struct {
			ExternalID           string  `json:"external_id"`
			BankCode             string  `json:"bank_code"`
			Amount               float64 `json:"amount"`
			TransactionTimestamp string  `json:"transaction_timestamp"`
		}

		if err := json.Unmarshal(body, &vaCallback); err == nil && vaCallback.ExternalID != "" {
			invoiceNumber := vaCallback.ExternalID

			log.Info().
				Str("invoice_number", invoiceNumber).
				Str("bank_code", vaCallback.BankCode).
				Float64("amount", vaCallback.Amount).
				Msg("Processing Xendit VA callback")

			// Update transaction status for VA payment with payment log
			paidAt := time.Now()
			result, err := deps.DB.Pool.Exec(ctx, `
				UPDATE transactions
				SET status = 'PROCESSING', payment_status = 'PAID', paid_at = $1, processed_at = $1,
				    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $2::jsonb, updated_at = NOW()
				WHERE invoice_number = $3 AND status NOT IN ('SUCCESS', 'FAILED')
			`, paidAt, string(paymentCallbackJSON), invoiceNumber)

			if err != nil {
				log.Error().
					Err(err).
					Str("invoice_number", invoiceNumber).
					Msg("Failed to update transaction from Xendit VA callback")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if result.RowsAffected() > 0 {
				// Update payment_data table
				rawCallbackJSON, _ := json.Marshal(vaCallback)
				_, _ = deps.DB.Pool.Exec(ctx, `
					UPDATE payment_data 
					SET status = 'PAID', paid_at = $1, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $2::jsonb), updated_at = NOW()
					WHERE invoice_number = $3
				`, paidAt, string(rawCallbackJSON), invoiceNumber)

				// Get transaction details
				var transactionID, providerID, accountInputs string
				var providerCode, providerSKU, paymentName, productName, skuName string
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT t.id, t.provider_id, t.account_inputs,
					       COALESCE(p.code, ''), COALESCE(s.provider_sku_code, ''),
					       COALESCE(pc.name, ''), COALESCE(pr.title, ''), COALESCE(s.name, '')
					FROM transactions t
					LEFT JOIN providers p ON t.provider_id = p.id
					LEFT JOIN skus s ON t.sku_id = s.id
					LEFT JOIN payment_channels pc ON t.payment_channel_id = pc.id
					LEFT JOIN products pr ON s.product_id = pr.id
					WHERE t.invoice_number = $1
				`, invoiceNumber).Scan(&transactionID, &providerID, &accountInputs, &providerCode, &providerSKU, &paymentName, &productName, &skuName)

				if err == nil && transactionID != "" {
					// Add timeline entry: Payment received via {payment.name}
					paymentReceivedMessage := fmt.Sprintf("Payment received via %s.", paymentName)
					if paymentName == "" {
						paymentReceivedMessage = "Payment received via " + vaCallback.BankCode + "."
					}
					_, _ = deps.DB.Pool.Exec(ctx, `
						INSERT INTO transaction_logs (transaction_id, status, message, created_at)
						VALUES ($1, 'PAYMENT', $2, NOW())
					`, transactionID, paymentReceivedMessage)

					// Process to provider
					if deps.ProviderManager != nil && providerCode != "" {
						var accInputs map[string]interface{}
						customerNo := ""
						if err := json.Unmarshal([]byte(accountInputs), &accInputs); err == nil {
							var userIdStr string
							if userId, ok := accInputs["userId"].(string); ok && userId != "" {
								userIdStr = userId
							} else if userIdFloat, ok := accInputs["userId"].(float64); ok {
								userIdStr = strconv.FormatFloat(userIdFloat, 'f', 0, 64)
							}
							if userIdStr != "" {
								customerNo = userIdStr
								if zoneId, ok := accInputs["zoneId"].(string); ok && zoneId != "" {
									customerNo = userIdStr + zoneId
								} else if zoneIdFloat, ok := accInputs["zoneId"].(float64); ok {
									customerNo = userIdStr + strconv.FormatFloat(zoneIdFloat, 'f', 0, 64)
								} else if serverId, ok := accInputs["serverId"].(string); ok && serverId != "" {
									customerNo = userIdStr + serverId
								}
							} else if phone, ok := accInputs["phoneNumber"].(string); ok && phone != "" {
								customerNo = phone
							}
						}

						if customerNo != "" && providerSKU != "" {
							prov, err := deps.ProviderManager.Get(strings.ToLower(providerCode))
							if err == nil {
								go func(invNum, txID, custNo, provCode, sku, prodName, skuName string) {
									providerCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
									defer cancel()

									// Add timeline entry: Processing order {product.name} {sku.name}
									processingMessage := fmt.Sprintf("Processing order %s %s", prodName, skuName)
									_, _ = deps.DB.Pool.Exec(providerCtx, `
										INSERT INTO transaction_logs (transaction_id, status, message, created_at)
										VALUES ($1, 'PROCESSING', $2, NOW())
									`, txID, processingMessage)

									req := &provider.OrderRequest{RefID: invNum, SKU: sku, CustomerNo: custNo}
									orderResp, err := prov.CreateOrder(providerCtx, req)

									if orderResp != nil && len(orderResp.RawRequest) > 0 {
										var rawReqData interface{}
										json.Unmarshal(orderResp.RawRequest, &rawReqData)
										orderReqLog := map[string]interface{}{"timestamp": time.Now().Format(time.RFC3339), "type": "ORDER_REQUEST", "data": rawReqData}
										orderReqJSON, _ := json.Marshal([]interface{}{orderReqLog})
										deps.DB.Pool.Exec(providerCtx, `UPDATE transactions SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb WHERE id = $2`, string(orderReqJSON), txID)
									}

									if err != nil {
										orderFailLog := map[string]interface{}{"timestamp": time.Now().Format(time.RFC3339), "type": "ORDER_FAILED", "data": map[string]interface{}{"error": err.Error()}}
										orderFailJSON, _ := json.Marshal([]interface{}{orderFailLog})
										deps.DB.Pool.Exec(providerCtx, `UPDATE transactions SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb WHERE id = $2`, string(orderFailJSON), txID)
										return
									}

									var rawRespData interface{}
									if len(orderResp.RawResponse) > 0 {
										json.Unmarshal(orderResp.RawResponse, &rawRespData)
									} else {
										rawRespData = map[string]interface{}{"ref_id": orderResp.RefID, "status": orderResp.Status, "message": orderResp.Message, "sn": orderResp.SN}
									}
									orderRespLog := map[string]interface{}{"timestamp": time.Now().Format(time.RFC3339), "type": "ORDER_RESPONSE", "data": rawRespData}
									orderRespJSON, _ := json.Marshal([]interface{}{orderRespLog})

									updateStatus := "PROCESSING"
									if orderResp.Status == "SUCCESS" {
										updateStatus = "SUCCESS"
									} else if orderResp.Status == "FAILED" {
										updateStatus = "FAILED"
									}

									providerRespJSON, _ := json.Marshal(map[string]interface{}{"ref_id": orderResp.RefID, "status": orderResp.Status, "message": orderResp.Message, "sn": orderResp.SN})
									deps.DB.Pool.Exec(providerCtx, `
										UPDATE transactions SET status = $1::transaction_status, provider_ref_id = $2, provider_serial_number = $3, provider_response = $4,
										provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $5::jsonb, completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END, updated_at = NOW()
										WHERE id = $6
									`, updateStatus, orderResp.ProviderRefID, orderResp.SN, string(providerRespJSON), string(orderRespJSON), txID)
									// Add timeline entry: Item has been successfully sent or failed to sent (only for final status)
									if updateStatus == "SUCCESS" || updateStatus == "FAILED" {
										var finalMessage string
										if updateStatus == "SUCCESS" {
											finalMessage = "Item has been successfully sent."
										} else {
											finalMessage = "Item has been failed to sent."
										}
										deps.DB.Pool.Exec(providerCtx, `INSERT INTO transaction_logs (transaction_id, status, message, created_at) VALUES ($1, $2, $3, NOW())`, txID, updateStatus, finalMessage)
									}
								}(invoiceNumber, transactionID, customerNo, providerCode, providerSKU, productName, skuName)
							}
						}
					}
				}

				log.Info().
					Str("invoice_number", invoiceNumber).
					Msg("Transaction updated from Xendit VA callback")
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleMidtransWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read Midtrans webhook body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		log.Info().
			Str("body", string(body)).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Received Midtrans webhook")

		// Parse webhook notification
		var notification struct {
			TransactionTime   string `json:"transaction_time"`
			TransactionStatus string `json:"transaction_status"`
			TransactionID     string `json:"transaction_id"`
			StatusMessage     string `json:"status_message"`
			StatusCode        string `json:"status_code"`
			SignatureKey      string `json:"signature_key"`
			SettlementTime    string `json:"settlement_time"`
			PaymentType       string `json:"payment_type"`
			OrderID           string `json:"order_id"`
			MerchantID        string `json:"merchant_id"`
			GrossAmount       string `json:"gross_amount"`
			FraudStatus       string `json:"fraud_status"`
			ExpiryTime        string `json:"expiry_time"`
			Currency          string `json:"currency"`
		}

		if err := json.Unmarshal(body, &notification); err != nil {
			log.Error().Err(err).Str("body", string(body)).Msg("Failed to parse Midtrans webhook")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Get invoice number (order_id is our invoice number)
		invoiceNumber := notification.OrderID
		if invoiceNumber == "" {
			log.Warn().Msg("Midtrans webhook: missing order_id")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Check if this is a deposit (starts with "SEAD") or transaction (starts with "SEAI")
		if strings.HasPrefix(invoiceNumber, "SEAD") {
			// Handle as deposit
			handleDepositPaymentCallback(ctx, deps, invoiceNumber, notification, body, w, "MIDTRANS")
			return
		}

		log.Info().
			Str("invoice_number", invoiceNumber).
			Str("transaction_id", notification.TransactionID).
			Str("status", notification.TransactionStatus).
			Str("gross_amount", notification.GrossAmount).
			Str("payment_type", notification.PaymentType).
			Msg("Processing Midtrans webhook")

		// Map Midtrans status to internal status according to requirements
		var newStatus, newPaymentStatus string
		var timelineMessage string
		var shouldProcessProvider bool
		var paidAtTime *time.Time

		switch notification.TransactionStatus {
		case "settlement":
			// settlement -> payment status: PAID, transaction status: PROCESSING -> process to provider
			newStatus = "PROCESSING"
			newPaymentStatus = "PAID"
			timelineMessage = "Payment received via " + notification.PaymentType + "."
			shouldProcessProvider = true
			// Parse settlement time or use current time
			if notification.SettlementTime != "" {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", notification.SettlementTime); err == nil {
					paidAtTime = &parsedTime
				}
			}
			if paidAtTime == nil {
				now := time.Now()
				paidAtTime = &now
			}
		case "expire":
			// expire -> payment status: EXPIRED, transaction status: FAILED
			newStatus = "FAILED"
			newPaymentStatus = "EXPIRED"
			timelineMessage = "Payment expired"
			shouldProcessProvider = false
		case "deny":
			// deny -> payment status: FAILED, transaction status: FAILED
			newStatus = "FAILED"
			newPaymentStatus = "FAILED"
			timelineMessage = "Payment denied"
			shouldProcessProvider = false
		case "pending":
			// pending -> payment status: UNPAID, transaction status: PENDING
			newStatus = "PENDING"
			newPaymentStatus = "UNPAID"
			timelineMessage = "Waiting for payment"
			shouldProcessProvider = false
		default:
			log.Info().
				Str("status", notification.TransactionStatus).
				Msg("Midtrans webhook: unhandled status, ignoring")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create payment log entry with full raw callback data
		paymentLogEntry := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"type":      "PAYMENT_CALLBACK",
			"data": map[string]interface{}{
				"transaction_time":   notification.TransactionTime,
				"transaction_status": notification.TransactionStatus,
				"transaction_id":     notification.TransactionID,
				"status_message":     notification.StatusMessage,
				"status_code":        notification.StatusCode,
				"settlement_time":    notification.SettlementTime,
				"payment_type":       notification.PaymentType,
				"order_id":           notification.OrderID,
				"merchant_id":        notification.MerchantID,
				"gross_amount":       notification.GrossAmount,
				"fraud_status":       notification.FraudStatus,
				"expiry_time":        notification.ExpiryTime,
				"currency":           notification.Currency,
			},
		}
		paymentLogJSON, _ := json.Marshal([]interface{}{paymentLogEntry})

		// Build update query
		var updateQuery string
		var updateArgs []interface{}
		if paidAtTime != nil && shouldProcessProvider {
			// For settlement, update paid_at and processed_at
			updateQuery = `
				UPDATE transactions
				SET status = $1, payment_status = $2, paid_at = $3, processed_at = $3, 
				    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $4::jsonb, updated_at = NOW()
				WHERE invoice_number = $5 AND status NOT IN ('SUCCESS', 'FAILED')
			`
			updateArgs = []interface{}{newStatus, newPaymentStatus, paidAtTime, string(paymentLogJSON), invoiceNumber}
		} else {
			// For other statuses, just update status
			updateQuery = `
				UPDATE transactions
				SET status = $1, payment_status = $2, 
				    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $3::jsonb, updated_at = NOW()
				WHERE invoice_number = $4 AND status NOT IN ('SUCCESS', 'FAILED')
			`
			updateArgs = []interface{}{newStatus, newPaymentStatus, string(paymentLogJSON), invoiceNumber}
		}

		// Update transaction status
		result, err := deps.DB.Pool.Exec(ctx, updateQuery, updateArgs...)

		if err != nil {
			log.Error().
				Err(err).
				Str("invoice_number", invoiceNumber).
				Msg("Failed to update transaction from Midtrans webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if result.RowsAffected() > 0 {
			// Update payment_data table status
			paymentDataStatus := "PENDING"
			switch newPaymentStatus {
			case "PAID":
				paymentDataStatus = "PAID"
			case "EXPIRED":
				paymentDataStatus = "EXPIRED"
			case "FAILED":
				paymentDataStatus = "FAILED"
			}
			rawCallbackJSON, _ := json.Marshal(notification)
			if paidAtTime != nil {
				_, _ = deps.DB.Pool.Exec(ctx, `
					UPDATE payment_data 
					SET status = $1, paid_at = $2, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $3::jsonb), updated_at = NOW()
					WHERE invoice_number = $4
				`, paymentDataStatus, paidAtTime, string(rawCallbackJSON), invoiceNumber)
			} else {
				_, _ = deps.DB.Pool.Exec(ctx, `
					UPDATE payment_data 
					SET status = $1, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $3::jsonb), updated_at = NOW()
					WHERE invoice_number = $2
				`, paymentDataStatus, invoiceNumber)
			}

			// Get transaction details for timeline and provider processing
			var transactionID, providerID, skuCode, accountInputs string
			var accountNickname *string
			var paymentName, productName, skuName string
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT t.id, t.provider_id, t.account_inputs, t.account_nickname,
				       COALESCE(pc.name, ''), COALESCE(pr.title, ''), COALESCE(s.name, '')
				FROM transactions t
				LEFT JOIN skus s ON t.sku_id = s.id
				LEFT JOIN payment_channels pc ON t.payment_channel_id = pc.id
				LEFT JOIN products pr ON s.product_id = pr.id
				WHERE t.invoice_number = $1
			`, invoiceNumber).Scan(&transactionID, &providerID, &accountInputs, &accountNickname, &paymentName, &productName, &skuName)

			if err != nil {
				log.Error().
					Err(err).
					Str("invoice_number", invoiceNumber).
					Msg("Failed to get transaction details")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Get provider SKU code and backup codes
			var providerSKUCode string
			var skuCodeBackup1, skuCodeBackup2 *string
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT s.provider_sku_code, s.provider_sku_code_backup1, s.provider_sku_code_backup2
				FROM skus s
				JOIN transactions t ON t.sku_id = s.id
				WHERE t.invoice_number = $1
			`, invoiceNumber).Scan(&providerSKUCode, &skuCodeBackup1, &skuCodeBackup2)
			skuCode = providerSKUCode

			if err != nil {
				log.Error().
					Err(err).
					Str("invoice_number", invoiceNumber).
					Msg("Failed to get SKU code")
			}

			// Add timeline entry: Payment received via {payment.name}
			paymentReceivedMessage := fmt.Sprintf("Payment received via %s.", paymentName)
			if paymentName == "" {
				paymentReceivedMessage = timelineMessage
			}
			_, _ = deps.DB.Pool.Exec(ctx, `
				INSERT INTO transaction_logs (transaction_id, status, message, created_at)
				VALUES ($1, 'PAYMENT', $2, NOW())
			`, transactionID, paymentReceivedMessage)

			// Process to provider if settlement
			if shouldProcessProvider && deps.ProviderManager != nil {
				log.Info().
					Str("invoice_number", invoiceNumber).
					Str("provider_id", providerID).
					Msg("Starting provider processing from payment callback")

				// Get provider code
				var providerCode string
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT code FROM providers WHERE id = $1
				`, providerID).Scan(&providerCode)

				if err != nil {
					log.Error().
						Err(err).
						Str("invoice_number", invoiceNumber).
						Str("provider_id", providerID).
						Msg("Failed to get provider code")
				}

				if err == nil && providerCode != "" {
					log.Info().
						Str("invoice_number", invoiceNumber).
						Str("provider_code", providerCode).
						Str("account_inputs", accountInputs).
						Str("sku_code", skuCode).
						Msg("Got provider code, parsing account inputs")

					// Parse account inputs
					var accountData map[string]interface{}
					if err := json.Unmarshal([]byte(accountInputs), &accountData); err == nil {
						// Extract customer number (userId + zoneId/serverId or phone number)
						customerNo := ""
						// Handle both string and float64 types for userId
						var userIdStr string
						if userId, ok := accountData["userId"].(string); ok && userId != "" {
							userIdStr = userId
						} else if userIdFloat, ok := accountData["userId"].(float64); ok {
							userIdStr = strconv.FormatFloat(userIdFloat, 'f', 0, 64)
						}

						if userIdStr != "" {
							customerNo = userIdStr
							// Append zoneId or serverId if exists
							if zoneId, ok := accountData["zoneId"].(string); ok && zoneId != "" {
								customerNo = userIdStr + zoneId
							} else if zoneIdFloat, ok := accountData["zoneId"].(float64); ok {
								customerNo = userIdStr + strconv.FormatFloat(zoneIdFloat, 'f', 0, 64)
							} else if serverId, ok := accountData["serverId"].(string); ok && serverId != "" {
								customerNo = userIdStr + serverId
							} else if serverIdFloat, ok := accountData["serverId"].(float64); ok {
								customerNo = userIdStr + strconv.FormatFloat(serverIdFloat, 'f', 0, 64)
							}
						} else if phone, ok := accountData["phoneNumber"].(string); ok && phone != "" {
							customerNo = phone
						}

						log.Info().
							Str("invoice_number", invoiceNumber).
							Str("customer_no", customerNo).
							Str("sku_code", skuCode).
							Msg("Parsed customer number from account inputs")

						if customerNo != "" && skuCode != "" {
							// Get provider (convert to lowercase as providers are registered with lowercase names)
							prov, err := deps.ProviderManager.Get(strings.ToLower(providerCode))
							if err == nil {
								// Process order asynchronously with retry logic
								go func(invNum, txID, custNo, provCode, sku, prodName, skuName string, backup1, backup2 *string, prov provider.Provider) {
									providerCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
									defer cancel()

									// Add timeline entry: Processing order {product.name} {sku.name} (only once before retry loop)
									processingMessage := fmt.Sprintf("Processing order %s %s", prodName, skuName)
									_, _ = deps.DB.Pool.Exec(providerCtx, `
										INSERT INTO transaction_logs (transaction_id, status, message, created_at)
										VALUES ($1, 'PROCESSING', $2, NOW())
									`, txID, processingMessage)

									// Retry loop - will try main SKU, then backup1, then backup2
									skuToUse := sku
									for attempt := 0; attempt < 3; attempt++ {
										// Determine which SKU to use
										if attempt == 1 && backup1 != nil && *backup1 != "" {
											skuToUse = *backup1
											// Update retry count
											_, _ = deps.DB.Pool.Exec(providerCtx, `
											UPDATE transactions SET retry_count = 1 WHERE invoice_number = $1
										`, invNum)
										} else if attempt == 2 && backup2 != nil && *backup2 != "" {
											skuToUse = *backup2
											// Update retry count
											_, _ = deps.DB.Pool.Exec(providerCtx, `
											UPDATE transactions SET retry_count = 2 WHERE invoice_number = $1
										`, invNum)
										} else if attempt > 0 {
											break // No more backups available
										}

										log.Info().
											Str("invoice_number", invNum).
											Str("provider", provCode).
											Str("sku", skuToUse).
											Str("customer_no", custNo).
											Int("attempt", attempt+1).
											Msg("Processing transaction to provider")

										orderReq := &provider.OrderRequest{
											RefID:      invNum,
											SKU:        skuToUse,
											CustomerNo: custNo,
										}

										orderResp, err := prov.CreateOrder(providerCtx, orderReq)

										// Log the raw request (from provider response)
										if orderResp != nil && len(orderResp.RawRequest) > 0 {
											var rawReqData interface{}
											json.Unmarshal(orderResp.RawRequest, &rawReqData)
											providerOrderLog := map[string]interface{}{
												"timestamp": time.Now().Format(time.RFC3339),
												"type":      "ORDER_REQUEST",
												"data":      rawReqData,
											}
											providerOrderJSON, _ := json.Marshal([]interface{}{providerOrderLog})
											_, reqErr := deps.DB.Pool.Exec(providerCtx, `
											UPDATE transactions
											SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb, updated_at = NOW()
											WHERE invoice_number = $2
										`, string(providerOrderJSON), invNum)
											if reqErr != nil {
												log.Error().Err(reqErr).Str("invoice_number", invNum).Msg("Failed to log ORDER_REQUEST")
											} else {
												log.Info().Str("invoice_number", invNum).Msg("Successfully logged ORDER_REQUEST")
											}
										}
										if err != nil {
											log.Error().Err(err).
												Str("invoice_number", invNum).
												Str("provider", provCode).
												Int("attempt", attempt+1).
												Msg("Failed to process transaction to provider")

											// Add provider log for failure
											providerFailLog := map[string]interface{}{
												"timestamp": time.Now().Format(time.RFC3339),
												"type":      "ORDER_FAILED",
												"data": map[string]interface{}{
													"error":   err.Error(),
													"sku":     skuToUse,
													"attempt": attempt + 1,
												},
											}
											providerFailJSON, _ := json.Marshal([]interface{}{providerFailLog})
											_, _ = deps.DB.Pool.Exec(context.Background(), `
											UPDATE transactions
											SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
											WHERE invoice_number = $2
										`, string(providerFailJSON), invNum)

											// Check if we have more backups to try
											if attempt == 0 && backup1 != nil && *backup1 != "" {
												log.Info().Msg("Main SKU failed, trying backup1...")
												continue
											}
											if attempt == 1 && backup2 != nil && *backup2 != "" {
												log.Info().Msg("Backup1 failed, trying backup2...")
												continue
											}

											// No more backups - mark as failed
											_, _ = deps.DB.Pool.Exec(context.Background(), `
											UPDATE transactions
											SET status = 'FAILED', updated_at = NOW()
											WHERE invoice_number = $1
										`, invNum)
											_, _ = deps.DB.Pool.Exec(context.Background(), `
											INSERT INTO transaction_logs (transaction_id, status, message, created_at)
											VALUES ($1, 'FAILED', $2, NOW())
										`, txID, "Item has been failed to sent.")
											return
										}

										// Check response status
										var updateStatus string
										switch orderResp.Status {
										case "SUCCESS":
											updateStatus = "SUCCESS"
										case "FAILED":
											updateStatus = "FAILED"
										default:
											updateStatus = "PROCESSING"
										}

										// Build provider response JSON for storage
										providerRespJSON, _ := json.Marshal(map[string]interface{}{
											"ref_id":          orderResp.RefID,
											"provider_ref_id": orderResp.ProviderRefID,
											"status":          orderResp.Status,
											"message":         orderResp.Message,
											"sn":              orderResp.SN,
										})

										// Add provider log for response (use raw response from provider)
										var rawRespData interface{}
										if len(orderResp.RawResponse) > 0 {
											json.Unmarshal(orderResp.RawResponse, &rawRespData)
										} else {
											rawRespData = map[string]interface{}{
												"ref_id":          orderResp.RefID,
												"provider_ref_id": orderResp.ProviderRefID,
												"status":          orderResp.Status,
												"message":         orderResp.Message,
												"sn":              orderResp.SN,
											}
										}
										providerRespLog := map[string]interface{}{
											"timestamp": time.Now().Format(time.RFC3339),
											"type":      "ORDER_RESPONSE",
											"data":      rawRespData,
										}
										providerRespLogJSON, _ := json.Marshal([]interface{}{providerRespLog})

										if updateStatus == "PROCESSING" || updateStatus == "SUCCESS" {
											// Success or pending - update and done
											_, err = deps.DB.Pool.Exec(context.Background(), `
											UPDATE transactions
											SET status = $1::transaction_status,
												provider_ref_id = $2,
												provider_serial_number = $3,
												provider_response = $4,
												provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $5::jsonb,
												completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END,
												updated_at = NOW()
											WHERE invoice_number = $6
										`, updateStatus, orderResp.ProviderRefID, orderResp.SN, string(providerRespJSON), string(providerRespLogJSON), invNum)

											if err != nil {
												log.Error().
													Err(err).
													Str("invoice_number", invNum).
													Str("provider_logs_json", string(providerRespLogJSON)).
													Msg("Failed to update transaction with provider response")
											} else {
												log.Info().
													Str("invoice_number", invNum).
													Msg("Successfully updated transaction with ORDER_RESPONSE")
												// Add timeline entry: Item has been successfully sent or failed to sent (only for final status)
												if updateStatus == "SUCCESS" || updateStatus == "FAILED" {
													var finalMessage string
													if updateStatus == "SUCCESS" {
														finalMessage = "Item has been successfully sent."
													} else {
														finalMessage = "Item has been failed to sent."
													}
													_, _ = deps.DB.Pool.Exec(context.Background(), `
													INSERT INTO transaction_logs (transaction_id, status, message, created_at)
													VALUES ($1, $2, $3, NOW())
												`, txID, updateStatus, finalMessage)
												}
											}

											log.Info().
												Str("invoice_number", invNum).
												Str("provider", provCode).
												Str("status", updateStatus).
												Str("serial_number", orderResp.SN).
												Int("attempt", attempt+1).
												Msg("Transaction processed to provider")
											return
										}

										// FAILED - try next backup if available
										_, _ = deps.DB.Pool.Exec(context.Background(), `
										UPDATE transactions
										SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
										WHERE invoice_number = $2
									`, string(providerRespLogJSON), invNum)

										if attempt == 0 && backup1 != nil && *backup1 != "" {
											log.Info().Msg("Main SKU failed, trying backup1...")
											continue
										}
										if attempt == 1 && backup2 != nil && *backup2 != "" {
											log.Info().Msg("Backup1 failed, trying backup2...")
											continue
										}

										// No more backups - mark as failed
										_, _ = deps.DB.Pool.Exec(context.Background(), `
										UPDATE transactions
										SET status = 'FAILED',
											provider_response = $1,
											updated_at = NOW()
										WHERE invoice_number = $2
									`, string(providerRespJSON), invNum)
										_, _ = deps.DB.Pool.Exec(context.Background(), `
										INSERT INTO transaction_logs (transaction_id, status, message, created_at)
										VALUES ($1, 'FAILED', $2, NOW())
									`, txID, "Item has been failed to sent.")
										return
									}
								}(invoiceNumber, transactionID, customerNo, providerCode, skuCode, productName, skuName, skuCodeBackup1, skuCodeBackup2, prov)
							} else {
								log.Error().
									Err(err).
									Str("invoice_number", invoiceNumber).
									Str("provider_code", providerCode).
									Msg("Provider not found")
							}
						} else {
							log.Warn().
								Str("invoice_number", invoiceNumber).
								Str("customer_no", customerNo).
								Str("sku_code", skuCode).
								Msg("Missing customer_no or sku_code, skipping provider processing")
						}
					} else {
						log.Error().
							Str("invoice_number", invoiceNumber).
							Str("account_inputs", accountInputs).
							Msg("Failed to parse account inputs JSON")
					}
				} else {
					log.Warn().
						Str("invoice_number", invoiceNumber).
						Str("provider_code", providerCode).
						Msg("Empty provider code, skipping provider processing")
				}
			} else if shouldProcessProvider {
				log.Warn().
					Str("invoice_number", invoiceNumber).
					Bool("provider_manager_nil", deps.ProviderManager == nil).
					Msg("Cannot process provider - ProviderManager is nil")
			}

			log.Info().
				Str("invoice_number", invoiceNumber).
				Str("new_status", newStatus).
				Str("payment_status", newPaymentStatus).
				Bool("will_process_provider", shouldProcessProvider).
				Msg("Transaction updated from Midtrans webhook")
		} else {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Msg("No transaction updated (already processed or not found)")
		}

		// Return OK to acknowledge receipt
		w.WriteHeader(http.StatusOK)
	}
}

func HandlePakaiLinkWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read PakaiLink webhook body")
			sendPakaiLinkResponse(w, "4002800", "Failed to read request body")
			return
		}
		defer r.Body.Close()

		// Get headers for signature verification
		timestamp := r.Header.Get("X-Timestamp")
		signature := r.Header.Get("X-Signature")

		log.Info().
			Str("body", string(body)).
			Str("timestamp", timestamp).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Received PakaiLink webhook")

		// Parse callback
		var callback struct {
			TransactionData struct {
				PaymentFlagStatus string `json:"paymentFlagStatus"`
				PaymentFlagReason struct {
					English   string `json:"english"`
					Indonesia string `json:"indonesia"`
				} `json:"paymentFlagReason"`
				CustomerNo         string `json:"customerNo"`
				VirtualAccountNo   string `json:"virtualAccountNo"`
				VirtualAccountName string `json:"virtualAccountName"`
				PartnerReferenceNo string `json:"partnerReferenceNo"`
				CallbackType       string `json:"callbackType"`
				PaidAmount         struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"paidAmount"`
				FeeAmount struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"feeAmount"`
				CreditBalance struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"creditBalance"`
				AdditionalInfo struct {
					CallbackUrl string `json:"callbackUrl"`
				} `json:"additionalInfo"`
			} `json:"transactionData"`
		}

		if err := json.Unmarshal(body, &callback); err != nil {
			log.Error().Err(err).Str("body", string(body)).Msg("Failed to parse PakaiLink webhook")
			sendPakaiLinkResponse(w, "4002800", "Failed to parse request body")
			return
		}

		invoiceNumber := callback.TransactionData.PartnerReferenceNo
		callbackType := callback.TransactionData.CallbackType

		// Check if this is a deposit (starts with "SEAD") or transaction (starts with "SEAI")
		if strings.HasPrefix(invoiceNumber, "SEAD") {
			// Handle as deposit
			handleDepositPaymentCallback(ctx, deps, invoiceNumber, callback, body, w, "PAKAILINK")
			return
		}

		log.Info().
			Str("invoice_number", invoiceNumber).
			Str("callback_type", callbackType).
			Str("payment_flag_status", callback.TransactionData.PaymentFlagStatus).
			Str("virtual_account_no", callback.TransactionData.VirtualAccountNo).
			Str("paid_amount", callback.TransactionData.PaidAmount.Value).
			Msg("Processing PakaiLink webhook")

		// If callback type is not "payment", just acknowledge and return
		if callbackType != "payment" {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Str("callback_type", callbackType).
				Msg("PakaiLink webhook: non-payment callback, acknowledging without processing")
			sendPakaiLinkResponse(w, "2002800", "Successful")
			return
		}

		// Verify signature if available
		if signature != "" && timestamp != "" && deps.Config != nil {
			// Get PakaiLink gateway for signature verification
			if deps.PaymentManager != nil {
				gw, err := deps.PaymentManager.Get("PAKAILINK")
				if err == nil {
					if pakaiLinkGw, ok := gw.(*payment.PakaiLinkGateway); ok {
						callbackURL := deps.Config.Payment.PakaiLink.CallbackURL
						if !pakaiLinkGw.VerifyCallbackSignature(callbackURL, body, timestamp, signature) {
							log.Warn().
								Str("invoice_number", invoiceNumber).
								Msg("PakaiLink webhook: signature verification failed, continuing anyway")
							// Continue processing even if signature verification fails
							// as PakaiLink might use different signing method
						}
					}
				}
			}
		}

		// For payment callback, inquiry transaction status first
		var statusVerified bool
		if deps.PaymentManager != nil {
			gw, err := deps.PaymentManager.Get("PAKAILINK")
			if err == nil {
				status, err := gw.CheckStatus(ctx, invoiceNumber)
				if err != nil {
					log.Warn().
						Err(err).
						Str("invoice_number", invoiceNumber).
						Msg("Failed to verify payment status from PakaiLink, using callback data")
				} else if status.Status == "PAID" {
					statusVerified = true
					log.Info().
						Str("invoice_number", invoiceNumber).
						Str("status", status.Status).
						Msg("Payment status verified from PakaiLink API")
				} else {
					log.Warn().
						Str("invoice_number", invoiceNumber).
						Str("status", status.Status).
						Msg("Payment status from API does not match callback")
				}
			}
		}

		// Process payment if payment flag status is "00" (success)
		if callback.TransactionData.PaymentFlagStatus != "00" {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Str("payment_flag_status", callback.TransactionData.PaymentFlagStatus).
				Msg("PakaiLink webhook: payment not successful")
			sendPakaiLinkResponse(w, "2002800", "Successful")
			return
		}

		// Log status verification result
		if statusVerified {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Msg("Processing verified payment from PakaiLink")
		} else {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Msg("Processing payment from PakaiLink callback (unverified)")
		}

		// Get transaction details for provider processing
		var transactionID, providerID, skuCode, accountInputs string
		var accountNickname *string
		var paymentName, productName, skuName string
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT t.id, t.provider_id, t.account_inputs, t.account_nickname,
			       COALESCE(pc.name, ''), COALESCE(pr.title, ''), COALESCE(s.name, '')
			FROM transactions t
			LEFT JOIN skus s ON t.sku_id = s.id
			LEFT JOIN payment_channels pc ON t.payment_channel_id = pc.id
			LEFT JOIN products pr ON s.product_id = pr.id
			WHERE t.invoice_number = $1
		`, invoiceNumber).Scan(&transactionID, &providerID, &accountInputs, &accountNickname, &paymentName, &productName, &skuName)

		if err != nil {
			log.Error().
				Err(err).
				Str("invoice_number", invoiceNumber).
				Msg("Failed to get transaction details")
			sendPakaiLinkResponse(w, "2002800", "Successful")
			return
		}

		// Get SKU code
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT s.code
			FROM skus s
			JOIN transactions t ON t.sku_id = s.id
			WHERE t.invoice_number = $1
		`, invoiceNumber).Scan(&skuCode)

		if err != nil {
			log.Error().
				Err(err).
				Str("invoice_number", invoiceNumber).
				Msg("Failed to get SKU code")
		}

		// Create payment callback log entry
		var fullCallbackData map[string]interface{}
		json.Unmarshal(body, &fullCallbackData)
		paymentCallbackLog := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"type":      "PAYMENT_CALLBACK",
			"data":      fullCallbackData,
		}
		paymentCallbackJSON, _ := json.Marshal([]interface{}{paymentCallbackLog})

		// Update transaction to PROCESSING and PAID with payment log
		paidAt := time.Now()
		result, err := deps.DB.Pool.Exec(ctx, `
			UPDATE transactions
			SET status = 'PROCESSING', payment_status = 'PAID', paid_at = $1, processed_at = $1, 
			    payment_logs = COALESCE(payment_logs, '[]'::jsonb) || $2::jsonb, updated_at = NOW()
			WHERE invoice_number = $3 AND status NOT IN ('SUCCESS', 'FAILED')
		`, paidAt, string(paymentCallbackJSON), invoiceNumber)

		if err != nil {
			log.Error().
				Err(err).
				Str("invoice_number", invoiceNumber).
				Msg("Failed to update transaction from PakaiLink webhook")
			sendPakaiLinkResponse(w, "2002800", "Successful")
			return
		}

		if result.RowsAffected() > 0 {
			// Update payment_data table
			rawCallbackJSON, _ := json.Marshal(callback)
			_, _ = deps.DB.Pool.Exec(ctx, `
				UPDATE payment_data 
				SET status = 'PAID', paid_at = $1, raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $2::jsonb), updated_at = NOW()
				WHERE invoice_number = $3
			`, paidAt, string(rawCallbackJSON), invoiceNumber)

			// Add timeline entry: Payment received via {payment.name}
			paymentReceivedMessage := fmt.Sprintf("Payment received via %s.", paymentName)
			if paymentName == "" {
				paymentReceivedMessage = "Payment received via Virtual Account."
			}
			_, _ = deps.DB.Pool.Exec(ctx, `
				INSERT INTO transaction_logs (transaction_id, status, message, created_at)
				VALUES ($1, 'PAYMENT', $2, NOW())
			`, transactionID, paymentReceivedMessage)

			// Process to provider
			if deps.ProviderManager != nil {
				// Get provider code
				var providerCode string
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT code FROM providers WHERE id = $1
				`, providerID).Scan(&providerCode)

				if err == nil && providerCode != "" {
					// Parse account inputs
					var accountData map[string]interface{}
					if err := json.Unmarshal([]byte(accountInputs), &accountData); err == nil {
						// Extract customer number (userId + zoneId/serverId or phone number)
						customerNo := ""
						if userId, ok := accountData["userId"].(string); ok && userId != "" {
							customerNo = userId
							// Append zoneId or serverId if exists
							if zoneId, ok := accountData["zoneId"].(string); ok && zoneId != "" {
								customerNo = userId + zoneId
							} else if serverId, ok := accountData["serverId"].(string); ok && serverId != "" {
								customerNo = userId + serverId
							}
						} else if phone, ok := accountData["phoneNumber"].(string); ok && phone != "" {
							customerNo = phone
						}

						if customerNo != "" && skuCode != "" {
							// Get provider (convert to lowercase as providers are registered with lowercase names)
							prov, err := deps.ProviderManager.Get(strings.ToLower(providerCode))
							if err == nil {
								// Create order request
								orderReq := &provider.OrderRequest{
									RefID:      invoiceNumber,
									SKU:        skuCode,
									CustomerNo: customerNo,
								}

								// Process order asynchronously
								go func(invNum, txID, custNo, provCode, sku, prodName, skuName string) {
									providerCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
									defer cancel()

									// Add timeline entry: Processing order {product.name} {sku.name}
									processingMessage := fmt.Sprintf("Processing order %s %s", prodName, skuName)
									_, _ = deps.DB.Pool.Exec(providerCtx, `
										INSERT INTO transaction_logs (transaction_id, status, message, created_at)
										VALUES ($1, 'PROCESSING', $2, NOW())
									`, txID, processingMessage)

									log.Info().
										Str("invoice_number", invNum).
										Str("provider", provCode).
										Str("sku", sku).
										Str("customer_no", custNo).
										Msg("Processing PakaiLink transaction to provider")

									orderResp, err := prov.CreateOrder(providerCtx, orderReq)

									// Log ORDER_REQUEST
									if orderResp != nil && len(orderResp.RawRequest) > 0 {
										var rawReqData interface{}
										json.Unmarshal(orderResp.RawRequest, &rawReqData)
										orderReqLog := map[string]interface{}{
											"timestamp": time.Now().Format(time.RFC3339),
											"type":      "ORDER_REQUEST",
											"data":      rawReqData,
										}
										orderReqJSON, _ := json.Marshal([]interface{}{orderReqLog})
										deps.DB.Pool.Exec(providerCtx, `
											UPDATE transactions
											SET provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb
											WHERE id = $2
										`, string(orderReqJSON), txID)
									}

									if err != nil {
										log.Error().
											Err(err).
											Str("invoice_number", invNum).
											Str("provider", provCode).
											Msg("Failed to process transaction to provider")

										// Log ORDER_FAILED
										orderFailLog := map[string]interface{}{
											"timestamp": time.Now().Format(time.RFC3339),
											"type":      "ORDER_FAILED",
											"data":      map[string]interface{}{"error": err.Error()},
										}
										orderFailJSON, _ := json.Marshal([]interface{}{orderFailLog})

										// Update transaction to failed
										_, _ = deps.DB.Pool.Exec(context.Background(), `
											UPDATE transactions
											SET status = 'FAILED', 
											    provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $1::jsonb,
											    updated_at = NOW()
											WHERE id = $2
										`, string(orderFailJSON), txID)

										// Add timeline entry
										_, _ = deps.DB.Pool.Exec(context.Background(), `
											INSERT INTO transaction_logs (transaction_id, status, message, created_at)
											VALUES ($1, $2, $3, NOW())
										`, txID, "FAILED", "Item has been failed to sent.")
										return
									}

									// Update transaction with provider response
									var updateStatus string
									switch orderResp.Status {
									case "SUCCESS":
										updateStatus = "SUCCESS"
									case "FAILED":
										updateStatus = "FAILED"
									default:
										updateStatus = "PROCESSING"
									}

									// Log ORDER_RESPONSE
									var rawRespData interface{}
									if len(orderResp.RawResponse) > 0 {
										json.Unmarshal(orderResp.RawResponse, &rawRespData)
									} else {
										rawRespData = map[string]interface{}{
											"ref_id":          orderResp.RefID,
											"provider_ref_id": orderResp.ProviderRefID,
											"status":          orderResp.Status,
											"message":         orderResp.Message,
											"sn":              orderResp.SN,
										}
									}
									orderRespLog := map[string]interface{}{
										"timestamp": time.Now().Format(time.RFC3339),
										"type":      "ORDER_RESPONSE",
										"data":      rawRespData,
									}
									orderRespJSON, _ := json.Marshal([]interface{}{orderRespLog})

									// Build provider response JSON
									providerRespJSON, _ := json.Marshal(map[string]interface{}{
										"ref_id":          orderResp.RefID,
										"provider_ref_id": orderResp.ProviderRefID,
										"status":          orderResp.Status,
										"message":         orderResp.Message,
										"sn":              orderResp.SN,
									})

									_, err = deps.DB.Pool.Exec(context.Background(), `
										UPDATE transactions
										SET status = $1::transaction_status,
											provider_ref_id = $2,
											provider_serial_number = $3,
											provider_response = $4,
											provider_logs = COALESCE(provider_logs, '[]'::jsonb) || $5::jsonb,
											completed_at = CASE WHEN $1::text = 'SUCCESS' THEN NOW() ELSE completed_at END,
											updated_at = NOW()
										WHERE id = $6
									`, updateStatus, orderResp.ProviderRefID, orderResp.SN, string(providerRespJSON), string(orderRespJSON), txID)

									if err != nil {
										log.Error().
											Err(err).
											Str("invoice_number", invNum).
											Msg("Failed to update transaction with provider response")
									} else {
										// Add timeline entry: Item has been successfully sent or failed to sent (only for final status)
										if updateStatus == "SUCCESS" || updateStatus == "FAILED" {
											var finalMessage string
											if updateStatus == "SUCCESS" {
												finalMessage = "Item has been successfully sent."
											} else {
												finalMessage = "Item has been failed to sent."
											}
											_, _ = deps.DB.Pool.Exec(context.Background(), `
												INSERT INTO transaction_logs (transaction_id, status, message, created_at)
												VALUES ($1, $2, $3, NOW())
											`, txID, updateStatus, finalMessage)
										}

										log.Info().
											Str("invoice_number", invNum).
											Str("provider", provCode).
											Str("status", updateStatus).
											Str("serial_number", orderResp.SN).
											Msg("PakaiLink transaction processed to provider successfully")
									}
								}(invoiceNumber, transactionID, customerNo, providerCode, skuCode, productName, skuName)
							} else {
								log.Error().
									Err(err).
									Str("invoice_number", invoiceNumber).
									Str("provider_code", providerCode).
									Msg("Provider not found")
							}
						}
					}
				}
			}

			log.Info().
				Str("invoice_number", invoiceNumber).
				Msg("Transaction updated from PakaiLink webhook")
		} else {
			log.Info().
				Str("invoice_number", invoiceNumber).
				Msg("No transaction updated (already processed or not found)")
		}

		// Send success response
		sendPakaiLinkResponse(w, "2002800", "Successful")
	}
}

// handleDepositPaymentCallback handles payment callbacks for deposits
// This function is called when a payment gateway sends a callback for a deposit transaction
func handleDepositPaymentCallback(
	ctx context.Context,
	deps *Dependencies,
	invoiceNumber string,
	notification interface{},
	body []byte,
	w http.ResponseWriter,
	gatewayName string,
) {
	// Parse the full notification for logging
	var fullNotification map[string]interface{}
	json.Unmarshal(body, &fullNotification)

	// Check current deposit status
	var depositID, status string
	var userID string
	var amount, totalAmount int64
	var currency string

	err := deps.DB.Pool.QueryRow(ctx, `
		SELECT id, status, user_id, amount, total_amount, currency
		FROM deposits
		WHERE invoice_number = $1
	`, invoiceNumber).Scan(&depositID, &status, &userID, &amount, &totalAmount, &currency)

	if err != nil {
		log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Deposit not found for payment callback")
		// Return success to gateway even if not found (idempotency)
		if gatewayName == "DANA" {
			sendDANAResponse(w, "2005600", "Successful")
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}

	// Idempotency: If already paid/success, return success immediately
	if status == "SUCCESS" || status == "PAID" {
		log.Info().
			Str("invoice", invoiceNumber).
			Str("status", status).
			Msg("Deposit already paid/success, ignoring callback")
		if gatewayName == "DANA" {
			sendDANAResponse(w, "2005600", "Successful")
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}

	// Determine payment status based on gateway and notification
	paidAt := time.Now()
	paymentStatus := "PAID"
	depositStatus := "SUCCESS"

	// Parse status based on gateway type using fullNotification map
	switch gatewayName {
	case "DANA":
		// DANA: latestTransactionStatus == "00" means success
		if status, ok := fullNotification["latestTransactionStatus"].(string); ok {
			if status == "00" {
				depositStatus = "SUCCESS"
				paymentStatus = "PAID"
				// Parse paid time if available
				if addInfo, ok := fullNotification["additionalInfo"].(map[string]interface{}); ok {
					if paidTimeStr, ok := addInfo["paidTime"].(string); ok && paidTimeStr != "" {
						if parsedTime, err := time.Parse(time.RFC3339, paidTimeStr); err == nil {
							paidAt = parsedTime
						}
					}
				}
			} else {
				depositStatus = "FAILED"
				paymentStatus = "FAILED"
			}
		}
	case "MIDTRANS":
		// Midtrans: transaction_status == "settlement" means success
		if status, ok := fullNotification["transaction_status"].(string); ok {
			if status == "settlement" {
				depositStatus = "SUCCESS"
				paymentStatus = "PAID"
				// Parse settlement time if available
				if settlementTime, ok := fullNotification["settlement_time"].(string); ok && settlementTime != "" {
					if parsedTime, err := time.Parse("2006-01-02 15:04:05", settlementTime); err == nil {
						paidAt = parsedTime
					}
				}
			} else if status == "expire" {
				depositStatus = "EXPIRED"
				paymentStatus = "EXPIRED"
			} else if status == "deny" || status == "cancel" {
				depositStatus = "FAILED"
				paymentStatus = "FAILED"
			}
		}
	case "XENDIT":
		// Xendit: event == "payment.capture" && data.status == "SUCCEEDED" means success
		if event, ok := fullNotification["event"].(string); ok {
			if event == "payment.capture" {
				if data, ok := fullNotification["data"].(map[string]interface{}); ok {
					if status, ok := data["status"].(string); ok && status == "SUCCEEDED" {
						depositStatus = "SUCCESS"
						paymentStatus = "PAID"
					}
				}
			} else if event == "payment.failure" {
				depositStatus = "FAILED"
				paymentStatus = "FAILED"
			}
		}
	case "PAKAILINK":
		// PakaiLink: transactionData.transactionStatus == "SUCCESS" means success
		if txData, ok := fullNotification["transactionData"].(map[string]interface{}); ok {
			if status, ok := txData["transactionStatus"].(string); ok {
				if status == "SUCCESS" {
					depositStatus = "SUCCESS"
					paymentStatus = "PAID"
				} else if status == "FAILED" {
					depositStatus = "FAILED"
					paymentStatus = "FAILED"
				} else if status == "EXPIRED" {
					depositStatus = "EXPIRED"
					paymentStatus = "EXPIRED"
				}
			}
		}
	default:
		// Default: assume success if callback received
		depositStatus = "SUCCESS"
		paymentStatus = "PAID"
	}

	// Create payment callback log entry
	paymentCallbackLog := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      "PAYMENT_CALLBACK",
		"data":      fullNotification,
	}

	// Get existing payment logs
	var existingLogs []interface{}
	var logsJSON []byte
	_ = deps.DB.Pool.QueryRow(ctx, `
		SELECT COALESCE(payment_logs, '[]'::jsonb) FROM deposits WHERE id = $1
	`, depositID).Scan(&logsJSON)
	json.Unmarshal(logsJSON, &existingLogs)
	existingLogs = append(existingLogs, paymentCallbackLog)
	updatedLogsJSON, _ := json.Marshal(existingLogs)

	// Begin transaction for atomic balance update and deposit status update
	// Timeline entry "Payment received" will be created inside transaction
	tx, err := deps.DB.Pool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Failed to begin transaction for deposit callback")
		if gatewayName == "DANA" {
			sendDANAResponse(w, "5005601", "Internal Server Error")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	defer tx.Rollback(ctx)

	// Create timeline entry: Payment received (before updating status)
	paymentReceivedMessage := fmt.Sprintf("Payment received via %s.", gatewayName)
	_, err = tx.Exec(ctx, `
		INSERT INTO deposit_logs (deposit_id, status, message, created_at)
		VALUES ($1, $2, $3, NOW())
	`, depositID, "PAYMENT", paymentReceivedMessage)

	if err != nil {
		log.Warn().
			Err(err).
			Str("invoice", invoiceNumber).
			Str("deposit_id", depositID).
			Msg("Failed to insert payment received timeline (non-fatal)")
	}

	// Only update balance if deposit status is SUCCESS
	var currentBalance, newBalance int64
	if depositStatus == "SUCCESS" {
		// Determine balance column based on currency
		balanceColumn := "balance_idr"
		if currency == "MYR" {
			balanceColumn = "balance_myr"
		} else if currency == "PHP" {
			balanceColumn = "balance_php"
		} else if currency == "SGD" {
			balanceColumn = "balance_sgd"
		} else if currency == "THB" {
			balanceColumn = "balance_thb"
		}

		// Get current balance
		err = tx.QueryRow(ctx, "SELECT "+balanceColumn+" FROM users WHERE id = $1", userID).Scan(&currentBalance)
		if err != nil {
			log.Error().Err(err).Str("invoice", invoiceNumber).Str("user_id", userID).Msg("Failed to get current balance")
			tx.Rollback(ctx)
			if gatewayName == "DANA" {
				sendDANAResponse(w, "5005601", "Internal Server Error")
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		// Calculate new balance (add deposit amount, not totalAmount to avoid double counting payment fee)
		newBalance = currentBalance + amount

		// Update user balance
		_, err = tx.Exec(ctx, "UPDATE users SET "+balanceColumn+" = $1 WHERE id = $2", newBalance, userID)
		if err != nil {
			log.Error().Err(err).Str("invoice", invoiceNumber).Str("user_id", userID).Msg("Failed to update user balance")
			tx.Rollback(ctx)
			if gatewayName == "DANA" {
				sendDANAResponse(w, "5005601", "Internal Server Error")
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		// Get payment channel name for mutation description
		var paymentChannelName sql.NullString
		_ = tx.QueryRow(ctx, `
			SELECT pc.name
			FROM deposits d
			LEFT JOIN payment_channels pc ON d.payment_channel_id = pc.id
			WHERE d.id = $1
		`, depositID).Scan(&paymentChannelName)

		paymentName := "Payment Gateway"
		if paymentChannelName.Valid && paymentChannelName.String != "" {
			paymentName = paymentChannelName.String
		}

		// Create balance mutation record
		mutationDesc := fmt.Sprintf("Isi Ulang Saldo via %s", paymentName)
		_, err = tx.Exec(ctx, `
			INSERT INTO mutations (
				user_id, invoice_number, mutation_type, amount,
				balance_before, balance_after, description,
				reference_type, reference_id, currency, created_at
			) VALUES ($1, $2, 'CREDIT', $3, $4, $5, $6, 'DEPOSIT', $7, $8, NOW())
		`, userID, invoiceNumber, amount, currentBalance, newBalance, mutationDesc, depositID, currency)

		if err != nil {
			log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Failed to create mutation record")
			tx.Rollback(ctx)
			if gatewayName == "DANA" {
				sendDANAResponse(w, "5005601", "Internal Server Error")
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		// Create timeline entry: Deposit successful (after balance updated)
		_, err = tx.Exec(ctx, `
			INSERT INTO deposit_logs (deposit_id, status, message, created_at)
			VALUES ($1, $2, $3, NOW())
		`, depositID, "SUCCESS", "Deposit successful, balance updated")

		if err != nil {
			log.Warn().
				Err(err).
				Str("invoice", invoiceNumber).
				Str("deposit_id", depositID).
				Msg("Failed to insert success timeline (non-fatal)")
		}
	}

	// Update deposit status and payment logs
	_, err = tx.Exec(ctx, `
		UPDATE deposits
		SET status = $1, paid_at = $2,
		    payment_logs = $3::jsonb,
		    updated_at = NOW()
		WHERE id = $4
	`, depositStatus, paidAt, string(updatedLogsJSON), depositID)

	if err != nil {
		log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Failed to update deposit status")
		tx.Rollback(ctx)
		if gatewayName == "DANA" {
			sendDANAResponse(w, "5005601", "Internal Server Error")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Update payment_data table if exists
	rawCallbackJSON, _ := json.Marshal(fullNotification)
	_, _ = tx.Exec(ctx, `
		UPDATE payment_data 
		SET status = $1::payment_status, paid_at = $2, 
		    raw_response = COALESCE(raw_response, '{}'::jsonb) || jsonb_build_object('callback', $3::jsonb), 
		    updated_at = NOW()
		WHERE invoice_number = $4
	`, paymentStatus, paidAt, string(rawCallbackJSON), invoiceNumber)

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Failed to commit deposit callback transaction")
		if gatewayName == "DANA" {
			sendDANAResponse(w, "5005601", "Internal Server Error")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if depositStatus == "SUCCESS" {
		log.Info().
			Str("invoice", invoiceNumber).
			Str("status", depositStatus).
			Str("gateway", gatewayName).
			Str("user_id", userID).
			Int64("amount", amount).
			Int64("balance_before", currentBalance).
			Int64("balance_after", newBalance).
			Msg("Deposit payment callback processed successfully, balance updated")
	} else {
		log.Info().
			Str("invoice", invoiceNumber).
			Str("status", depositStatus).
			Str("gateway", gatewayName).
			Msg("Deposit payment callback processed (status not SUCCESS, balance not updated)")
	}

	// Send success response
	if gatewayName == "DANA" {
		sendDANAResponse(w, "2005600", "Successful")
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func sendPakaiLinkResponse(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"responseCode":    code,
		"responseMessage": message,
	})
}
