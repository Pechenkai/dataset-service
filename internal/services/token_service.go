package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

type tokenClaims struct {
	userID    uint64
	expiresAt time.Time
	jti       string
}

type revokedEntry struct {
	expiresAt time.Time
}

type hmacTokenService struct {
	secret     []byte
	defaultTTL time.Duration
	revoked    sync.Map
}

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

	payload := fmt.Sprintf("%d|%d|%s", userID, expiresAt.Unix(), jti)
	payloadEncoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	signature := s.sign(payload)
	token := payloadEncoded + "." + signature

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
	s.revoked.Store(claims.jti, revokedEntry{expiresAt: claims.expiresAt})
	return nil
}

func (s *hmacTokenService) GetToken(_ context.Context, token string) (*TokenRecord, error) {
	claims, err := s.parse(token)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if time.Now().UTC().After(claims.expiresAt) {
		s.revoked.Delete(claims.jti)
		return nil, ErrTokenNotFound
	}
	if s.isRevoked(claims.jti) {
		return nil, ErrTokenNotFound
	}
	return &TokenRecord{
		ID:        token,
		UserID:    claims.userID,
		ExpiresAt: claims.expiresAt,
	}, nil
}

func (s *hmacTokenService) parse(token string) (*tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrTokenInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if !s.verify(string(payloadBytes), parts[1]) {
		return nil, ErrTokenInvalid
	}

	fields := strings.Split(string(payloadBytes), "|")
	if len(fields) != 3 {
		return nil, ErrTokenInvalid
	}
	userID, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	expUnix, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	expiresAt := time.Unix(expUnix, 0).UTC()
	return &tokenClaims{
		userID:    userID,
		expiresAt: expiresAt,
		jti:       fields[2],
	}, nil
}

func (s *hmacTokenService) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *hmacTokenService) verify(payload, signature string) bool {
	expected := s.sign(payload)
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
