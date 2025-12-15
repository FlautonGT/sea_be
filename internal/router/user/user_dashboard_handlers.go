package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"seaply/internal/middleware"
	"seaply/internal/payment"
	"seaply/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
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
		paymentStatus := r.URL.Query().Get("paymentStatus")
		search := r.URL.Query().Get("search")
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		regionParam := r.URL.Query().Get("region")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user's current region if region param not provided
		var userRegion string
		if regionParam == "" {
			_ = deps.DB.Pool.QueryRow(ctx, `SELECT current_region FROM users WHERE id = $1`, userID).Scan(&userRegion)
		} else {
			userRegion = strings.ToUpper(regionParam)
		}

		// Build WHERE clause with all filters
		whereClause := "WHERE t.user_id = $1"
		args := []interface{}{userID}
		argCount := 1

		if userRegion != "" {
			argCount++
			whereClause += " AND t.region = $" + strconv.Itoa(argCount) + "::region_code"
			args = append(args, userRegion)
		}

		// Build filter conditions for overview (same as list query)
		overviewWhereClause := whereClause
		overviewArgs := make([]interface{}, len(args))
		copy(overviewArgs, args)
		overviewArgCount := argCount

		if status != "" && status != "ALL" {
			overviewArgCount++
			overviewWhereClause += " AND t.status = $" + strconv.Itoa(overviewArgCount) + "::transaction_status"
			overviewArgs = append(overviewArgs, status)
		}

		if paymentStatus != "" {
			overviewArgCount++
			overviewWhereClause += " AND t.payment_status = $" + strconv.Itoa(overviewArgCount) + "::payment_status"
			overviewArgs = append(overviewArgs, paymentStatus)
		}

		if search != "" {
			overviewArgCount++
			overviewWhereClause += " AND t.invoice_number ILIKE $" + strconv.Itoa(overviewArgCount)
			overviewArgs = append(overviewArgs, "%"+search+"%")
		}

		if startDate != "" {
			overviewArgCount++
			overviewWhereClause += " AND t.created_at >= $" + strconv.Itoa(overviewArgCount) + "::timestamp"
			overviewArgs = append(overviewArgs, startDate)
		}

		if endDate != "" {
			overviewArgCount++
			overviewWhereClause += " AND t.created_at <= $" + strconv.Itoa(overviewArgCount) + "::timestamp"
			overviewArgs = append(overviewArgs, endDate+" 23:59:59")
		}

		// Get overview stats with all filters applied
		// transaction_status: PENDING, PROCESSING, SUCCESS, FAILED
		// payment_status: UNPAID, PAID, FAILED, EXPIRED, REFUNDED
		var totalTransactions, successCount, processingCount, pendingCount, failedCount int
		var totalPurchase int64

		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN t.status = 'SUCCESS'::transaction_status THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN t.status = 'PROCESSING'::transaction_status THEN 1 ELSE 0 END), 0) as processing,
				COALESCE(SUM(CASE WHEN t.status = 'PENDING'::transaction_status THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN t.status = 'FAILED'::transaction_status THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN t.status = 'SUCCESS'::transaction_status THEN t.total_amount ELSE 0 END), 0) as total_purchase
			FROM transactions t
			`+overviewWhereClause+`
		`, overviewArgs...).Scan(&totalTransactions, &successCount, &processingCount, &pendingCount, &failedCount, &totalPurchase)

		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to get transaction overview")
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for transactions list
		query := `
			SELECT
				t.id, t.invoice_number, t.status, t.payment_status,
				p.code as product_code, p.title as product_name,
				s.code as sku_code, s.name as sku_name,
				t.account_inputs, t.account_nickname,
				t.quantity, t.sell_price, COALESCE(t.discount_amount, 0), COALESCE(t.payment_fee, 0), t.total_amount,
				t.currency, t.provider_serial_number,
				pc.code as payment_code, pc.name as payment_name,
				t.created_at
			FROM transactions t
			JOIN products p ON t.product_id = p.id
			JOIN skus s ON t.sku_id = s.id
			JOIN payment_channels pc ON t.payment_channel_id = pc.id
			` + whereClause + `
		`

		// Add filters
		if status != "" && status != "ALL" {
			argCount++
			query += " AND t.status = $" + strconv.Itoa(argCount) + "::transaction_status"
			args = append(args, status)
		}

		if paymentStatus != "" {
			argCount++
			query += " AND t.payment_status = $" + strconv.Itoa(argCount) + "::payment_status"
			args = append(args, paymentStatus)
		}

		if search != "" {
			argCount++
			query += " AND t.invoice_number ILIKE $" + strconv.Itoa(argCount)
			args = append(args, "%"+search+"%")
		}

		if startDate != "" {
			argCount++
			query += " AND t.created_at >= $" + strconv.Itoa(argCount) + "::timestamp"
			args = append(args, startDate)
		}

		if endDate != "" {
			argCount++
			query += " AND t.created_at <= $" + strconv.Itoa(argCount) + "::timestamp"
			args = append(args, endDate+" 23:59:59")
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
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to query transactions")
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		transactions := []map[string]interface{}{}
		for rows.Next() {
			var id, invoiceNumber, txStatus, pStatus, productCode, productName, skuCode, skuName string
			var accountInputs, accountNickname sql.NullString
			var quantity int
			var sellPrice, discountAmount, paymentFee, totalAmount int64
			var currency string
			var serialNumber sql.NullString
			var paymentCode, paymentName string
			var createdAt time.Time

			err := rows.Scan(
				&id, &invoiceNumber, &txStatus, &pStatus,
				&productCode, &productName, &skuCode, &skuName,
				&accountInputs, &accountNickname,
				&quantity, &sellPrice, &discountAmount, &paymentFee, &totalAmount,
				&currency, &serialNumber,
				&paymentCode, &paymentName,
				&createdAt,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan transaction row")
				continue
			}

			// Build account object
			account := map[string]interface{}{}
			if accountNickname.Valid && accountNickname.String != "" {
				account["nickname"] = accountNickname.String
			}
			if accountInputs.Valid && accountInputs.String != "" {
				account["inputs"] = parseAccountInputsToString(accountInputs.String)
			}

			transaction := map[string]interface{}{
				"invoiceNumber": invoiceNumber,
				"product": map[string]interface{}{
					"code": productCode,
					"name": productName,
				},
				"sku": map[string]interface{}{
					"code": skuCode,
					"name": skuName,
				},
				"quantity": quantity,
				"status": map[string]interface{}{
					"transaction": txStatus,
					"payment":     pStatus,
				},
				"account": account,
				"pricing": map[string]interface{}{
					"subtotal":   sellPrice,
					"discount":   discountAmount,
					"paymentFee": paymentFee,
					"total":      totalAmount,
					"currency":   currency,
				},
				"payment": map[string]interface{}{
					"code": paymentCode,
					"name": paymentName,
				},
				"createdAt": createdAt.Format(time.RFC3339),
			}

			transactions = append(transactions, transaction)
		}

		// Get total count with same filters as list query (but without limit/offset)
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM transactions t " + whereClause
		countArgs := []interface{}{userID}
		countArgCount := 1

		// Add region filter if exists
		if userRegion != "" {
			countArgCount++
			countQuery += " AND t.region = $" + strconv.Itoa(countArgCount) + "::region_code"
			countArgs = append(countArgs, userRegion)
		}

		// Add all filters
		if status != "" && status != "ALL" {
			countArgCount++
			countQuery += " AND t.status = $" + strconv.Itoa(countArgCount) + "::transaction_status"
			countArgs = append(countArgs, status)
		}

		if paymentStatus != "" {
			countArgCount++
			countQuery += " AND t.payment_status = $" + strconv.Itoa(countArgCount) + "::payment_status"
			countArgs = append(countArgs, paymentStatus)
		}

		if search != "" {
			countArgCount++
			countQuery += " AND t.invoice_number ILIKE $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, "%"+search+"%")
		}

		if startDate != "" {
			countArgCount++
			countQuery += " AND t.created_at >= $" + strconv.Itoa(countArgCount) + "::timestamp"
			countArgs = append(countArgs, startDate)
		}

		if endDate != "" {
			countArgCount++
			countQuery += " AND t.created_at <= $" + strconv.Itoa(countArgCount) + "::timestamp"
			countArgs = append(countArgs, endDate+" 23:59:59")
		}

		err = deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to count transactions")
			utils.WriteInternalServerError(w)
			return
		}

		// Calculate totalPages: (totalRows + limit - 1) / limit
		// Examples: 30 rows, limit 10 = 3 pages; 5 rows, limit 10 = 1 page; 0 rows = 0 pages
		totalPages := 0
		if totalRows > 0 {
			totalPages = (totalRows + limit - 1) / limit
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalTransaction": totalTransactions,
				"totalPurchase":    totalPurchase,
				"success":          successCount,
				"processing":       processingCount,
				"pending":          pendingCount,
				"failed":           failedCount,
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

// parseAccountInputsToString converts account inputs JSON to readable string format
func parseAccountInputsToString(inputs string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(inputs), &data); err != nil {
		return inputs
	}

	// Try to format as "userId - zoneId" or just "userId"
	userId, hasUserId := data["userId"]
	zoneId, hasZoneId := data["zoneId"]
	serverId, hasServerId := data["serverId"]

	if hasUserId {
		result := formatValue(userId)
		if hasZoneId {
			result += " - " + formatValue(zoneId)
		} else if hasServerId {
			result += " - " + formatValue(serverId)
		}
		return result
	}

	return inputs
}

func formatValue(v interface{}) string {
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

func nullStringOrEmpty(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
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
		search := r.URL.Query().Get("search")     // invoice number
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		regionParam := r.URL.Query().Get("region")
		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user's current region if region param not provided
		var userRegion string
		if regionParam == "" {
			_ = deps.DB.Pool.QueryRow(ctx, `SELECT current_region FROM users WHERE id = $1`, userID).Scan(&userRegion)
		} else {
			userRegion = strings.ToUpper(regionParam)
		}

		// Get currency for region
		var regionCurrency string
		if userRegion != "" {
			_ = deps.DB.Pool.QueryRow(ctx, `SELECT currency FROM regions WHERE code = $1`, userRegion).Scan(&regionCurrency)
		}

		// Build WHERE clause with all filters
		whereClause := "WHERE m.user_id = $1"
		args := []interface{}{userID}
		argCount := 1

		if regionCurrency != "" {
			argCount++
			whereClause += " AND m.currency = $" + strconv.Itoa(argCount)
			args = append(args, regionCurrency)
		}

		// Build filter conditions for overview (same as list query)
		overviewWhereClause := whereClause
		overviewArgs := make([]interface{}, len(args))
		copy(overviewArgs, args)
		overviewArgCount := argCount

		if mutationType != "" && mutationType != "ALL" {
			overviewArgCount++
			overviewWhereClause += " AND m.mutation_type = $" + strconv.Itoa(overviewArgCount)
			overviewArgs = append(overviewArgs, mutationType)
		}

		if search != "" {
			overviewArgCount++
			overviewWhereClause += " AND m.invoice_number ILIKE $" + strconv.Itoa(overviewArgCount)
			overviewArgs = append(overviewArgs, "%"+search+"%")
		}

		if startDate != "" {
			overviewArgCount++
			overviewWhereClause += " AND m.created_at >= $" + strconv.Itoa(overviewArgCount) + "::timestamp"
			overviewArgs = append(overviewArgs, startDate)
		}

		if endDate != "" {
			overviewArgCount++
			overviewWhereClause += " AND m.created_at <= $" + strconv.Itoa(overviewArgCount) + "::timestamp"
			overviewArgs = append(overviewArgs, endDate+" 23:59:59")
		}

		// Get overview (mutations table) with all filters applied
		var totalCredit, totalDebit, netBalance int64
		var transactionCount int
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT
				COALESCE(SUM(CASE WHEN mutation_type = 'CREDIT' THEN amount ELSE 0 END), 0) AS total_credit,
				COALESCE(SUM(CASE WHEN mutation_type = 'DEBIT'  THEN amount ELSE 0 END), 0) AS total_debit,
				COALESCE(SUM(CASE WHEN mutation_type = 'CREDIT' THEN amount ELSE -amount END), 0) AS net_balance,
				COUNT(*) AS txn_count
			FROM mutations m
			`+overviewWhereClause+`
		`, overviewArgs...).Scan(&totalCredit, &totalDebit, &netBalance, &transactionCount)

		if err != nil && err != pgx.ErrNoRows {
			utils.WriteInternalServerError(w)
			return
		}

		// Build query for mutations list
		query := `
			SELECT
				m.invoice_number, m.description, m.amount, m.mutation_type,
				m.balance_before, m.balance_after, m.currency, m.created_at
			FROM mutations m
			` + whereClause + `
		`

		if mutationType != "" && mutationType != "ALL" {
			argCount++
			query += " AND m.mutation_type = $" + strconv.Itoa(argCount)
			args = append(args, mutationType)
		}

		if search != "" {
			argCount++
			query += " AND m.invoice_number ILIKE $" + strconv.Itoa(argCount)
			args = append(args, "%"+search+"%")
		}

		if startDate != "" {
			argCount++
			query += " AND m.created_at >= $" + strconv.Itoa(argCount) + "::timestamp"
			args = append(args, startDate)
		}

		if endDate != "" {
			argCount++
			query += " AND m.created_at <= $" + strconv.Itoa(argCount) + "::timestamp"
			args = append(args, endDate+" 23:59:59")
		}

		query += " ORDER BY m.created_at DESC"
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
			var invoiceNumber sql.NullString
			var description sql.NullString
			var amount, balanceBefore, balanceAfter int64
			var mType, currency string
			var createdAt time.Time

			err := rows.Scan(
				&invoiceNumber, &description, &amount, &mType,
				&balanceBefore, &balanceAfter, &currency, &createdAt,
			)
			if err != nil {
				continue
			}

			mutations = append(mutations, map[string]interface{}{
				"invoiceNumber": nullStringOrEmpty(invoiceNumber),
				"description":   nullStringOrEmpty(description),
				"amount":        amount,
				"type":          mType,
				"balanceBefore": balanceBefore,
				"balanceAfter":  balanceAfter,
				"currency":      currency,
				"createdAt":     createdAt.Format(time.RFC3339),
			})
		}

		// Get total count with same filters (but without limit/offset)
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM mutations m " + whereClause
		countArgs := []interface{}{userID}
		countArgCount := 1

		// Add currency filter if exists
		if regionCurrency != "" {
			countArgCount++
			countQuery += " AND m.currency = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, regionCurrency)
		}

		// Add all filters
		if mutationType != "" && mutationType != "ALL" {
			countArgCount++
			countQuery += " AND m.mutation_type = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, mutationType)
		}

		if search != "" {
			countArgCount++
			countQuery += " AND m.invoice_number ILIKE $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, "%"+search+"%")
		}

		if startDate != "" {
			countArgCount++
			countQuery += " AND m.created_at >= $" + strconv.Itoa(countArgCount) + "::timestamp"
			countArgs = append(countArgs, startDate)
		}

		if endDate != "" {
			countArgCount++
			countQuery += " AND m.created_at <= $" + strconv.Itoa(countArgCount) + "::timestamp"
			countArgs = append(countArgs, endDate+" 23:59:59")
		}

		err = deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to count mutations")
			utils.WriteInternalServerError(w)
			return
		}

		// Calculate totalPages: (totalRows + limit - 1) / limit
		// Examples: 30 rows, limit 10 = 3 pages; 5 rows, limit 10 = 1 page; 0 rows = 0 pages
		totalPages := 0
		if totalRows > 0 {
			totalPages = (totalRows + limit - 1) / limit
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalDebit":       totalDebit,
				"totalCredit":      totalCredit,
				"netBalance":       netBalance,
				"transactionCount": transactionCount,
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

// handleGetReportsImpl returns user's daily transaction reports
func HandleGetReportsImpl(deps *Dependencies) http.HandlerFunc {
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

		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		regionParam := r.URL.Query().Get("region")

		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user's current region if region param not provided
		var userRegion string
		if regionParam == "" {
			_ = deps.DB.Pool.QueryRow(ctx, `SELECT current_region FROM users WHERE id = $1`, userID).Scan(&userRegion)
		} else {
			userRegion = strings.ToUpper(regionParam)
		}

		// Get user currency from region
		currency := getCurrencyByRegion(userRegion)

		// Build WHERE clause for date and region filters
		whereClause := "WHERE t.user_id = $1 AND t.status = 'SUCCESS'::transaction_status"
		args := []interface{}{userID}
		argCount := 1

		if userRegion != "" {
			argCount++
			whereClause += " AND t.region = $" + strconv.Itoa(argCount) + "::region_code"
			args = append(args, userRegion)
		}

		if startDate != "" {
			argCount++
			whereClause += " AND DATE(t.created_at) >= $" + strconv.Itoa(argCount)
			args = append(args, startDate)
		}

		if endDate != "" {
			argCount++
			whereClause += " AND DATE(t.created_at) <= $" + strconv.Itoa(argCount)
			args = append(args, endDate)
		}

		// Get overview stats
		var totalDays, totalTransactions int
		var totalAmount int64
		var avgPerDay float64
		var highestDate, lowestDate sql.NullString
		var highestAmount, lowestAmount sql.NullInt64

		// Get overview stats - split into two queries for clarity
		overviewQuery := `
			WITH daily_stats AS (
				SELECT
					DATE(t.created_at)::text as date,
					COUNT(*) as tx_count,
					SUM(t.total_amount) as amount
				FROM transactions t
				` + whereClause + `
				GROUP BY DATE(t.created_at)
			)
			SELECT
				COUNT(DISTINCT date) as total_days,
				COALESCE(SUM(tx_count), 0) as total_transactions,
				COALESCE(SUM(amount), 0) as total_amount,
				COALESCE(AVG(amount), 0) as avg_per_day
			FROM daily_stats
		`
		err := deps.DB.Pool.QueryRow(ctx, overviewQuery, args...).Scan(&totalDays, &totalTransactions, &totalAmount, &avgPerDay)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to get reports overview")
			utils.WriteInternalServerError(w)
			return
		}

		// Get highest and lowest day separately
		highestQuery := `
			SELECT
				DATE(t.created_at)::text as date,
				SUM(t.total_amount) as amount
			FROM transactions t
			` + whereClause + `
			GROUP BY DATE(t.created_at)
			ORDER BY amount DESC
			LIMIT 1
		`
		_ = deps.DB.Pool.QueryRow(ctx, highestQuery, args...).Scan(&highestDate, &highestAmount)

		lowestQuery := `
			SELECT
				DATE(t.created_at)::text as date,
				SUM(t.total_amount) as amount
			FROM transactions t
			` + whereClause + `
			GROUP BY DATE(t.created_at)
			HAVING SUM(t.total_amount) > 0
			ORDER BY amount ASC
			LIMIT 1
		`
		_ = deps.DB.Pool.QueryRow(ctx, lowestQuery, args...).Scan(&lowestDate, &lowestAmount)

		// Build highestDay and lowestDay
		highestDay := map[string]interface{}{}
		if highestDate.Valid && highestAmount.Valid {
			highestDay["date"] = highestDate.String
			highestDay["amount"] = highestAmount.Int64
		}

		lowestDay := map[string]interface{}{}
		if lowestDate.Valid && lowestAmount.Valid {
			lowestDay["date"] = lowestDate.String
			lowestDay["amount"] = lowestAmount.Int64
		}

		// Get daily reports with pagination
		query := `
			SELECT
				DATE(t.created_at) as date,
				COUNT(*) as total_transactions,
				SUM(t.total_amount) as total_amount
			FROM transactions t
			` + whereClause + `
			GROUP BY DATE(t.created_at)
			ORDER BY date DESC
			LIMIT $` + strconv.Itoa(argCount+1) + `
			OFFSET $` + strconv.Itoa(argCount+2) + `
		`
		queryArgs := append(args, limit, offset)

		rows, err := deps.DB.Pool.Query(ctx, query, queryArgs...)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to query daily reports")
			utils.WriteInternalServerError(w)
			return
		}
		defer rows.Close()

		reports := []map[string]interface{}{}
		for rows.Next() {
			var date time.Time
			var txCount int
			var amount int64

			err := rows.Scan(&date, &txCount, &amount)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan report row")
				continue
			}

			reports = append(reports, map[string]interface{}{
				"date":              date.Format("2006-01-02"),
				"totalTransactions": txCount,
				"totalAmount":       amount,
				"currency":          currency,
			})
		}

		// Get total count of distinct days
		var totalRows int
		countQuery := `
			SELECT COUNT(DISTINCT DATE(t.created_at))
			FROM transactions t
			` + whereClause
		err = deps.DB.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalRows)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to count report days")
			utils.WriteInternalServerError(w)
			return
		}

		// Calculate totalPages: (totalRows + limit - 1) / limit
		// Examples: 30 rows, limit 10 = 3 pages; 5 rows, limit 10 = 1 page; 0 rows = 0 pages
		totalPages := 0
		if totalRows > 0 {
			totalPages = (totalRows + limit - 1) / limit
		}

		utils.WriteSuccessJSON(w, map[string]interface{}{
			"overview": map[string]interface{}{
				"totalDays":         totalDays,
				"totalTransactions": totalTransactions,
				"totalAmount":       totalAmount,
				"averagePerDay":     avgPerDay,
				"highestDay":        highestDay,
				"lowestDay":         lowestDay,
			},
			"reports": reports,
			"pagination": map[string]interface{}{
				"limit":      limit,
				"page":       page,
				"totalRows":  totalRows,
				"totalPages": totalPages,
			},
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
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		regionParam := r.URL.Query().Get("region")
		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get user's current region if region param not provided
		var userRegion string
		if regionParam == "" {
			_ = deps.DB.Pool.QueryRow(ctx, `SELECT current_region FROM users WHERE id = $1`, userID).Scan(&userRegion)
		} else {
			userRegion = strings.ToUpper(regionParam)
		}

		// Build WHERE clause with filters
		whereClause := "WHERE d.user_id = $1"
		args := []interface{}{userID}
		argCount := 1

		// Add region filter (deposits table has region column)
		if userRegion != "" {
			argCount++
			whereClause += " AND d.region = $" + strconv.Itoa(argCount) + "::region_code"
			args = append(args, userRegion)
		}

		if status != "" {
			argCount++
			whereClause += " AND d.status = $" + strconv.Itoa(argCount)
			args = append(args, status)
		}

		if startDate != "" {
			argCount++
			whereClause += " AND d.created_at >= $" + strconv.Itoa(argCount) + "::timestamp"
			args = append(args, startDate)
		}

		if endDate != "" {
			argCount++
			whereClause += " AND d.created_at <= $" + strconv.Itoa(argCount) + "::timestamp"
			args = append(args, endDate+" 23:59:59")
		}

		// Get overview with filters
		var totalDeposits, successCount, pendingCount, failedCount int
		var totalAmount int64

		overviewQuery := `
			SELECT
				COUNT(*) as total,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END), 0) as success,
				COALESCE(SUM(CASE WHEN status = 'PENDING' THEN 1 ELSE 0 END), 0) as pending,
				COALESCE(SUM(CASE WHEN status IN ('FAILED', 'EXPIRED') THEN 1 ELSE 0 END), 0) as failed,
				COALESCE(SUM(CASE WHEN status = 'SUCCESS' THEN amount ELSE 0 END), 0) as total_amount
			FROM deposits d
			` + whereClause

		err := deps.DB.Pool.QueryRow(ctx, overviewQuery, args...).Scan(&totalDeposits, &successCount, &pendingCount, &failedCount, &totalAmount)

		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("endpoint", "/v2/deposits").Msg("Failed to get deposit overview")
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
			` + whereClause

		query += " ORDER BY d.created_at DESC"
		argCount++
		query += " LIMIT $" + strconv.Itoa(argCount)
		args = append(args, limit)

		argCount++
		query += " OFFSET $" + strconv.Itoa(argCount)
		args = append(args, offset)

		// Reset argCount for count query (same filters)
		countArgCount := 1
		countArgs := []interface{}{userID}
		countWhereClause := "WHERE d.user_id = $1"

		if userRegion != "" {
			countArgCount++
			countWhereClause += " AND d.region = $" + strconv.Itoa(countArgCount) + "::region_code"
			countArgs = append(countArgs, userRegion)
		}

		if status != "" {
			countArgCount++
			countWhereClause += " AND d.status = $" + strconv.Itoa(countArgCount)
			countArgs = append(countArgs, status)
		}

		if startDate != "" {
			countArgCount++
			countWhereClause += " AND d.created_at >= $" + strconv.Itoa(countArgCount) + "::timestamp"
			countArgs = append(countArgs, startDate)
		}

		if endDate != "" {
			countArgCount++
			countWhereClause += " AND d.created_at <= $" + strconv.Itoa(countArgCount) + "::timestamp"
			countArgs = append(countArgs, endDate+" 23:59:59")
		}

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

		// Get total count with same filters
		var totalRows int
		countQuery := "SELECT COUNT(*) FROM deposits d " + countWhereClause
		err = deps.DB.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&totalRows)
		if err != nil && err != pgx.ErrNoRows {
			log.Error().Err(err).Str("user_id", userID).Msg("Failed to count deposits")
			utils.WriteInternalServerError(w)
			return
		}

		// Calculate totalPages: (totalRows + limit - 1) / limit
		// Examples: 30 rows, limit 10 = 3 pages; 5 rows, limit 10 = 1 page; 0 rows = 0 pages
		totalPages := 0
		if totalRows > 0 {
			totalPages = (totalRows + limit - 1) / limit
		}

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

		// Get region from context or query param
		region := middleware.GetRegionFromContext(r.Context())
		if region == "" {
			region = r.URL.Query().Get("region")
			if region == "" {
				region = "ID" // Default
			}
		}
		region = strings.ToUpper(region)

		// Get currency from region
		var currency string
		_ = deps.DB.Pool.QueryRow(ctx, `SELECT currency FROM regions WHERE code = $1`, region).Scan(&currency)
		if currency == "" {
			currency = "IDR" // Default
		}

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

		// Generate validation token using JWT
		tokenData := map[string]interface{}{
			"amount":      req.Amount,
			"paymentCode": req.PaymentCode,
			"region":      region,
			"currency":    currency,
			"pricing": map[string]interface{}{
				"subtotal":   req.Amount,
				"paymentFee": paymentFee,
				"total":      total,
			},
		}

		validationToken, err := deps.JWTService.GenerateValidationToken(tokenData)
		if err != nil {
			utils.WriteInternalServerError(w)
			return
		}

		expiresAt := time.Now().Add(5 * time.Minute)

		// Build response - same structure as order inquiry
		response := map[string]interface{}{
			"validationToken": validationToken,
			"expiresAt":       expiresAt.Format(time.RFC3339),
			"deposit": map[string]interface{}{
				"amount": float64(req.Amount),
				"pricing": map[string]interface{}{
					"subtotal":   float64(req.Amount),
					"paymentFee": float64(paymentFee),
					"total":      float64(total),
					"currency":   currency,
				},
				"payment": map[string]interface{}{
					"code":          req.PaymentCode,
					"name":          paymentChannelName,
					"currency":      currency,
					"minAmount":     float64(minAmount),
					"maxAmount":     float64(maxAmount),
					"feeAmount":     float64(feeAmount),
					"feePercentage": float64(feePercentage),
				},
			},
		}

		utils.WriteSuccessJSON(w, response)
	}
}

// CreateDepositRequest represents the request body for creating deposit
type CreateDepositRequest struct {
	ValidationToken string `json:"validationToken" validate:"required"`
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

		if req.ValidationToken == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"validationToken": "Validation token is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Validate and decode validation token
		tokenData, err := deps.JWTService.ValidateValidationToken(req.ValidationToken)
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "INVALID_TOKEN").
				Msg("Failed to validate validation token")
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INVALID_TOKEN",
				"Invalid or expired validation token", "Please create a new deposit inquiry")
			return
		}

		// Check if token was already used (via Redis)
		tokenKey := deps.Redis.ValidationTokenKey(req.ValidationToken)
		exists, err := deps.Redis.Client.Exists(ctx, tokenKey).Result()
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "REDIS_ERROR").
				Str("token_key", tokenKey).
				Msg("Failed to check token usage in Redis")
		}
		if err == nil && exists > 0 {
			log.Warn().
				Str("endpoint", "/v2/deposits").
				Str("error_type", "TOKEN_ALREADY_USED").
				Str("token_key", tokenKey).
				Msg("Validation token has already been used")
			utils.WriteErrorJSON(w, http.StatusBadRequest, "TOKEN_ALREADY_USED",
				"Validation token has already been used", "Please create a new deposit inquiry")
			return
		}

		// Extract deposit data from token
		amount := int64(tokenData["amount"].(float64))
		paymentCode, _ := tokenData["paymentCode"].(string)
		region, _ := tokenData["region"].(string)
		currency, _ := tokenData["currency"].(string)
		pricingData, _ := tokenData["pricing"].(map[string]interface{})
		paymentFee := int64(pricingData["paymentFee"].(float64))
		totalAmount := int64(pricingData["total"].(float64))

		// Mark token as used in Redis (expires in 1 hour)
		_ = deps.Redis.Client.Set(ctx, tokenKey, "used", time.Hour)

		// Get region from context if not in token
		if region == "" {
			region = middleware.GetRegionFromContext(r.Context())
			if region == "" {
				region = "ID" // Default
			}
		}

		// Start database transaction
		tx, err := deps.DB.Pool.Begin(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "DB_TRANSACTION_ERROR").
				Msg("Failed to begin database transaction")
			utils.WriteInternalServerError(w)
			return
		}
		defer tx.Rollback(ctx)

		// Get payment channel details
		var paymentChannelID string
		var paymentName, paymentInstruction string
		var feeAmount, feePercentage float64
		err = tx.QueryRow(ctx, `
			SELECT pc.id, pc.name, pc.instruction, pc.fee_amount, pc.fee_percentage
			FROM payment_channels pc
			WHERE pc.code = $1 AND pc.is_active = true
			LIMIT 1
		`, paymentCode).Scan(&paymentChannelID, &paymentName, &paymentInstruction, &feeAmount, &feePercentage)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "PAYMENT_CHANNEL_NOT_FOUND").
				Str("payment_code", paymentCode).
				Bool("is_no_rows", err == pgx.ErrNoRows).
				Msg("Failed to fetch payment channel from database")
			utils.WriteErrorJSON(w, http.StatusNotFound, "PAYMENT_CHANNEL_NOT_FOUND",
				"Payment channel not found", "")
			return
		}

		// Generate invoice number
		invoiceNumber := utils.GenerateDepositInvoiceNumber()

		// Set expiry time based on payment type
		var expiredAt time.Time
		if paymentCode == "QRIS" {
			// QRIS expires in 30 minutes
			expiredAt = time.Now().Add(30 * time.Minute)
		} else if strings.HasSuffix(paymentCode, "_VA") {
			// VA expires in 24 hours
			expiredAt = time.Now().Add(24 * time.Hour)
		} else {
			// E-wallet expires in 1 hour
			expiredAt = time.Now().Add(1 * time.Hour)
		}

		// Get IP address and user agent
		ipAddress := extractIPAddress(r)
		userAgent := r.UserAgent()

		// Determine gateway based on payment channel code
		gatewayName := getGatewayForChannel(paymentCode)

		// Get gateway ID if gateway name is available
		var gatewayID *string
		if gatewayName != "" {
			var gID string
			err = tx.QueryRow(ctx, `SELECT id FROM payment_gateways WHERE code = $1`, gatewayName).Scan(&gID)
			if err == nil {
				gatewayID = &gID
			}
		}

		// Insert deposit record
		var depositID string
		err = tx.QueryRow(ctx, `
			INSERT INTO deposits (
				invoice_number, user_id, payment_channel_id, payment_gateway_id,
				amount, payment_fee, total_amount, currency, status, region,
				ip_address, user_agent, created_at, expired_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), $13)
			RETURNING id
		`, invoiceNumber, userID, paymentChannelID, gatewayID,
			amount, paymentFee, totalAmount, currency, "PENDING", region,
			ipAddress, userAgent, expiredAt,
		).Scan(&depositID)

		if err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "DEPOSIT_INSERT_ERROR").
				Msg("Failed to insert deposit")
			utils.WriteInternalServerError(w)
			return
		}

		// Create initial timeline entry: Deposit created
		_, err = tx.Exec(ctx, `
			INSERT INTO deposit_logs (deposit_id, status, message, created_at)
			VALUES ($1, $2, $3, NOW())
		`, depositID, "PENDING", "Deposit created, waiting for payment")

		if err != nil {
			log.Warn().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "DEPOSIT_LOG_INSERT_ERROR").
				Str("deposit_id", depositID).
				Msg("Failed to insert deposit timeline (non-fatal)")
		}

		// Integrate with payment gateway (similar to create order)
		var paymentData map[string]interface{}
		var gatewayRefID *string

		// For BALANCE payment, skip gateway call
		if paymentCode != "BALANCE" {
			// Check if payment manager is available
			if deps.PaymentManager == nil {
				log.Error().
					Str("endpoint", "/v2/deposits").
					Str("payment_code", paymentCode).
					Msg("Payment manager not configured")
				tx.Rollback(ctx)
				utils.WriteErrorJSON(w, http.StatusServiceUnavailable, "PAYMENT_GATEWAY_UNAVAILABLE",
					"Payment gateway is not available", "Please try again later or use a different payment method")
				return
			}

			// Build description for payment
			paymentDesc := fmt.Sprintf("Deposit/Top-up %s", currency)

			// Build return URL with locale (default to id-id)
			locale := "id-id"
			if region == "MY" {
				locale = "en-my"
			} else if region == "PH" {
				locale = "en-ph"
			} else if region == "SG" {
				locale = "en-sg"
			} else if region == "TH" {
				locale = "en-th"
			}
			frontendInvoiceURL := fmt.Sprintf("%s/%s/deposit/%s", deps.Config.App.FrontendBaseURL, locale, invoiceNumber)

			// Determine callback URL based on gateway
			callbackURL := deps.Config.App.BaseURL + "/v2/webhooks/payment"
			if gatewayName == "DANA_DIRECT" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/dana"
			} else if gatewayName == "MIDTRANS" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/midtrans"
			} else if gatewayName == "XENDIT" {
				callbackURL = deps.Config.App.BaseURL + "/v2/webhooks/xendit"
			} else if gatewayName == "PAKAILINK" {
				callbackURL = deps.Config.App.BaseURL + "/webhooks/pakailink"
			}

			// Create payment request
			paymentReq := &payment.PaymentRequest{
				RefID:          invoiceNumber,
				Amount:         float64(totalAmount),
				Currency:       currency,
				Channel:        paymentCode,
				GatewayName:    gatewayName,
				GatewayCode:    paymentCode,
				Description:    paymentDesc,
				CustomerName:   "",
				CustomerEmail:  "",
				CustomerPhone:  "",
				ExpiryDuration: time.Until(expiredAt),
				CallbackURL:    callbackURL,
				SuccessURL:     frontendInvoiceURL,
				FailureURL:     frontendInvoiceURL,
				Metadata: map[string]string{
					"deposit_id": depositID,
					"type":       "deposit",
					"user_id":    userID,
					"sku_code":   invoiceNumber, // Use invoice number as sku_code for Midtrans item_details.id
				},
			}

			log.Info().
				Str("endpoint", "/v2/deposits").
				Str("ref_id", invoiceNumber).
				Str("channel", paymentCode).
				Str("gateway_name", gatewayName).
				Float64("amount", paymentReq.Amount).
				Str("callback_url", paymentReq.CallbackURL).
				Msg("Calling payment gateway for deposit")

			// Create payment via gateway
			paymentResult, paymentErr := deps.PaymentManager.CreatePayment(ctx, paymentReq)
			if paymentErr != nil {
				log.Error().
					Err(paymentErr).
					Str("endpoint", "/v2/deposits").
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
				Str("endpoint", "/v2/deposits").
				Str("ref_id", paymentResult.RefID).
				Str("gateway_ref", paymentResult.GatewayRefID).
				Str("status", paymentResult.Status).
				Str("payment_url", paymentResult.PaymentURL).
				Str("qr_code", paymentResult.QRCode).
				Msg("Payment created successfully via gateway for deposit")

			// Map gateway response to payment data
			paymentData = mapGatewayResponseToPaymentDataDeposit(paymentResult, expiredAt, paymentName, paymentInstruction)
			gatewayRefID = &paymentResult.GatewayRefID

			// Note: payment_data table requires transaction_id which deposits don't have
			// For now, we store payment data in payment_logs only for deposits
			// The payment code will be extracted from payment_logs in the invoice handler

			// Update deposits table with gateway ref and payment logs
			paymentLogEntry := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"type":      "PAYMENT_CREATED",
				"data":      paymentData,
			}
			paymentLogJSON, _ := json.Marshal([]interface{}{paymentLogEntry})
			_, err = tx.Exec(ctx, `
				UPDATE deposits
				SET payment_gateway_ref_id = $1, payment_logs = $2::jsonb, updated_at = NOW()
				WHERE id = $3
			`, gatewayRefID, string(paymentLogJSON), depositID)

			if err != nil {
				log.Warn().
					Err(err).
					Str("endpoint", "/v2/deposits").
					Str("error_type", "PAYMENT_LOG_UPDATE_ERROR").
					Str("deposit_id", depositID).
					Msg("Failed to update payment logs (non-fatal)")
			}
		} else {
			// For BALANCE payment, create basic payment data
			paymentData = map[string]interface{}{
				"code":        paymentCode,
				"name":        paymentName,
				"instruction": paymentInstruction,
				"expiredAt":   expiredAt.Format(time.RFC3339),
			}
		}

		// Commit transaction
		if err = tx.Commit(ctx); err != nil {
			log.Error().
				Err(err).
				Str("endpoint", "/v2/deposits").
				Str("error_type", "DB_COMMIT_ERROR").
				Msg("Failed to commit deposit transaction")
			utils.WriteInternalServerError(w)
			return
		}

		// Build response - same structure as create order
		response := map[string]interface{}{
			"step": "SUCCESS",
			"deposit": map[string]interface{}{
				"invoiceNumber": invoiceNumber,
				"status":        "PENDING",
				"amount":        float64(amount),
				"pricing": map[string]interface{}{
					"subtotal":   float64(amount),
					"paymentFee": float64(paymentFee),
					"total":      float64(totalAmount),
					"currency":   currency,
				},
				"payment":   paymentData,
				"createdAt": time.Now().Format(time.RFC3339),
				"expiredAt": expiredAt.Format(time.RFC3339),
			},
		}

		utils.WriteSuccessJSON(w, response)
	}
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
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				// Check if it's an IPv6 address
				if strings.Count(ip, ":") > 1 {
					// IPv6 address, don't remove last colon
					return ip
				}
				return ip[:idx]
			}
			return ip
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		// Remove port if present
		if idx := strings.LastIndex(xri, ":"); idx != -1 {
			return xri[:idx]
		}
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in format "IP:port" or "[IPv6]:port"
	remoteAddr := r.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		// Check if it's an IPv6 address
		if strings.HasPrefix(remoteAddr, "[") {
			// IPv6 address format: [::1]:port
			if closingBracket := strings.LastIndex(remoteAddr, "]"); closingBracket != -1 {
				return remoteAddr[1:closingBracket]
			}
		}
		return remoteAddr[:idx]
	}
	return remoteAddr
}

// getGatewayForChannel returns gateway name for payment channel code
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

		// Virtual Accounts via PakaiLink SNAP
		// BRI uses PAKAILINK (BRI_DIRECT API not available yet)
		"BRI_VA":      "PAKAILINK",
		"VA_BRI":      "PAKAILINK", // alias
		"BCA_VA":      "PAKAILINK",
		"VA_BCA":      "PAKAILINK",
		"BNI_VA":      "PAKAILINK",
		"VA_BNI":      "PAKAILINK",
		"BSI_VA":      "PAKAILINK",
		"VA_BSI":      "PAKAILINK",
		"CIMB_VA":     "PAKAILINK",
		"VA_CIMB":     "PAKAILINK",
		"DANAMON_VA":  "PAKAILINK",
		"VA_DANAMON":  "PAKAILINK",
		"MANDIRI_VA":  "PAKAILINK",
		"VA_MANDIRI":  "PAKAILINK",
		"BMI_VA":      "PAKAILINK",
		"VA_BMI":      "PAKAILINK",
		"BNC_VA":      "PAKAILINK",
		"VA_BNC":      "PAKAILINK",
		"OCBC_VA":     "PAKAILINK",
		"VA_OCBC":     "PAKAILINK",
		"PERMATA_VA":  "PAKAILINK",
		"VA_PERMATA":  "PAKAILINK",
		"SINARMAS_VA": "PAKAILINK",
		"VA_SINARMAS": "PAKAILINK",
	}

	gateway, exists := gatewayMap[channelCode]
	if !exists {
		// Default to XENDIT for unknown channels
		return "XENDIT"
	}
	return gateway
}

// mapGatewayResponseToPaymentDataDeposit maps payment gateway response to payment data format for deposits
func mapGatewayResponseToPaymentDataDeposit(resp *payment.PaymentResponse, expiredAt time.Time, paymentName, paymentInstruction string) map[string]interface{} {
	data := make(map[string]interface{})

	// Common fields
	data["code"] = resp.Channel
	data["name"] = paymentName
	if paymentInstruction != "" {
		data["instruction"] = paymentInstruction
	}

	// Expiry
	if resp.ExpiresAt.IsZero() {
		data["expiredAt"] = expiredAt.Format(time.RFC3339)
	} else {
		data["expiredAt"] = resp.ExpiresAt.Format(time.RFC3339)
	}

	// Gateway reference
	if resp.GatewayRefID != "" {
		data["gatewayRefId"] = resp.GatewayRefID
	}

	// Determine payment code based on channel type
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

	return data
}
