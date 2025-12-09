package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

const (
	invoicePrefix       = "GATE"
	depositPrefix       = "DEP"
	invoiceRandomLength = 20
)

// GenerateInvoiceNumber generates a unique invoice number
// Format: GATE + random alphanumeric (total 23 chars)
func GenerateInvoiceNumber() string {
	return generateRandomID(invoicePrefix, invoiceRandomLength)
}

// GenerateDepositInvoiceNumber generates a unique deposit invoice number
// Format: DEP + random alphanumeric (total 23 chars)
func GenerateDepositInvoiceNumber() string {
	return generateRandomID(depositPrefix, invoiceRandomLength)
}

func generateRandomID(prefix string, length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return prefix + string(result)
}

// GenerateMFASecret generates a new TOTP secret
func GenerateMFASecret(email string, issuer string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate TOTP key: %w", err)
	}
	return key.Secret(), key.URL(), nil
}

// ValidateMFACode validates a TOTP code
func ValidateMFACode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateQRCode generates a QR code image as base64
func GenerateQRCode(content string, size int) ([]byte, error) {
	return qrcode.Encode(content, qrcode.Medium, size)
}

// GenerateQRCodeURL generates a URL for QR code API
func GenerateQRCodeURL(baseURL, data string) string {
	return fmt.Sprintf("%s/v2/qr/generate?data=%s", baseURL, data)
}

// CalculateFee calculates payment fee
func CalculateFee(amount, feeAmount, feePercentage float64) float64 {
	percentageFee := amount * feePercentage / 100
	return feeAmount + percentageFee
}

// CalculateDiscount calculates promo discount
func CalculateDiscount(amount, promoFlat, promoPercentage, maxPromoAmount float64) float64 {
	discount := promoFlat + (amount * promoPercentage / 100)
	if maxPromoAmount > 0 && discount > maxPromoAmount {
		discount = maxPromoAmount
	}
	return discount
}

// CalculateMargin calculates profit margin percentage
func CalculateMargin(buyPrice, sellPrice float64) float64 {
	if buyPrice == 0 {
		return 0
	}
	return (sellPrice - buyPrice) / buyPrice * 100
}

// CalculateDiscountPercentage calculates discount percentage
func CalculateDiscountPercentage(originalPrice, sellPrice float64) float64 {
	if originalPrice == 0 {
		return 0
	}
	return (originalPrice - sellPrice) / originalPrice * 100
}

// FormatCurrency formats amount with currency symbol
func FormatCurrency(amount float64, currency string) string {
	symbols := map[string]string{
		"IDR": "Rp",
		"MYR": "RM",
		"PHP": "₱",
		"SGD": "S$",
		"THB": "฿",
	}
	
	symbol := symbols[currency]
	if symbol == "" {
		symbol = currency
	}
	
	// Format number with thousand separator
	formatted := formatNumber(amount, currency)
	return symbol + formatted
}

func formatNumber(amount float64, currency string) string {
	if currency == "IDR" || currency == "PHP" {
		return formatWithThousands(int64(amount), ".")
	}
	return fmt.Sprintf("%.2f", amount)
}

func formatWithThousands(n int64, sep string) string {
	str := strconv.FormatInt(n, 10)
	if len(str) <= 3 {
		return str
	}
	
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(sep)
		}
		result.WriteRune(c)
	}
	return result.String()
}

// ParseQueryInt parses integer from query string with default value
func ParseQueryInt(value string, defaultVal int) int {
	if value == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return parsed
}

// ParseQueryBool parses boolean from query string
func ParseQueryBool(value string) *bool {
	if value == "" {
		return nil
	}
	b := value == "true" || value == "1"
	return &b
}

// ParseQueryDate parses date from query string (YYYY-MM-DD)
func ParseQueryDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &t
}

// MD5Hash generates MD5 hash
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to a float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// GetStringValue returns string value from pointer, or default if nil
func GetStringValue(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// GetIntValue returns int value from pointer, or default if nil
func GetIntValue(i *int, def int) int {
	if i == nil {
		return def
	}
	return *i
}

// GetBoolValue returns bool value from pointer, or default if nil
func GetBoolValue(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// GetFloat64Value returns float64 value from pointer, or default if nil
func GetFloat64Value(f *float64, def float64) float64 {
	if f == nil {
		return def
	}
	return *f
}

// Truncate truncates a string to max length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// SlugToCode converts slug to code (uppercase, replace - with _)
func SlugToCode(slug string) string {
	return strings.ToUpper(strings.ReplaceAll(slug, "-", "_"))
}

// CodeToSlug converts code to slug (lowercase, replace _ with -)
func CodeToSlug(code string) string {
	return strings.ToLower(strings.ReplaceAll(code, "_", "-"))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

