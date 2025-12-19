package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TokenRecord struct {
	ID        string
	UserID    uint64
	ExpiresAt time.Time
}

type TokenService interface {
	IssueToken(ctx context.Context, userID uint64, ttl time.Duration) (TokenRecord, error)
	RevokeToken(ctx context.Context, token string) error
	GetToken(ctx context.Context, token string) (*TokenRecord, error)
}

type jwtClaims struct {
	UserID  uint64 `json:"user_id"`
	Exp     int64  `json:"exp"`
	Iat     int64  `json:"iat"`
	JTI     string `json:"jti"`
	Subject string `json:"sub"`
}

type revokedEntry struct {
	expiresAt time.Time
}

type hmacTokenService struct {
	secret     []byte
	defaultTTL time.Duration
	revoked    sync.Map
}

const jwtHeaderEncoded = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" // base64({"alg":"HS256","typ":"JWT"})

func NewHMACTokenService(secret string, defaultTTL time.Duration) TokenService {
	if defaultTTL <= 0 {
		defaultTTL = 24 * time.Hour
	}
	return &hmacTokenService{
		secret:     []byte(secret),
		defaultTTL: defaultTTL,
	}
}

func (s *hmacTokenService) IssueToken(_ context.Context, userID uint64, ttl time.Duration) (TokenRecord, error) {
	if len(s.secret) == 0 {
		return TokenRecord{}, errors.New("token secret is not configured")
	}
	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	jti := uuid.NewString()
	claims := jwtClaims{
		UserID:  userID,
		Subject: strconv.FormatUint(userID, 10),
		Exp:     expiresAt.Unix(),
		Iat:     now.Unix(),
		JTI:     jti,
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return TokenRecord{}, fmt.Errorf("marshal claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	unsigned := jwtHeaderEncoded + "." + payloadEncoded
	signature := s.sign(unsigned)
	token := unsigned + "." + signature

	return TokenRecord{
		ID:        token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *hmacTokenService) RevokeToken(_ context.Context, token string) error {
	claims, err := s.parse(token)
	if err != nil {
		return ErrTokenInvalid
	}
	s.revoked.Store(claims.JTI, revokedEntry{expiresAt: time.Unix(claims.Exp, 0)})
	return nil
}

func (s *hmacTokenService) GetToken(_ context.Context, token string) (*TokenRecord, error) {
	claims, err := s.parse(token)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if time.Now().UTC().Unix() > claims.Exp {
		s.revoked.Delete(claims.JTI)
		return nil, ErrTokenNotFound
	}
	if s.isRevoked(claims.JTI) {
		return nil, ErrTokenNotFound
	}
	return &TokenRecord{
		ID:        token,
		UserID:    claims.UserID,
		ExpiresAt: time.Unix(claims.Exp, 0),
	}, nil
}

func (s *hmacTokenService) parse(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenInvalid
	}
	unsigned := parts[0] + "." + parts[1]
	if !s.verify(unsigned, parts[2]) {
		return nil, ErrTokenInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrTokenInvalid
	}
	if claims.Exp == 0 || claims.JTI == "" {
		return nil, ErrTokenInvalid
	}
	return &claims, nil
}

func (s *hmacTokenService) sign(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *hmacTokenService) verify(value, signature string) bool {
	expected := s.sign(value)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func (s *hmacTokenService) isRevoked(jti string) bool {
	raw, ok := s.revoked.Load(jti)
	if !ok {
		return false
	}
	entry := raw.(revokedEntry)
	if time.Now().UTC().After(entry.expiresAt) {
		s.revoked.Delete(jti)
		return false
	}
	return true
}
