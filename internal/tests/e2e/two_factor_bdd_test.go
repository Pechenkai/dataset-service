//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"ppo/internal/dataaccess/repositories/postqbuild"
	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/services"
)

type twoFASteps struct {
	baseURL      string
	client       *http.Client
	user         *entities.User
	password     string
	oldPassword  string
	challengeID  string
	lastCode     string
	token        string
	debugSecret  string
	blockTimeout time.Duration
}

type challengeResponse struct {
	ChallengeID  string    `json:"challenge_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	AttemptsLeft int       `json:"attempts_left"`
	Delivery     string    `json:"delivery"`
}

type tokenResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      userEnvelope `json:"user"`
}

type userEnvelope struct {
	ID       uint64 `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func TestTwoFactorFeatures(t *testing.T) {
	output := colors.Colored(os.Stdout)
	suite := godog.TestSuite{
		Name:                 "two_factor",
		ScenarioInitializer:  initTwoFactorScenario,
		TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {},
		Options: &godog.Options{
			Format:   "progress",
			Paths:    []string{"features/two_factor.feature"},
			TestingT: t,
			Output:   output,
		},
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}

func initTwoFactorScenario(sc *godog.ScenarioContext) {
	steps := &twoFASteps{
		baseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("E2E_BASE_URL")), "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
	if steps.baseURL == "" {
		steps.baseURL = "http://localhost:8080/api/v2"
	}
	if e2eApp != nil {
		steps.debugSecret = e2eApp.Config.Auth.TwoFA.DebugSecret
		steps.blockTimeout = e2eApp.Config.Auth.TwoFA.BlockDuration
	}
	if steps.blockTimeout == 0 {
		steps.blockTimeout = 3 * time.Second
	}

	sc.Step(`^существует технический пользователь для 2FA$`, steps.ensureTechnicalUser)
	sc.Step(`^я запрашиваю 2FA-челлендж с текущим паролем$`, steps.requestChallengeWithPassword)
	sc.Step(`^я получаю идентификатор челленджа$`, steps.challengeIdReturned)
	sc.Step(`^я могу получить одноразовый код через debug-канал$`, steps.fetchOtpViaDebug)
	sc.Step(`^я подтверждаю челлендж полученным кодом$`, steps.confirmChallengeWithCode)
	sc.Step(`^я получаю access token$`, steps.receiveAccessToken)
	sc.Step(`^я пытаюсь подтвердить челлендж неверным кодом (\d+) раз$`, steps.confirmWithWrongCodeMultiple)
	sc.Step(`^челлендж временно блокируется$`, steps.challengeBlocked)
	sc.Step(`^я жду окончания блокировки$`, steps.waitForCooldown)
	sc.Step(`^я меняю пароль технического пользователя на новый$`, steps.rotatePassword)
	sc.Step(`^аутентификация со старым паролем отклоняется$`, steps.oldPasswordRejected)
}

func (s *twoFASteps) ensureTechnicalUser() error {
	if e2eApp == nil || e2eApp.DB == nil {
		return fmt.Errorf("app is not bootstrapped for e2e")
	}
	email := strings.TrimSpace(os.Getenv("E2E_TWOFA_USER_EMAIL"))
	if email == "" {
		email = "bdd-2fa@example.com"
	}
	pass := strings.TrimSpace(os.Getenv("E2E_TWOFA_USER_PASSWORD"))
	if pass == "" {
		pass = "bdd-pass-123"
	}

	ctx := context.Background()
	userRepo := postqbuild.NewUserRepo(e2eApp.DB, e2eApp.Logger)

	user, err := userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return err
	}
	svc := services.NewUserService(userRepo, e2eApp.Logger)
	if user == nil {
		id, err := svc.Register(ctx, services.RegisterUserCmd{
			Username: "bdd-2fa",
			Email:    email,
			Password: pass,
			Country:  "RU",
			Role:     entities.RoleUser,
		})
		if err != nil {
			return fmt.Errorf("register technical user: %w", err)
		}
		user, err = userRepo.FindByID(ctx, id)
		if err != nil {
			return err
		}
	} else {
		err = svc.UpdateUser(ctx, services.UpdateUserCmd{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Password:  pass,
			Country:   user.Country,
			Role:      user.Role,
			IsBlocked: false,
		})
		if err != nil {
			return err
		}
		user, err = userRepo.FindByID(ctx, user.ID)
		if err != nil {
			return err
		}
	}
	s.user = user
	s.password = pass
	return nil
}

func (s *twoFASteps) requestChallengeWithPassword() error {
	if s.user == nil {
		return fmt.Errorf("user is not initialized")
	}
	payload := map[string]string{
		"email":    s.user.Email,
		"password": s.password,
	}
	var resp challengeResponse
	status, err := s.postJSON("/auth/2fa/challenge", payload, &resp, nil)
	if err != nil {
		return err
	}
	if status != http.StatusAccepted {
		return fmt.Errorf("expected 202 from challenge request, got %d", status)
	}
	if strings.TrimSpace(resp.ChallengeID) == "" {
		return fmt.Errorf("empty challenge id")
	}
	s.challengeID = resp.ChallengeID
	s.lastCode = ""
	return nil
}

func (s *twoFASteps) challengeIdReturned() error {
	if s.challengeID == "" {
		return fmt.Errorf("challenge id is empty")
	}
	return nil
}

func (s *twoFASteps) fetchOtpViaDebug() error {
	if s.challengeID == "" {
		return fmt.Errorf("challenge id missing")
	}
	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/auth/2fa/challenges/"+s.challengeID+"/code", nil)
	if err != nil {
		return err
	}
	if s.debugSecret != "" {
		req.Header.Set("X-Debug-Secret", s.debugSecret)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch code: status=%d body=%s", resp.StatusCode, string(body))
	}
	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	code := strings.TrimSpace(payload["code"])
	if code == "" {
		return fmt.Errorf("debug code is empty")
	}
	s.lastCode = code
	return nil
}

func (s *twoFASteps) confirmChallengeWithCode() error {
	if s.challengeID == "" || s.lastCode == "" {
		return fmt.Errorf("missing challenge or code")
	}
	var resp tokenResponse
	status, err := s.postJSON("/auth/tokens", map[string]string{
		"challenge_id": s.challengeID,
		"code":         s.lastCode,
	}, &resp, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("expected 200 on token issuance, got %d", status)
	}
	s.token = strings.TrimSpace(resp.Token)
	if s.token == "" {
		return fmt.Errorf("empty token in response")
	}
	return nil
}

func (s *twoFASteps) receiveAccessToken() error {
	if s.token == "" {
		return fmt.Errorf("token not issued")
	}
	return nil
}

func (s *twoFASteps) confirmWithWrongCodeMultiple(times int) error {
	if s.challengeID == "" {
		return fmt.Errorf("challenge id missing")
	}
	var lastStatus int
	for i := 0; i < times; i++ {
		status, err := s.postJSON("/auth/tokens", map[string]string{
			"challenge_id": s.challengeID,
			"code":         "000000",
		}, nil, nil)
		if err != nil {
			return err
		}
		lastStatus = status
		if status == http.StatusForbidden {
			break
		}
	}
	if lastStatus == 0 {
		return fmt.Errorf("no attempts were made")
	}
	if lastStatus != http.StatusUnauthorized && lastStatus != http.StatusForbidden {
		return fmt.Errorf("unexpected status after wrong code: %d", lastStatus)
	}
	return nil
}

func (s *twoFASteps) challengeBlocked() error {
	if s.challengeID == "" {
		return fmt.Errorf("challenge id missing")
	}
	status, err := s.postJSON("/auth/tokens", map[string]string{
		"challenge_id": s.challengeID,
		"code":         "000000",
	}, nil, nil)
	if err != nil {
		return err
	}
	if status != http.StatusForbidden && status != http.StatusNotFound {
		return fmt.Errorf("expected forbidden/not found after block, got %d", status)
	}
	return nil
}

func (s *twoFASteps) waitForCooldown() error {
	time.Sleep(s.blockTimeout + 1*time.Second)
	return nil
}

func (s *twoFASteps) rotatePassword() error {
	if s.user == nil {
		return fmt.Errorf("user not prepared")
	}
	if s.token == "" {
		return fmt.Errorf("token is required to change password")
	}
	newPass := fmt.Sprintf("bdd-%d", time.Now().UnixNano())
	_, err := s.patchJSON("/users/"+fmt.Sprint(s.user.ID), map[string]string{
		"password": newPass,
	}, s.token)
	if err != nil {
		return err
	}
	s.oldPassword, s.password = s.password, newPass
	return nil
}

func (s *twoFASteps) oldPasswordRejected() error {
	if s.user == nil || s.oldPassword == "" {
		return fmt.Errorf("old password is missing")
	}
	status, err := s.postJSON("/auth/2fa/challenge", map[string]string{
		"email":    s.user.Email,
		"password": s.oldPassword,
	}, nil, nil)
	if err != nil {
		return err
	}
	if status != http.StatusUnauthorized && status != http.StatusForbidden {
		return fmt.Errorf("expected old password to fail, got %d", status)
	}
	return nil
}

func (s *twoFASteps) postJSON(path string, payload interface{}, out interface{}, headers map[string]string) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, s.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}
	return resp.StatusCode, nil
}

func (s *twoFASteps) patchJSON(path string, payload interface{}, token string) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPatch, s.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("patch failed: %s", string(body))
	}
	return resp.StatusCode, nil
}
