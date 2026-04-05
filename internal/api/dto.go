package api

import (
	"time"

	"effective-mobile/internal/models"
)

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"999" minimum:"0"`
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	StartDate   string `json:"start_date" example:"01-2024" description:"Month-year MM-YYYY"`
	EndDate     string `json:"end_date,omitempty" example:"12-2024" description:"Optional end month MM-YYYY"`
}

type SubscriptionResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"999"`
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000" format:"uuid"`
	StartDate   string `json:"start_date" example:"01-2024" description:"Subscription period start, MM-YYYY (calendar month)"`
	EndDate     string `json:"end_date,omitempty" example:"12-2024" description:"Optional subscription period end, MM-YYYY"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T12:00:00Z" description:"Record creation time, RFC3339 UTC"`
	UpdatedAt   string `json:"updated_at" example:"2024-01-01T12:00:00Z" description:"Last update time, RFC3339 UTC"`
}

func SubscriptionResponseFromModel(s *models.Subscription) SubscriptionResponse {
	out := SubscriptionResponse{
		ID:          s.ID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   toMonthYear(s.StartDate),
		CreatedAt:   s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if s.EndDate != nil {
		out.EndDate = toMonthYear(*s.EndDate)
	}
	return out
}

func SubscriptionResponsesFromModels(subs []*models.Subscription) []SubscriptionResponse {
	out := make([]SubscriptionResponse, len(subs))
	for i := range subs {
		out[i] = SubscriptionResponseFromModel(subs[i])
	}
	return out
}

func toMonthYear(t time.Time) string {
	u := t.In(time.UTC)
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC).Format("01-2006")
}

type SummaryResponse struct {
	Total int `json:"total" example:"1000"`
}
