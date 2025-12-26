package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ppo/internal/integrations/catfacts"
	"ppo/internal/repositories"
)

type DatasetFact struct {
	DatasetID   uint64
	Text        string
	Length      int
	Source      string
	Mode        string
	RetrievedAt time.Time
}

type FactProvider interface {
	RandomFact(ctx context.Context) (catfacts.Fact, error)
}

type DatasetFactService interface {
	GetDatasetFact(ctx context.Context, datasetID uint64) (DatasetFact, error)
}

type datasetFactService struct {
	dsRepo   repositories.DatasetRepository
	provider FactProvider
	logger   *zap.Logger
}

func NewDatasetFactService(
	dsRepo repositories.DatasetRepository,
	provider FactProvider,
	logger *zap.Logger,
) DatasetFactService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &datasetFactService{
		dsRepo:   dsRepo,
		provider: provider,
		logger:   logger,
	}
}

func (s *datasetFactService) GetDatasetFact(ctx context.Context, datasetID uint64) (DatasetFact, error) {
	ds, err := s.dsRepo.FindByID(ctx, datasetID)
	if err != nil {
		if errors.Is(err, repositories.ErrDatasetNotFound) {
			return DatasetFact{}, ErrDatasetNotFound
		}
		return DatasetFact{}, fmt.Errorf("fetch dataset: %w", err)
	}
	if ds == nil {
		return DatasetFact{}, ErrDatasetNotFound
	}

	fact, err := s.provider.RandomFact(ctx)
	if err != nil {
		s.logger.Warn("failed to fetch external fact",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return DatasetFact{}, ErrExternalServiceUnavailable
	}

	return DatasetFact{
		DatasetID:   datasetID,
		Text:        fact.Text,
		Length:      fact.Length,
		Source:      fact.Source,
		Mode:        fact.Mode,
		RetrievedAt: time.Now().UTC(),
	}, nil
}
