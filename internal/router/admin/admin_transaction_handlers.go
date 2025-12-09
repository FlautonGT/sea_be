package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gate-v2/internal/middleware"
	"gate-v2/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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
		paymentCode := r.URL.Query().Get("paymentCode")
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Get overview stats
		var totalTransactions, successCount, pendingCount, failedCount int
		var totalRevenue, totalProfit int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN status IN ('PENDING', 'PAID', 'PROCESSING') THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status IN ('FAILED', 'EXPIRED', 'REFUNDED') THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN total_amount ELSE 0 END), 0) as revenue,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN (sell_price - buy_price) ELSE 0 END), 0) as profit
			FROM transactions
		`).Scan(&totalTransactions, &successCount, &pendingCount, &failedCount, &totalRevenue, &totalProfit)

		if err != nil && err != pgx.ErrNoRows {
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
				t.quantity, t.total_amount, t.currency,
				pc.code as payment_code, pc.name as payment_name,
				t.created_at, t.paid_at, t.completed_at
			FROM transactions t
			LEFT JOIN users u ON t.user_id = u.id
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
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
			query += " OR u.phone_number ILIKE $" + strconv.Itoa(argCount) + ")"
			args = append(args, "%"+search+"%")
		}

		// Add status filter
		if status != "" {
			argCount++
			query += " AND t.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

		// Add payment filter
		if paymentCode != "" {
			argCount++
			query += " AND pc.code = $" + strconv.Itoa(argCount)
			args = append(args, paymentCode)
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
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		transactions := []map[string]interface{}{}
		for rows.Next() {
			var id, invoiceNumber, status, paymentStatus, productCode, productName, skuCode, skuName string
			var userID, userName, userEmail, paymentCode, paymentName, currency string
			var quantity int
			var totalAmount int64
			var createdAt time.Time
			var paidAt, completedAt *time.Time

			err := rows.Scan(
				&id, &invoiceNumber, &status, &paymentStatus,
				&userID, &userName, &userEmail,
				&productCode, &productName, &skuCode, &skuName,
				&quantity, &totalAmount, &currency,
				&paymentCode, &paymentName,
				&createdAt, &paidAt, &completedAt,
			)
			if err != nil {
				continue
			}

			transaction := map[string]interface{}{
				"id":            id,
				"invoiceNumber": invoiceNumber,
				"status":        status,
				"paymentStatus": paymentStatus,
				"user": map[string]interface{}{
					"id":    userID,
					"name":  userName,
					"email": userEmail,
				},
				"product": map[string]interface{}{
					"code": productCode,
					"name": productName,
				},
				"sku": map[string]interface{}{
					"code": skuCode,
					"name": skuName,
				},
				"quantity": quantity,
				"total":    totalAmount,
				"currency": currency,
				"payment": map[string]interface{}{
					"code": paymentCode,
					"name": paymentName,
				},
				"createdAt": createdAt.Format(time.RFC3339),
			}

			if paidAt != nil {
				transaction["paidAt"] = (*paidAt).Format(time.RFC3339)
			}
			if completedAt != nil {
				transaction["completedAt"] = (*completedAt).Format(time.RFC3339)
			}

			transactions = append(transactions, transaction)
		}

		// Get total count for pagination
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM transactions t LEFT JOIN users u ON t.user_id = u.id JOIN payment_channels pc ON t.payment_channel_id = pc.id WHERE 1=1"
		countArgs := []interface{}{}
		countArgCount := 0

		if search != "" {
			countArgCount++
			countQuery += " AND (t.invoice_number ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR u.email ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR u.phone_number ILIKE $" + strconv.Itoa(countArgCount) + ")"
			countArgs = append(countArgs, "%"+search+"%")
		}

		if status != "" {
			countArgCount++
			countQuery += " AND t.status = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, status)
		}

		if paymentCode != "" {
			countArgCount++
			countQuery += " AND pc.code = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, paymentCode)
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
		var id, invoiceNumber, status, paymentStatus string
		var userID, userName, userEmail string
		var productCode, productName, skuCode, skuName string
		var quantity int
		var buyPrice, sellPrice, subtotal, discountAmount, paymentFee, totalAmount int64
		var currency, paymentCode, paymentName, providerCode, providerName string
		var serialNumber, accountNickname, contactEmail, contactPhone string
		var accountInputs, paymentData, providerResponse json.RawMessage
		var createdAt, expiredAt time.Time
		var paidAt, completedAt *time.Time

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				t.id, t.invoice_number, t.status, t.payment_status,
				u.id, u.first_name || ' ' || u.last_name, u.email,
				p.code, p.title, s.code, s.name,
				t.quantity, t.buy_price, t.sell_price, t.subtotal,
				t.discount_amount, t.payment_fee, t.total_amount, t.currency,
				pc.code, pc.name,
				pr.code, pr.name,
				t.serial_number, t.account_nickname,
				t.contact_email, t.contact_phone,
				t.account_inputs, t.payment_data, t.provider_response,
				t.created_at, t.paid_at, t.completed_at, t.expired_at
			FROM transactions t
			LEFT JOIN users u ON t.user_id = u.id
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			LEFT JOIN providers pr ON t.provider_id = pr.id
			WHERE t.id = $1
		`, transactionID).Scan(
			&id, &invoiceNumber, &status, &paymentStatus,
			&userID, &userName, &userEmail,
			&productCode, &productName, &skuCode, &skuName,
			&quantity, &buyPrice, &sellPrice, &subtotal,
			&discountAmount, &paymentFee, &totalAmount, &currency,
			&paymentCode, &paymentName,
			&providerCode, &providerName,
			&serialNumber, &accountNickname,
			&contactEmail, &contactPhone,
			&accountInputs, &paymentData, &providerResponse,
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
		var paymentInfo map[string]interface{}
		var providerData map[string]interface{}

		json.Unmarshal(accountInputs, &accountData)
		json.Unmarshal(paymentData, &paymentInfo)
		json.Unmarshal(providerResponse, &providerData)

		// Get timeline
		timelineRows, _ := deps.DB.Pool.Query(ctx, `
			SELECT status, message, created_at
			FROM transaction_logs
			WHERE transaction_id = $1
			ORDER BY created_at ASC
		`, transactionID)

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

		response := map[string]interface{}{
			"id":            id,
			"invoiceNumber": invoiceNumber,
			"status":        status,
			"paymentStatus": paymentStatus,
			"user": map[string]interface{}{
				"id":    userID,
				"name":  userName,
				"email": userEmail,
			},
			"product": map[string]interface{}{
				"code": productCode,
				"name": productName,
			},
			"sku": map[string]interface{}{
				"code": skuCode,
				"name": skuName,
			},
			"quantity": quantity,
			"pricing": map[string]interface{}{
				"buyPrice":       buyPrice,
				"sellPrice":      sellPrice,
				"subtotal":       subtotal,
				"discountAmount": discountAmount,
				"paymentFee":     paymentFee,
				"total":          totalAmount,
				"currency":       currency,
				"profit":         sellPrice - buyPrice,
			},
			"payment": map[string]interface{}{
				"code": paymentCode,
				"name": paymentName,
				"data": paymentInfo,
			},
			"provider": map[string]interface{}{
				"code":     providerCode,
				"name":     providerName,
				"response": providerData,
			},
			"account": map[string]interface{}{
				"nickname": accountNickname,
				"data":     accountData,
			},
			"contact": map[string]interface{}{
				"email": contactEmail,
				"phone": contactPhone,
			},
			"serialNumber": serialNumber,
			"timeline":     timeline,
			"createdAt":    createdAt.Format(time.RFC3339),
			"expiredAt":    expiredAt.Format(time.RFC3339),
		}

		if paidAt != nil {
			response["paidAt"] = (*paidAt).Format(time.RFC3339)
		}
		if completedAt != nil {
			response["completedAt"] = (*completedAt).Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, response)
	}
}

// UpdateTransactionStatusRequest represents the request to update transaction status
type UpdateTransactionStatusRequest struct {
	Status string `json:"status" validate:"required"`
	Reason string `json:"reason"`
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

		// Validate status value
		validStatuses := []string{"PENDING", "PAID", "PROCESSING", "SUCCESS", "FAILED", "EXPIRED", "REFUNDED"}
		isValid := false
		for _, s := range validStatuses {
			if req.Status == s {
				isValid = true
				break
			}
		}
		if !isValid {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"status": "Invalid status value",
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

		// Update transaction status
		_, err = tx.Exec(ctx, `
			UPDATE transactions
			SET status = $1, updated_at = NOW()
			WHERE id = $2
		`, req.Status, transactionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add log entry
		message := "Status updated to " + req.Status
		if req.Reason != "" {
			message += ": " + req.Reason
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
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, "Updated transaction status to "+req.Status)

		if err != nil {
			// Don't fail if audit log fails
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Transaction status updated successfully",
			"status":  req.Status,
		})
	}
}

// RefundTransactionRequest represents the request to refund a transaction
type RefundTransactionRequest struct {
	Reason string `json:"reason" validate:"required"`
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
		var userID, currency string
		var totalAmount int64
		var status string

		err = tx.QueryRow(ctx, `
			SELECT user_id, total_amount, currency, status
			FROM transactions
			WHERE id = $1
		`, transactionID).Scan(&userID, &totalAmount, &currency, &status)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND",
					"Transaction not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Validate transaction can be refunded
		if status != "SUCCESS" && status != "PAID" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "TRANSACTION_NOT_REFUNDABLE",
				"Transaction must be in SUCCESS or PAID status to be refunded", "")
			return
		}

		// Get current balance
		var currentBalance int64
		balanceColumn := "balance_" + currency
		if currency == "IDR" {
			balanceColumn = "balance_idr"
		}

		err = tx.QueryRow(ctx, `SELECT `+balanceColumn+` FROM users WHERE id = $1`, userID).Scan(&currentBalance)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add refund amount to user balance
		newBalance := currentBalance + totalAmount

		_, err = tx.Exec(ctx, `
			UPDATE users
			SET `+balanceColumn+` = $1
			WHERE id = $2
		`, newBalance, userID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create balance mutation record
		_, err = tx.Exec(ctx, `
			INSERT INTO balance_mutations (
				user_id, type, amount, balance_before, balance_after,
				description, reference_type, reference_id, currency, created_at
			) VALUES ($1, 'CREDIT', $2, $3, $4, $5, 'REFUND', $6, $7, NOW())
		`, userID, totalAmount, currentBalance, newBalance,
			"Refund for transaction: "+req.Reason, transactionID, currency)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

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
		_, err = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, 'REFUNDED', $2, NOW())
		`, transactionID, "Transaction refunded: "+req.Reason)

		// Create audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'REFUND', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, "Refunded transaction: "+req.Reason)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":       "Transaction refunded successfully",
			"refundAmount":  totalAmount,
			"balanceBefore": currentBalance,
			"balanceAfter":  newBalance,
		})
	}
}

// RetryTransactionRequest represents the request to retry a transaction
type RetryTransactionRequest struct {
	ProviderCode string `json:"providerCode" validate:"required"`
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

		// Get provider ID
		var providerID string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id FROM providers WHERE code = $1 AND is_active = true
		`, req.ProviderCode).Scan(&providerID)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PROVIDER_NOT_FOUND",
					"Provider not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Begin transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Update transaction with new provider and reset status
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
		message := "Retrying transaction with provider " + req.ProviderCode
		if req.Reason != "" {
			message += ": " + req.Reason
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, 'PROCESSING', $2, NOW())
		`, transactionID, message)

		// Create audit log
		_, err = tx.Exec(ctx, `
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
			"provider": req.ProviderCode,
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
			SET status = 'SUCCESS', serial_number = $1, completed_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`, req.SerialNumber, transactionID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add transaction log
		_, err = tx.Exec(ctx, `
			INSERT INTO transaction_logs (transaction_id, status, message, created_at)
			VALUES ($1, 'SUCCESS', $2, NOW())
		`, transactionID, "Manually processed: "+req.Reason)

		// Create audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'MANUAL_PROCESS', 'TRANSACTION', $2, $3, NOW())
		`, adminID, transactionID, "Manually processed: "+req.Reason)

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
