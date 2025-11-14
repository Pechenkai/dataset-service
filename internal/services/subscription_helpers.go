package services

func EncodeSubscriptionID(userID, datasetID uint64) uint64 {
	return (userID << 32) | datasetID
}

func DecodeSubscriptionID(id uint64) (uint64, uint64) {
	userID := id >> 32
	datasetID := id & 0xffffffff
	return userID, datasetID
}
