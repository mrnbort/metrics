// Package api provides rest-like api
package api

import (
	"context"
	"fmt"
	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/umputun/metrics/metric"
	"log"
	"net/http"
	"time"
)

//go:generate moq -out storage_mock.go . Storage

// Service provides access to the db
type Service struct {
	Storage Storage
	Port    string
	Auth    AuthMidlwr
}

// Storage interface updates, deletes and gets metrics from the memory and db
type Storage interface {
	Update(ctx context.Context, m metric.Entry) error
	Delete(ctx context.Context, m metric.Entry) error
	GetList(ctx context.Context) ([]string, error)
	GetOneMetric(ctx context.Context, name string, from, to time.Time, interval time.Duration) ([]metric.Entry, error)
	GetAll(ctx context.Context, from, to time.Time, interval time.Duration) ([]metric.Entry, error)
}

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

// Run the listener and request's router, activates the rest server
func (s Service) Run() error {
	log.Printf("[INFO] activate rest service")
	if err := http.ListenAndServe(s.Port, s.routes()); err != http.ErrServerClosed {
		return fmt.Errorf("service failed to run, err:%v", err)
	}
	return nil
}

func (s Service) routes() chi.Router {
	mux := chi.NewRouter()
	mux.Use(middleware.Throttle(100), middleware.Timeout(60*time.Second))
	mux.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))

	limiter := func(limit float64) func(http.Handler) http.Handler {
		return tollbooth_chi.LimitHandler(tollbooth.NewLimiter(limit, nil))
	}

	mux.Group(func(r chi.Router) { // protected routes
		r.Use(s.Auth.Handler)
		r.With(limiter(1000)).Post("/metric", s.postMetric)
		r.With(limiter(10)).Delete("/metric", s.deleteMetric)
	})

	mux.Get("/get-metrics-list", s.getMetricsList)
	mux.Get("/get-metric?name={name}&from={from}&to={to}&interval={int}", s.getMetric)
	mux.Get("/get-metrics?from={from}&to={to}&interval={int}", s.getMetrics)

	return mux
}

// POST /metric
func (s Service) postMetric(w http.ResponseWriter, r *http.Request) {
	request := metric.Entry{}
	ctx := r.Context()

	if err := render.DecodeJSON(r.Body, &request); err != nil {
		log.Printf("[WARN] can't bind request %+v: %v", request, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
	if err := s.Storage.Update(ctx, request); err != nil {
		log.Printf("[WARN] can't update request %v: %v", request, err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
	render.JSON(w, r, JSON{"status": "ok"})
}

// DELETE /metric?name={metric}
func (s Service) deleteMetric(w http.ResponseWriter, r *http.Request) {
	entry := metric.Entry{Name: r.URL.Query().Get("name")}
	ctx := r.Context()

	if err := s.Storage.Delete(ctx, entry); err != nil {
		log.Printf("[WARN] can't delete %v: %v", entry, err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
	render.JSON(w, r, JSON{"status": "ok"})
}

// GET /get-metrics-list
func (s Service) getMetricsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := s.Storage.GetList(ctx)
	if err != nil {
		log.Printf("[WARN] can't get a list of metrics: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		// no metrics in db
		render.JSON(w, r, JSON{"error": "no metrics in db"})
	}

	render.JSON(w, r, result)

}

// GET /get-metric?name={name}&from={from}&to={to}&interval={int}
func (s Service) getMetric(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()

	//result, err := s.Storage.GetOneMetric(ctx)
}

// GET /get-metrics?from={from}&to={to}&interval={int}
func (s Service) getMetrics(w http.ResponseWriter, r *http.Request) {

}
