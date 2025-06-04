package dto

type SubscriptionDTO struct {
	DatasetID uint64
}

type SubscriberDTO struct {
	UserID uint64
}

func ToSubscriptionDTOs(list []uint64) []*SubscriptionDTO {
	out := make([]*SubscriptionDTO, 0, len(list))
	for _, dsid := range list {
		out = append(out, &SubscriptionDTO{DatasetID: dsid})
	}
	return out
}

func ToSubscriberDTOs(list []uint64) []*SubscriberDTO {
	out := make([]*SubscriberDTO, 0, len(list))
	for _, uid := range list {
		out = append(out, &SubscriberDTO{UserID: uid})
	}
	return out
}

type CreateSubscriptionForm struct {
	DatasetID uint64
}
