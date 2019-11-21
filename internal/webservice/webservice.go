package webservice

import (
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	port          int
	httpEndpoints []Endpoint
	httpServer    *http.Server
	httpRouter    *chi.Mux
}

type Endpoint struct {
	Path        string
	HandlerFunc http.HandlerFunc
	Method      string
}

func New(port int) *Server {
	s := &Server{
		port:       port,
		httpRouter: chi.NewRouter(),
	}

	s.httpEndpoints = []Endpoint{
		// root and healthchecks
		{Method: "GET", Path: "/", HandlerFunc: instrumentHandler("/", s.rootHandler)},
		{Method: "GET", Path: "/metrics", HandlerFunc: promhttp.Handler().ServeHTTP},
	}

	s.httpRouter.NotFound(instrumentHandler("NotFound", s.codeHandler(404)))
	for _, code := range []int{200, 404, 500} {
		path := fmt.Sprintf("/%d", code)
		s.httpEndpoints = append(s.httpEndpoints, Endpoint{
			Method:      "GET",
			Path:        path,
			HandlerFunc: instrumentHandler(path, s.codeHandler(code)),
		})
	}

	for _, ep := range s.httpEndpoints {
		s.httpRouter.MethodFunc(ep.Method, ep.Path, ep.HandlerFunc)
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.httpRouter,
	}

	return s

}

func (s *Server) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	return s.httpServer.Serve(listener)
}

var (
	inflightCount = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "webservice",
		Subsystem: "http",
		Name:      "in_flight_requests_total",
	})

	totalRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webservice",
			Subsystem: "http",
			Name:      "api_requests_total",
		},
		[]string{"handler", "code", "method"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "webservice",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Buckets:   []float64{0.0125, 0.025, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 2.0},
		},
		[]string{"handler", "code", "method"},
	)

	timeToWrite = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "webservice",
			Subsystem:  "http",
			Name:       "time_to_write_seconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.01, 0.99: 0.001},
		},
		[]string{"handler", "code", "method"},
	)
)

func instrumentHandler(path string, handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerInFlight(inflightCount,
		promhttp.InstrumentHandlerDuration(
			requestDuration.MustCurryWith(
				prometheus.Labels{
					"handler": path,
				}),
			promhttp.InstrumentHandlerCounter(totalRequests.MustCurryWith(
				prometheus.Labels{
					"handler": path,
				}),
				promhttp.InstrumentHandlerTimeToWriteHeader(timeToWrite.MustCurryWith(
					prometheus.Labels{
						"handler": path,
					},
				), handler)))).ServeHTTP
}
