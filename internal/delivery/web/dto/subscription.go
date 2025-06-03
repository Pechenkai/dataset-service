package dto

// SubscriptionDTO — данные одной подписки (DatasetID).
type SubscriptionDTO struct {
	DatasetID uint64
}

// SubscriberDTO — данные одного подписчика (UserID).
type SubscriberDTO struct {
	UserID uint64
}

// ToSubscriptionDTOs конвертирует срез datasetID ([]uint64) → []*SubscriptionDTO.
func ToSubscriptionDTOs(list []uint64) []*SubscriptionDTO {
	out := make([]*SubscriptionDTO, 0, len(list))
	for _, dsid := range list {
		out = append(out, &SubscriptionDTO{DatasetID: dsid})
	}
	return out
}

// ToSubscriberDTOs конвертирует срез userID ([]uint64) → []*SubscriberDTO.
func ToSubscriberDTOs(list []uint64) []*SubscriberDTO {
	out := make([]*SubscriberDTO, 0, len(list))
	for _, uid := range list {
		out = append(out, &SubscriberDTO{UserID: uid})
	}
	return out
}

// CreateSubscriptionForm — поля для формы “Новая подписка”.
type CreateSubscriptionForm struct {
	DatasetID uint64
}
