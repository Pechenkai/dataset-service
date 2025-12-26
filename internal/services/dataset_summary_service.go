package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/integrations/openai"
	"ppo/internal/repositories"
)

type DatasetSummary struct {
	DatasetID uint64
	Text      string
	Model     string
	Mode      string
	Source    string
	CreatedAt time.Time
}

type SummaryProvider interface {
	CreateSummary(ctx context.Context, prompt string) (openai.Summary, error)
}

type DatasetSummaryService interface {
	GenerateSummary(ctx context.Context, datasetID uint64) (DatasetSummary, error)
}

type datasetSummaryService struct {
	dsRepo   repositories.DatasetRepository
	provider SummaryProvider
	logger   *zap.Logger
}

func NewDatasetSummaryService(dsRepo repositories.DatasetRepository, provider SummaryProvider, logger *zap.Logger) DatasetSummaryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &datasetSummaryService{
		dsRepo:   dsRepo,
		provider: provider,
		logger:   logger,
	}
}

func (s *datasetSummaryService) GenerateSummary(ctx context.Context, datasetID uint64) (DatasetSummary, error) {
	ds, err := s.dsRepo.FindByID(ctx, datasetID)
	if err != nil {
		if errors.Is(err, repositories.ErrDatasetNotFound) {
			return DatasetSummary{}, ErrDatasetNotFound
		}
		return DatasetSummary{}, fmt.Errorf("fetch dataset: %w", err)
	}
	if ds == nil {
		return DatasetSummary{}, ErrDatasetNotFound
	}

	prompt := buildPrompt(ds)
	resp, err := s.provider.CreateSummary(ctx, prompt)
	if err != nil {
		s.logger.Warn("failed to generate dataset summary",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
		)
		return DatasetSummary{}, ErrExternalServiceUnavailable
	}

	return DatasetSummary{
		DatasetID: datasetID,
		Text:      resp.Text,
		Model:     resp.Model,
		Mode:      resp.Mode,
		Source:    resp.Source,
		CreatedAt: resp.CreatedAt,
	}, nil
}

func buildPrompt(ds *entities.Dataset) string {
	parts := []string{
		fmt.Sprintf("Dataset: %s", ds.Name),
		fmt.Sprintf("Description: %s", ds.Description),
		fmt.Sprintf("Category ID: %d", ds.CategoryID),
	}
	var cleaned []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			cleaned = append(cleaned, p)
		}
	}
	return strings.Join(cleaned, "\n")
}
