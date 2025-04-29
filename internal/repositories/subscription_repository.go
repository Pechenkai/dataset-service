package repositories

//import "ppo/internal/entities"

type SubscriptionRepository interface {
	Subscribe(userID, datasetID uint64) error
	IsSubscribed(userID, datasetID uint64) (bool, error)
	GetSubscribers(datasetID uint64) ([]uint64, error)
}
