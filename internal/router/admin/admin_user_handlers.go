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
// ADMIN USER MANAGEMENT
// ============================================

// handleAdminGetUsersImpl returns all users with filters
func HandleAdminGetUsersImpl(deps *Dependencies) http.HandlerFunc {
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
		membership := r.URL.Query().Get("membership")
		region := r.URL.Query().Get("region")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Build query
		query := `
			SELECT
				u.id, u.first_name, u.last_name, u.email, u.phone_number,
				u.profile_picture, u.status, u.primary_region, u.membership_level,
				u.balance_idr, u.balance_myr, u.balance_php, u.balance_sgd, u.balance_thb,
				u.total_transactions, u.total_spent_idr, u.last_login_at, u.created_at
			FROM users u
			WHERE 1=1
		`

		args := []interface{}{}
		argCount := 0

		// Add search filter
		if search != "" {
			argCount++
			query += " AND (u.first_name ILIKE $" + strconv.Itoa(argCount)
			query += " OR u.last_name ILIKE $" + strconv.Itoa(argCount)
			query += " OR u.email ILIKE $" + strconv.Itoa(argCount)
			query += " OR u.phone_number ILIKE $" + strconv.Itoa(argCount) + ")"
			args = append(args, "%"+search+"%")
		}

		// Add filters
		if status != "" {
			argCount++
			query += " AND u.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

		if membership != "" {
			argCount++
			query += " AND u.membership_level = $" + strconv.Itoa(argCount)
			args = append(args, membership)
		}

		if region != "" {
			argCount++
			query += " AND u.primary_region = $" + strconv.Itoa(argCount)
			args = append(args, region)
		}

		query += " ORDER BY u.created_at DESC"
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

		users := []map[string]interface{}{}
		for rows.Next() {
			var id, firstName, lastName, email, phoneNumber, profilePicture string
			var status, primaryRegion, membershipLevel string
			var balanceIDR, balanceMYR, balancePHP, balanceSGD, balanceTHB int64
			var totalTransactions int
			var totalSpentIDR int64
			var lastLoginAt, createdAt *time.Time

			err := rows.Scan(
				&id, &firstName, &lastName, &email, &phoneNumber,
				&profilePicture, &status, &primaryRegion, &membershipLevel,
				&balanceIDR, &balanceMYR, &balancePHP, &balanceSGD, &balanceTHB,
				&totalTransactions, &totalSpentIDR, &lastLoginAt, &createdAt,
			)
			if err != nil {
				continue
			}

			user := map[string]interface{}{
				"id":            id,
				"firstName":     firstName,
				"lastName":      lastName,
				"email":         email,
				"phoneNumber":   phoneNumber,
				"profilePicture": profilePicture,
				"status":        status,
				"primaryRegion": primaryRegion,
				"membership": map[string]interface{}{
					"level": membershipLevel,
					"name":  getMembershipName(membershipLevel),
				},
				"balance": map[string]interface{}{
					"IDR": balanceIDR,
					"MYR": balanceMYR,
					"PHP": balancePHP,
					"SGD": balanceSGD,
					"THB": balanceTHB,
				},
				"stats": map[string]interface{}{
					"totalTransactions": totalTransactions,
					"totalSpent":        totalSpentIDR,
				},
				"createdAt": createdAt.Format(time.RFC3339),
			}

			if lastLoginAt != nil {
				user["lastLoginAt"] = lastLoginAt.Format(time.RFC3339)
			}

			users = append(users, user)
		}

		// Get total count
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM users WHERE 1=1"
		countArgs := []interface{}{}
		countArgCount := 0

		if search != "" {
			countArgCount++
			countQuery += " AND (first_name ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR last_name ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR email ILIKE $" + strconv.Itoa(countArgCount)
			countQuery += " OR phone_number ILIKE $" + strconv.Itoa(countArgCount) + ")"
			countArgs = append(countArgs, "%"+search+"%")
		}

		if status != "" {
			countArgCount++
			countQuery += " AND status = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, status)
		}

		if membership != "" {
			countArgCount++
			countQuery += " AND membership_level = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, membership)
		}

		if region != "" {
			countArgCount++
			countQuery += " AND primary_region = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, region)
		}

		deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"users": users,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}

// handleAdminGetUserImpl returns detailed user info
func HandleAdminGetUserImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")
		if userID == "" {
			utils.WriteBadRequestError(w, "User ID is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user details
		var id, firstName, lastName, email, phoneNumber, profilePicture string
		var status, primaryRegion, membershipLevel string
		var balanceIDR, balanceMYR, balancePHP, balanceSGD, balanceTHB int64
		var totalTransactions int
		var totalSpentIDR int64
		var emailVerifiedAt, lastLoginAt, createdAt *time.Time

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				id, first_name, last_name, email, phone_number, profile_picture,
				status, primary_region, membership_level,
				balance_idr, balance_myr, balance_php, balance_sgd, balance_thb,
				total_transactions, total_spent_idr,
				email_verified_at, last_login_at, created_at
			FROM users
			WHERE id = $1
		`, userID).Scan(
			&id, &firstName, &lastName, &email, &phoneNumber, &profilePicture,
			&status, &primaryRegion, &membershipLevel,
			&balanceIDR, &balanceMYR, &balancePHP, &balanceSGD, &balanceTHB,
			&totalTransactions, &totalSpentIDR,
			&emailVerifiedAt, &lastLoginAt, &createdAt,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND",
					"User not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Get last transaction
		var lastTransactionAt *time.Time
		deps.DB.Pool.QueryRow(ctx, `
			SELECT MAX(created_at)
			FROM transactions
			WHERE user_id = $1 AND status = 'SUCCESS'
		`, userID).Scan(&lastTransactionAt)

		response := map[string]interface{}{
			"id":            id,
			"firstName":     firstName,
			"lastName":      lastName,
			"email":         email,
			"phoneNumber":   phoneNumber,
			"profilePicture": profilePicture,
			"status":        status,
			"primaryRegion": primaryRegion,
			"membership": map[string]interface{}{
				"level": membershipLevel,
				"name":  getMembershipName(membershipLevel),
			},
			"balance": map[string]interface{}{
				"IDR": balanceIDR,
				"MYR": balanceMYR,
				"PHP": balancePHP,
				"SGD": balanceSGD,
				"THB": balanceTHB,
			},
			"stats": map[string]interface{}{
				"totalTransactions": totalTransactions,
				"totalSpent":        totalSpentIDR,
			},
			"createdAt": createdAt.Format(time.RFC3339),
		}

		if emailVerifiedAt != nil {
			response["emailVerifiedAt"] = emailVerifiedAt.Format(time.RFC3339)
		}
		if lastLoginAt != nil {
			response["lastLoginAt"] = lastLoginAt.Format(time.RFC3339)
		}
		if lastTransactionAt != nil {
			response["lastTransactionAt"] = (*lastTransactionAt).Format(time.RFC3339)
		}

		utils.WriteSuccessJSON(w, response)
	}
}

// UpdateUserStatusRequest represents the request to update user status
type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required"`
	Reason string `json:"reason"`
}

// handleUpdateUserStatusImpl updates user status (suspend/activate)
func HandleUpdateUserStatusImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req UpdateUserStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate status
		validStatuses := []string{"ACTIVE", "INACTIVE", "SUSPENDED"}
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

		// Update user status
		_, err = tx.Exec(ctx, `
			UPDATE users
			SET status = $1
			WHERE id = $2
		`, req.Status, userID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		description := "Updated user status to " + req.Status
		if req.Reason != "" {
			description += ": " + req.Reason
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'UPDATE', 'USER', $2, $3, NOW())
		`, adminID, userID, description)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"message": "User status updated successfully",
			"status":  req.Status,
		})
	}
}

// AdjustBalanceRequest represents the request to adjust user balance
type AdjustBalanceRequest struct {
	Type     string `json:"type" validate:"required"`     // CREDIT or DEBIT
	Amount   int64  `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
	Reason   string `json:"reason" validate:"required"`
}

// handleAdjustBalanceImpl adjusts user balance (credit/debit)
func HandleAdjustBalanceImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")
		adminID := middleware.GetAdminIDFromContext(r.Context())

		var req AdjustBalanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate
		if req.Type != "CREDIT" && req.Type != "DEBIT" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"type": "Type must be CREDIT or DEBIT",
			})
			return
		}

		if req.Amount <= 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"amount": "Amount must be greater than 0",
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

		// Get current balance
		balanceColumn := "balance_idr"
		if req.Currency == "MYR" {
			balanceColumn = "balance_myr"
		} else if req.Currency == "PHP" {
			balanceColumn = "balance_php"
		} else if req.Currency == "SGD" {
			balanceColumn = "balance_sgd"
		} else if req.Currency == "THB" {
			balanceColumn = "balance_thb"
		}

		var currentBalance int64
		err = tx.QueryRow(ctx, "SELECT "+balanceColumn+" FROM users WHERE id = $1", userID).Scan(&currentBalance)
		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "USER_NOT_FOUND",
					"User not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Calculate new balance
		var newBalance int64
		if req.Type == "CREDIT" {
			newBalance = currentBalance + req.Amount
		} else {
			newBalance = currentBalance - req.Amount
			if newBalance < 0 {
				utils.WriteErrorJSON(w, http.StatusBadRequest, "INSUFFICIENT_BALANCE",
					"Insufficient balance", "")
				return
			}
		}

		// Update balance
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
			) VALUES ($1, $2, $3, $4, $5, $6, 'ADMIN_ADJUSTMENT', $7, $8, NOW())
		`, userID, req.Type, req.Amount, currentBalance, newBalance, req.Reason, adminID, req.Currency)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Create audit log
		_, err = tx.Exec(ctx, `
			INSERT INTO audit_logs (admin_id, action, resource, resource_id, description, created_at)
			VALUES ($1, 'ADJUST_BALANCE', 'USER', $2, $3, NOW())
		`, adminID, userID, req.Type+" "+strconv.FormatInt(req.Amount, 10)+" "+req.Currency+": "+req.Reason)

		if err := tx.Commit(ctx); err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// Get admin name for response
		var adminName string
		deps.DB.Pool.QueryRow(context.Background(), "SELECT first_name || ' ' || last_name FROM admins WHERE id = $1", adminID).Scan(&adminName)

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"userId":        userID,
			"type":          req.Type,
			"amount":        req.Amount,
			"currency":      req.Currency,
			"balanceBefore": currentBalance,
			"balanceAfter":  newBalance,
			"reason":        req.Reason,
			"processedBy": map[string]interface{}{
				"id":   adminID,
				"name": adminName,
			},
			"createdAt": time.Now().Format(time.RFC3339),
		})
	}
}

// handleUserTransactionsImpl returns user's transaction history
func HandleUserTransactionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")

		// Parse pagination
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get transactions
		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT
				t.id, t.invoice_number, t.status, t.payment_status,
				p.code, p.title, s.code, s.name,
				t.quantity, t.total_amount, t.currency,
				t.created_at, t.completed_at
			FROM transactions t
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			WHERE t.user_id = $1
			ORDER BY t.created_at DESC
			LIMIT $2 OFFSET $3
		`, userID, limit, offset)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		transactions := []map[string]interface{}{}
		for rows.Next() {
			var id, invoiceNumber, status, paymentStatus string
			var productCode, productName, skuCode, skuName, currency string
			var quantity int
			var totalAmount int64
			var createdAt time.Time
			var completedAt *time.Time

			rows.Scan(
				&id, &invoiceNumber, &status, &paymentStatus,
				&productCode, &productName, &skuCode, &skuName,
				&quantity, &totalAmount, &currency,
				&createdAt, &completedAt,
			)

			tx := map[string]interface{}{
				"id":            id,
				"invoiceNumber": invoiceNumber,
				"status":        status,
				"paymentStatus": paymentStatus,
				"product":       productName,
				"sku":           skuName,
				"quantity":      quantity,
				"total":         totalAmount,
				"currency":      currency,
				"createdAt":     createdAt.Format(time.RFC3339),
			}

			if completedAt != nil {
				tx["completedAt"] = completedAt.Format(time.RFC3339)
			}

			transactions = append(transactions, tx)
		}

		// Get total count
		var totalRows int
		deps.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM transactions WHERE user_id = $1", userID).Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
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

// handleUserMutationsImpl returns user's balance mutations
func HandleUserMutationsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")

		// Parse pagination
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get mutations
		rows, err := deps.DB.Pool.Query(ctx, `
			SELECT
				id, type, amount, balance_before, balance_after,
				description, reference_type, currency, created_at
			FROM balance_mutations
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, userID, limit, offset)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		mutations := []map[string]interface{}{}
		for rows.Next() {
			var id, mType, description, refType, currency string
			var amount, balanceBefore, balanceAfter int64
			var createdAt time.Time

			rows.Scan(
				&id, &mType, &amount, &balanceBefore, &balanceAfter,
				&description, &refType, &currency, &createdAt,
			)

			mutations = append(mutations, map[string]interface{}{
				"id":            id,
				"type":          mType,
				"amount":        amount,
				"balanceBefore": balanceBefore,
				"balanceAfter":  balanceAfter,
				"description":   description,
				"referenceType": refType,
				"currency":      currency,
				"createdAt":     createdAt.Format(time.RFC3339),
			})
		}

		// Get total count
		var totalRows int
		deps.DB.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM balance_mutations WHERE user_id = $1", userID).Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"mutations": mutations,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
		})
	}
}
