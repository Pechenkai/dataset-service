package api

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"ppo/internal/entities"
	"ppo/internal/services"
)

func (c *Client) CreateReview(ctx context.Context, cmd services.CreateReviewCmd) (uint64, error) {
	payload := map[string]interface{}{
		"user_id":    cmd.UserID,
		"dataset_id": cmd.DatasetID,
		"rating":     cmd.Rating,
		"text":       cmd.Text,
	}
	var resp reviewPayload
	if err := c.postJSON(ctx, "/reviews", payload, true, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (c *Client) UpdateReview(ctx context.Context, cmd services.UpdateReviewCmd) error {
	path := "/reviews/" + strconv.FormatUint(cmd.ReviewID, 10)
	payload := map[string]interface{}{}
	if cmd.Rating != 0 {
		payload["rating"] = cmd.Rating
	}
	if cmd.Text != "" {
		payload["text"] = cmd.Text
	}
	if len(payload) == 0 {
		return errors.New("nothing to update")
	}
	return c.patchJSON(ctx, path, payload, true, nil)
}

func (c *Client) DeleteReview(ctx context.Context, id uint64) error {
	return c.delete(ctx, "/reviews/"+strconv.FormatUint(id, 10), true)
}

func (c *Client) ListReviewsByDataset(ctx context.Context, datasetID uint64) ([]*entities.Review, error) {
	return c.listReviews(ctx, url.Values{"dataset_id": []string{strconv.FormatUint(datasetID, 10)}})
}

func (c *Client) ListReviewsByUser(ctx context.Context, userID uint64) ([]*entities.Review, error) {
	return c.listReviews(ctx, url.Values{"user_id": []string{strconv.FormatUint(userID, 10)}})
}

func (c *Client) listReviews(ctx context.Context, params url.Values) ([]*entities.Review, error) {
	if params.Get("limit") == "" {
		params.Set("limit", "200")
	}
	var resp reviewsResponse
	if err := c.get(ctx, "/reviews?"+params.Encode(), true, &resp); err != nil {
		return nil, err
	}
	return mapReviews(resp.Items), nil
}

func (c *Client) GetRatingSummary(ctx context.Context, datasetID uint64) (services.RatingSummary, error) {
	return c.GetDatasetRatingSummary(ctx, datasetID)
}

func mapReviews(items []reviewPayload) []*entities.Review {
	list := make([]*entities.Review, 0, len(items))
	for _, item := range items {
		list = append(list, &entities.Review{
			ID:        item.ID,
			DatasetID: item.DatasetID,
			UserID:    item.UserID,
			Rating:    entities.Rating(item.Rating),
			Text:      item.Text,
			CreatedAt: item.CreatedAt,
		})
	}
	return list
}
