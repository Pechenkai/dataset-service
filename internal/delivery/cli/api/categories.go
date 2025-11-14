package api

import (
	"context"
	"strconv"

	"ppo/internal/entities"
	"ppo/internal/services"
)

func (c *Client) CreateCategory(ctx context.Context, cmd services.CreateCategoryCmd) (uint64, error) {
	var cat categoryPayload
	if err := c.postJSON(ctx, "/categories", map[string]string{
		"name":        cmd.Name,
		"description": cmd.Description,
	}, true, &cat); err != nil {
		return 0, err
	}
	return cat.ID, nil
}

func (c *Client) ListCategories(ctx context.Context) ([]*entities.Category, error) {
	var resp categoriesResponse
	if err := c.get(ctx, "/categories?limit=200", true, &resp); err != nil {
		return nil, err
	}
	list := make([]*entities.Category, 0, len(resp.Items))
	for _, item := range resp.Items {
		list = append(list, &entities.Category{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		})
	}
	return list, nil
}

func (c *Client) DeleteCategory(ctx context.Context, id uint64) error {
	return c.delete(ctx, "/categories/"+strconv.FormatUint(id, 10), true)
}

func (c *Client) UpdateCategory(ctx context.Context, cmd services.UpdateCategoryCmd) error {
	payload := map[string]string{}
	if cmd.Name != "" {
		payload["name"] = cmd.Name
	}
	if cmd.Description != "" {
		payload["description"] = cmd.Description
	}
	if len(payload) == 0 {
		return nil
	}
	return c.patchJSON(ctx, "/categories/"+strconv.FormatUint(cmd.ID, 10), payload, true, nil)
}
