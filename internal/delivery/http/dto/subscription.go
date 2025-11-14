package dto

import "ppo/internal/entities"

type SubscribeRequest struct {
	UserID    uint64 `json:"user_id" validate:"required"`
	DatasetID uint64 `json:"dataset_id" validate:"required"`
}

type ListSubscribersResponse struct {
	Subscribers []uint64 `json:"subscribers"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []uint64 `json:"subscriptions"`
}

func ToSubscriberIDs(list []*entities.Subscription) []uint64 {
	out := make([]uint64, len(list))
	for i, sub := range list {
		out[i] = sub.UserID
	}
	return out
}

func ToDatasetIDs(list []*entities.Subscription) []uint64 {
	out := make([]uint64, len(list))
	for i, sub := range list {
		out[i] = sub.DatasetID
	}
	return out
}
