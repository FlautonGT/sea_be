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
// ADMIN DEPOSIT MANAGEMENT
// ============================================

// handleAdminGetDepositsImpl returns all deposits with filters
func HandleAdminGetDepositsImpl(deps *Dependencies) http.HandlerFunc {
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
		gatewayCode := r.URL.Query().Get("gatewayCode")
		region := r.URL.Query().Get("region")
		userID := r.URL.Query().Get("userId")
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Get overview stats
		var totalDeposits, successCount, pendingCount, expiredCount, failedCount int
		var totalAmount int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0) as expired,
				COALESCE(SUM(CASE WHEN status IN ('FAILED', 'CANCELLED') THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN total_amount ELSE 0 END), 0) as total_amount
			FROM deposits
		`).Scan(&totalDeposits, &successCount, &pendingCount, &expiredCount, &failedCount, &totalAmount)

		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for deposits list
		query := `
			SELECT
				d.id, d.invoice_number, d.status, d.amount, d.payment_fee, d.total_amount, d.currency,
				u.id, u.first_name || ' ' || u.last_name, u.email,
				pc.code, pc.name,
				pg.code, pg.name,
				u.primary_region,
				d.created_at, d.paid_at, d.expired_at
			FROM deposits d
			JOIN users u ON d.user_id = u.id
			JOIN payment_channels pc ON d.payment_channel_id = pc.id
			LEFT JOIN payment_gateways pg ON d.payment_gateway_id = pg.id
			WHERE 1=1
		`

		args := []interface{}{}
		argCount := 0

		// Add search filter
		if search != "" {
			argCount++
			query += " AND (d.invoice_number ILIKE $" + strconv.Itoa(argCount)
			query += " OR u.email ILIKE $" + strconv.Itoa(argCount) + ")"
			args = append(args, "%"+search+"%")
		}

		// Add filters
		if status != "" {
			argCount++
			query += " AND d.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

		if paymentCode != "" {
			argCount++
			query += " AND pc.code = $" + strconv.Itoa(argCount)
			args = append(args, paymentCode)
		}

		if gatewayCode != "" {
			argCount++
			query += " AND pg.code = $" + strconv.Itoa(argCount)
			args = append(args, gatewayCode)
		}

		if region != "" {
			argCount++
			query += " AND u.primary_region = $" + strconv.Itoa(argCount)
			args = append(args, region)
		}

		if userID != "" {
			argCount++
			query += " AND d.user_id = $" + strconv.Itoa(argCount)
			args = append(args, userID)
		}

		if startDate != "" {
			argCount++
			query += " AND d.created_at >= $" + strconv.Itoa(argCount)
			args = append(args, startDate)
		}

		if endDate != "" {
			argCount++
			query += " AND d.created_at <= $" + strconv.Itoa(argCount)
			args = append(args, endDate)
		}

		query += " ORDER BY d.created_at DESC"
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

		deposits := []map[string]interface{}{}
		for rows.Next() {
			var id, invoiceNumber, status, currency string
			var amount, paymentFee, totalAmount int64
			var userID, userName, userEmail, paymentCode, paymentName, gatewayCode, gatewayName, region string
			var createdAt, expiredAt time.Time
			var paidAt *time.Time

			err := rows.Scan(
				&id, &invoiceNumber, &status, &amount, &paymentFee, &totalAmount, &currency,
				&userID, &userName, &userEmail,
				&paymentCode, &paymentName,
				&gatewayCode, &gatewayName,
				&region,
				&createdAt, &paidAt, &expiredAt,
			)
			if err != nil {
				continue
			}

			deposit := map[string]interface{}{
				"id":            id,
				"invoiceNumber": invoiceNumber,
				"status":        status,
				"amount":        amount,
				"paymentFee":    paymentFee,
				"total":         totalAmount,
				"currency":      currency,
				"user": map[string]interface{}{
					"id":    userID,
					"name":  userName,
					"email": userEmail,
				},
				"payment": map[string]interface{}{
					"code": paymentCode,
					"name": paymentName,
				},
				"gateway": map[string]interface{}{
					"code": gatewayCode,
					"name": gatewayName,
				},
				"region":    region,
				"createdAt": createdAt.Format(time.RFC3339),
				"expiredAt": expiredAt.Format(time.RFC3339),
			}

			if paidAt != nil {
				deposit["paidAt"] = (*paidAt).Format(time.RFC3339)
			}

			deposits = append(deposits, deposit)
		}

		// Get total count for pagination
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM deposits d JOIN users u ON d.user_id = u.id JOIN payment_channels pc ON d.payment_channel_id = pc.id LEFT JOIN payment_gateways pg ON d.payment_gateway_id = pg.id WHERE 1=1"
		countArgs := []interface{}{}
		countArgCount := 0

		if search != "" {
			countArgCount++
			countQuery += " AND (d.invoice_number ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR u.email ILIKE $" + strconv.Itoa(countArgCount) + ")"
			countArgs = append(countArgs, "%"+search+"%")
		}

		if status != "" {
			countArgCount++
			countQuery += " AND d.status = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, status)
		}

		deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalDeposits": totalDeposits,
				"totalAmount":   totalAmount,
				"successCount":  successCount,
				"pendingCount":  pendingCount,
				"expiredCount":  expiredCount,
				"failedCount":   failedCount,
			},
			"deposits": deposits,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

// handleAdminGetDepositImpl returns detailed deposit info
func HandleAdminGetDepositImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := chi.URLParam(r, "depositId")
		if depositID == "" {
			utils.WriteBadRequestError(w, "Deposit ID is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get deposit details
		var id, invoiceNumber, status, currency string
		var amount, paymentFee, totalAmount int64
		var userID, userName, userEmail, userPhone string
		var paymentCode, paymentName, gatewayCode, gatewayName string
		var paymentData json.RawMessage
		var createdAt, expiredAt time.Time
		var paidAt *time.Time

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				d.id, d.invoice_number, d.status, d.amount, d.payment_fee, d.total_amount, d.currency,
				u.id, u.first_name || ' ' || u.last_name, u.email, u.phone_number,
				pc.code, pc.name,
				pg.code, pg.name,
				d.payment_data,
				d.created_at, d.paid_at, d.expired_at
			FROM deposits d
			JOIN users u ON d.user_id = u.id
			JOIN payment_channels pc ON d.payment_channel_id = pc.id
			LEFT JOIN payment_gateways pg ON d.payment_gateway_id = pg.id
			WHERE d.id = $1
		`, depositID).Scan(
			&id, &invoiceNumber, &status, &amount, &paymentFee, &totalAmount, &currency,
			&userID, &userName, &userEmail, &userPhone,
			&paymentCode, &paymentName,
			&gatewayCode, &gatewayName,
			&paymentData,
			&createdAt, &paidAt, &expiredAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "DEPOSIT_NOT_FOUND",
					"Deposit not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Parse payment data
		var paymentInfo map[string]interface{}
		json.Unmarshal(paymentData, &paymentInfo)

		response := map[string]interface{}{
			"id":            id,
			"invoiceNumber": invoiceNumber,
			"status":        status,
			"user": map[string]interface{}{
				"id":          userID,
				"name":        userName,
				"email":       userEmail,
				"phoneNumber": userPhone,
			},
			"pricing": map[string]interface{}{
				"amount":     amount,
				"paymentFee": paymentFee,
				"total":      totalAmount,
				"currency":   currency,
			},
			"payment": map[string]interface{}{
				"code": paymentCode,
				"name": paymentName,
				"data": paymentInfo,
			},
			"gateway": map[string]interface{}{
				"code": gatewayCode,
				"name": gatewayName,
			},
			"createdAt": createdAt.Format(time.RFC3339),
			"expiredAt": expiredAt.Format(time.RFC3339),
		}

		if paidAt != nil {
			response["paidAt"] = (*paidAt).Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, response)
	}
}

// ConfirmDepositRequest represents the request to confirm a deposit
type ConfirmDepositRequest struct {
	Reason       string `json:"reason" validate:"required"`
	GatewayRefID string `json:"gatewayRefId"`
}

// handleConfirmDepositImpl manually confirms a pending deposit
func HandleConfirmDepositImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := chi.URLParam(r, "depositId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req ConfirmDepositRequest
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

		// Get deposit details
		var userID, currency, status string
		var amount int64

		err = tx.QueryRow(ctx, `
			SELECT user_id, amount, currency, status
			FROM deposits
			WHERE id = $1
		`, depositID).Scan(&userID, &amount, &currency, &status)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "DEPOSIT_NOT_FOUND",
					"Deposit not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Validate deposit status
		if status != "PENDING" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "DEPOSIT_NOT_PENDING",
				"Deposit must be in PENDING status", "")
			return
		}

		// Get current balance
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

		var currentBalance int64
		err = tx.QueryRow(ctx, "SELECT "+balanceColumn+" FROM users WHERE id = $1", userID).Scan(&currentBalance)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Add deposit amount to user balance
		newBalance := currentBalance + amount

		_, err = tx.Exec(ctx, "UPDATE users SET "+balanceColumn+" = $1 WHERE id = $2", newBalance, userID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create balance mutation record
		_, err = tx.Exec(ctx, `
			INSERT INTO balance_mutations (
				user_id, type, amount, balance_before, balance_after,
				description, reference_type, reference_id, currency, created_at
			) VALUES ($1, 'CREDIT', $2, $3, $4, $5, 'DEPOSIT', $6, $7, NOW())
		`, userID, amount, currentBalance, newBalance,
			"Deposit confirmed: "+req.Reason, depositID, currency)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Update deposit status to SUCCESS
		_, err = tx.Exec(ctx, `
			UPDATE deposits
			SET status = 'SUCCESS', paid_at = NOW()
			WHERE id = $1
		`, depositID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CONFIRM', 'DEPOSIT', $2, $3, NOW())
		`, adminID, depositID, "Manually confirmed deposit: "+req.Reason)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Get admin name
		var adminName string
		deps.DB.Pool.QueryRow(context.Background(), "SELECT first_name || ' ' || last_name FROM admins WHERE id = $1", adminID).Scan(&adminName)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":       "Deposit confirmed successfully",
			"depositAmount": amount,
			"balanceBefore": currentBalance,
			"balanceAfter":  newBalance,
			"confirmedBy": map[string]interface{}{
				"id":   adminID,
				"name": adminName,
			},
			"confirmedAt": time.Now().Format(time.RFC3339),
		})
	}
}

// CancelDepositRequest represents the request to cancel a deposit
type CancelDepositRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// handleCancelDepositImpl cancels a pending deposit
func HandleCancelDepositImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := chi.URLParam(r, "depositId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req CancelDepositRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
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

		// Get deposit status
		var status string
		err = tx.QueryRow(ctx, "SELECT status FROM deposits WHERE id = $1", depositID).Scan(&status)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "DEPOSIT_NOT_FOUND",
					"Deposit not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Validate deposit status
		if status != "PENDING" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "DEPOSIT_NOT_PENDING",
				"Deposit must be in PENDING status", "")
			return
		}

		// Update deposit status to CANCELLED
		_, err = tx.Exec(ctx, "UPDATE deposits SET status = 'CANCELLED' WHERE id = $1", depositID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'CANCEL', 'DEPOSIT', $2, $3, NOW())
		`, adminID, depositID, "Cancelled deposit: "+req.Reason)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "Deposit cancelled successfully",
		})
	}
}

// RefundDepositRequest represents the request to refund a deposit
type RefundDepositRequest struct {
	Reason   string `json:"reason" validate:"required"`
	RefundTo string `json:"refundTo" validate:"required"` // ORIGINAL_METHOD or BALANCE
	Amount   int64  `json:"amount" validate:"required"`
}

// handleRefundDepositImpl refunds a completed deposit
func HandleRefundDepositImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		depositID := chi.URLParam(r, "depositId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req RefundDepositRequest
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

		// Get deposit details
		var userID, currency, status string
		var amount int64

		err = tx.QueryRow(ctx, `
			SELECT user_id, amount, currency, status
			FROM deposits
			WHERE id = $1
		`, depositID).Scan(&userID, &amount, &currency, &status)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "DEPOSIT_NOT_FOUND",
					"Deposit not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Validate deposit status
		if status != "SUCCESS" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "DEPOSIT_NOT_SUCCESS",
				"Can only refund SUCCESS deposits", "")
			return
		}

		// Get current balance
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

		var currentBalance int64
		err = tx.QueryRow(ctx, "SELECT "+balanceColumn+" FROM users WHERE id = $1", userID).Scan(&currentBalance)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Check if user has sufficient balance
		if currentBalance < req.Amount {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE",
				"User has insufficient balance for refund", "")
			return
		}

		// Deduct refund amount from user balance
		newBalance := currentBalance - req.Amount

		_, err = tx.Exec(ctx, "UPDATE users SET "+balanceColumn+" = $1 WHERE id = $2", newBalance, userID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create balance mutation record
		_, err = tx.Exec(ctx, `
			INSERT INTO balance_mutations (
				user_id, type, amount, balance_before, balance_after,
				description, reference_type, reference_id, currency, created_at
			) VALUES ($1, 'DEBIT', $2, $3, $4, $5, 'DEPOSIT_REFUND', $6, $7, NOW())
		`, userID, req.Amount, currentBalance, newBalance,
			"Deposit refunded: "+req.Reason, depositID, currency)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Update deposit status to REFUNDED
		_, err = tx.Exec(ctx, "UPDATE deposits SET status = 'REFUNDED' WHERE id = $1", depositID)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'REFUND', 'DEPOSIT', $2, $3, NOW())
		`, adminID, depositID, "Refunded deposit: "+req.Reason)

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Get admin name
		var adminName string
		deps.DB.Pool.QueryRow(context.Background(), "SELECT first_name || ' ' || last_name FROM admins WHERE id = $1", adminID).Scan(&adminName)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message":       "Deposit refunded successfully",
			"refundAmount":  req.Amount,
			"refundTo":      req.RefundTo,
			"balanceBefore": currentBalance,
			"balanceAfter":  newBalance,
			"processedBy": map[string]interface{}{
				"id":   adminID,
				"name": adminName,
			},
			"createdAt": time.Now().Format(time.RFC3339),
		})
	}
}
