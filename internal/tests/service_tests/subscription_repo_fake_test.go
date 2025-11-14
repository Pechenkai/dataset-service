package services_test

import (
	"context"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type fakeSubscriptionRepo struct {
	subs map[uint64][]*entities.Subscription
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{subs: make(map[uint64][]*entities.Subscription)}
}

func (r *fakeSubscriptionRepo) ensure() {
	if r.subs == nil {
		r.subs = make(map[uint64][]*entities.Subscription)
	}
}

func (r *fakeSubscriptionRepo) Create(ctx context.Context, s *entities.Subscription) error {
	r.ensure()
	r.subs[s.DatasetID] = append(r.subs[s.DatasetID], &entities.Subscription{
		UserID:    s.UserID,
		DatasetID: s.DatasetID,
		CreatedAt: s.CreatedAt,
	})
	return nil
}

func (r *fakeSubscriptionRepo) IsSubscribed(ctx context.Context, userID, datasetID uint64) (bool, error) {
	r.ensure()
	for _, sub := range r.subs[datasetID] {
		if sub.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeSubscriptionRepo) GetSubscribers(ctx context.Context, datasetID uint64) ([]*entities.Subscription, error) {
	r.ensure()
	list := r.subs[datasetID]
	cp := make([]*entities.Subscription, len(list))
	copy(cp, list)
	return cp, nil
}

func (r *fakeSubscriptionRepo) Unsubscribe(ctx context.Context, userID, datasetID uint64) error {
	r.ensure()
	list := r.subs[datasetID]
	for i, sub := range list {
		if sub.UserID == userID {
			r.subs[datasetID] = append(list[:i], list[i+1:]...)
			return nil
		}
	}
	return repositories.ErrSubscriptionNotFound
}

func (r *fakeSubscriptionRepo) GetByUser(ctx context.Context, userID uint64) ([]*entities.Subscription, error) {
	r.ensure()
	var result []*entities.Subscription
	for _, list := range r.subs {
		for _, sub := range list {
			if sub.UserID == userID {
				result = append(result, &entities.Subscription{
					UserID:    sub.UserID,
					DatasetID: sub.DatasetID,
					CreatedAt: sub.CreatedAt,
				})
			}
		}
	}
	return result, nil
}

func (r *fakeSubscriptionRepo) GetAll(ctx context.Context) ([]*entities.Subscription, error) {
	r.ensure()
	var result []*entities.Subscription
	for _, list := range r.subs {
		for _, sub := range list {
			result = append(result, &entities.Subscription{
				UserID:    sub.UserID,
				DatasetID: sub.DatasetID,
				CreatedAt: sub.CreatedAt,
			})
		}
	}
	return result, nil
}

var _ repositories.SubscriptionRepository = (*fakeSubscriptionRepo)(nil)
