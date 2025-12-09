package public

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"unsafe"

	"seaply/internal/domain"
	"seaply/internal/middleware"
	"seaply/internal/router/user"
	"seaply/internal/utils"

	"github.com/rs/zerolog/log"
)

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
		if req.PromoCode == "" || req.ProductCode == "" || req.PaymentCode == "" || req.Region == "" {
			utils.WriteBadRequestError(w, "Missing required fields: promoCode, productCode, paymentCode, region, amount")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get region from context or use provided region
		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = req.Region
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

		err := deps.DB.Pool.QueryRow(ctx, `
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
		var productCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM promo_products pp
			JOIN products p ON pp.product_id = p.id
			WHERE pp.promo_id = $1 AND p.code = $2
		`, promoID, req.ProductCode).Scan(&productCount)

		if err != nil || productCount == 0 {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "PRODUCT_NOT_APPLICABLE",
			})
			return
		}

		// Validation 5: Check if payment channel is in allowed payment channels
		var channelCount int
		err = deps.DB.Pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM promo_payment_channels ppc
			JOIN payment_channels pc ON ppc.channel_id = pc.id
			WHERE ppc.promo_id = $1 AND pc.code = $2
		`, promoID, req.PaymentCode).Scan(&channelCount)

		if err != nil || channelCount == 0 {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "PAYMENT_NOT_APPLICABLE",
			})
			return
		}

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
		if req.Amount < minAmount {
			utils.WriteSuccessJSON(w, map[string]interface{}{
				"valid":  false,
				"reason": "MIN_AMOUNT_NOT_MET",
			})
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
			discountAmount = (req.Amount * promoPercentage) / 100
		}

		// Cap discount by max promo amount
		if maxPromoAmount > 0 && discountAmount > maxPromoAmount {
			discountAmount = maxPromoAmount
		}

		// Fetch applicable products and payment channels for response
		var applicableProducts []string
		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT DISTINCT p.code
			FROM promo_products pp
			JOIN products p ON pp.product_id = p.id
			WHERE pp.promo_id = $1
			ORDER BY p.code
		`, promoID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var code string
				if err := rows.Scan(&code); err == nil {
					applicableProducts = append(applicableProducts, code)
				}
			}
		}

		var applicablePayments []string
		rows, err = deps.DB.Pool.Query(ctx, `
			SELECT DISTINCT pc.code
			FROM promo_payment_channels ppc
			JOIN payment_channels pc ON ppc.channel_id = pc.id
			WHERE ppc.promo_id = $1
			ORDER BY pc.code
		`, promoID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var code string
				if err := rows.Scan(&code); err == nil {
					applicablePayments = append(applicablePayments, code)
				}
			}
		}

		var applicableRegions []string
		rows, err = deps.DB.Pool.Query(ctx, `
			SELECT DISTINCT region_code
			FROM promo_regions
			WHERE promo_id = $1
			ORDER BY region_code
		`, promoID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var regionCode string
				if err := rows.Scan(&regionCode); err == nil {
					applicableRegions = append(applicableRegions, regionCode)
				}
			}
		}

		// Calculate remaining usages
		remaining := maxUsage - totalUsage
		if remaining < 0 {
			remaining = 0
		}

		// Build success response
		response := map[string]interface{}{
			"valid":          true,
			"promoCode":      promoCode,
			"title":          promoTitle,
			"description":    promoDesc,
			"discountAmount": math.Round(discountAmount*100) / 100,
			"requirements": map[string]interface{}{
				"minAmount":          minAmount,
				"applicableProducts": applicableProducts,
				"applicablePayments": applicablePayments,
				"applicableRegions":  applicableRegions,
				"validDays":          daysAvailable,
			},
			"usage": map[string]interface{}{
				"used":      totalUsage,
				"limit":     maxUsage,
				"remaining": remaining,
			},
		}

		utils.WriteSuccessJSON(w, response)
	}
}

func HandleGetInvoice(deps *Dependencies) http.HandlerFunc {
	return handleGetInvoiceImpl(deps)
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

func HandleVerifyMFA(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccessJSON(w, map[string]string{"step": "SUCCESS"})
	}
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

func HandleEnableMFA(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccessJSON(w, map[string]string{"step": "SETUP"})
	}
}

func HandleVerifyMFASetup(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccessJSON(w, map[string]string{"step": "SUCCESS"})
	}
}

func HandleDisableMFA(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteSuccessJSON(w, map[string]string{"message": "MFA disabled"})
	}
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

// Webhook Handlers
func HandleDigiflazzWebhook(_ *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement webhook signature validation and processing
		w.WriteHeader(http.StatusOK)
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

		if err := json.Unmarshal(body, &webhookEvent); err == nil && webhookEvent.Event != "" {
			// This is a Payment Requests API webhook
			invoiceNumber := webhookEvent.Data.ReferenceID
			if invoiceNumber == "" {
				log.Warn().Msg("Xendit webhook: missing reference_id")
				w.WriteHeader(http.StatusOK)
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

			switch webhookEvent.Event {
			case "payment.capture":
				if webhookEvent.Data.Status == "SUCCEEDED" {
					newStatus = "PROCESSING"
					newPaymentStatus = "PAID"
					timelineMessage = "Payment received via " + webhookEvent.Data.ChannelCode
				}
			case "payment.failure":
				newStatus = "FAILED"
				newPaymentStatus = "FAILED"
				timelineMessage = "Payment failed: " + webhookEvent.Data.FailureCode
			default:
				log.Info().
					Str("event", webhookEvent.Event).
					Msg("Xendit webhook: unhandled event, ignoring")
				w.WriteHeader(http.StatusOK)
				return
			}

			// Update transaction status
			result, err := deps.DB.Pool.Exec(ctx, `
				UPDATE transactions
				SET status = $1, payment_status = $2, updated_at = NOW()
				WHERE invoice_number = $3 AND status NOT IN ('SUCCESS', 'COMPLETED', 'FAILED')
			`, newStatus, newPaymentStatus, invoiceNumber)

			if err != nil {
				log.Error().
					Err(err).
					Str("invoice_number", invoiceNumber).
					Msg("Failed to update transaction from Xendit webhook")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if result.RowsAffected() > 0 {
				// Get transaction ID for timeline
				var transactionID string
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT id FROM transactions WHERE invoice_number = $1
				`, invoiceNumber).Scan(&transactionID)

				if err == nil && transactionID != "" {
					// Add timeline entry
					_, _ = deps.DB.Pool.Exec(ctx, `
						INSERT INTO transaction_logs (transaction_id, status, message, created_at)
						VALUES ($1, $2, $3, NOW())
					`, transactionID, newStatus, timelineMessage)
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

			// Update transaction status for VA payment
			result, err := deps.DB.Pool.Exec(ctx, `
				UPDATE transactions
				SET status = 'PROCESSING', payment_status = 'PAID', updated_at = NOW()
				WHERE invoice_number = $1 AND status NOT IN ('SUCCESS', 'COMPLETED', 'FAILED')
			`, invoiceNumber)

			if err != nil {
				log.Error().
					Err(err).
					Str("invoice_number", invoiceNumber).
					Msg("Failed to update transaction from Xendit VA callback")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if result.RowsAffected() > 0 {
				var transactionID string
				err = deps.DB.Pool.QueryRow(ctx, `
					SELECT id FROM transactions WHERE invoice_number = $1
				`, invoiceNumber).Scan(&transactionID)

				if err == nil && transactionID != "" {
					_, _ = deps.DB.Pool.Exec(ctx, `
						INSERT INTO transaction_logs (transaction_id, status, message, created_at)
						VALUES ($1, $2, $3, NOW())
					`, transactionID, "PROCESSING", "Payment received via "+vaCallback.BankCode)
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

		log.Info().
			Str("invoice_number", invoiceNumber).
			Str("transaction_id", notification.TransactionID).
			Str("status", notification.TransactionStatus).
			Str("gross_amount", notification.GrossAmount).
			Str("payment_type", notification.PaymentType).
			Msg("Processing Midtrans webhook")

		// Map Midtrans status to internal status
		var newStatus, newPaymentStatus string
		var timelineMessage string

		switch notification.TransactionStatus {
		case "capture", "settlement":
			newStatus = "PROCESSING"
			newPaymentStatus = "PAID"
			timelineMessage = "Payment received via " + notification.PaymentType
		case "pending":
			newStatus = "PENDING"
			newPaymentStatus = "UNPAID"
			timelineMessage = "Waiting for payment"
		case "deny", "cancel":
			newStatus = "CANCELLED"
			newPaymentStatus = "FAILED"
			timelineMessage = "Payment denied or cancelled"
		case "expire":
			newStatus = "EXPIRED"
			newPaymentStatus = "EXPIRED"
			timelineMessage = "Payment expired"
		default:
			log.Info().
				Str("status", notification.TransactionStatus).
				Msg("Midtrans webhook: unhandled status, ignoring")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Update transaction status
		result, err := deps.DB.Pool.Exec(ctx, `
			UPDATE transactions
			SET status = $1, payment_status = $2, updated_at = NOW()
			WHERE invoice_number = $3 AND status NOT IN ('SUCCESS', 'COMPLETED', 'FAILED')
		`, newStatus, newPaymentStatus, invoiceNumber)

		if err != nil {
			log.Error().
				Err(err).
				Str("invoice_number", invoiceNumber).
				Msg("Failed to update transaction from Midtrans webhook")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if result.RowsAffected() > 0 {
			// Get transaction ID for timeline
			var transactionID string
			err = deps.DB.Pool.QueryRow(ctx, `
				SELECT id FROM transactions WHERE invoice_number = $1
			`, invoiceNumber).Scan(&transactionID)

			if err == nil && transactionID != "" {
				// Add timeline entry
				_, _ = deps.DB.Pool.Exec(ctx, `
					INSERT INTO transaction_logs (transaction_id, status, message, created_at)
					VALUES ($1, $2, $3, NOW())
				`, transactionID, newStatus, timelineMessage)
			}

			log.Info().
				Str("invoice_number", invoiceNumber).
				Str("new_status", newStatus).
				Str("payment_status", newPaymentStatus).
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

func HandleDANAWebhook(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read DANA webhook body")
			w.WriteHeader(http.StatusBadRequest)
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Get invoice number (originalPartnerReferenceNo is our invoice number)
		invoiceNumber := notification.OriginalPartnerReferenceNo
		if invoiceNumber == "" {
			log.Warn().Msg("DANA webhook: missing originalPartnerReferenceNo")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Map DANA status to our status
		// 00 = SUCCESS, 01/02 = PENDING, 05 = FAILED
		var newStatus, newPaymentStatus string
		switch notification.LatestTransactionStatus {
		case "00":
			newStatus = "PAID"
			newPaymentStatus = "PAID"
		case "05":
			newStatus = "FAILED"
			newPaymentStatus = "FAILED"
		default:
			// Still pending, just acknowledge
			w.WriteHeader(http.StatusOK)
			return
		}

		// Update transaction status
		_, err = deps.DB.Pool.Exec(ctx, `
			UPDATE transactions
			SET status = $1, payment_status = $2, updated_at = NOW()
			WHERE invoice_number = $3 AND status = 'PENDING'
		`, newStatus, newPaymentStatus, invoiceNumber)

		if err != nil {
			log.Error().Err(err).Str("invoice", invoiceNumber).Msg("Failed to update transaction from DANA webhook")
		} else {
			log.Info().
				Str("invoice", invoiceNumber).
				Str("status", newStatus).
				Str("dana_status", notification.LatestTransactionStatus).
				Msg("Transaction updated from DANA webhook")

			// Add transaction log
			_, _ = deps.DB.Pool.Exec(ctx, `
				INSERT INTO transaction_logs (transaction_id, status, message, created_at)
				SELECT id, $2, $3, NOW()
				FROM transactions
				WHERE invoice_number = $1
			`, invoiceNumber, newStatus, "Payment status updated from DANA webhook")
		}

		w.WriteHeader(http.StatusOK)
	}
}
