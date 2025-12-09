package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gate-v2/internal/middleware"
	"gate-v2/internal/utils"

	"github.com/jackc/pgx/v5"
)

// ============================================
// USER DASHBOARD ENDPOINTS
// ============================================

// handleGetUserTransactionsImpl returns user's transaction history with pagination
func HandleGetUserTransactionsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (required authentication)
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteUnauthorizedError(w)
			return
		}

		// Parse query parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		status := r.URL.Query().Get("status")
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get overview stats
		var totalTransactions, successCount, pendingCount, failedCount int
		var totalSpent int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN status IN ('PENDING', 'PAID', 'PROCESSING') THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status IN ('FAILED', 'EXPIRED', 'REFUNDED') THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN total_amount ELSE 0 END), 0) as total_spent
			FROM transactions
			WHERE user_id = $1
		`, userID).Scan(&totalTransactions, &successCount, &pendingCount, &failedCount, &totalSpent)

		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for transactions list
		query := `
			SELECT
				t.id, t.invoice_number, t.status, t.payment_status,
				p.code as product_code, p.title as product_name,
				s.code as sku_code, s.name as sku_name,
				t.quantity, t.subtotal, t.discount_amount, t.payment_fee, t.total_amount,
				t.currency, t.serial_number, t.created_at, t.completed_at
			FROM transactions t
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			WHERE t.user_id = $1
		`

		args := []interface{}{userID}
		argCount := 1

		// Add filters
		if status != "" {
			argCount++
			query += " AND t.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

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
			var quantity int
			var subtotal, discountAmount, paymentFee, totalAmount int64
			var currency, serialNumber string
			var createdAt, completedAt time.Time
			var completedAtPtr *time.Time

			err := rows.Scan(
				&id, &invoiceNumber, &status, &paymentStatus,
				&productCode, &productName, &skuCode, &skuName,
				&quantity, &subtotal, &discountAmount, &paymentFee, &totalAmount,
				&currency, &serialNumber, &createdAt, &completedAtPtr,
			)
			if err != nil {
				continue
			}

			if completedAtPtr != nil {
				completedAt = *completedAtPtr
			}

			transaction := map[string]interface{}{
				"id":             id,
				"invoiceNumber":  invoiceNumber,
				"status":         status,
				"paymentStatus":  paymentStatus,
				"productCode":    productCode,
				"productName":    productName,
				"skuCode":        skuCode,
				"skuName":        skuName,
				"quantity":       quantity,
				"subtotal":       subtotal,
				"discountAmount": discountAmount,
				"paymentFee":     paymentFee,
				"total":          totalAmount,
				"currency":       currency,
				"createdAt":      createdAt.Format(time.RFC3339),
			}

			if serialNumber != "" {
				transaction["serialNumber"] = serialNumber
			}

			if completedAtPtr != nil {
				transaction["completedAt"] = completedAt.Format(time.RFC3339)
			}

			transactions = append(transactions, transaction)
		}

		// Calculate total pages
		totalPages := (totalTransactions + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalTransactions": totalTransactions,
				"totalSpent":        totalSpent,
				"successCount":      successCount,
				"pendingCount":      pendingCount,
				"failedCount":       failedCount,
			},
			"transactions": transactions,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalTransactions,
				"totalPages": totalPages,
			},
		})
	}
}

// handleGetMutationsImpl returns user's balance mutation history
func HandleGetMutationsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteUnauthorizedError(w)
			return
		}

		// Parse query parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		mutationType := r.URL.Query().Get("type") // CREDIT or DEBIT
		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get overview
		var totalCredit, totalDebit, netBalance int64
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COALESCE(SUM(CASE WHEN type = 'CREDIT' THEN amount ELSE 0 END), 0) as total_credit,
				COALESCE(SUM(CASE WHEN type = 'DEBIT' THEN amount ELSE 0 END), 0) as total_debit,
				COALESCE(SUM(CASE WHEN type = 'CREDIT' THEN amount ELSE -amount END), 0) as net_balance
			FROM balance_mutations
			WHERE user_id = $1 AND currency = 'IDR'
		`, userID).Scan(&totalCredit, &totalDebit, &netBalance)

		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for mutations list
		query := `
			SELECT
				id, type, amount, balance_before, balance_after,
				description, reference_type, reference_id, currency, created_at
			FROM balance_mutations
			WHERE user_id = $1
		`

		args := []interface{}{userID}
		argCount := 1

		if mutationType != "" {
			argCount++
			query += " AND type = $" + strconv.Itoa(argCount)
			args = append(args, mutationType)
		}

		query += " ORDER BY created_at DESC"
		argCount++
		query += " LIMIT $" + strconv.Itoa(argCount)
		args = append(args, limit)

		argCount++
		query += " OFFSET $" + strconv.Itoa(argCount)
		args = append(args, offset)

		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		mutations := []map[string]interface{}{}
		for rows.Next() {
			var id, mType, description, referenceType, referenceID, currency string
			var amount, balanceBefore, balanceAfter int64
			var createdAt time.Time

			err := rows.Scan(
				&id, &mType, &amount, &balanceBefore, &balanceAfter,
				&description, &referenceType, &referenceID, &currency, &createdAt,
			)
			if err != nil {
				continue
			}

			mutations = append(mutations, map[string]interface{}{
				"id":            id,
				"type":          mType,
				"amount":        amount,
				"balanceBefore": balanceBefore,
				"balanceAfter":  balanceAfter,
				"description":   description,
				"reference":     referenceID,
				"currency":      currency,
				"createdAt":     createdAt.Format(time.RFC3339),
			})
		}

		// Get total count
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM balance_mutations WHERE user_id = $1"
		countArgs := []interface{}{userID}

		if mutationType != "" {
			countQuery += " AND type = $2"
			countArgs = append(countArgs, mutationType)
		}

		deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		totalPages := (totalRows + limit - 1) / limit

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalCredit": totalCredit,
				"totalDebit":  totalDebit,
				"netBalance":  netBalance,
			},
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

// handleGetReportsImpl returns user's spending analytics
func HandleGetReportsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteUnauthorizedError(w)
			return
		}

		period := r.URL.Query().Get("period") // daily, weekly, monthly, yearly
		if period == "" {
			period = "monthly"
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get overview
		var totalSpent, totalTransactions int64
		var avgPerTransaction float64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COALESCE(SUM(total_amount), 0) as total_spent,
				COUNT(*) as total_transactions,
				COALESCE(AVG(total_amount), 0) as avg_per_transaction
			FROM transactions
			WHERE user_id = $1 AND status = 'SUCCESS'
		`, userID).Scan(&totalSpent, &totalTransactions, &avgPerTransaction)

		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

		// Get spending by product
		byProductRows, _ := deps.DB.Pool.Query(ctx, `
			SELECT
				p.code, p.title,
				COUNT(*) as transaction_count,
				SUM(t.total_amount) as total_spent
			FROM transactions t
			JOIN products p ON t.product_id = p.id
			WHERE t.user_id = $1 AND t.status = 'SUCCESS'
			GROUP BY p.code, p.title
			ORDER BY total_spent DESC
			LIMIT 10
		`, userID)

		byProduct := []map[string]interface{}{}
		if byProductRows != nil {
			defer byProductRows.Close()
			for byProductRows.Next() {
				var code, title string
				var count int
				var spent int64
				byProductRows.Scan(&code, &title, &count, &spent)
				byProduct = append(byProduct, map[string]interface{}{
					"product":     title,
					"productCode": code,
					"count":       count,
					"spent":       spent,
				})
			}
		}

		// Get spending by month (last 12 months)
		byMonthRows, _ := deps.DB.Pool.Query(ctx, `
			SELECT
				TO_CHAR(created_at, 'YYYY-MM') as month,
				COUNT(*) as transaction_count,
				SUM(total_amount) as total_spent
			FROM transactions
			WHERE user_id = $1 AND status = 'SUCCESS'
			AND created_at >= CURRENT_DATE - INTERVAL '12 months'
			GROUP BY TO_CHAR(created_at, 'YYYY-MM')
			ORDER BY month DESC
			LIMIT 12
		`, userID)

		byMonth := []map[string]interface{}{}
		if byMonthRows != nil {
			defer byMonthRows.Close()
			for byMonthRows.Next() {
				var month string
				var count int
				var spent int64
				byMonthRows.Scan(&month, &count, &spent)
				byMonth = append(byMonth, map[string]interface{}{
					"month": month,
					"count": count,
					"spent": spent,
				})
			}
		}

		// Get spending by payment method
		byPaymentRows, _ := deps.DB.Pool.Query(ctx, `
			SELECT
				pc.code, pc.name,
				COUNT(*) as transaction_count,
				SUM(t.total_amount) as total_spent
			FROM transactions t
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			WHERE t.user_id = $1 AND t.status = 'SUCCESS'
			GROUP BY pc.code, pc.name
			ORDER BY total_spent DESC
			LIMIT 10
		`, userID)

		byPaymentMethod := []map[string]interface{}{}
		if byPaymentRows != nil {
			defer byPaymentRows.Close()
			for byPaymentRows.Next() {
				var code, name string
				var count int
				var spent int64
				byPaymentRows.Scan(&code, &name, &count, &spent)
				byPaymentMethod = append(byPaymentMethod, map[string]interface{}{
					"payment": name,
					"code":    code,
					"count":   count,
					"spent":   spent,
				})
			}
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalSpent":            totalSpent,
				"totalTransactions":     totalTransactions,
				"averagePerTransaction": avgPerTransaction,
			},
			"byProduct":       byProduct,
			"byMonth":         byMonth,
			"byPaymentMethod": byPaymentMethod,
		})
	}
}

// handleGetDepositsImpl returns user's deposit history
func HandleGetDepositsImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteUnauthorizedError(w)
			return
		}

		// Parse query parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 100 {
			limit = 10
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page <= 0 {
			page = 1
		}

		status := r.URL.Query().Get("status")
		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get overview
		var totalDeposits, successCount, pendingCount, failedCount int
		var totalAmount int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status IN ('FAILED', 'EXPIRED', 'CANCELLED') THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN amount ELSE 0 END), 0) as total_amount
			FROM deposits
			WHERE user_id = $1
		`, userID).Scan(&totalDeposits, &successCount, &pendingCount, &failedCount, &totalAmount)

		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for deposits list
		query := `
			SELECT
				d.id, d.invoice_number, d.status, d.amount, d.payment_fee, d.total_amount,
				d.currency, pc.code as payment_code, pc.name as payment_name,
				d.created_at, d.paid_at, d.expired_at
			FROM deposits d
			JOIN payment_channels pc ON d.payment_channel_id = pc.id
			WHERE d.user_id = $1
		`

		args := []interface{}{userID}
		argCount := 1

		if status != "" {
			argCount++
			query += " AND d.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

		query += " ORDER BY d.created_at DESC"
		argCount++
		query += " LIMIT $" + strconv.Itoa(argCount)
		args = append(args, limit)

		argCount++
		query += " OFFSET $" + strconv.Itoa(argCount)
		args = append(args, offset)

		rows, err := deps.DB.Pool.Query(ctx, query, args...)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		deposits := []map[string]interface{}{}
		for rows.Next() {
			var id, invoiceNumber, status, paymentCode, paymentName, currency string
			var amount, paymentFee, totalAmount int64
			var createdAt, expiredAt time.Time
			var paidAtPtr *time.Time

			err := rows.Scan(
				&id, &invoiceNumber, &status, &amount, &paymentFee, &totalAmount,
				&currency, &paymentCode, &paymentName,
				&createdAt, &paidAtPtr, &expiredAt,
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
				"payment": map[string]interface{}{
					"code": paymentCode,
					"name": paymentName,
				},
				"createdAt": createdAt.Format(time.RFC3339),
				"expiredAt": expiredAt.Format(time.RFC3339),
			}

			if paidAtPtr != nil {
				deposit["paidAt"] = (*paidAtPtr).Format(time.RFC3339)
			}

			deposits = append(deposits, deposit)
		}

		// Get total count
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM deposits WHERE user_id = $1"
		countArgs := []interface{}{userID}

		if status != "" {
			countQuery += " AND status = $2"
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

// DepositInquiryRequest represents the request body for deposit inquiry
type DepositInquiryRequest struct {
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Currency    string `json:"currency" validate:"required"`
	PaymentCode string `json:"paymentCode" validate:"required"`
}

// handleDepositInquiryImpl validates deposit request and returns pricing
func HandleDepositInquiryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DepositInquiryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate required fields
		if req.Amount <= 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"amount": "Amount must be greater than 0",
			})
			return
		}

		if req.PaymentCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"paymentCode": "Payment code is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get payment channel info
		var paymentChannelID, paymentChannelName string
		var feeAmount, feePercentage int
		var minAmount, maxAmount int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT id, name, fee_amount, fee_percentage, min_amount, max_amount
			FROM payment_channels
			WHERE code = $1 AND is_active = true AND 'deposit' = ANY(supported_types)
		`, req.PaymentCode).Scan(
			&paymentChannelID, &paymentChannelName,
			&feeAmount, &feePercentage, &minAmount, &maxAmount,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_NOT_FOUND",
					"Payment method not found or not supported for deposit", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Validate amount limits
		if req.Amount < minAmount {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "AMOUNT_TOO_LOW",
				"Deposit amount is below minimum", "Minimum: "+strconv.FormatInt(minAmount, 10))
			return
		}

		if maxAmount > 0 && req.Amount > maxAmount {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "AMOUNT_TOO_HIGH",
				"Deposit amount exceeds maximum", "Maximum: "+strconv.FormatInt(maxAmount, 10))
			return
		}

		// Calculate payment fee
		paymentFee := int64(feeAmount)
		if feePercentage > 0 {
			paymentFee += (req.Amount * int64(feePercentage)) / 100
		}

		total := req.Amount + paymentFee

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"amount":      req.Amount,
			"paymentFee":  paymentFee,
			"total":       total,
			"currency":    req.Currency,
			"paymentCode": req.PaymentCode,
			"paymentName": paymentChannelName,
			"limits": map[string]interface{}{
				"minAmount": minAmount,
				"maxAmount": maxAmount,
			},
		})
	}
}

// CreateDepositRequest represents the request body for creating deposit
type CreateDepositRequest struct {
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Currency    string `json:"currency" validate:"required"`
	PaymentCode string `json:"paymentCode" validate:"required"`
}

// handleCreateDepositImpl creates a new deposit/top-up request
func HandleCreateDepositImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		if userID == "" {
			utils.WriteUnauthorizedError(w)
			return
		}

		var req CreateDepositRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate
		if req.Amount <= 0 {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"amount": "Amount must be greater than 0",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get payment channel
		var paymentChannelID, paymentChannelName, instruction string
		var feeAmount, feePercentage int
		var minAmount, maxAmount int64
		var gatewayID *string

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				pc.id, pc.name, pc.instruction,
				pc.fee_amount, pc.fee_percentage,
				pc.min_amount, pc.max_amount,
				pcg.gateway_id
			FROM payment_channels pc
			LEFT JOIN payment_channel_gateway_assignment pcg
				ON pc.id = pcg.payment_channel_id AND pcg.transaction_type = 'deposit'
			WHERE pc.code = $1 AND pc.is_active = true
		`, req.PaymentCode).Scan(
			&paymentChannelID, &paymentChannelName, &instruction,
			&feeAmount, &feePercentage, &minAmount, &maxAmount,
			&gatewayID,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_NOT_FOUND",
					"Payment method not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Validate limits
		if req.Amount < minAmount {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "AMOUNT_TOO_LOW",
				"Amount below minimum", "")
			return
		}

		if maxAmount > 0 && req.Amount > maxAmount {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "AMOUNT_TOO_HIGH",
				"Amount exceeds maximum", "")
			return
		}

		// Calculate fee
		paymentFee := int64(feeAmount)
		if feePercentage > 0 {
			paymentFee += (req.Amount * int64(feePercentage)) / 100
		}
		total := req.Amount + paymentFee

		// Generate invoice number
		invoiceNumber := utils.GenerateDepositInvoiceNumber()

		// Set expiry (1 hour for e-wallet, 24 hours for VA/QRIS)
		expiryDuration := 1 * time.Hour
		if req.PaymentCode == "BCA_VA" || req.PaymentCode == "BRI_VA" || req.PaymentCode == "QRIS" {
			expiryDuration = 24 * time.Hour
		}
		expiredAt := time.Now().Add(expiryDuration)

		// Insert deposit record
		var depositID string
		err = deps.DB.Pool.QueryRow(ctx, `
			INSERT INTO deposits (
				invoice_number, user_id, payment_channel_id, payment_gateway_id,
				amount, payment_fee, total_amount, currency, status,
				created_at, expired_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
		`, invoiceNumber, userID, paymentChannelID, gatewayID,
			req.Amount, paymentFee, total, req.Currency, "PENDING",
			time.Now(), expiredAt,
		).Scan(&depositID)

		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		// TODO: Call payment gateway to generate payment data
		// For now, return mock payment data
		paymentData := map[string]interface{}{
			"code":        req.PaymentCode,
			"name":        paymentChannelName,
			"instruction": instruction,
			"expiredAt":   expiredAt.Format(time.RFC3339),
		}

		// Mock payment data based on payment type
		if req.PaymentCode == "QRIS" {
			paymentData["qrCode"] = "00020101021126660016ID.CO.QRIS.WWW..."
			paymentData["qrCodeImage"] = "https://nos.jkt-1.neo.id/gate/qr/" + invoiceNumber + ".png"
		} else if req.PaymentCode == "BCA_VA" || req.PaymentCode == "BRI_VA" {
			paymentData["accountNumber"] = "80777" + invoiceNumber[3:18]
			paymentData["accountName"] = "GATE TOPUP"
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"step": "SUCCESS",
			"deposit": map[string]interface{}{
				"invoiceNumber": invoiceNumber,
				"status":        "PENDING",
				"amount":        req.Amount,
				"paymentFee":    paymentFee,
				"total":         total,
				"currency":      req.Currency,
				"payment":       paymentData,
				"createdAt":     time.Now().Format(time.RFC3339),
				"expiredAt":     expiredAt.Format(time.RFC3339),
			},
		})
	}
}
