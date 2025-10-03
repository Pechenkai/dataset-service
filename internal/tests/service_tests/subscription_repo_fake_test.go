package services_test

import (
	"context"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type fakeSubscriptionRepo struct {
	subs map[uint64][]uint64
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{subs: make(map[uint64][]uint64)}
}

func (r *fakeSubscriptionRepo) ensure() {
	if r.subs == nil {
		r.subs = make(map[uint64][]uint64)
	}
}

func (r *fakeSubscriptionRepo) Create(ctx context.Context, s *entities.Subscription) error {
	r.ensure()
	r.subs[s.DatasetID] = append(r.subs[s.DatasetID], s.UserID)
	return nil
}

func (r *fakeSubscriptionRepo) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	r.ensure()
	for _, u := range r.subs[datasetID] {
		if u == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeSubscriptionRepo) GetSubscribers(ctx context.Context, datasetID uint64) ([]uint64, error) {
	r.ensure()
	return append([]uint64(nil), r.subs[datasetID]...), nil
}

func (r *fakeSubscriptionRepo) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	r.ensure()
	users := r.subs[datasetID]
	for i, u := range users {
		if u == userID {
			r.subs[datasetID] = append(users[:i], users[i+1:]...)
			return nil
		}
	}
	return repositories.ErrSubscriptionNotFound
}

func (r *fakeSubscriptionRepo) GetByUser(ctx context.Context, userID uint64) ([]uint64, error) {
	r.ensure()
	var result []uint64
	for datasetID, users := range r.subs {
		for _, u := range users {
			if u == userID {
				result = append(result, datasetID)
			}
		}
	}
	return result, nil
}

var _ repositories.SubscriptionRepository = (*fakeSubscriptionRepo)(nil)
