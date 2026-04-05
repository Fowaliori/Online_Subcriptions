package service

import (
	"context"
	"time"

	storage "effective-mobile/internal/db"
	"effective-mobile/internal/models"

	"github.com/google/uuid"
)

type SubscriptionUpdate struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   time.Time
	EndDate     *time.Time
}

type SubscriptionService struct {
	store storage.Store
}

func NewSubscriptionService(store storage.Store) *SubscriptionService {
	return &SubscriptionService{store: store}
}

func (svc *SubscriptionService) Create(ctx context.Context, s *models.Subscription) error {
	return svc.store.CreateSubscription(ctx, s)
}

func (svc *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	return svc.store.GetSubscription(ctx, id)
}

func (svc *SubscriptionService) List(ctx context.Context) ([]*models.Subscription, error) {
	return svc.store.ListSubscriptions(ctx)
}

func (svc *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, u SubscriptionUpdate) (*models.Subscription, error) {
	s, err := svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	s.ServiceName = u.ServiceName
	s.Price = u.Price
	s.UserID = u.UserID
	s.StartDate = u.StartDate
	s.EndDate = u.EndDate
	if err := svc.store.UpdateSubscription(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (svc *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return svc.store.DeleteSubscription(ctx, id)
}

func (svc *SubscriptionService) Summary(ctx context.Context, start, end time.Time, userID, serviceName string) (int, error) {
	subs, err := svc.store.QuerySubscriptions(ctx, userID, serviceName)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, sub := range subs {
		total += monthsOverlap(start, end, sub.StartDate, sub.EndDate) * sub.Price
	}
	return total, nil
}

func monthsOverlap(queryStart, queryEnd, subStart time.Time, subEnd *time.Time) int {
	s := subStart
	e := time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)
	if subEnd != nil {
		e = *subEnd
	}
	if e.Before(queryStart) || s.After(queryEnd) {
		return 0
	}
	if s.Before(queryStart) {
		s = queryStart
	}
	if e.After(queryEnd) {
		e = queryEnd
	}
	sy, sm, _ := s.Date()
	ey, em, _ := e.Date()
	return (ey-sy)*12 + int(em-sm) + 1
}
