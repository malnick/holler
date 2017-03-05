package holler

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

func (h *HollerProxy) HealthSupervisor() {
	var log = logrus.WithFields(logrus.Fields{"holler": "health"})
	for {
		for name, backend := range h.Backends {
			log.Debugf("executing health check for %s targets", name)

			for _, target := range backend.Targets {
				log.Debugf("checking %s target: %+v", name, target)

				if len(target.HealthRoute) == 0 {
					log.Warnf("%+v health route empty, ignoring", target)
					break
				}

				client := &http.Client{}
				resp, err := client.Head(target.URL)
				if err != nil {
					log.Error(err)
					target.Healthy = false
					break
				}

				if resp.StatusCode != 200 {
					target.Healthy = false
					log.Warnf("backend %s target %s unhealthy", name, target.URL)
					break
				}
				log.Infof("target %s for backend %s is healthy", target.URL, name)
				target.Healthy = true
			}
			time.Sleep(time.Duration(backend.HealthCheckInterval) * time.Second)
		}
	}
}
