package api

import (
	"context"
	"fmt"
	"strconv"

	"ppo/internal/entities"
	"ppo/internal/services"
)

func (c *Client) Subscribe(ctx context.Context, userID, datasetID uint64) error {
	payload := map[string]uint64{
		"user_id":    userID,
		"dataset_id": datasetID,
	}
	return c.postJSON(ctx, "/subscriptions", payload, true, nil)
}

func (c *Client) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	id := services.EncodeSubscriptionID(userID, datasetID)
	return c.delete(ctx, "/subscriptions/"+strconv.FormatUint(id, 10), true)
}

func (c *Client) ListSubscriptions(ctx context.Context, userID uint64) ([]*entities.Subscription, error) {
	path := fmt.Sprintf("/subscriptions?user_id=%d", userID)
	var resp subscriptionsResponse
	if err := c.get(ctx, path, true, &resp); err != nil {
		return nil, err
	}
	return c.mapSubscriptions(resp.Items), nil
}

func (c *Client) ListSubscribers(ctx context.Context, datasetID uint64) ([]*entities.Subscription, error) {
	path := fmt.Sprintf("/datasets/%d/subscribers", datasetID)
	var resp struct {
		Items []struct {
			UserID uint64 `json:"user_id"`
		} `json:"items"`
	}
	if err := c.get(ctx, path, true, &resp); err != nil {
		return nil, err
	}
	list := make([]*entities.Subscription, 0, len(resp.Items))
	for _, item := range resp.Items {
		list = append(list, &entities.Subscription{
			UserID:    item.UserID,
			DatasetID: datasetID,
		})
	}
	return list, nil
}

func (c *Client) mapSubscriptions(items []subscriptionPayload) []*entities.Subscription {
	list := make([]*entities.Subscription, 0, len(items))
	for _, item := range items {
		list = append(list, &entities.Subscription{
			UserID:    item.UserID,
			DatasetID: item.DatasetID,
			CreatedAt: item.CreatedAt,
		})
	}
	return list
}
