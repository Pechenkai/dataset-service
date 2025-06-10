package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type accessService struct {
	dsRepo   repositories.DatasetRepository
	arRepo   repositories.AccessRequestRepository
	notifSvc NotificationService
	logger   *zap.Logger
}

func NewAccessService(
	dsRepo repositories.DatasetRepository,
	arRepo repositories.AccessRequestRepository,
	logger *zap.Logger,
) AccessService {
	logger.Debug("NewAccessService initialized")
	return &accessService{dsRepo: dsRepo, arRepo: arRepo, logger: logger}
}

func (s *accessService) Request(ctx context.Context, cmd RequestAccessCmd) error {
	s.logger.Debug("Request access called", zap.Uint64("dataset_id", cmd.DatasetID), zap.Uint64("user_id", cmd.UserID))

	ds, err := s.dsRepo.FindByID(ctx, cmd.DatasetID)
	if err != nil {
		s.logger.Error("dataset not found", zap.Uint64("dataset_id", cmd.DatasetID), zap.Error(err))
		return ErrRequestNotFound
	}
	if ds.IsPublic {
		return ErrBadRequest
	}

	if ds.OwnerID == cmd.UserID {
		return ErrBadRequest
	}

	existing, err := s.arRepo.Find(ctx, cmd.DatasetID, cmd.UserID)
	if err == nil && existing != nil {
		switch existing.Status {
		case entities.AccessStatusDenied:
			if err := s.arRepo.UpdateStatus(ctx, existing.ID, string(entities.AccessStatusPending)); err != nil {
				s.logger.Error("reset denied to pending", zap.Error(err))
				return fmt.Errorf("reset denied to pending: %w", err)
			}
			return nil
		default:
			return ErrRequestAlreadyExists
		}
	}

	ar, err := entities.NewAccessRequest(cmd.DatasetID, cmd.UserID)
	if err != nil {
		s.logger.Error("invalid access request data", zap.Error(err))
		return err
	}
	if err := s.arRepo.Create(ctx, ar); err != nil {
		s.logger.Error("failed to create access request in repo", zap.Error(err))
		return err
	}
	s.logger.Info("access request created", zap.Uint64("request_id", ar.ID))

	return nil
}

func (s *accessService) Approve(ctx context.Context, requestID uint64, ownerID uint64) error {
	s.logger.Debug("Approve access called", zap.Uint64("request_id", requestID), zap.Uint64("owner_id", ownerID))

	ar, err := s.arRepo.FindByRequestID(ctx, requestID)
	if err != nil {
		s.logger.Error("access request not found", zap.Uint64("request_id", requestID), zap.Error(err))
		return ErrRequestNotFound
	}

	ds, err := s.dsRepo.FindByID(ctx, ar.DatasetID)
	if err != nil {
		return ErrRequestNotFound
	}
	if ds.OwnerID != ownerID {
		return ErrRequestForbidden
	}

	if err := s.arRepo.UpdateStatus(ctx, requestID, string(entities.AccessStatusApproved)); err != nil {
		s.logger.Error("failed to update access request status", zap.Error(err))
		return err
	}
	s.logger.Info("access request approved", zap.Uint64("request_id", requestID))

	return nil
}

func (s *accessService) Deny(ctx context.Context, requestID uint64, ownerID uint64) error {
	s.logger.Debug("Deny access called", zap.Uint64("request_id", requestID), zap.Uint64("owner_id", ownerID))

	ar, err := s.arRepo.FindByRequestID(ctx, requestID)
	if err != nil {
		s.logger.Error("access request not found", zap.Uint64("request_id", requestID), zap.Error(err))
		return ErrRequestNotFound
	}
	ds, err := s.dsRepo.FindByID(ctx, ar.DatasetID)
	if err != nil {
		return ErrRequestNotFound
	}
	if ds.OwnerID != ownerID {
		return ErrRequestForbidden
	}
	if err := s.arRepo.UpdateStatus(ctx, requestID, string(entities.AccessStatusDenied)); err != nil {
		s.logger.Error("failed to update access request status", zap.Error(err))
		return err
	}
	s.logger.Info("access request denied", zap.Uint64("request_id", requestID))

	return nil
}

func (s *accessService) ListPending(ctx context.Context, ownerID uint64) ([]*entities.AccessRequest, error) {
	s.logger.Debug("ListPending called", zap.Uint64("owner_id", ownerID))

	reqs, err := s.arRepo.ListPendingByOwner(ctx, ownerID)
	if err != nil {
		s.logger.Error("failed to list pending access requests", zap.Uint64("owner_id", ownerID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("ListPending completed", zap.Uint64("owner_id", ownerID), zap.Int("count", len(reqs)))
	return reqs, nil
}

func (s *accessService) FindByRequestID(ctx context.Context, requestID uint64) (*entities.AccessRequest, error) {
	s.logger.Debug("FindByRequestID called", zap.Uint64("request_id", requestID))
	ar, err := s.arRepo.FindByRequestID(ctx, requestID)
	if err != nil {
		s.logger.Error("failed to find access request by ID", zap.Uint64("request_id", requestID), zap.Error(err))
		return nil, err
	}
	s.logger.Info("access request found by ID", zap.Uint64("request_id", ar.ID))
	return ar, nil
}

func (s *accessService) Find(ctx context.Context, datasetID, userID uint64) (*entities.AccessRequest, error) {
	s.logger.Debug("Find access request called",
		zap.Uint64("dataset_id", datasetID),
		zap.Uint64("user_id", userID),
	)

	ar, err := s.arRepo.Find(ctx, datasetID, userID)
	if err != nil {
		s.logger.Error("failed to find access request in repo",
			zap.Error(err),
			zap.Uint64("dataset_id", datasetID),
			zap.Uint64("user_id", userID),
		)
		return nil, err
	}

	if ar == nil {
		s.logger.Debug("no access request found",
			zap.Uint64("dataset_id", datasetID),
			zap.Uint64("user_id", userID),
		)
		return nil, nil
	}

	s.logger.Info("access request found", zap.Uint64("id", ar.ID))
	return ar, nil
}
