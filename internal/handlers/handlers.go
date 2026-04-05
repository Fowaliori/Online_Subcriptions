package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	api "effective-mobile/internal/api"
	"effective-mobile/internal/errs"
	"effective-mobile/internal/models"
	"effective-mobile/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	svc *service.SubscriptionService
}

func NewHandler(svc *service.SubscriptionService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/subscriptions", h.createSubscription).Methods("POST")
	r.HandleFunc("/subscriptions", h.listSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions/summary", h.summary).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.getSubscription).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", h.updateSubscription).Methods("PUT")
	r.HandleFunc("/subscriptions/{id}", h.deleteSubscription).Methods("DELETE")
}

// createSubscription creates a subscription.
// @Summary      Create subscription
// @Description  Request body uses MM-YYYY for start_date and optional end_date.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      api.CreateSubscriptionRequest  true  "Payload"
// @Success      201   {object}  api.SubscriptionResponse
// @Failure      400   {string}  string  "Bad request"
// @Failure      500   {string}  string  "Internal error"
// @Router       /subscriptions [post]
func (h *Handler) createSubscription(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
		UserID      string `json:"user_id"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверное тело запроса", http.StatusBadRequest)
		return
	}
	userUUID, err := parseUserID(in.UserID)
	if err != nil {
		http.Error(w, "неверный user_id", http.StatusBadRequest)
		return
	}
	svcName := strings.TrimSpace(in.ServiceName)
	if err := validatePrice(in.Price); err != nil {
		http.Error(w, "неверная цена", http.StatusBadRequest)
		return
	}
	sd, err := time.Parse("01-2006", in.StartDate)
	if err != nil {
		http.Error(w, "неверный start_date", http.StatusBadRequest)
		return
	}
	var ed *time.Time
	if in.EndDate != "" {
		t, err := time.Parse("01-2006", in.EndDate)
		if err != nil {
			http.Error(w, "неверный end_date", http.StatusBadRequest)
			return
		}
		ed = &t
	}
	if err := validateSubscriptionMonthRange(sd, ed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s := &models.Subscription{
		ServiceName: svcName,
		Price:       in.Price,
		UserID:      userUUID.String(),
		StartDate:   sd,
		EndDate:     ed,
	}
	if err := h.svc.Create(r.Context(), s); err != nil {
		if errors.Is(err, errs.ErrServiceNotFound) {
			http.Error(w, "сервис не найден", http.StatusBadRequest)
			return
		}
		http.Error(w, "не удалось сохранить", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(api.SubscriptionResponseFromModel(s))
}

// getSubscription returns one subscription by id.
// @Summary      Get subscription
// @Tags         subscriptions
// @Produce      json
// @Param        id    path      string  true  "Subscription UUID"
// @Success      200   {object}  api.SubscriptionResponse
// @Failure      404   {string}  string  "Not found"
// @Router       /subscriptions/{id} [get]
func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	parsedID, err := parseSubscriptionID(id)
	if err != nil {
		http.Error(w, "неверный идентификатор, нужен UUID", http.StatusBadRequest)
		return
	}
	s, err := h.svc.Get(r.Context(), parsedID)
	if err != nil {
		if errors.Is(err, errs.ErrSubscriptionNotFound) {
			http.Error(w, "не найдено", http.StatusNotFound)
			return
		}
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(api.SubscriptionResponseFromModel(s))
}

// listSubscriptions returns all subscriptions.
// @Summary      List subscriptions
// @Tags         subscriptions
// @Produce      json
// @Success      200  {array}   api.SubscriptionResponse
// @Failure      500  {string}  string  "Internal error"
// @Router       /subscriptions [get]
func (h *Handler) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(api.SubscriptionResponsesFromModels(subs))
}

// updateSubscription replaces fields of an existing subscription.
// @Summary      Update subscription
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "Subscription UUID"
// @Param        body  body      api.CreateSubscriptionRequest  true  "Payload (same shape as create)"
// @Success      200   {object}  api.SubscriptionResponse
// @Failure      400   {string}  string  "Bad request"
// @Failure      404   {string}  string  "Not found"
// @Failure      500   {string}  string  "Internal error"
// @Router       /subscriptions/{id} [put]
func (h *Handler) updateSubscription(w http.ResponseWriter, r *http.Request) {
	parsedID, err := parseSubscriptionID(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "неверный идентификатор, нужен UUID", http.StatusBadRequest)
		return
	}
	var in struct {
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
		UserID      string `json:"user_id"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "неверное тело запроса", http.StatusBadRequest)
		return
	}
	userUUID, err := parseUserID(in.UserID)
	if err != nil {
		http.Error(w, "неверный user_id", http.StatusBadRequest)
		return
	}
	svcName := strings.TrimSpace(in.ServiceName)
	if err := validatePrice(in.Price); err != nil {
		http.Error(w, "неверная цена", http.StatusBadRequest)
		return
	}
	sd, err := time.Parse("01-2006", in.StartDate)
	if err != nil {
		http.Error(w, "неверный start_date", http.StatusBadRequest)
		return
	}
	var ed *time.Time
	if in.EndDate != "" {
		t, err := time.Parse("01-2006", in.EndDate)
		if err != nil {
			http.Error(w, "неверный end_date", http.StatusBadRequest)
			return
		}
		ed = &t
	}
	if err := validateSubscriptionMonthRange(sd, ed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updated, err := h.svc.UpdateSubscription(r.Context(), parsedID, service.SubscriptionUpdate{
		ServiceName: svcName,
		Price:       in.Price,
		UserID:      userUUID.String(),
		StartDate:   sd,
		EndDate:     ed,
	})
	if err != nil {
		if errors.Is(err, errs.ErrSubscriptionNotFound) {
			http.Error(w, "не найдено", http.StatusNotFound)
			return
		}
		if errors.Is(err, errs.ErrServiceNotFound) {
			http.Error(w, "сервис не найден", http.StatusBadRequest)
			return
		}
		http.Error(w, "не удалось обновить", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(api.SubscriptionResponseFromModel(updated))
}

// deleteSubscription removes a subscription.
// @Summary      Delete subscription
// @Tags         subscriptions
// @Param        id   path  string  true  "Subscription UUID"
// @Success      204  "No Content"
// @Failure      404  {string}  string  "Not found"
// @Failure      500  {string}  string  "Internal error"
// @Router       /subscriptions/{id} [delete]
func (h *Handler) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	parsedID, err := parseSubscriptionID(id)
	if err != nil {
		http.Error(w, "неверный идентификатор, нужен UUID", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), parsedID); err != nil {
		if errors.Is(err, errs.ErrSubscriptionNotFound) {
			http.Error(w, "не найдено", http.StatusNotFound)
			return
		}
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// summary aggregates spend over a month range with optional filters.
// @Summary      Subscription spend summary
// @Description  Computes total price × overlapping months for subscriptions in range; query dates are MM-YYYY.
// @Tags         subscriptions
// @Produce      json
// @Param        start         query  string  true   "Range start MM-YYYY"
// @Param        end           query  string  true   "Range end MM-YYYY"
// @Param        user_id       query  string  false  "Filter by user UUID"
// @Param        service_name  query  string  false  "Filter by service name"
// @Success      200  {object}  api.SummaryResponse
// @Failure      400  {string}  string  "Bad request"
// @Failure      500  {string}  string  "Internal error"
// @Router       /subscriptions/summary [get]
func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	userIDRaw := strings.TrimSpace(r.URL.Query().Get("user_id"))
	svcName := strings.TrimSpace(r.URL.Query().Get("service_name"))
	var userID string
	if userIDRaw != "" {
		u, err := parseUserID(userIDRaw)
		if err != nil {
			http.Error(w, "неверный user_id", http.StatusBadRequest)
			return
		}
		userID = u.String()
	}
	if start == "" || end == "" {
		http.Error(w, "необходимо указать start и end", http.StatusBadRequest)
		return
	}
	sdate, err := time.Parse("01-2006", start)
	if err != nil {
		http.Error(w, "неверный параметр start", http.StatusBadRequest)
		return
	}
	edate, err := time.Parse("01-2006", end)
	if err != nil {
		http.Error(w, "неверный параметр end", http.StatusBadRequest)
		return
	}
	if err := validateSubscriptionMonthRange(sdate, &edate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, err := h.svc.Summary(r.Context(), sdate, edate, userID, svcName)
	if err != nil {
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(api.SummaryResponse{Total: total})
}
