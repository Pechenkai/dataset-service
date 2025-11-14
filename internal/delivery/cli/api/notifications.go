package api

import (
	"context"
	"fmt"
	"strconv"

	"ppo/internal/entities"
)

func (c *Client) NotifySubscribers(ctx context.Context, datasetID uint64, message string) (int, error) {
	var resp struct {
		Sent int `json:"sent"`
	}
	path := fmt.Sprintf("/datasets/%d/notifications", datasetID)
	if err := c.postJSON(ctx, path, map[string]string{"message": message}, true, &resp); err != nil {
		return 0, err
	}
	return resp.Sent, nil
}

func (c *Client) ListNotifications(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	path := fmt.Sprintf("/notifications?user_id=%d", userID)
	var resp notificationsResponse
	if err := c.get(ctx, path, true, &resp); err != nil {
		return nil, err
	}
	list := make([]*entities.Notification, 0, len(resp.Items))
	for _, n := range resp.Items {
		list = append(list, &entities.Notification{
			ID:        n.ID,
			UserID:    n.UserID,
			DatasetID: n.DatasetID,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		})
	}
	return list, nil
}

func (c *Client) MarkNotification(ctx context.Context, notifID uint64, isRead bool) error {
	path := "/notifications/" + strconv.FormatUint(notifID, 10)
	return c.patchJSON(ctx, path, map[string]bool{"is_read": isRead}, true, nil)
}
