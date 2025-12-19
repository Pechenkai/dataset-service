package api

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"ppo/internal/entities"
	"ppo/internal/services"
)

func (c *Client) RegisterUser(ctx context.Context, cmd services.RegisterUserCmd) (uint64, error) {
	var resp userPayload
	if err := c.postJSON(ctx, "/users", cmd, false, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (c *Client) Authenticate(ctx context.Context, cmd services.AuthenticateUserCmd) (*entities.User, error) {
	var challenge twoFAChallengeResponse
	if err := c.postJSON(ctx, "/auth/2fa/challenge", cmd, false, &challenge); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(os.Getenv("TWOFA_CODE"))
	if code == "" {
		fmt.Print("Enter 2FA code: ")
		_, _ = fmt.Scanln(&code)
	}

	var resp authenticateResponse
	if err := c.postJSON(ctx, "/auth/tokens", map[string]string{
		"challenge_id": challenge.ChallengeID,
		"code":         code,
	}, false, &resp); err != nil {
		return nil, err
	}
	if err := c.store.Set(resp.Token); err != nil {
		return nil, err
	}
	return resp.User.toEntity(), nil
}

func (c *Client) GetUserByID(ctx context.Context, id uint64) (*entities.User, error) {
	var payload userPayload
	if err := c.get(ctx, "/users/"+strconv.FormatUint(id, 10), true, &payload); err != nil {
		return nil, err
	}
	return payload.toEntity(), nil
}

func (c *Client) UpdateUser(ctx context.Context, cmd services.UpdateUserCmd) error {
	path := "/users/" + strconv.FormatUint(cmd.ID, 10)
	return c.patchJSON(ctx, path, cmd, true, nil)
}

func (c *Client) DeleteUser(ctx context.Context, id uint64) error {
	return c.delete(ctx, "/users/"+strconv.FormatUint(id, 10), true)
}

func (u userPayload) toEntity() *entities.User {
	return &entities.User{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		Country:          u.Country,
		Role:             u.Role,
		RegistrationDate: u.RegistrationDate,
		IsBlocked:        u.IsBlocked,
	}
}
