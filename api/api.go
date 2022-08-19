package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/umputun/metrics/metric"
	"log"
	"net/http"
	"time"
)

//go:generate moq -out storage_mock.go . Storage

// Service blah blah
type Service struct {
	Storage Storage
	Port    string
}

type Storage interface {
	Update(m metric.Entry) error
	Delete(m metric.Entry) error
	GetAll(from, to time.Time, interval time.Duration) ([]metric.Entry, error)
}

// JSON is a map alias, just for convenience
type JSON map[string]interface{}

func (s Service) Run() error {
	log.Printf("[INFO] activate rest service")
	if err := http.ListenAndServe(s.Port, s.routes()); err != http.ErrServerClosed {
		return fmt.Errorf("service failed to run, err:%v", err)
	}
	return nil
}

func (s Service) routes() chi.Router {
	mux := chi.NewRouter()

	mux.Route("/api/metrics", func(mux chi.Router) {
		mux.Use(Auth)
		mux.Post("/post-metric", s.postMetric)
		mux.Delete("/delete-metric", s.deleteMetric)
	})
	mux.Get("/get-metrics?from={from}&to={to}&interval={int}", s.getMetrics)

	return mux
}

func (s Service) postMetric(w http.ResponseWriter, r *http.Request) {
	request := metric.Entry{}

	//r.URL.Query().Get("blah")
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

// DELETE /delete-metric?name=blah
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

func (s Service) getMetrics(w http.ResponseWriter, r *http.Request) {

}
