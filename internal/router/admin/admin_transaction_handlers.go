package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"seaply/internal/middleware"
	"seaply/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// ============================================
// ADMIN TRANSACTION MANAGEMENT
// ============================================

// handleAdminGetTransactionsImpl returns all transactions with filters
func HandleAdminGetTransactionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		search := r.URL.Query().Get("search")
		status := r.URL.Query().Get("status")
		paymentStatus := r.URL.Query().Get("paymentStatus")
		productCode := r.URL.Query().Get("productCode")
		providerCode := r.URL.Query().Get("providerCode")
		paymentCode := r.URL.Query().Get("paymentCode")
		region := r.URL.Query().Get("region")
		userID := r.URL.Query().Get("userId")
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Get overview stats
		var totalTransactions, successCount, processingCount, pendingCount, failedCount int
		var totalRevenue, totalProfit int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN status = 'PROCESSING' THEN 1 ELSE 0 END), 0) as processing,
				COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status = 'FAILED' OR payment_status IN ('EXPIRED', 'REFUNDED', 'FAILED') THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN total_amount ELSE 0 END), 0) as revenue,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN (sell_price - buy_price) ELSE 0 END), 0) as profit
			FROM transactions
		`).Scan(&totalTransactions, &successCount, &processingCount, &pendingCount, &failedCount, &totalRevenue, &totalProfit)

		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Msg("Failed to get transaction overview stats")
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for transactions list
		query := `
			SELECT
				t.id, t.invoice_number, t.status, t.payment_status,
				u.id as user_id, u.first_name || ' ' || u.last_name as user_name, u.email as user_email,
				p.code as product_code, p.title as product_name,
				s.code as sku_code, s.name as sku_name,
				pr.code as provider_code, pr.name as provider_name, t.provider_ref_id,
				t.account_nickname, t.account_inputs,
				t.buy_price, t.sell_price, COALESCE(t.discount_amount, 0), COALESCE(t.payment_fee, 0), t.total_amount, t.currency,
				pc.code as payment_code, pc.name as payment_name, pc.gateway_code,
				t.promo_code, COALESCE(t.discount_amount, 0) as promo_discount,
				t.region, t.ip_address, t.user_agent,
				t.created_at, t.paid_at, t.completed_at
			FROM transactions t
			LEFT JOIN users u ON t.user_id = u.id
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			LEFT JOIN providers pr ON t.provider_id = pr.id
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			WHERE 1=1
		`

		args := []interface{}{}
		argCount := 0

		// Add search filter
		if search != "" {
			argCount++
			query += " AND (t.invoice_number ILIKE $" + strconv.Itoa(argCount)
			query += " OR u.email ILIKE $" + strconv.Itoa(argCount)
			query += " OR t.contact_email ILIKE $" + strconv.Itoa(argCount)
			query += " OR t.contact_phone ILIKE $" + strconv.Itoa(argCount) + ")"
			args = append(args, "%"+search+"%")
		}

		// Add status filter
		if status != "" {
			argCount++
			query += " AND t.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

		// Add payment status filter
		if paymentStatus != "" {
			argCount++
			query += " AND t.payment_status = $" + strconv.Itoa(argCount)
			args = append(args, paymentStatus)
		}

		// Add product filter
		if productCode != "" {
			argCount++
			query += " AND p.code = $" + strconv.Itoa(argCount)
			args = append(args, productCode)
		}

		// Add provider filter
		if providerCode != "" {
			argCount++
			query += " AND pr.code = $" + strconv.Itoa(argCount)
			args = append(args, providerCode)
		}

		// Add payment filter
		if paymentCode != "" {
			argCount++
			query += " AND pc.code = $" + strconv.Itoa(argCount)
			args = append(args, paymentCode)
		}

		// Add region filter
		if region != "" {
			argCount++
			query += " AND t.region = $" + strconv.Itoa(argCount)
			args = append(args, region)
		}

		// Add user filter
		if userID != "" {
			argCount++
			query += " AND t.user_id = $" + strconv.Itoa(argCount)
			args = append(args, userID)
		}

		// Add date filters
		if startDate != "" {
			argCount++
			query += " AND t.created_at >= $" + strconv.Itoa(argCount)
			args = append(args, startDate)
		}

		if endDate != "" {
			argCount++
			query += " AND t.created_at <= $" + strconv.Itoa(argCount)
			args = append(args, endDate)
		}

		query += " ORDER BY t.created_at DESC"
		argCount++
		query += " LIMIT $" + strconv.Itoa(argCount)
		args = append(args, limit)

		argCount++
		query += " OFFSET $" + strconv.Itoa(argCount)
		args = append(args, offset)

		// Execute query
		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			log.Error().Err(err).Str("query", query).Msg("Failed to query transactions")
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		transactions := []map[string]interface{}{}
		for rows.Next() {
			var id, invoiceNumber, txStatus, txPaymentStatus string
			var userIDVal, userName, userEmail sql.NullString
			var productCodeVal, productName, skuCodeVal, skuName string
			var providerCodeVal, providerName, providerRefID sql.NullString
			var accountNickname sql.NullString
			var accountInputs []byte
			var buyPrice, sellPrice, discountAmount, paymentFee, totalAmount int64
			var currency string
			var paymentCodeVal, paymentName string
			var gatewayCode sql.NullString
			var promoCode sql.NullString
			var promoDiscount int64
			var regionVal string
			var ipAddress, userAgent sql.NullString
			var createdAt time.Time
			var paidAt, completedAt *time.Time

			err := rows.Scan(
				&id, &invoiceNumber, &txStatus, &txPaymentStatus,
				&userIDVal, &userName, &userEmail,
				&productCodeVal, &productName, &skuCodeVal, &skuName,
				&providerCodeVal, &providerName, &providerRefID,
				&accountNickname, &accountInputs,
				&buyPrice, &sellPrice, &discountAmount, &paymentFee, &totalAmount, &currency,
				&paymentCodeVal, &paymentName, &gatewayCode,
				&promoCode, &promoDiscount,
				&regionVal, &ipAddress, &userAgent,
				&createdAt, &paidAt, &completedAt,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan transaction row")
				continue
			}

			// Parse account inputs and format as string
			var accountData map[string]interface{}
			json.Unmarshal(accountInputs, &accountData)

			// Format account inputs as string like "656696292 - 8610"
			var accountInputsStr string
			if accountData != nil {
				var parts []string
				if uid, ok := accountData["userId"]; ok {
					parts = append(parts, formatAccountValue(uid))
				}
				if zid, ok := accountData["zoneId"]; ok {
					parts = append(parts, formatAccountValue(zid))
				}
				if sid, ok := accountData["serverId"]; ok {
					parts = append(parts, formatAccountValue(sid))
				}
				if len(parts) > 0 {
					accountInputsStr = joinStrings(parts, " - ")
				}
			}

			transaction := map[string]interface{}{
				"id":            id,
				"invoiceNumber": invoiceNumber,
				"status":        txStatus,
				"paymentStatus": txPaymentStatus,
				"product": map[string]interface{}{
					"code": productCodeVal,
					"name": productName,
				},
				"sku": map[string]interface{}{
					"code": skuCodeVal,
					"name": skuName,
				},
				"account": map[string]interface{}{
					"nickname": accountNickname.String,
					"inputs":   accountInputsStr,
				},
				"pricing": map[string]interface{}{
					"buyPrice":   buyPrice,
					"sellPrice":  sellPrice,
					"discount":   discountAmount,
					"paymentFee": paymentFee,
					"total":      totalAmount,
					"profit":     sellPrice - buyPrice,
					"currency":   currency,
				},
				"payment": map[string]interface{}{
					"code":    paymentCodeVal,
					"name":    paymentName,
					"gateway": gatewayCode.String,
				},
				"region":    regionVal,
				"createdAt": createdAt.Format(time.RFC3339),
			}

			// Add provider info
			if providerCodeVal.Valid {
				transaction["provider"] = map[string]interface{}{
					"code":  providerCodeVal.String,
					"name":  providerName.String,
					"refId": providerRefID.String,
				}
			} else {
				transaction["provider"] = nil
			}

			// Add user info only if exists
			if userIDVal.Valid {
				transaction["user"] = map[string]interface{}{
					"id":    userIDVal.String,
					"name":  userName.String,
					"email": userEmail.String,
				}
			} else {
				transaction["user"] = nil
			}

			// Add promo info if exists
			if promoCode.Valid && promoCode.String != "" {
				transaction["promo"] = map[string]interface{}{
					"code":           promoCode.String,
					"discountAmount": promoDiscount,
				}
			}

			// Add optional fields
			if ipAddress.Valid {
				transaction["ipAddress"] = ipAddress.String
			}
			if userAgent.Valid {
				transaction["userAgent"] = userAgent.String
			}
			if paidAt != nil {
				// Add paidAt to payment object
				paymentMap := transaction["payment"].(map[string]interface{})
				paymentMap["paidAt"] = (*paidAt).Format(time.RFC3339)
			}
			if completedAt != nil {
				transaction["completedAt"] = (*completedAt).Format(time.RFC3339)
			}

			transactions = append(transactions, transaction)
		}

		// Get total count for pagination
		var totalRows int
		countQuery := `SELECT COUNT(*) FROM transactions t 
			LEFT JOIN users u ON t.user_id = u.id 
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			LEFT JOIN providers pr ON t.provider_id = pr.id
			JOIN payment_channels pc ON t.payment_channel_id = pc.id 
			WHERE 1=1`
		countArgs := []interface{}{}
		countArgCount := 0

		if search != "" {
			countArgCount++
			countQuery += " AND (t.invoice_number ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR u.email ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR t.contact_email ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR t.contact_phone ILIKE $" + strconv.Itoa(countArgCount) + ")"
			countArgs = append(countArgs, "%"+search+"%")
		}

		if status != "" {
			countArgCount++
			countQuery += " AND t.status = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, status)
		}

		if paymentStatus != "" {
			countArgCount++
			countQuery += " AND t.payment_status = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, paymentStatus)
		}

		if productCode != "" {
			countArgCount++
			countQuery += " AND p.code = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, productCode)
		}

		if providerCode != "" {
			countArgCount++
			countQuery += " AND pr.code = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, providerCode)
		}

		if paymentCode != "" {
			countArgCount++
			countQuery += " AND pc.code = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, paymentCode)
		}

		if region != "" {
			countArgCount++
			countQuery += " AND t.region = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, region)
		}

		if userID != "" {
			countArgCount++
			countQuery += " AND t.user_id = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, userID)
		}

		if startDate != "" {
			countArgCount++
			countQuery += " AND t.created_at >= $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, startDate)
		}

		if endDate != "" {
			countArgCount++
			countQuery += " AND t.created_at <= $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, endDate)
		}

		deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalTransactions": totalTransactions,
				"totalRevenue":      totalRevenue,
				"totalProfit":       totalProfit,
				"successCount":      successCount,
				"processingCount":   processingCount,
				"pendingCount":      pendingCount,
				"failedCount":       failedCount,
			},
			"transactions": transactions,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

// Helper functions for transaction handlers
func formatAccountValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

// handleAdminGetTransactionImpl returns detailed transaction info
func HandleAdminGetTransactionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionID := chi.URLParam(r, "transactionId")
		if transactionID == "" {
			utils.WriteBadRequestError(w, "Transaction ID is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get transaction details
		var id, invoiceNumber, txStatus, paymentStatus string
		var userID, userName, userEmail sql.NullString
		var productCode, productName, skuCode, skuName string
		var quantity int
		var buyPrice, sellPrice, discountAmount, paymentFee, totalAmount int64
		var currency, paymentCode, paymentName string
		var gatewayCode sql.NullString
		var providerCode, providerName, providerRefID, serialNumber sql.NullString
		var accountNickname, contactEmail, contactPhone sql.NullString
		var promoCode sql.NullString
		var regionVal string
		var ipAddress, userAgent sql.NullString
		var accountInputs []byte
		var paymentGatewayRefID sql.NullString
		var paymentLogs, providerLogs []byte
		var createdAt time.Time
		var expiredAt sql.NullTime
		var paidAt, completedAt *time.Time

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				t.id, t.invoice_number, t.status, t.payment_status,
				u.id, u.first_name || ' ' || u.last_name, u.email,
				p.code, p.title, s.code, s.name,
				t.quantity, t.buy_price, t.sell_price,
				COALESCE(t.discount_amount, 0), COALESCE(t.payment_fee, 0), t.total_amount, t.currency,
				pc.code, pc.name, pc.gateway_code,
				pr.code, pr.name, t.provider_ref_id, t.provider_serial_number,
				t.account_nickname, t.contact_email, t.contact_phone,
				t.promo_code,
				t.region, t.ip_address, t.user_agent,
				t.account_inputs,
				t.payment_gateway_ref_id,
				COALESCE(t.payment_logs, '[]'::jsonb), COALESCE(t.provider_logs, '[]'::jsonb),
				t.created_at, t.paid_at, t.completed_at, t.expired_at
			FROM transactions t
			LEFT JOIN users u ON t.user_id = u.id
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			LEFT JOIN providers pr ON t.provider_id = pr.id
			WHERE t.id::text = $1 OR t.invoice_number = $1
		`, transactionID).Scan(
			&id, &invoiceNumber, &txStatus, &paymentStatus,
			&userID, &userName, &userEmail,
			&productCode, &productName, &skuCode, &skuName,
			&quantity, &buyPrice, &sellPrice,
			&discountAmount, &paymentFee, &totalAmount, &currency,
			&paymentCode, &paymentName, &gatewayCode,
			&providerCode, &providerName, &providerRefID, &serialNumber,
			&accountNickname, &contactEmail, &contactPhone,
			&promoCode,
			&regionVal, &ipAddress, &userAgent,
			&accountInputs,
			&paymentGatewayRefID,
			&paymentLogs, &providerLogs,
			&createdAt, &paidAt, &completedAt, &expiredAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND",
					"Transaction not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Parse JSON fields
		var accountData map[string]interface{}
		json.Unmarshal(accountInputs, &accountData)

		// Format account inputs as string
		var accountInputsStr string
		if accountData != nil {
			var parts []string
			if uid, ok := accountData["userId"]; ok {
				parts = append(parts, formatAccountValue(uid))
			}
			if zid, ok := accountData["zoneId"]; ok {
				parts = append(parts, formatAccountValue(zid))
			}
			if sid, ok := accountData["serverId"]; ok {
				parts = append(parts, formatAccountValue(sid))
			}
			if len(parts) > 0 {
				accountInputsStr = joinStrings(parts, " - ")
			}
		}

		// Get timeline
		timelineRows, _ := deps.DB.Pool.Query(ctx, `
			SELECT status, message, created_at
			FROM transaction_logs
			WHERE transaction_id = $1
			ORDER BY created_at ASC
		`, id)

		timeline := []map[string]interface{}{}
		if timelineRows != nil {
			defer timelineRows.Close()
			for timelineRows.Next() {
				var tlStatus, tlMessage string
				var tlCreatedAt time.Time
				timelineRows.Scan(&tlStatus, &tlMessage, &tlCreatedAt)
				timeline = append(timeline, map[string]interface{}{
					"status":    tlStatus,
					"message":   tlMessage,
					"timestamp": tlCreatedAt.Format(time.RFC3339),
				})
			}
		}

		// Build response
		response := map[string]interface{}{
			"id":            id,
			"invoiceNumber": invoiceNumber,
			"status":        txStatus,
			"paymentStatus": paymentStatus,
			"product": map[string]interface{}{
				"code": productCode,
				"name": productName,
			},
			"sku": map[string]interface{}{
				"code": skuCode,
				"name": skuName,
			},
			"account": map[string]interface{}{
				"nickname": accountNickname.String,
				"inputs":   accountInputsStr,
			},
			"pricing": map[string]interface{}{
				"buyPrice":   buyPrice,
				"sellPrice":  sellPrice,
				"discount":   discountAmount,
				"paymentFee": paymentFee,
				"total":      totalAmount,
				"profit":     sellPrice - buyPrice,
				"currency":   currency,
			},
			"timeline":  timeline,
			"region":    regionVal,
			"createdAt": createdAt.Format(time.RFC3339),
		}

		// Parse logs as JSON arrays
		var paymentLogsData []interface{}
		var providerLogsData []interface{}
		json.Unmarshal(paymentLogs, &paymentLogsData)
		json.Unmarshal(providerLogs, &providerLogsData)

		// Add provider info with logs
		if providerCode.Valid {
			response["provider"] = map[string]interface{}{
				"code":         providerCode.String,
				"name":         providerName.String,
				"refId":        providerRefID.String,
				"serialNumber": serialNumber.String,
				"logs":         providerLogsData,
			}
		} else {
			response["provider"] = map[string]interface{}{
				"logs": providerLogsData,
			}
		}

		// Add user info only if exists
		if userID.Valid {
			response["user"] = map[string]interface{}{
				"id":    userID.String,
				"name":  userName.String,
				"email": userEmail.String,
			}
		} else {
			response["user"] = nil
		}

		// Build payment info with logs
		paymentObj := map[string]interface{}{
			"code":    paymentCode,
			"name":    paymentName,
			"gateway": gatewayCode.String,
			"logs":    paymentLogsData,
		}
		if paymentGatewayRefID.Valid {
			paymentObj["gatewayRefId"] = paymentGatewayRefID.String
		}
		if paidAt != nil {
			paymentObj["paidAt"] = (*paidAt).Format(time.RFC3339)
		}
		response["payment"] = paymentObj

		// Add promo info if exists
		if promoCode.Valid && promoCode.String != "" {
			response["promo"] = map[string]interface{}{
				"code":           promoCode.String,
				"discountAmount": discountAmount,
			}
		}

		// Add optional fields
		if ipAddress.Valid {
			response["ipAddress"] = ipAddress.String
		}
		if userAgent.Valid {
			response["userAgent"] = userAgent.String
		}
		if completedAt != nil {
			response["completedAt"] = (*completedAt).Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, response)
	}
}

// UpdateTransactionStatusRequest represents the request to update transaction status
type UpdateTransactionStatusRequest struct {
	Status       string `json:"status" validate:"required"`
	Reason       string `json:"reason"`
	SerialNumber string `json:"serialNumber"`
}

// handleUpdateTransactionStatusImpl updates transaction status
func HandleUpdateTransactionStatusImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionID := chi.URLParam(r, "transactionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req UpdateTransactionStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.Status == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"status": "Status is required",
			})
			return
		}

		// Validate status value - only valid TransactionStatus values
		validStatuses := []string{"PENDING", "PROCESSING", "SUCCESS", "FAILED"}
		isValid := false
		for _, s := range validStatuses {
			if req.Status == s {
				isValid = true
				break
			}
		}
		if !isValid {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"status": "Invalid status value. Must be one of: PENDING, PROCESSING, SUCCESS, FAILED",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Build update query
		updateQuery := `UPDATE transactions SET status = $1, updated_at = NOW()`
		args := []interface{}{req.Status}
		argCount := 1

		// Add serial number if provided
		if req.SerialNumber != "" {
			argCount++
			updateQuery += `, provider_serial_number = $` + strconv.Itoa(argCount)
			args = append(args, req.SerialNumber)
		}

		// Add completed_at if status is SUCCESS
		if req.Status == "SUCCESS" {
			updateQuery += `, completed_at = NOW()`
		}

		argCount++
		updateQuery += ` WHERE id = $` + strconv.Itoa(argCount)
		args = append(args, transactionID)

		_, err = tx.Exec(ctx, updateQuery, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add log entry
		message := "Status updated to " + req.Status
		if req.Reason != "" {
			message += ": " + req.Reason
		}
		if req.SerialNumber != "" {
			message += " (SN: " + req.SerialNumber + ")"
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, $2, $3, NOW())
		`, transactionID, req.Status, message)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		_, _ = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, "Updated transaction status to "+req.Status)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":      "Transaction status updated successfully",
			"status":       req.Status,
			"serialNumber": req.SerialNumber,
		})
	}
}

// RefundTransactionRequest represents the request to refund a transaction
type RefundTransactionRequest struct {
	Reason   string `json:"reason" validate:"required"`
	RefundTo string `json:"refundTo"` // BALANCE or ORIGINAL
	Amount   *int64 `json:"amount"`   // Optional, defaults to total_amount
}

// handleRefundTransactionImpl processes transaction refund
func HandleRefundTransactionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionID := chi.URLParam(r, "transactionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req RefundTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.Reason == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"reason": "Reason is required",
			})
			return
		}

		// Default refundTo to BALANCE
		if req.RefundTo == "" {
			req.RefundTo = "BALANCE"
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Get transaction details
		var userID sql.NullString
		var invoiceNumber, currency, status string
		var totalAmount int64

		err = tx.QueryRow(ctx, `
			SELECT user_id, invoice_number, total_amount, currency, status
			FROM transactions
			WHERE id = $1
		`, transactionID).Scan(&userID, &invoiceNumber, &totalAmount, &currency, &status)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND",
					"Transaction not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get payment status
		var txPaymentStatus string
		_ = tx.QueryRow(ctx, `SELECT payment_status FROM transactions WHERE id = $1`, transactionID).Scan(&txPaymentStatus)

		// Validate transaction can be refunded
		// Payment status must be PAID
		if txPaymentStatus != "PAID" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "TRANSACTION_NOT_REFUNDABLE",
				"Transaction payment status must be PAID to be refunded", "")
			return
		}

		// Transaction status must be PROCESSING or FAILED
		if status != "PROCESSING" && status != "FAILED" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "TRANSACTION_NOT_REFUNDABLE",
				"Transaction status must be PROCESSING or FAILED to be refunded", "")
			return
		}

		// Determine refund amount
		refundAmount := totalAmount
		if req.Amount != nil && *req.Amount > 0 {
			if *req.Amount > totalAmount {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_AMOUNT",
					"Refund amount cannot exceed transaction total", "")
				return
			}
			refundAmount = *req.Amount
		}

		var refundID string
		refundStatus := "SUCCESS"

		// Refund to balance if user exists
		if req.RefundTo == "BALANCE" && userID.Valid {
			// Get current balance
			var currentBalance int64
			err = tx.QueryRow(ctx, `SELECT balance_idr FROM users WHERE id = $1`, userID.String).Scan(&currentBalance)
			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Add refund amount to user balance
			newBalance := currentBalance + refundAmount

			_, err = tx.Exec(ctx, `
				UPDATE users SET balance_idr = $1 WHERE id = $2
			`, newBalance, userID.String)

			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}

			// Create balance mutation record
			_, err = tx.Exec(ctx, `
				INSERT INTO mutations (
					user_id, invoice_number, mutation_type, amount, balance_before, balance_after,
					description, reference_type, reference_id, currency, created_at
				) VALUES ($1, $2, 'CREDIT', $3, $4, $5, $6, 'REFUND', $7, $8, NOW())
			`, userID.String, invoiceNumber, refundAmount, currentBalance, newBalance,
				"Pengembalian Dana - "+req.Reason, transactionID, currency)

			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		// Generate refund ID
		randomStr, _ := utils.GenerateRandomString(12)
		refundID = "ref_" + randomStr

		// Update transaction status to REFUNDED
		_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET status = 'REFUNDED', updated_at = NOW()
			WHERE id = $1
		`, transactionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add transaction log
		_, _ = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, 'REFUNDED', $2, NOW())
		`, transactionID, "Transaction refunded: "+req.Reason)

		// Create audit log
		_, _ = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'REFUND', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, "Refunded transaction: "+req.Reason)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Get admin name
		var adminName string
		_ = deps.DB.Pool.QueryRow(ctx, `SELECT first_name || ' ' || last_name FROM admins WHERE id = $1`, adminID).Scan(&adminName)
		if adminName == "" {
			adminName = "Admin"
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"refundId":      refundID,
			"transactionId": transactionID,
			"invoiceNumber": invoiceNumber,
			"amount":        refundAmount,
			"currency":      currency,
			"refundTo":      req.RefundTo,
			"status":        refundStatus,
			"reason":        req.Reason,
			"processedBy": map[string]interface{}{
				"id":   adminID,
				"name": adminName,
			},
			"createdAt": time.Now().Format(time.RFC3339),
		})
	}
}

// RetryTransactionRequest represents the request to retry a transaction
type RetryTransactionRequest struct {
	ProviderCode string `json:"providerCode"` // Optional, uses existing provider if empty
	Reason       string `json:"reason"`
}

// handleRetryTransactionImpl retries failed transaction with different provider
func HandleRetryTransactionImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionID := chi.URLParam(r, "transactionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req RetryTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var providerID, providerCode string

		if req.ProviderCode != "" {
			// Get new provider ID
			err := deps.DB.Pool.QueryRow(ctx, `
				SELECT id, code FROM providers WHERE code = $1 AND is_active = true
			`, req.ProviderCode).Scan(&providerID, &providerCode)

			if err != nil {
				if err == pgx.ErrNoRows {
					utils.WriteErrorJSON(w, http.StatusNotFound, "PROVIDER_NOT_FOUND",
						"Provider not found", "")
					return
				}
				utils.WriteInternalServerError(w)
				return
			}
		} else {
			// Use existing provider
			err := deps.DB.Pool.QueryRow(ctx, `
				SELECT t.provider_id, p.code 
				FROM transactions t
				JOIN providers p ON t.provider_id = p.id
				WHERE t.id = $1
			`, transactionID).Scan(&providerID, &providerCode)

			if err != nil {
				utils.WriteInternalServerError(w)
				return
			}
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Update transaction with provider and reset status
		_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET provider_id = $1, status = 'PROCESSING', updated_at = NOW()
			WHERE id = $2
		`, providerID, transactionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add transaction log
		message := "Retrying transaction with provider " + providerCode
		if req.Reason != "" {
			message += ": " + req.Reason
		}

		_, _ = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, 'PROCESSING', $2, NOW())
		`, transactionID, message)

		// Create audit log
		_, _ = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'RETRY', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, message)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// TODO: Trigger actual provider processing here
		// deps.ProviderManager.ProcessTransaction(transactionID)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":  "Transaction retry initiated",
			"provider": providerCode,
			"status":   "PROCESSING",
		})
	}
}

// ManualProcessRequest represents the request to manually process a transaction
type ManualProcessRequest struct {
	SerialNumber string `json:"serialNumber" validate:"required"`
	Reason       string `json:"reason" validate:"required"`
}

// handleManualProcessImpl manually marks transaction as complete
func HandleManualProcessImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionID := chi.URLParam(r, "transactionId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req ManualProcessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		if req.SerialNumber == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"serialNumber": "Serial number is required",
			})
			return
		}

		if req.Reason == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"reason": "Reason is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Update transaction to SUCCESS with serial number
		_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET status = 'SUCCESS', provider_serial_number = $1, completed_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`, req.SerialNumber, transactionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add transaction log
		logMessage := "Manually processed: " + req.Reason + " (SN: " + req.SerialNumber + ")"
		_, _ = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, 'SUCCESS', $2, NOW())
		`, transactionID, logMessage)

		// Create audit log
		_, _ = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'MANUAL_PROCESS', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, logMessage)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":      "Transaction manually processed successfully",
			"status":       "SUCCESS",
			"serialNumber": req.SerialNumber,
		})
	}
}
