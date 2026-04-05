package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"effective-mobile/internal/errs"
	"effective-mobile/internal/models"

	"github.com/google/uuid"
)

type Store interface {
	CreateSubscription(ctx context.Context, s *models.Subscription) error
	GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, s *models.Subscription) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context) ([]*models.Subscription, error)
	QuerySubscriptions(ctx context.Context, userID, serviceName string) ([]*models.Subscription, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }

func (p *PostgresStore) CreateSubscription(ctx context.Context, s *models.Subscription) error {
	if err := p.checkService(ctx, s.ServiceName); err != nil {
		return err
	}
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	_, err := p.db.ExecContext(ctx, `INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, s.CreatedAt, s.UpdatedAt)
	return err
}

func (p *PostgresStore) GetSubscription(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	s := &models.Subscription{}
	var endDate sql.NullTime
	row := p.db.QueryRowContext(ctx, `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id=$1`, id)
	err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrSubscriptionNotFound
		}
		return nil, err
	}
	if endDate.Valid {
		t := endDate.Time
		s.EndDate = &t
	}
	return s, nil
}

func (p *PostgresStore) UpdateSubscription(ctx context.Context, s *models.Subscription) error {
	if err := p.checkService(ctx, s.ServiceName); err != nil {
		return err
	}
	s.UpdatedAt = time.Now().UTC()
	_, err := p.db.ExecContext(ctx, `UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5, updated_at=$6 WHERE id=$7`,
		s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, s.UpdatedAt, s.ID)
	return err
}

func (p *PostgresStore) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	res, err := p.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errs.ErrSubscriptionNotFound
	}
	return nil
}

func (p *PostgresStore) ListSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Subscription
	for rows.Next() {
		s := &models.Subscription{}
		var endDate sql.NullTime
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if endDate.Valid {
			t := endDate.Time
			s.EndDate = &t
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (p *PostgresStore) QuerySubscriptions(ctx context.Context, userID, serviceName string) ([]*models.Subscription, error) {
	q := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE 1=1`
	var args []any
	idx := 1
	if userID != "" {
		q += fmt.Sprintf(" AND user_id=$%d", idx)
		args = append(args, userID)
		idx++
	}
	if serviceName != "" {
		q += fmt.Sprintf(" AND service_name ILIKE $%d", idx)
		args = append(args, serviceName)
		idx++
	}
	rows, err := p.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Subscription
	for rows.Next() {
		s := &models.Subscription{}
		var endDate sql.NullTime
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &endDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if endDate.Valid {
			t := endDate.Time
			s.EndDate = &t
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (p *PostgresStore) checkService(ctx context.Context, serviceName string) error {
	var found int
	err := p.db.QueryRowContext(ctx, `SELECT 1 FROM services WHERE service_name = $1`, serviceName).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrServiceNotFound
	}
	return err
}
