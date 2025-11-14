package dto

import "ppo/internal/entities"

type SubscriptionDTO struct {
	DatasetID uint64
	CreatedAt string
}

type SubscriberDTO struct {
	UserID uint64
}

func ToSubscriptionDTOs(list []*entities.Subscription) []*SubscriptionDTO {
	out := make([]*SubscriptionDTO, 0, len(list))
	for _, sub := range list {
		out = append(out, &SubscriptionDTO{
			DatasetID: sub.DatasetID,
			CreatedAt: sub.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return out
}

func ToSubscriberDTOs(list []*entities.Subscription) []*SubscriberDTO {
	out := make([]*SubscriberDTO, 0, len(list))
	for _, sub := range list {
		out = append(out, &SubscriberDTO{UserID: sub.UserID})
	}
	return out
}

type CreateSubscriptionForm struct {
	DatasetID uint64
}
