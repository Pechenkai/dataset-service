package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"go.uber.org/zap"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"ppo/internal/entities"
)

type TwoFAConfig struct {
	CodeTTL       time.Duration
	MaxAttempts   int
	BlockDuration time.Duration
	CodeLength    int
	Delivery      string
	DebugSecret   string
}

type TwoFAChallenge struct {
	ID           string
	UserID       uint64
	ExpiresAt    time.Time
	AttemptsLeft int
	Delivery     string
}

type TwoFANotifier interface {
	SendCode(ctx context.Context, user *entities.User, code string, expiresAt time.Time) error
}

type TwoFactorService interface {
	StartChallenge(ctx context.Context, user *entities.User) (*TwoFAChallenge, error)
	VerifyCode(ctx context.Context, challengeID, code string) (*entities.User, error)
	DebugCode(challengeID, secret string) (string, error)
}

type twoFAService struct {
	cfg        TwoFAConfig
	notifier   TwoFANotifier
	logger     *zap.Logger
	mu         sync.Mutex
	challenges map[string]*challengeState
	lockouts   map[uint64]time.Time
}

type challengeState struct {
	challenge TwoFAChallenge
	user      *entities.User
	code      string
}

func NewTwoFAService(cfg TwoFAConfig, notifier TwoFANotifier, logger *zap.Logger) TwoFactorService {
	cfg = normalizeTwoFACfg(cfg)
	if notifier == nil {
		notifier = &loggingNotifier{logger: logger}
	}
	return &twoFAService{
		cfg:        cfg,
		notifier:   notifier,
		logger:     logger,
		challenges: make(map[string]*challengeState),
		lockouts:   make(map[uint64]time.Time),
	}
}

func (s *twoFAService) StartChallenge(ctx context.Context, user *entities.User) (*TwoFAChallenge, error) {
	if user == nil {
		return nil, ErrNilUser
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	if until, blocked := s.lockouts[user.ID]; blocked {
		if now.Before(until) {
			return nil, ErrUserBlocked
		}
		delete(s.lockouts, user.ID)
	}
	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	code := s.generateCode()
	id := uuid.NewString()

	state := &challengeState{
		code: code,
		user: user,
		challenge: TwoFAChallenge{
			ID:           id,
			UserID:       user.ID,
			ExpiresAt:    now.Add(s.cfg.CodeTTL),
			AttemptsLeft: s.cfg.MaxAttempts,
			Delivery:     s.cfg.Delivery,
		},
	}
	s.challenges[id] = state
	s.logger.Info("2fa challenge created",
		zap.String("challenge_id", id),
		zap.Uint64("user_id", user.ID),
		zap.Time("expires_at", state.challenge.ExpiresAt),
		zap.Int("max_attempts", state.challenge.AttemptsLeft),
	)

	if err := s.notifier.SendCode(ctx, user, code, state.challenge.ExpiresAt); err != nil {
		s.logger.Warn("failed to send 2fa code",
			zap.Error(err),
			zap.Uint64("user_id", user.ID),
		)
	}

	return &state.challenge, nil
}

func (s *twoFAService) VerifyCode(ctx context.Context, challengeID, code string) (*entities.User, error) {
	_ = ctx // reserved for future audit hooks
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.challenges[strings.TrimSpace(challengeID)]
	if !ok {
		return nil, ErrTwoFAChallengeNotFound
	}
	if until, blocked := s.lockouts[state.user.ID]; blocked && now.Before(until) {
		return nil, ErrUserBlocked
	}
	if now.After(state.challenge.ExpiresAt) {
		delete(s.challenges, challengeID)
		return nil, ErrTwoFAExpired
	}

	if strings.TrimSpace(code) != state.code {
		state.challenge.AttemptsLeft--
		if state.challenge.AttemptsLeft <= 0 {
			delete(s.challenges, challengeID)
			s.lockouts[state.user.ID] = now.Add(s.cfg.BlockDuration)
			s.logger.Warn("2fa attempts exhausted; user temporarily blocked",
				zap.Uint64("user_id", state.user.ID),
				zap.Duration("block_duration", s.cfg.BlockDuration),
			)
			return nil, ErrUserBlocked
		}
		s.challenges[challengeID] = state
		return nil, ErrInvalidTwoFACode
	}

	delete(s.challenges, challengeID)
	return state.user, nil
}

func (s *twoFAService) DebugCode(challengeID, secret string) (string, error) {
	if s.cfg.DebugSecret == "" {
		return "", ErrTwoFADebugDisabled
	}
	if strings.TrimSpace(secret) != s.cfg.DebugSecret {
		return "", ErrInvalidCredentials
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.challenges[strings.TrimSpace(challengeID)]
	if !ok {
		return "", ErrTwoFAChallengeNotFound
	}
	return state.code, nil
}

func (s *twoFAService) generateCode() string {
	length := s.cfg.CodeLength
	var digits strings.Builder
	for i := 0; i < length; i++ {
		n := randomDigit()
		digits.WriteString(fmt.Sprintf("%d", n))
	}
	return digits.String()
}

func randomDigit() int {
	n, err := rand.Int(rand.Reader, big.NewInt(10))
	if err != nil {
		return time.Now().Nanosecond() % 10
	}
	return int(n.Int64())
}

func normalizeTwoFACfg(cfg TwoFAConfig) TwoFAConfig {
	if cfg.CodeTTL <= 0 {
		cfg.CodeTTL = 5 * time.Minute
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.BlockDuration <= 0 {
		cfg.BlockDuration = 2 * time.Minute
	}
	if cfg.CodeLength <= 0 {
		cfg.CodeLength = 6
	}
	if strings.TrimSpace(cfg.Delivery) == "" {
		cfg.Delivery = "email"
	}
	return cfg
}

type loggingNotifier struct {
	logger *zap.Logger
}

func (n *loggingNotifier) SendCode(_ context.Context, user *entities.User, code string, expiresAt time.Time) error {
	if n.logger != nil {
		n.logger.Info("2fa code issued",
			zap.Uint64("user_id", user.ID),
			zap.String("email", user.Email),
			zap.String("code", code),
			zap.Time("expires_at", expiresAt),
		)
	}
	return nil
}
