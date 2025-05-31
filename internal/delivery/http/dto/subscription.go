package dto

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
