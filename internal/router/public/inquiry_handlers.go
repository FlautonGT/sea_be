package public

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"seaply/internal/utils"

	"github.com/jackc/pgx/v5"
)

// AccountInquiryRequest represents the request body for account inquiry
type AccountInquiryRequest struct {
	ProductCode string `json:"productCode" validate:"required"`
	UserID      string `json:"userId" validate:"required"`
	ZoneID      string `json:"zoneId,omitempty"`
	ServerID    string `json:"serverId,omitempty"`
}

// InquiryResponse represents the response from external inquiry API
type InquiryResponse struct {
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

// handleAccountInquiryImpl implements account inquiry using external API
func HandleAccountInquiryImpl(deps *Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AccountInquiryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteBadRequestError(w, "Invalid request body")
			return
		}

		// Validate request
		if req.ProductCode == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"productCode": "Product code is required",
			})
			return
		}

		if req.UserID == "" {
			utils.WriteValidationErrorJSON(w, "Validation failed", map[string]string{
				"userId": "User ID is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		// Get product info from database
		var productCode, productTitle, inquirySlug string
		err := deps.DB.Pool.QueryRow(ctx, `
			SELECT code, title, COALESCE(inquiry_slug, '') as inquiry_slug
			FROM products
			WHERE code = $1 AND is_active = true
		`, req.ProductCode).Scan(&productCode, &productTitle, &inquirySlug)

		if err != nil {
			if err == pgx.ErrNoRows {
				utils.WriteErrorJSON(w, http.StatusNotFound, "PRODUCT_NOT_FOUND",
					"Product not found", "")
				return
			}
			utils.WriteInternalServerError(w)
			return
		}

		// Check if inquiry_slug is set
		if inquirySlug == "" {
			utils.WriteErrorJSON(w, http.StatusBadRequest, "INQUIRY_NOT_CONFIGURED",
				"Inquiry slug is not configured for this product", "")
			return
		}

		// Determine zone value (zoneId or serverId)
		zoneValue := req.ZoneID
		if zoneValue == "" {
			zoneValue = req.ServerID
		}

		// Build inquiry URL
		inquiryURL := fmt.Sprintf("%s/%s", deps.Config.App.InquiryBaseURL, inquirySlug)

		// Build query parameters
		queryParams := url.Values{}
		queryParams.Set("id", req.UserID)
		if zoneValue != "" {
			queryParams.Set("zone", zoneValue)
		}
		queryParams.Set("key", deps.Config.App.InquiryKey)

		fullURL := fmt.Sprintf("%s?%s", inquiryURL, queryParams.Encode())

		// Make HTTP request to external inquiry API
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
		var inquiryResp InquiryResponse
		if err := json.Unmarshal(body, &inquiryResp); err != nil {
			utils.WriteErrorJSON(w, http.StatusInternalServerError, "INQUIRY_RESPONSE_ERROR",
				"Failed to parse inquiry response", "")
			return
		}

		// Handle error response from inquiry API
		if inquiryResp.Error != nil {
			errorCode := inquiryResp.Error.Code
			errorMessage := inquiryResp.Error.Message

			// Map error codes to user-friendly messages
			var userMessage string
			var httpStatus int

			var details string
			switch errorCode {
			case "NOT_FOUND":
				userMessage = "Account not found"
				details = "The provided User ID and Zone ID combination does not exist"
				httpStatus = http.StatusNotFound
			case "BAD_REQUEST":
				userMessage = "Invalid request"
				details = "The request data is invalid. Please check your input."
				httpStatus = http.StatusBadRequest
			case "TOO_MANY_REQUESTS":
				userMessage = "Too many requests"
				details = "Please try again later."
				httpStatus = http.StatusTooManyRequests
			case "INTERNAL_ERROR":
				userMessage = "Internal server error"
				details = "An error occurred on the inquiry server. Please try again later."
				httpStatus = http.StatusInternalServerError
			default:
				userMessage = errorMessage
				if userMessage == "" {
					userMessage = "An error occurred while checking the account."
				}
				details = "Please try again or contact support if the problem persists."
				httpStatus = http.StatusInternalServerError
			}

			utils.WriteErrorJSON(w, httpStatus, errorCode, userMessage, details)
			return
		}

		// Check if userName is null/empty (account not found)
		if inquiryResp.Data.UserName == "" {
			utils.WriteErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND",
				"Account not found",
				"The provided User ID and Zone ID combination does not exist")
			return
		}

		// Determine region value (from inquiry response or empty)
		region := inquiryResp.Data.Region
		if region == "" {
			region = "ID" // Default region
		}

		// Success response according to documentation format
		responseData := map[string]interface{}{
			"product": map[string]interface{}{
				"name": productTitle,
				"code": productCode,
			},
			"account": map[string]interface{}{
				"region":   region,
				"nickname": inquiryResp.Data.UserName,
			},
		}

		utils.WriteSuccessJSON(w, responseData)
	}
}
