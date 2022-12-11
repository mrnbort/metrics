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
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

//go:generate moq -out storage_mock.go . Storage

// Service provides access to the db
type Service struct {
	Storage    Storage
	Port       string
	Auth       AuthMidlwr
	templates  *template.Template
	httpServer *http.Server
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
func (s Service) Run(ctx context.Context) error {

	s.templates = template.Must(template.ParseGlob("web/templates/*.tmpl"))

	s.httpServer = &http.Server{
		Addr:         s.Port,
		Handler:      s.routes(),
		ReadTimeout:  time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		err := s.httpServer.Close()
		if err != nil {
			log.Printf("[WARN] can't close server: %v", err)
		}
	}()

	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("service failed to run, err:%v", err)
	}

	return nil
}

func (s Service) routes() chi.Router {
	mux := chi.NewRouter()
	mux.Use(middleware.Throttle(100), middleware.Timeout(60*time.Second))
	mux.Use(PingMiddleware)
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
	mux.Post("/get-metric", s.getMetric)
	mux.Post("/get-metrics", s.getMetrics)

	fs := http.FileServer(http.Dir("./web/static"))
	mux.Route("/web", func(r chi.Router) {
		r.Get("/metrics-list", s.webGetMetricsList)
		r.Get("/metric-details", s.webGetMetricsDetails)
		r.Handle("/static/*", http.StripPrefix("/web/static/", fs))
	})

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
		render.Status(r, http.StatusOK)
		render.JSON(w, r, JSON{"error": "no metrics in db"})
	}
	render.JSON(w, r, result)
}

// POST /get-metric
func (s Service) getMetric(w http.ResponseWriter, r *http.Request) {
	request := metric.Lookup{}
	ctx := r.Context()

	if err := render.DecodeJSON(r.Body, &request); err != nil {
		log.Printf("[WARN] can't bind request %+v: %v", request, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	result, err := s.Storage.GetOneMetric(ctx, request.Name, request.From, request.To, time.Duration(request.Interval))
	if err != nil {
		log.Printf("[WARN] can't get metric data: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		// no metric in db
		render.Status(r, http.StatusOK)
		render.JSON(w, r, JSON{"error": "no metric in db"})
	}
	render.JSON(w, r, result)
}

// POST /get-metrics
func (s Service) getMetrics(w http.ResponseWriter, r *http.Request) {
	request := metric.Lookup{}
	ctx := r.Context()

	if err := render.DecodeJSON(r.Body, &request); err != nil {
		log.Printf("[WARN] can't bind request %+v: %v", request, err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	result, err := s.Storage.GetAll(ctx, request.From, request.To, time.Duration(request.Interval))
	if err != nil {
		log.Printf("[WARN] can't get metrics data: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		// no metric in db
		render.Status(r, http.StatusOK)
		render.JSON(w, r, JSON{"error": "no metrics in db"})
	}
	render.JSON(w, r, result)
}

// GET /metrics-list
func (s Service) webGetMetricsList(w http.ResponseWriter, r *http.Request) {
	metrxList, err := s.Storage.GetList(r.Context())
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	tmplData := struct {
		Metrics []string
	}{
		Metrics: metrxList,
	}

	err = s.templates.ExecuteTemplate(w, "metrics-list.tmpl", &tmplData)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}

	//w.Header().Set("Content-Type", "text/html")
	//w.WriteHeader(http.StatusOK)
	//_, _ = w.Write([]byte("<html>metrics-list</html>"))
}

// GET /metric-details?name={metric}
func (s Service) webGetMetricsDetails(w http.ResponseWriter, r *http.Request) {
	mname := r.URL.Query().Get("name")
	metrs, _ := s.Storage.GetOneMetric(r.Context(), mname, time.Now().Add(-24*time.Hour), time.Now(), time.Minute*30)

	sort.Slice(metrs, func(i, j int) bool {
		return metrs[i].TimeStamp.Before(metrs[j].TimeStamp)
	})

	tmplData := struct {
		Metrics  []metric.Entry
		Name     string
		From, To string
	}{
		Metrics: metrs,
		Name:    mname,
		From:    time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05"),
		To:      time.Now().Format("2006-01-02 15:04:05"),
	}

	err := s.templates.ExecuteTemplate(w, "metric-details.tmpl", &tmplData)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, JSON{"error": err.Error()})
		return
	}
}
