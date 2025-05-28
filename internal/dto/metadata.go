package dto

type MetadataResponse struct {
	ID               uint64 `json:"id"`
	Format           string `json:"format"`
	Size             uint64 `json:"size"`
	Tags             string `json:"tags"`
	DatasetVersionID uint64 `json:"dataset_version_id"`
}

type ListMetadataResponse struct {
	Metadata []MetadataResponse `json:"metadata"`
}
