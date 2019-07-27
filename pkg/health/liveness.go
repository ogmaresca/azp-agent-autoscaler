package health

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/logging"
)

var (
	livenessProbeCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "azp_agent_autoscaler_liveness_probe_count",
		Help: "The total number of liveness probes",
	})
)

// LivenessCheck is an HTTP Handler
type LivenessCheck struct {
}

func (c LivenessCheck) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	logging.Logger.Trace("Liveness probe")

	livenessProbeCounter.Inc()

	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
