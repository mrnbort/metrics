// Package api provides rest-like api
package api

import (
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
}

// Storage interface updates, deletes and gets metrics from the memory and db
type Storage interface {
	Update(m metric.Entry) error
	Delete(m metric.Entry) error
	GetAll(from, to time.Time, interval time.Duration) ([]metric.Entry, error)
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

	mux.Route("/protected-post", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(1000, nil)))
		mux.Post("/metric", s.postMetric)
	})

	mux.Route("/protected-delete", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
		mux.Delete("/metric", s.deleteMetric)
	})

	mux.Get("/get-metrics?from={from}&to={to}&interval={int}", s.getMetrics)

	return mux
}

// POST /protected-post/metric
func (s Service) postMetric(w http.ResponseWriter, r *http.Request) {
	request := metric.Entry{}

	if err := render.DecodeJSON(r.Body, &request); err != nil {
		log.Printf("[WARN] can't bind request %+v: %v", request, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
	if err := s.Storage.Update(request); err != nil {
		log.Printf("[WARN] can't update request %v: %v", request, err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
	render.JSON(w, r, JSON{"status": "ok"})
}

// DELETE /delete-metric?name={metric}
func (s Service) deleteMetric(w http.ResponseWriter, r *http.Request) {
	entry := metric.Entry{Name: r.URL.Query().Get("name")}

	if err := s.Storage.Delete(entry); err != nil {
		log.Printf("[WARN] can't delete %v: %v", entry, err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
	render.JSON(w, r, JSON{"status": "ok"})
}

// GET /get-metrics?from={from}&to={to}&interval={int}
func (s Service) getMetrics(w http.ResponseWriter, r *http.Request) {

}
