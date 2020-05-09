package exporter

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/tcardonne/restic-controller/conf"
	"github.com/tcardonne/restic-controller/controller"
)

// Exporter represents a Prometheus Exporter
type Exporter struct {
	config              conf.ExporterConfig
	repositories        []*conf.Repository
	integrityController *controller.IntegrityController
	retentionController *controller.RetentionController
}

// NewExporter creates a new exporter
func NewExporter(config conf.ExporterConfig,
	repositories []*conf.Repository,
	integrityController *controller.IntegrityController,
	retentionController *controller.RetentionController,
) *Exporter {
	return &Exporter{config, repositories, integrityController, retentionController}
}

func (exp *Exporter) handler(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(newRepositoryCollector(r.Context(), exp.repositories, exp.integrityController, exp.retentionController))

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

// ListenAndServe starts the Prometheus exporter endpoint
func (exp *Exporter) ListenAndServe() error {
	http.HandleFunc("/metrics", exp.handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Restic Controller Exporter</title></head>
			<body>
			<h1>Restic Controller Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})

	log.WithFields(log.Fields{
		"component": "exporter",
		"addr":      exp.config.BindAddress,
	}).Info("Starting http server")

	return http.ListenAndServe(exp.config.BindAddress, nil)
}
