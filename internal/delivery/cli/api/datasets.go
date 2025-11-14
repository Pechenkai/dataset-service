package api

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"ppo/internal/entities"
	"ppo/internal/services"
)

func (c *Client) CreateDataset(ctx context.Context, cmd services.CreateDatasetCmd, r io.Reader, size int64) (uint64, error) {
	fields := map[string]string{
		"name":        cmd.Name,
		"description": cmd.Description,
		"category":    strconv.FormatUint(cmd.CategoryID, 10),
		"public":      boolToString(cmd.IsPublic),
	}
	if cmd.MetaFormat != "" {
		fields["metaFormat"] = cmd.MetaFormat
	}
	if cmd.MetaTags != "" {
		fields["metaTags"] = cmd.MetaTags
	}
	if cmd.MetaSize != 0 {
		fields["metaSize"] = strconv.FormatUint(cmd.MetaSize, 10)
	}
	var payload datasetPayload
	if err := c.writeMultipart(ctx, "/datasets", fields, "file", cmd.FileName, r, true, &payload); err != nil {
		return 0, err
	}
	return payload.ID, nil
}

func (c *Client) ListDatasets(ctx context.Context, publicOnly bool, ownerID *uint64) ([]*entities.Dataset, error) {
	params := url.Values{}
	params.Set("limit", "200")
	if publicOnly {
		params.Set("is_public", "true")
	}
	if ownerID != nil {
		params.Set("owner_id", strconv.FormatUint(*ownerID, 10))
	}
	var resp datasetsResponse
	if err := c.get(ctx, "/datasets?"+params.Encode(), true, &resp); err != nil {
		return nil, err
	}
	list := make([]*entities.Dataset, 0, len(resp.Items))
	for _, item := range resp.Items {
		list = append(list, &entities.Dataset{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			OwnerID:     item.OwnerID,
			CategoryID:  item.CategoryID,
			IsPublic:    item.IsPublic,
			CreatedAt:   item.CreatedAt,
		})
	}
	return list, nil
}

func (c *Client) AddDatasetVersion(ctx context.Context, cmd services.AddVersionCmd, r io.Reader, size int64) (uint64, error) {
	fields := map[string]string{}
	if cmd.ChangeLog != "" {
		fields["changeLog"] = cmd.ChangeLog
	}
	if cmd.MetaFormat != "" {
		fields["metaFormat"] = cmd.MetaFormat
	}
	if cmd.MetaTags != "" {
		fields["metaTags"] = cmd.MetaTags
	}
	if cmd.MetaSize != 0 {
		fields["metaSize"] = strconv.FormatUint(cmd.MetaSize, 10)
	}
	var resp struct {
		ID uint64 `json:"id"`
	}
	path := fmt.Sprintf("/datasets/%d/versions", cmd.DatasetID)
	if err := c.writeMultipart(ctx, path, fields, "file", cmd.FileName, r, true, &resp); err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (c *Client) GetDatasetRatingSummary(ctx context.Context, datasetID uint64) (services.RatingSummary, error) {
	var payload datasetWithRating
	if err := c.get(ctx, fmt.Sprintf("/datasets/%d", datasetID), true, &payload); err != nil {
		return services.RatingSummary{}, err
	}
	return services.RatingSummary{
		Average: payload.RatingSummary.Average,
		Count:   payload.RatingSummary.Count,
	}, nil
}
