package utils

import (
	"fmt"
	"time"

	"gate-v2/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateAccessToken(claims TokenClaims) (string, error)
	GenerateRefreshToken(subject string, tokenType string) (string, error)
	GenerateMFAToken(subject string, tokenType string) (string, error)
	GenerateValidationToken(data map[string]interface{}) (string, error)
	ValidateAccessToken(tokenString string) (*TokenClaims, error)
	ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error)
	ValidateMFAToken(tokenString string) (*jwt.RegisteredClaims, error)
	ValidateValidationToken(tokenString string) (map[string]interface{}, error)
}

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"userId"` // Subject (user/admin ID)
	Type        string   `json:"type"` // user or admin
	Email       string   `json:"email,omitempty"`
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// Subject returns the subject (user/admin ID) from the token
func (t *TokenClaims) Subject() string {
	if t.UserID != "" {
		return t.UserID
	}
	sub, _ := t.RegisteredClaims.GetSubject()
	return sub
}

type ValidationTokenClaims struct {
	jwt.RegisteredClaims
	Data map[string]interface{} `json:"data"`
}

type jwtService struct {
	cfg config.JWTConfig
}

func NewJWTService(cfg config.JWTConfig) JWTService {
	return &jwtService{cfg: cfg}
}

func (s *jwtService) GenerateAccessToken(claims TokenClaims) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.UserID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessTokenExpiry)),
		Issuer:    "gate.co.id",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.SecretKey))
}

func (s *jwtService) GenerateRefreshToken(subject string, tokenType string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.RefreshTokenExpiry)),
		Issuer:    "gate.co.id",
		Audience:  jwt.ClaimStrings{tokenType + "_refresh"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.SecretKey))
}

func (s *jwtService) GenerateMFAToken(subject string, tokenType string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.MFATokenExpiry)),
		Issuer:    "gate.co.id",
		Audience:  jwt.ClaimStrings{tokenType + "_mfa"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.SecretKey))
}

func (s *jwtService) GenerateValidationToken(data map[string]interface{}) (string, error) {
	claims := ValidationTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.ValidationTokenExpiry)),
			Issuer:    "gate.co.id",
			Audience:  jwt.ClaimStrings{"validation"},
		},
		Data: data,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.SecretKey))
}

func (s *jwtService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *jwtService) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *jwtService) ValidateMFAToken(tokenString string) (*jwt.RegisteredClaims, error) {
	return s.ValidateRefreshToken(tokenString)
}

func (s *jwtService) ValidateValidationToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ValidationTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*ValidationTokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims.Data, nil
}

