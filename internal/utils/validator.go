package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Use JSON tag names for error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

// ValidateStruct validates a struct using go-playground/validator
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		errors[field] = getValidationErrorMessage(err)
	}
	return errors
}

// DecodeAndValidate decodes JSON body and validates the struct
func DecodeAndValidate(r *http.Request, dst interface{}) (map[string]string, error) {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return map[string]string{"body": "Invalid JSON format"}, err
	}
	
	if errors := ValidateStruct(dst); errors != nil {
		return errors, fmt.Errorf("validation failed")
	}
	
	return nil, nil
}

func getValidationErrorMessage(err validator.FieldError) string {
	field := err.Field()
	
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, err.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	return validate.Var(email, "email") == nil
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) bool {
	// Basic phone validation - starts with + or 0, 10-15 digits
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	for i, c := range phone {
		if i == 0 && (c == '+' || c == '0') {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ValidateUUID validates UUID format
func ValidateUUID(id string) bool {
	return validate.Var(id, "uuid") == nil
}

// ValidateRegion validates region code
func ValidateRegion(region string) bool {
	validRegions := map[string]bool{
		"ID": true,
		"MY": true,
		"PH": true,
		"SG": true,
		"TH": true,
	}
	return validRegions[region]
}

// ValidateCurrency validates currency code
func ValidateCurrency(currency string) bool {
	validCurrencies := map[string]bool{
		"IDR": true,
		"MYR": true,
		"PHP": true,
		"SGD": true,
		"THB": true,
	}
	return validCurrencies[currency]
}

// SanitizeString removes leading/trailing whitespace
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// SanitizeEmail normalizes email to lowercase and trims whitespace
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

