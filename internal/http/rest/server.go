package rest

import (
	"net/http"
	"time"

	jobprocessor "github.com/KompiTech/itsm-reporting-service/internal/domain/job/processor"
	jobsvc "github.com/KompiTech/itsm-reporting-service/internal/domain/job/service"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/presenters"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server is a http.Handler with dependencies
type Server struct {
	Addr                    string
	URISchema               string
	router                  *httprouter.Router
	logger                  *zap.SugaredLogger
	jobsService             jobsvc.JobService
	jobsPresenter           presenters.JobPresenter
	jobsProcessor           jobprocessor.JobProcessor
	ExternalLocationAddress string
}

// Config contains server configuration and dependencies
type Config struct {
	Addr                    string
	URISchema               string
	Logger                  *zap.SugaredLogger
	JobsService             jobsvc.JobService
	JobsProcessor           jobprocessor.JobProcessor
	ExternalLocationAddress string
}

// NewServer creates new server with the necessary dependencies
func NewServer(cfg Config) *Server {
	r := httprouter.New()

	URISchema := "http://"
	if cfg.URISchema != "" {
		URISchema = cfg.URISchema
	}

	s := &Server{
		Addr:                    cfg.Addr,
		URISchema:               URISchema,
		router:                  r,
		logger:                  cfg.Logger,
		jobsService:             cfg.JobsService,
		jobsProcessor:           cfg.JobsProcessor,
		ExternalLocationAddress: cfg.ExternalLocationAddress,
	}
	if s.jobsProcessor == nil {
		s.logger.Fatal("jobsProcessor not set")
	}
	s.jobsProcessor.WaitForJobs()
	s.registerPresenters()
	s.registerRoutes()

	// Expose Prometheus metrics
	s.router.Handler(http.MethodGet, "/metrics", promhttp.Handler())

	return s
}

// ServeHTTP makes the server implement the http.Handler interface
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Infow(r.Method,
		"time", time.Now().Format(time.RFC3339),
		"url", r.URL.String(),
	)

	s.router.ServeHTTP(w, r)
}
