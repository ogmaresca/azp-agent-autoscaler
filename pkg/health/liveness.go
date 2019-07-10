package health

import (
	"net/http"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/logging"
)

type LivenessCheck struct {
}

func (c LivenessCheck) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	logging.Logger.Trace("Liveness probe")

	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
